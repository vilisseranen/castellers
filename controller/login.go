package controller

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/vilisseranen/castellers/common"
	"github.com/vilisseranen/castellers/model"
)

const (
	ERRORCREATETOKEN      = "error creating the token"
	ERRORREFRESHTOKEN     = "error refreshing token"
	ERRORTOKENEXPIRED     = "token has expired"
	ERRORTOKENINVALID     = "token is invalid"
	ERRORGUESTCANNOTLOGIN = "guests cannot login"
)

type TokenDetails struct {
	AccessToken  string
	RefreshToken string
	AccessUuid   string
	RefreshUuid  string
	AtExpires    int64
	RtExpires    int64
}

type AccessTokenDetails struct {
	TokenUuid   string
	RefreshUuid string
	UserId      string
	Permissions []string
}

const ResetCredentialsPermission = "reset_credentials"
const ParticipateEventPermission = "participate_event"

func Login(w http.ResponseWriter, r *http.Request) {
	ctx, span := tracer.Start(r.Context(), "Login")
	defer span.End()

	var credentialsInRequest model.Credentials
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&credentialsInRequest); err != nil {
		common.Debug("Error decoding login request: %s", err.Error())
		RespondWithError(w, http.StatusBadRequest, ERRORINVALIDPAYLOAD)
		return
	}
	credentialsInDB := model.Credentials{Username: credentialsInRequest.Username}
	if err := credentialsInDB.GetCredentials(ctx); err != nil {
		switch err {
		case sql.ErrNoRows:
			common.Info("User has no credentials: %s", err.Error())
			RespondWithError(w, http.StatusUnauthorized, ERRORUNAUTHORIZED)
			return
		default:
			common.Warn("Error getting credentials: %s", err.Error())
			RespondWithError(w, http.StatusInternalServerError, ERRORINTERNAL)
			return
		}
	}
	err := common.CompareHashAndPassword(credentialsInDB.PasswordHashed, credentialsInRequest.Password)
	if err != nil {
		common.Debug("Wrong password: %s", err.Error())
		RespondWithError(w, http.StatusUnauthorized, ERRORUNAUTHORIZED)
		return
	}
	tokens, err := createMemberToken(ctx, credentialsInDB.UUID)
	if err != nil {
		common.Warn("Error creating the token: %s", err.Error())
		RespondWithError(w, http.StatusInternalServerError, ERRORCREATETOKEN)
		return
	}
	RespondWithJSON(w, http.StatusOK, tokens)
}

func createMemberToken(ctx context.Context, uuid string) (map[string]string, error) {
	ctx, span := tracer.Start(ctx, "createMemberToken")
	defer span.End()

	permissions, err := getMemberPermissions(ctx, uuid)
	if err != nil {
		return nil, err
	}
	token, err := createToken(ctx, uuid, "", permissions, common.GetConfigInt("jwt.access_ttl_minutes"), common.GetConfigInt("jwt.refresh_ttl_days"))
	if err != nil {
		return nil, err
	}
	tokens := map[string]string{
		"access_token":  token.AccessToken,
		"refresh_token": token.RefreshToken,
	}
	return tokens, nil
}

// Logout revokes the session (refresh token) of the access token. It also
// works with an expired access token, and succeeds if the session was already
// revoked, so clients can always call it.
func Logout(w http.ResponseWriter, r *http.Request) {
	ctx, span := tracer.Start(r.Context(), "Logout")
	defer span.End()

	token, err := parseTokenIgnoringExpiry(extractTokenString(ctx, r), common.GetConfigString("jwt.access_secret"))
	if err != nil {
		common.Debug("Invalid token on logout: %s", err.Error())
		RespondWithError(w, http.StatusUnauthorized, ERRORUNAUTHORIZED)
		return
	}
	claims, _ := token.Claims.(jwt.MapClaims)
	if refreshUuid, ok := claims["refresh_uuid"].(string); ok && refreshUuid != "" {
		if _, err := deleteTokenInCache(ctx, refreshUuid); err != nil {
			common.Warn("Cannot delete refresh token in cache: %s", err.Error())
			RespondWithError(w, http.StatusInternalServerError, ERRORINTERNAL)
			return
		}
	}
	RespondWithJSON(w, http.StatusAccepted, "Successfully logged out")
}

// parseTokenIgnoringExpiry checks the signature of an HS256 token but accepts it even if expired.
func parseTokenIgnoringExpiry(tokenString, secret string) (*jwt.Token, error) {
	return jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return []byte(secret), nil
	}, jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}), jwt.WithoutClaimsValidation())
}

