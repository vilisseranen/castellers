package controller

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
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
//	legacy_revoked:<member>   -> set when the member's sessions are revoked: refresh
//	                             tokens issued before sessions existed are refused
const (
	sessionKeyPrefix       = "session:"
	userSessionsKeyPrefix  = "user_sessions:"
	legacyRevokedKeyPrefix = "legacy_revoked:"
)

// ErrSessionRevoked is returned for a token whose session was revoked (logout,
// password change, member deleted). It wraps jwt.ErrTokenExpired on purpose:
// clients then try to refresh, the refresh fails and they go back to the login
// page, which is what should happen.
var ErrSessionRevoked = fmt.Errorf("session revoked: %w", jwt.ErrTokenExpired)

// ErrSessionCheck is returned when the session store (Redis) cannot be reached.
// It is a server error, not an authentication failure.
var ErrSessionCheck = errors.New("cannot check the session")

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
	pruneMemberSessions(ctx, userKey)
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
	extended, err := RedisClient.Expire(ctx, sessionKeyPrefix+sessionID, ttl).Result()
	if err != nil {
		return err
	}
	if !extended {
		// Revoked between the GET and the EXPIRE.
		return ErrSessionRevoked
	}
	return RedisClient.Expire(ctx, userSessionsKeyPrefix+memberUUID, ttl).Err()
}

// checkSession returns ErrSessionRevoked if the session does not exist anymore.
func checkSession(ctx context.Context, sessionID string) error {
	exists, err := RedisClient.Exists(ctx, sessionKeyPrefix+sessionID).Result()
	if err != nil {
		return fmt.Errorf("%w: %v", ErrSessionCheck, err)
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
// (empty to revoke them all). Refresh tokens issued before sessions existed are
// refused from then on.
func revokeMemberSessions(ctx context.Context, memberUUID, keepSessionID string) error {
	if err := RedisClient.Set(ctx, legacyRevokedKeyPrefix+memberUUID, "1", sessionTTL()).Err(); err != nil {
		return err
	}
	userKey := userSessionsKeyPrefix + memberUUID
	sessionIDs, err := RedisClient.SMembers(ctx, userKey).Result()
	if err != nil {
		return err
	}
	var keys []string
	var ids []interface{}
	for _, sessionID := range sessionIDs {
		if sessionID != keepSessionID {
			keys = append(keys, sessionKeyPrefix+sessionID)
			ids = append(ids, sessionID)
		}
	}
	if len(keys) == 0 {
		return nil
	}
	if err := RedisClient.Del(ctx, keys...).Err(); err != nil {
		return err
	}
	return RedisClient.SRem(ctx, userKey, ids...).Err()
}

// pruneMemberSessions removes from the member's index the sessions that expired.
func pruneMemberSessions(ctx context.Context, userKey string) {
	sessionIDs, err := RedisClient.SMembers(ctx, userKey).Result()
	if err != nil {
		return
	}
	for _, sessionID := range sessionIDs {
		if exists, err := RedisClient.Exists(ctx, sessionKeyPrefix+sessionID).Result(); err == nil && exists == 0 {
			RedisClient.SRem(ctx, userKey, sessionID)
		}
	}
}

// legacySession returns the session that a refresh token issued before
// sessions existed was moved to, creating it the first time. The legacy Redis
// key then points to the session, so revoking the session also revokes the
// legacy token, and refreshing it again reuses the same session.
func legacySession(ctx context.Context, legacyKey, memberUUID string) (string, error) {
	revoked, err := RedisClient.Exists(ctx, legacyRevokedKeyPrefix+memberUUID).Result()
	if err != nil {
		return "", err
	}
	if revoked > 0 {
		return "", ErrSessionRevoked
	}
	value, err := RedisClient.Get(ctx, legacyKey).Result()
	if err == redis.Nil {
		return "", ErrSessionRevoked
	}
	if err != nil {
		return "", err
	}
	if sessionID, ok := strings.CutPrefix(value, sessionKeyPrefix); ok {
		return sessionID, extendSession(ctx, sessionID, memberUUID)
	}
	if value != memberUUID {
		return "", ErrSessionRevoked
	}
	ttl, err := RedisClient.TTL(ctx, legacyKey).Result()
	if err != nil {
		return "", err
	}
	sessionID, err := createSession(ctx, memberUUID)
	if err != nil {
		return "", err
	}
	if ttl <= 0 {
		ttl = sessionTTL()
	}
	if err := RedisClient.Set(ctx, legacyKey, sessionKeyPrefix+sessionID, ttl).Err(); err != nil {
		return "", err
	}
	return sessionID, nil
}

// RespondWithTokenError answers a request whose access token was rejected by ExtractToken.
func RespondWithTokenError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, ErrSessionCheck):
		common.Warn("Cannot check session: %s", err.Error())
		RespondWithError(w, http.StatusServiceUnavailable, ERRORINTERNAL)
	case IsTokenExpired(err):
		RespondWithTokenExpired(w, ERRORTOKENEXPIRED)
	default:
		common.Debug("Invalid token: %s", err.Error())
		RespondWithError(w, http.StatusUnauthorized, ERRORTOKENINVALID)
	}
}
