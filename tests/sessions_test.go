package tests

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/vilisseranen/castellers/controller"
)

func loginTokens(t *testing.T, username, password string) (string, string) {
	t.Helper()
	req, _ := http.NewRequest("POST", "/api/v1/login", bytes.NewBufferString(`{"username":"`+username+`","password":"`+password+`"}`))
	response := h.executeRequest(req)
	if err := h.checkResponseCode(http.StatusOK, response.Code); err != nil {
		t.Fatal(err)
	}
	var tokens map[string]string
	if err := json.Unmarshal(response.Body.Bytes(), &tokens); err != nil {
		t.Fatal(err)
	}
	return tokens["access_token"], tokens["refresh_token"]
}

func refreshTokens(t *testing.T, refreshToken string) (string, string) {
	t.Helper()
	payload, _ := json.Marshal(map[string]string{"refresh_token": refreshToken})
	req, _ := http.NewRequest("POST", "/api/v1/refresh", bytes.NewBuffer(payload))
	response := h.executeRequest(req)
	if err := h.checkResponseCode(http.StatusCreated, response.Code); err != nil {
		t.Fatal(err)
	}
	var tokens map[string]string
	json.Unmarshal(response.Body.Bytes(), &tokens)
	return tokens["access_token"], tokens["refresh_token"]
}

// expectRevoked checks that an access token is refused with the code that makes
// clients refresh (and then fail and log out).
func expectRevoked(t *testing.T, name, accessToken string) {
	t.Helper()
	req, _ := http.NewRequest("GET", "/api/v1/castells/models", nil)
	req.Header.Add("Authorization", "Bearer "+accessToken)
	response := h.executeRequest(req)
	var body map[string]string
	json.Unmarshal(response.Body.Bytes(), &body)
	if response.Code != http.StatusForbidden || body["code"] != "token_expired" {
		t.Errorf("%s: expected 403 token_expired, got %d %v", name, response.Code, body)
	}
}

func expectAccepted(t *testing.T, name, accessToken string) {
	t.Helper()
	if err := h.checkResponseCode(http.StatusOK, getWithAccessToken("/api/v1/castells/models", accessToken)); err != nil {
		t.Errorf("%s: %v", name, err)
	}
}

// Logout ends the whole session: tokens from the login and from every refresh since.
func TestLogoutRevokesWholeSession(t *testing.T) {
	h.clearTables()
	access1, refresh1 := loginAsMember(t)
	access2, refresh2 := refreshTokens(t, refresh1)
	expectAccepted(t, "access token after refresh", access2)

	if err := h.checkResponseCode(http.StatusAccepted, logout("Bearer "+access1)); err != nil {
		t.Fatal(err)
	}
	expectRevoked(t, "access token from login", access1)
	expectRevoked(t, "access token from refresh", access2)
	for name, token := range map[string]string{"login": refresh1, "refresh": refresh2} {
		if err := h.checkResponseCode(http.StatusUnauthorized, refresh(token)); err != nil {
			t.Errorf("refresh token from %s: %v", name, err)
		}
	}
}

// Logging out of one session does not end the member's other sessions.
func TestLogoutKeepsOtherSessions(t *testing.T) {
	h.clearTables()
	access1, _ := loginAsMember(t)
	access2, refresh2 := loginTokens(t, "member", "member")

	logout("Bearer " + access1)
	expectAccepted(t, "other session", access2)
	refreshTokens(t, refresh2)
}

// Changing the password ends the member's other sessions but keeps the current one.
func TestChangePasswordRevokesOtherSessions(t *testing.T) {
	h.clearTables()
	current, currentRefresh := loginAsMember(t)
	other, otherRefresh := loginTokens(t, "member", "member")

	req, _ := http.NewRequest("POST", "/api/v1/change_password", bytes.NewBufferString(`{"password":"changed"}`))
	req.Header.Add("Authorization", "Bearer "+current)
	if err := h.checkResponseCode(http.StatusOK, h.executeRequest(req).Code); err != nil {
		t.Fatal(err)
	}
	expectAccepted(t, "current session", current)
	refreshTokens(t, currentRefresh)
	expectRevoked(t, "other session", other)
	if err := h.checkResponseCode(http.StatusUnauthorized, refresh(otherRefresh)); err != nil {
		t.Error("other session refresh:", err)
	}
}

