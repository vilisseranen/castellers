package tests

import (
	"bytes"
	"encoding/json"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func loginAsMember(t *testing.T) (string, string) {
	t.Helper()
	h.addMember("deadbeef", "Ramon", "Gerard", "", "", "", "segon", "member", "ramon@gerard.ca", "")
	h.addCredentials("deadbeef", "member", "member")
	req, _ := http.NewRequest("POST", "/api/v1/login", bytes.NewBufferString(`{"username":"member","password":"member"}`))
	response := h.executeRequest(req)
	if err := h.checkResponseCode(http.StatusOK, response.Code); err != nil {
		t.Fatal(err)
	}
	var tokens map[string]string
	if err := json.Unmarshal(response.Body.Bytes(), &tokens); err != nil {
		t.Fatal(err)
	}
	if tokens["access_token"] == "" || tokens["refresh_token"] == "" {
		t.Fatalf("login did not return tokens: %s", response.Body.String())
	}
	return tokens["access_token"], tokens["refresh_token"]
}

// claimsOf decodes a token without verifying it, so tests can forge variants of a real token.
func claimsOf(t *testing.T, token string) jwt.MapClaims {
	t.Helper()
	claims := jwt.MapClaims{}
	if _, _, err := jwt.NewParser().ParseUnverified(token, claims); err != nil {
		t.Fatal(err)
	}
	return claims
}

func sign(t *testing.T, method jwt.SigningMethod, claims jwt.MapClaims, key interface{}) string {
	t.Helper()
	s, err := jwt.NewWithClaims(method, claims).SignedString(key)
	if err != nil {
		t.Fatal(err)
	}
	return s
}

func refresh(refreshToken string) int {
	payload, _ := json.Marshal(map[string]string{"refresh_token": refreshToken})
	req, _ := http.NewRequest("POST", "/api/v1/refresh", bytes.NewBuffer(payload))
	return h.executeRequest(req).Code
}

func getWithAccessToken(path, accessToken string) int {
	req, _ := http.NewRequest("GET", path, nil)
	req.Header.Add("Authorization", "Bearer "+accessToken)
	return h.executeRequest(req).Code
}

func TestRefreshToken(t *testing.T) {
	h.clearTables()
	_, refreshToken := loginAsMember(t)

	if err := h.checkResponseCode(http.StatusCreated, refresh(refreshToken)); err != nil {
		t.Error(err)
	}
}

func TestAccessTokenAccepted(t *testing.T) {
	h.clearTables()
	accessToken, _ := loginAsMember(t)

	if err := h.checkResponseCode(http.StatusOK, getWithAccessToken("/api/v1/castells/models", accessToken)); err != nil {
		t.Error(err)
	}
}

func TestForgedRefreshTokensRejected(t *testing.T) {
	h.clearTables()
	_, refreshToken := loginAsMember(t)
	secret := []byte(os.Getenv("APP_REFRESH_SECRET"))

	expired := claimsOf(t, refreshToken)
	expired["exp"] = time.Now().Add(-time.Minute).Unix()

	noExp := claimsOf(t, refreshToken)
	delete(noExp, "exp")

	cases := map[string]string{
		"expired":      sign(t, jwt.SigningMethodHS256, expired, secret),
		"missing exp":  sign(t, jwt.SigningMethodHS256, noExp, secret),
		"wrong secret": sign(t, jwt.SigningMethodHS256, claimsOf(t, refreshToken), []byte("not-the-secret")),
		"alg none":     sign(t, jwt.SigningMethodNone, claimsOf(t, refreshToken), jwt.UnsafeAllowNoneSignatureType),
		"alg HS512":    sign(t, jwt.SigningMethodHS512, claimsOf(t, refreshToken), secret),
		"garbage":      "not.a.token",
	}
	for name, token := range cases {
		if err := h.checkResponseCode(http.StatusUnauthorized, refresh(token)); err != nil {
			t.Errorf("%s: %v", name, err)
		}
	}
}

func TestForgedAccessTokensRejected(t *testing.T) {
	h.clearTables()
	accessToken, refreshToken := loginAsMember(t)
	secret := []byte(os.Getenv("APP_ACCESS_SECRET"))

	noExp := claimsOf(t, accessToken)
	delete(noExp, "exp")

	cases := map[string]string{
		"missing exp":   sign(t, jwt.SigningMethodHS256, noExp, secret),
		"wrong secret":  sign(t, jwt.SigningMethodHS256, claimsOf(t, accessToken), []byte("not-the-secret")),
		"alg none":      sign(t, jwt.SigningMethodNone, claimsOf(t, accessToken), jwt.UnsafeAllowNoneSignatureType),
		"alg HS512":     sign(t, jwt.SigningMethodHS512, claimsOf(t, accessToken), secret),
		"refresh token": refreshToken,
	}
	for name, token := range cases {
		if err := h.checkResponseCode(http.StatusUnauthorized, getWithAccessToken("/api/v1/castells/models", token)); err != nil {
			t.Errorf("%s: %v", name, err)
		}
	}
}

// The UI calls /refresh when it gets a 403, so an expired access token must give exactly 403.
func TestExpiredAccessTokenReturnsForbidden(t *testing.T) {
	h.clearTables()
	accessToken, _ := loginAsMember(t)

	expired := claimsOf(t, accessToken)
	expired["exp"] = time.Now().Add(-time.Minute).Unix()
	token := sign(t, jwt.SigningMethodHS256, expired, []byte(os.Getenv("APP_ACCESS_SECRET")))

	for _, path := range []string{"/api/v1/castells/models", "/api/v1/events"} {
		if err := h.checkResponseCode(http.StatusForbidden, getWithAccessToken(path, token)); err != nil {
			t.Errorf("%s: %v", path, err)
		}
	}
}