func createToken(ctx context.Context, uuid string, email string, permissions []string, access_ttl_minutes, refresh_ttl_days int) (*TokenDetails, error) {
	ctx, span := tracer.Start(ctx, "createToken")
	defer span.End()

	td := &TokenDetails{}
	td.AtExpires = time.Now().Add(time.Minute * time.Duration(access_ttl_minutes)).Unix()
	td.AccessUuid = common.GenerateUUID()

	td.RtExpires = time.Now().Add(time.Hour * 24 * time.Duration(refresh_ttl_days)).Unix()
	td.RefreshUuid = common.GenerateUUID()

	var err error
	//Creating Access Token
	atClaims := jwt.MapClaims{}
	atClaims["token_uuid"] = td.AccessUuid
	atClaims["refresh_uuid"] = td.RefreshUuid
	atClaims["user_uuid"] = uuid
	atClaims["permissions"] = permissions
	atClaims["exp"] = td.AtExpires
	if email != "" {
		atClaims["email"] = email
	}
	at := jwt.NewWithClaims(jwt.SigningMethodHS256, atClaims)
	td.AccessToken, err = at.SignedString([]byte(common.GetConfigString("jwt.access_secret")))
	if err != nil {
		return nil, err
	}
	//Creating Refresh Token
	rtClaims := jwt.MapClaims{}
	rtClaims["token_uuid"] = td.RefreshUuid
	rtClaims["user_uuid"] = uuid
	rtClaims["exp"] = td.RtExpires
	rt := jwt.NewWithClaims(jwt.SigningMethodHS256, rtClaims)
	td.RefreshToken, err = rt.SignedString([]byte(common.GetConfigString("jwt.refresh_secret")))
	if err != nil {
		return nil, err
	}
	// save in cache
	if err := saveTokenInCache(ctx, uuid, td); err != nil {
		return nil, err
	}
	return td, nil
}

// extractTokenString returns the token of a "Bearer <token>" Authorization
// header. "Bearer: <token>", which the UI used to send, is also accepted.
// Any other scheme is rejected.
func extractTokenString(ctx context.Context, r *http.Request) string {
	scheme, token, found := strings.Cut(strings.TrimSpace(r.Header.Get("Authorization")), " ")
	if !found {
		return ""
	}
	scheme = strings.TrimSuffix(scheme, ":")
	token = strings.TrimSpace(token)
	if !strings.EqualFold(scheme, "Bearer") || token == "" || strings.Contains(token, " ") {
		return ""
	}
	return token
}

func requestHasAuthorizationToken(ctx context.Context, r *http.Request) bool {
	ctx, span := tracer.Start(ctx, "requestHasAuthorizationToken")
	defer span.End()

	return extractTokenString(ctx, r) != ""
}

func verifyToken(ctx context.Context, tokenString, tokenType string) (*jwt.Token, error) {
	_, span := tracer.Start(ctx, "requestHasAuthorizationToken")
	defer span.End()

	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			common.Debug("Incorrect signing method")
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		if tokenType == "access" {
			return []byte(common.GetConfigString("jwt.access_secret")), nil
		} else if tokenType == "refresh" {
			return []byte(common.GetConfigString("jwt.refresh_secret")), nil
		}
		return nil, errors.New("unsupported token type")
	}, jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}), jwt.WithExpirationRequired())
	if err != nil {
		return nil, err
	}
	return token, nil
}

// IsTokenExpired reports whether err, as returned by ExtractToken, means the token has expired.
// Clients rely on this to know when to call /refresh.
func IsTokenExpired(err error) bool {
	return errors.Is(err, jwt.ErrTokenExpired)
}

func ExtractToken(ctx context.Context, r *http.Request) (*AccessTokenDetails, error) {
	ctx, span := tracer.Start(ctx, "ExtractToken")
	defer span.End()
	tokenString := extractTokenString(ctx, r)
	token, err := verifyToken(ctx, tokenString, "access")
	if err != nil {
		return nil, err
	}
	claims, ok := token.Claims.(jwt.MapClaims)
	if ok && token.Valid {
		err = errors.New("error decoding token")
		tokenUuid, ok := claims["token_uuid"].(string)
		if !ok {
			return nil, err
		}
		refreshUuid, ok := claims["refresh_uuid"].(string)
		if !ok {
			return nil, err
		}
		userUuid, ok := claims["user_uuid"].(string)
		if !ok {
			return nil, err
		}
		// The "permissions" claim is a slice of interfaces
		// We need to cast each element to a string
		tokenPermissionsInterface, ok := claims["permissions"].([]interface{})
		tokenPermissions := make([]string, len(tokenPermissionsInterface))
		for i, v := range tokenPermissionsInterface {
			tokenPermissions[i] = v.(string)
		}
		if !ok {
			return nil, err
		}
		return &AccessTokenDetails{
			TokenUuid:   tokenUuid,
			RefreshUuid: refreshUuid,
			UserId:      userUuid,
			Permissions: tokenPermissions,
		}, nil
	}
	return nil, err
}

