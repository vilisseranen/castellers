package controller

import (
	"context"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/redis/go-redis/v9"
	"github.com/vilisseranen/castellers/common"
)

// Each login creates a session in Redis. Access and refresh tokens carry its
// id in the "session_id" claim and are only accepted while the session exists,
// so deleting a session revokes all its tokens at once: those issued at login
// and by every refresh since.
//
//	session:<session_id>      -> member UUID, expires after refresh_ttl_days without refresh
//	user_sessions:<member>    -> set of the member's session ids
const (
	sessionKeyPrefix      = "session:"
	userSessionsKeyPrefix = "user_sessions:"
)

// ErrSessionRevoked is returned for a token whose session was revoked (logout,
// password change, member deleted). It wraps jwt.ErrTokenExpired on purpose:
// clients then try to refresh, the refresh fails and they go back to the login
// page, which is what should happen.
var ErrSessionRevoked = fmt.Errorf("session revoked: %w", jwt.ErrTokenExpired)

func sessionTTL() time.Duration {
	return 24 * time.Hour * time.Duration(common.GetConfigInt("jwt.refresh_ttl_days"))
}

// createSession starts a new session for a member and returns its id.
func createSession(ctx context.Context, memberUUID string) (string, error) {
	sessionID := common.GenerateUUID()
	ttl := sessionTTL()
	if err := RedisClient.Set(ctx, sessionKeyPrefix+sessionID, memberUUID, ttl).Err(); err != nil {
		return "", err
	}
	userKey := userSessionsKeyPrefix + memberUUID
	if err := RedisClient.SAdd(ctx, userKey, sessionID).Err(); err != nil {
		return "", err
	}
	if err := RedisClient.Expire(ctx, userKey, ttl).Err(); err != nil {
		return "", err
	}
	return sessionID, nil
}

// extendSession keeps an existing session alive after a refresh. It returns
// ErrSessionRevoked if the session does not exist or belongs to someone else.
func extendSession(ctx context.Context, sessionID, memberUUID string) error {
	owner, err := RedisClient.Get(ctx, sessionKeyPrefix+sessionID).Result()
	if err == redis.Nil {
		return ErrSessionRevoked
	}
	if err != nil {
		return err
	}
	if owner != memberUUID {
		return ErrSessionRevoked
	}
	ttl := sessionTTL()
	if err := RedisClient.Expire(ctx, sessionKeyPrefix+sessionID, ttl).Err(); err != nil {
		return err
	}
	return RedisClient.Expire(ctx, userSessionsKeyPrefix+memberUUID, ttl).Err()
}

// checkSession returns ErrSessionRevoked if the session does not exist anymore.
func checkSession(ctx context.Context, sessionID string) error {
	exists, err := RedisClient.Exists(ctx, sessionKeyPrefix+sessionID).Result()
	if err != nil {
		return err
	}
	if exists == 0 {
		return ErrSessionRevoked
	}
	return nil
}

// revokeSession deletes one session. It succeeds if the session is already gone.
func revokeSession(ctx context.Context, sessionID string) error {
	memberUUID, err := RedisClient.Get(ctx, sessionKeyPrefix+sessionID).Result()
	if err == redis.Nil {
		return nil
	}
	if err != nil {
		return err
	}
	if err := RedisClient.Del(ctx, sessionKeyPrefix+sessionID).Err(); err != nil {
		return err
	}
	return RedisClient.SRem(ctx, userSessionsKeyPrefix+memberUUID, sessionID).Err()
}

// revokeMemberSessions deletes all the sessions of a member, except keepSessionID
// (empty to revoke them all).
func revokeMemberSessions(ctx context.Context, memberUUID, keepSessionID string) error {
	userKey := userSessionsKeyPrefix + memberUUID
	sessionIDs, err := RedisClient.SMembers(ctx, userKey).Result()
	if err != nil {
		return err
	}
	for _, sessionID := range sessionIDs {
		if sessionID == keepSessionID {
			continue
		}
		if err := RedisClient.Del(ctx, sessionKeyPrefix+sessionID).Err(); err != nil {
			return err
		}
		if err := RedisClient.SRem(ctx, userKey, sessionID).Err(); err != nil {
			return err
		}
	}
	return nil
}
