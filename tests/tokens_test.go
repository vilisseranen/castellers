package tests

import (
	"bytes"
	"context"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/vilisseranen/castellers/controller"
)

func resetCredentials(token, password string) int {
	req, _ := http.NewRequest("POST", "/api/v1/reset_credentials", bytes.NewBufferString(`{"password":"`+password+`"}`))
	req.Header.Add("Authorization", "Bearer "+token)
	return h.executeRequest(req).Code
}

func login(username, password string) int {
	req, _ := http.NewRequest("POST", "/api/v1/login", bytes.NewBufferString(`{"username":"`+username+`","password":"`+password+`"}`))
	return h.executeRequest(req).Code
}

func logout(authorization string) int {
	req, _ := http.NewRequest("POST", "/api/v1/logout", nil)
	if authorization != "" {
		req.Header.Add("Authorization", authorization)
	}
	return h.executeRequest(req).Code
}

// Registration and "forgot password" links can only be used once.
func TestResetCredentialsTokenIsSingleUse(t *testing.T) {
	h.clearTables()
	loginAsMember(t)
	token, err := controller.ResetCredentialsToken(context.Background(), "deadbeef", "ramon@gerard.ca", 60)
	if err != nil {
		t.Fatal(err)
	}

	if err := h.checkResponseCode(http.StatusOK, resetCredentials(token, "first")); err != nil {
		t.Fatal(err)
	}
	if err := h.checkResponseCode(http.StatusUnauthorized, resetCredentials(token, "second")); err != nil {
		t.Error("token reused:", err)
	}
	if err := h.checkResponseCode(http.StatusOK, login("member", "first")); err != nil {
		t.Error(err)
	}
	if err := h.checkResponseCode(http.StatusUnauthorized, login("member", "second")); err != nil {
		t.Error(err)
	}
}

// Reset links created before this version came with a refresh token stored in
// Redis: they keep working once, then are rejected.
func TestLegacyResetCredentialsTokenIsSingleUse(t *testing.T) {
	h.clearTables()
	loginAsMember(t)
	refreshUuid := "legacyrefreshuuid"
	if err := controller.RedisClient.Set(context.Background(), refreshUuid, "deadbeef", 0).Err(); err != nil {
		t.Fatal(err)
	}
	token := sign(t, jwt.SigningMethodHS256, jwt.MapClaims{
		"token_uuid":   "legacytokenuuid",
		"refresh_uuid": refreshUuid,
		"user_uuid":    "deadbeef",
		"permissions":  []string{controller.ResetCredentialsPermission},
		"exp":          time.Now().Add(time.Hour).Unix(),
	}, []byte(os.Getenv("APP_ACCESS_SECRET")))

	if err := h.checkResponseCode(http.StatusOK, resetCredentials(token, "first")); err != nil {
		t.Fatal(err)
	}
	if err := h.checkResponseCode(http.StatusUnauthorized, resetCredentials(token, "second")); err != nil {
		t.Error("legacy token reused:", err)
	}
}

// Link tokens must not leave Redis keys without expiry behind.
func TestLinkTokensRedisKeys(t *testing.T) {
	h.clearTables()
	ctx := context.Background()
	before, _ := controller.RedisClient.DBSize(ctx).Result()

	if _, err := controller.ParticipateEventToken(ctx, "deadbeef", 2880); err != nil {
		t.Fatal(err)
	}
	afterParticipation, _ := controller.RedisClient.DBSize(ctx).Result()
	if afterParticipation != before {
		t.Errorf("participation token created %d Redis keys", afterParticipation-before)
	}

	token, err := controller.ResetCredentialsToken(ctx, "deadbeef", "ramon@gerard.ca", 60)
	if err != nil {
		t.Fatal(err)
	}
	key := "link:" + claimsOf(t, token)["token_uuid"].(string)
	ttl, err := controller.RedisClient.TTL(ctx, key).Result()
	if err != nil {
		t.Fatal(err)
	}
	if ttl <= 0 || ttl > time.Hour {
		t.Errorf("reset token key %s has TTL %v, want (0, 1h]", key, ttl)
	}
}

// Logout revokes the session, even with an expired access token, and can be called twice.
func TestLogoutRevokesSession(t *testing.T) {
	h.clearTables()
	accessToken, refreshToken := loginAsMember(t)

	if err := h.checkResponseCode(http.StatusAccepted, logout("Bearer "+accessToken)); err != nil {
		t.Fatal(err)
	}
	if err := h.checkResponseCode(http.StatusUnauthorized, refresh(refreshToken)); err != nil {
		t.Error("refresh after logout:", err)
	}
	if err := h.checkResponseCode(http.StatusAccepted, logout("Bearer "+accessToken)); err != nil {
		t.Error("second logout:", err)
	}
	if err := h.checkResponseCode(http.StatusUnauthorized, logout("Bearer not.a.token")); err != nil {
		t.Error("logout with an invalid token:", err)
	}
	if err := h.checkResponseCode(http.StatusUnauthorized, logout("")); err != nil {
		t.Error("logout without token:", err)
	}
}

func TestLogoutWithExpiredAccessToken(t *testing.T) {
	h.clearTables()
	accessToken, refreshToken := loginAsMember(t)
	expired := claimsOf(t, accessToken)
	expired["exp"] = time.Now().Add(-time.Minute).Unix()
	token := sign(t, jwt.SigningMethodHS256, expired, []byte(os.Getenv("APP_ACCESS_SECRET")))

	if err := h.checkResponseCode(http.StatusAccepted, logout("Bearer "+token)); err != nil {
		t.Fatal(err)
	}
	if err := h.checkResponseCode(http.StatusUnauthorized, refresh(refreshToken)); err != nil {
		t.Error("refresh after logout:", err)
	}
}

func TestAuthorizationHeaderSchemes(t *testing.T) {
	h.clearTables()
	accessToken, _ := loginAsMember(t)
	cases := map[string]int{
		"Bearer " + accessToken:        http.StatusOK,
		"bearer " + accessToken:        http.StatusOK,
		"Bearer: " + accessToken:       http.StatusOK, // sent by the UI so far
		"Basic " + accessToken:         http.StatusUnauthorized,
		accessToken:                    http.StatusUnauthorized,
		"Bearer " + accessToken + " x": http.StatusUnauthorized,
	}
	for header, want := range cases {
		req, _ := http.NewRequest("GET", "/api/v1/castells/models", nil)
		req.Header.Add("Authorization", header)
		if err := h.checkResponseCode(want, h.executeRequest(req).Code); err != nil {
			t.Errorf("%.20q...: %v", header, err)
		}
	}
}