func saveTokenInCache(ctx context.Context, uuid string, td *TokenDetails) error {
	ctx, span := tracer.Start(ctx, "saveTokenInCache")
	defer span.End()

	rt := time.Unix(td.RtExpires, 0)
	now := time.Now()

	errRefresh := RedisClient.Set(ctx, td.RefreshUuid, uuid, rt.Sub(now)).Err()
	if errRefresh != nil {
		return errRefresh
	}
	return nil
}

func deleteTokenInCache(ctx context.Context, uuid string) (int64, error) {
	ctx, span := tracer.Start(ctx, "deleteTokenInCache")
	defer span.End()

	deleted, err := RedisClient.Del(ctx, uuid).Result()
	if err != nil {
		return 0, err
	}
	return deleted, nil
}

func checkTokenInCache(ctx context.Context, token *jwt.Token) (string, error) {
	ctx, span := tracer.Start(ctx, "checkTokenInCache")
	defer span.End()

	claims, ok := token.Claims.(jwt.MapClaims)
	err := errors.New("error decoding token")
	if ok && token.Valid {
		tokenUuid, ok := claims["token_uuid"].(string)
		if !ok {
			return "", err
		}
		userUuid, err := RedisClient.Get(ctx, tokenUuid).Result()
		if err != nil {
			return "", err
		}
		return userUuid, nil
	}
	return "", err
}

func RefreshToken(w http.ResponseWriter, r *http.Request) {
	ctx, span := tracer.Start(r.Context(), "ExtractToken")
	defer span.End()
	mapToken := map[string]string{}
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&mapToken); err != nil {
		common.Warn("Error decoding token: %s", err.Error())
		RespondWithError(w, http.StatusUnprocessableEntity, ERRORREFRESHTOKEN)
		return
	}
	refreshToken := mapToken["refresh_token"]

	token, err := verifyToken(ctx, refreshToken, "refresh")
	//if there is an error, the token must have expired
	if err != nil {
		common.Debug("Token verification failed: %s", err.Error())
		RespondWithError(w, http.StatusUnauthorized, ERRORTOKENEXPIRED)
		return
	}
	_, err = checkTokenInCache(ctx, token)
	if err != nil {
		common.Debug("Cannot find token in cache: %s", err.Error())
		RespondWithError(w, http.StatusUnauthorized, ERRORTOKENINVALID)
		return
	}
	if !token.Valid {
		common.Debug("Token invalid: %s", err.Error())
		RespondWithError(w, http.StatusUnauthorized, ERRORTOKENINVALID)
		return
	}
	claims, ok := token.Claims.(jwt.MapClaims)
	if ok && token.Valid {
		userUuid, ok := claims["user_uuid"].(string)
		if !ok {
			common.Info("Token claim 'user_uuid' is invalid")
			RespondWithError(w, http.StatusUnprocessableEntity, ERRORTOKENINVALID)
			return
		}
		// TODO: Link refresh token to a parent DEVICE ID.
		//       After that, we only delete the refresh tokens
		//       from this device ID when the device is logged out.
		// deleted, delErr := deleteTokenInCache(ctx, refreshUuid)
		// if delErr != nil || deleted == 0 { //if any goes wrong
		// 	common.Warn("Error deleting token in cache: %s", delErr.Error())
		// 	RespondWithError(w, http.StatusUnauthorized, ERRORUNAUTHORIZED)
		// 	return
		// }
		permissions, err := getMemberPermissions(ctx, userUuid)
		if err != nil {
			common.Warn("Error getting permissions: %s", err.Error())
			RespondWithError(w, http.StatusInternalServerError, ERRORINTERNAL)
			return
		}
		//Create new pairs of refresh and access tokens
		ts, createErr := createToken(ctx, userUuid, "", permissions, common.GetConfigInt("jwt.access_ttl_minutes"), common.GetConfigInt("jwt.refresh_ttl_days"))
		if createErr != nil {
			common.Warn("Error creating a new token pair: %s", createErr.Error())
			RespondWithError(w, http.StatusInternalServerError, ERRORINTERNAL)
			return
		}
		tokens := map[string]string{
			"access_token":  ts.AccessToken,
			"refresh_token": ts.RefreshToken,
		}
		RespondWithJSON(w, http.StatusCreated, tokens)
	} else {
		common.Debug("Refresh token has expired")
		RespondWithError(w, http.StatusUnauthorized, ERRORTOKENEXPIRED)
	}
}