// A password reset through an email link ends all the member's sessions.
func TestResetLinkRevokesAllSessions(t *testing.T) {
	h.clearTables()
	access1, refresh1 := loginAsMember(t)
	access2, _ := loginTokens(t, "member", "member")
	token, err := controller.ResetCredentialsToken(context.Background(), "deadbeef", "ramon@gerard.ca", 60)
	if err != nil {
		t.Fatal(err)
	}

	if err := h.checkResponseCode(http.StatusOK, resetCredentials(token, "reset")); err != nil {
		t.Fatal(err)
	}
	expectRevoked(t, "session 1", access1)
	expectRevoked(t, "session 2", access2)
	if err := h.checkResponseCode(http.StatusUnauthorized, refresh(refresh1)); err != nil {
		t.Error(err)
	}
}

// Deleting a member ends their sessions.
func TestDeleteMemberRevokesSessions(t *testing.T) {
	h.clearTables()
	admin := h.addAnAdmin()
	access, refreshToken := loginAsMember(t)

	req, _ := http.NewRequest("DELETE", "/api/v1/members/deadbeef", nil)
	req.Header.Add("Authorization", "Bearer "+admin)
	if err := h.checkResponseCode(http.StatusOK, h.executeRequest(req).Code); err != nil {
		t.Fatal(err)
	}
	expectRevoked(t, "deleted member", access)
	if err := h.checkResponseCode(http.StatusUnauthorized, refresh(refreshToken)); err != nil {
		t.Error(err)
	}
	expectAccepted(t, "admin", admin)
}

// Refresh tokens issued before sessions existed keep working and move to a session.
func TestLegacyRefreshTokenMovesToSession(t *testing.T) {
	h.clearTables()
	loginAsMember(t)
	ctx := context.Background()
	if err := controller.RedisClient.Set(ctx, "legacyrefresh", "deadbeef", time.Hour).Err(); err != nil {
		t.Fatal(err)
	}
	legacy := sign(t, jwt.SigningMethodHS256, jwt.MapClaims{
		"token_uuid": "legacyrefresh",
		"user_uuid":  "deadbeef",
		"exp":        time.Now().Add(time.Hour).Unix(),
	}, []byte(os.Getenv("APP_REFRESH_SECRET")))

	access, newRefresh := refreshTokens(t, legacy)
	sessionID, _ := claimsOf(t, access)["session_id"].(string)
	if sessionID == "" || claimsOf(t, newRefresh)["session_id"] != sessionID {
		t.Fatalf("tokens not attached to a session: %v", claimsOf(t, access))
	}
	expectAccepted(t, "migrated session", access)
	logout("Bearer " + access)
	expectRevoked(t, "migrated session after logout", access)
}

// Sessions expire in Redis after refresh_ttl_days without refresh.
func TestSessionKeysHaveTTL(t *testing.T) {
	h.clearTables()
	access, _ := loginAsMember(t)
	ctx := context.Background()
	sessionID := claimsOf(t, access)["session_id"].(string)
	for _, key := range []string{"session:" + sessionID, "user_sessions:deadbeef"} {
		ttl, err := controller.RedisClient.TTL(ctx, key).Result()
		if err != nil {
			t.Fatal(err)
		}
		if ttl <= 0 || ttl > 15*24*time.Hour {
			t.Errorf("%s has TTL %v", key, ttl)
		}
	}
}

// GET /events/{uuid} tells clients to refresh an expired access token.
func TestGetEventWithExpiredToken(t *testing.T) {
	h.clearTables()
	h.addEvent("deadbeef", "Diada", int(time.Now().Unix())+3600, int(time.Now().Unix())+7200)
	accessToken, _ := loginAsMember(t)
	expired := claimsOf(t, accessToken)
	expired["exp"] = time.Now().Add(-time.Minute).Unix()
	token := sign(t, jwt.SigningMethodHS256, expired, []byte(os.Getenv("APP_ACCESS_SECRET")))

	req, _ := http.NewRequest("GET", "/api/v1/events/deadbeef", nil)
	req.Header.Add("Authorization", "Bearer "+token)
	response := h.executeRequest(req)
	var body map[string]string
	json.Unmarshal(response.Body.Bytes(), &body)
	if response.Code != http.StatusForbidden || body["code"] != "token_expired" {
		t.Errorf("expected 403 token_expired, got %d %v", response.Code, body)
	}
}