// linkTokenKeyPrefix prefixes the Redis keys of single-use link tokens.
const linkTokenKeyPrefix = "link:"

// createLinkToken creates an access token for a link sent by email. Unlike a
// login, it has no refresh token and no session in Redis.
func createLinkToken(uuid, email, permission string, ttlMinutes int) (token string, tokenUuid string, err error) {
	tokenUuid = common.GenerateUUID()
	claims := jwt.MapClaims{
		"token_uuid":   tokenUuid,
		"refresh_uuid": "",
		"user_uuid":    uuid,
		"permissions":  []string{permission},
		"exp":          time.Now().Add(time.Minute * time.Duration(ttlMinutes)).Unix(),
	}
	if email != "" {
		claims["email"] = email
	}
	token, err = jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(common.GetConfigString("jwt.access_secret")))
	return token, tokenUuid, err
}

// ResetCredentialsToken creates the token of a registration or password reset
// link. It can be used only once (see consumeResetCredentialsToken).
func ResetCredentialsToken(ctx context.Context, uuid string, email string, ttl int) (string, error) {
	ctx, span := tracer.Start(ctx, "ResetCredentialsToken")
	defer span.End()
	token, tokenUuid, err := createLinkToken(uuid, email, ResetCredentialsPermission, ttl)
	if err != nil {
		return "", err
	}
	if err := RedisClient.Set(ctx, linkTokenKeyPrefix+tokenUuid, uuid, time.Minute*time.Duration(ttl)).Err(); err != nil {
		return "", err
	}
	return token, nil
}

// consumeResetCredentialsToken marks a reset credentials token as used. It
// returns false if the token was already used.
func consumeResetCredentialsToken(ctx context.Context, tokenAuth *AccessTokenDetails) (bool, error) {
	deleted, err := RedisClient.Del(ctx, linkTokenKeyPrefix+tokenAuth.TokenUuid).Result()
	if err != nil || deleted == 1 {
		return deleted == 1, err
	}
	// Tokens created before version 0.24.4 came with a refresh token stored in
	// Redis: consume that key instead.
	if tokenAuth.RefreshUuid == "" {
		return false, nil
	}
	deleted, err = RedisClient.Del(ctx, tokenAuth.RefreshUuid).Result()
	return deleted == 1, err
}

// ParticipateEventToken creates the token of the participation links sent in reminders.
func ParticipateEventToken(ctx context.Context, uuid string, ttl int) (string, error) {
	_, span := tracer.Start(ctx, "ParticipateEventToken")
	defer span.End()
	token, _, err := createLinkToken(uuid, "", ParticipateEventPermission, ttl)
	return token, err
}

func getMemberPermissions(ctx context.Context, uuid string) ([]string, error) {
	ctx, span := tracer.Start(ctx, "getMemberPermissions")
	defer span.End()

	member := model.Member{UUID: uuid}
	if err := member.Get(ctx); err != nil {
		return []string{}, err
	}
	var permissions []string
	if member.Type == model.MEMBERSTYPEREGULAR {
		permissions = append(permissions, model.MEMBERSTYPEREGULAR)
	} else if member.Type == model.MEMBERSTYPEADMIN {
		permissions = append(permissions, model.MEMBERSTYPEREGULAR)
		permissions = append(permissions, model.MEMBERSTYPEADMIN)
	} else {
		return []string{}, errors.New(ERRORGUESTCANNOTLOGIN)
	}
	return permissions, nil
}

func ForgotPassword(w http.ResponseWriter, r *http.Request) {
	ctx, span := tracer.Start(r.Context(), "ForgotPassword")
	defer span.End()
	var member model.Member
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&member); err != nil {
		common.Debug("Cannot decode forget password request: %s", err.Error())
		RespondWithError(w, http.StatusUnprocessableEntity, ERRORINVALIDPAYLOAD)
		return
	}
	err := member.GetByEmail(ctx)
	// Even if the member is not found, we cannot give an error because that will be a way to
	// find which users are in the system
	if err != nil {
		common.Warn("Cannot get user by email: %s", err.Error())
		RespondWithJSON(w, http.StatusAccepted, "")
		return
	}
	if err == nil {
		n := model.Notification{NotificationType: model.TypeForgotPassword, ObjectUUID: member.UUID, SendDate: int(time.Now().Unix())}
		if err := n.CreateNotification(ctx); err != nil {
			common.Warn("Error creating notification: %s", err.Error())
			RespondWithError(w, http.StatusInternalServerError, ERRORNOTIFICATION)
			return
		}
	}
	RespondWithJSON(w, http.StatusAccepted, "")
}
