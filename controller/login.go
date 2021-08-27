package controller

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt"
	"github.com/vilisseranen/castellers/common"
	"github.com/vilisseranen/castellers/model"
)

const (
	ERRORCREATETOKEN  = "Error creating the token"
	ERRORREFRESHTOKEN = "Error refreshing token"
	ERRORTOKENEXPIRED = "Token has expired"
	ERRORTOKENINVALID = "Token is invalid"
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
	var credentialsInRequest model.Credentials
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&credentialsInRequest); err != nil {
		common.Debug("Error decoding login request: %s", err.Error())
		RespondWithError(w, http.StatusBadRequest, ERRORINVALIDPAYLOAD)
		return
	}
	credentialsInDB := model.Credentials{Username: credentialsInRequest.Username}
	if err := credentialsInDB.GetCredentials(); err != nil {
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
	tokens, err := createMemberToken(credentialsInDB.UUID)
	if err != nil {
		common.Warn("Error creating the token: %s", err.Error())
		RespondWithError(w, http.StatusInternalServerError, ERRORCREATETOKEN)
		return
	}
	RespondWithJSON(w, http.StatusOK, tokens)
}

func createMemberToken(uuid string) (map[string]string, error) {
	permissions, err := getMemberPermissions(uuid)
	if err != nil {
		return nil, err
	}
	token, err := createToken(uuid, permissions, common.GetConfigInt("jwt.access_ttl_minutes"), common.GetConfigInt("jwt.refresh_ttl_days"))
	if err != nil {
		return nil, err
	}
	tokens := map[string]string{
		"access_token":  token.AccessToken,
		"refresh_token": token.RefreshToken,
	}
	return tokens, nil
}

func Logout(w http.ResponseWriter, r *http.Request) {
	au, err := ExtractToken(r)
	if err != nil {
		common.Warn("Invalid token: %s", err.Error())
		RespondWithError(w, http.StatusBadRequest, ERRORUNAUTHORIZED)
		return
	}
	deleted, delErr := deleteTokenInCache(au.RefreshUuid)
	if delErr != nil || deleted == 0 {
		common.Warn("Cannot delete refresh token in cache: %s", err.Error())
		RespondWithError(w, http.StatusBadRequest, ERRORUNAUTHORIZED)
		return
	}
	deleted, delErr = deleteTokenInCache(au.TokenUuid)
	if delErr != nil || deleted == 0 {
		common.Warn("Cannot delete access token in cache: %s", err.Error())
		RespondWithError(w, http.StatusBadRequest, ERRORUNAUTHORIZED)
		return
	}
	RespondWithJSON(w, http.StatusAccepted, "Successfully logged out")
}

func createToken(uuid string, permissions []string, access_ttl_minutes, refresh_ttl_days int) (*TokenDetails, error) {
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
	saveErr := saveTokenInCache(uuid, td)
	if saveErr != nil {
		return nil, err
	}
	return td, nil
}

func extractTokenString(r *http.Request) string {
	bearToken := r.Header.Get("Authorization")
	strArr := strings.Split(bearToken, " ")
	if len(strArr) == 2 {
		return strArr[1]
	}
	return ""
}

func requestHasAuthorizationToken(r *http.Request) bool {
	return extractTokenString(r) != ""
}

func verifyToken(tokenString, tokenType string) (*jwt.Token, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			common.Debug("Incorrect signing method")
			return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
		}
		if tokenType == "access" {
			return []byte(common.GetConfigString("jwt.access_secret")), nil
		} else if tokenType == "refresh" {
			return []byte(common.GetConfigString("jwt.refresh_secret")), nil
		}
		return nil, errors.New("Unsupported token type")
	})
	if err != nil {
		// if err.(*jwt.ValidationError).Errors == jwt.ValidationErrorExpired {

		// }
		return nil, err
	}
	_, err = checkTokenInCache(token)
	if err != nil {
		common.Debug("Cannot find token in cache: %s", err.Error())
		return nil, err
	}
	return token, nil
}

func ExtractToken(r *http.Request) (*AccessTokenDetails, error) {
	tokenString := extractTokenString(r)
	token, err := verifyToken(tokenString, "access")
	if err != nil {
		return nil, err
	}
	claims, ok := token.Claims.(jwt.MapClaims)
	if ok && token.Valid {
		err = errors.New("Error decoding token")
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

func saveTokenInCache(uuid string, td *TokenDetails) error {
	// We add 1 second because we check in redis after we check the token
	// The token could be removed from redis right after we do the static validation
	at := time.Unix(td.AtExpires+1, 0)
	rt := time.Unix(td.RtExpires+1, 0)
	now := time.Now()

	errAccess := RedisClient.Set(td.AccessUuid, uuid, at.Sub(now)).Err()
	if errAccess != nil {
		return errAccess
	}
	errRefresh := RedisClient.Set(td.RefreshUuid, uuid, rt.Sub(now)).Err()
	if errRefresh != nil {
		return errRefresh
	}
	return nil
}

func deleteTokenInCache(uuid string) (int64, error) {
	deleted, err := RedisClient.Del(uuid).Result()
	if err != nil {
		return 0, err
	}
	return deleted, nil
}

func checkTokenInCache(token *jwt.Token) (string, error) {
	claims, ok := token.Claims.(jwt.MapClaims)
	err := errors.New("Error decoding token")
	if ok && token.Valid {
		tokenUuid, ok := claims["token_uuid"].(string)
		if !ok {
			return "", err
		}
		userUuid, err := RedisClient.Get(tokenUuid).Result()
		if err != nil {
			return "", err
		}
		return userUuid, nil
	}
	return "", err
}

func RefreshToken(w http.ResponseWriter, r *http.Request) {
	mapToken := map[string]string{}
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&mapToken); err != nil {
		common.Warn("Error decoding token: %s", err.Error())
		RespondWithError(w, http.StatusUnprocessableEntity, ERRORREFRESHTOKEN)
		return
	}
	refreshToken := mapToken["refresh_token"]

	token, err := verifyToken(refreshToken, "refresh")
	//if there is an error, the token must have expired
	if err != nil {
		common.Debug("Token verification failed: %s", err.Error())
		RespondWithError(w, http.StatusUnauthorized, ERRORTOKENEXPIRED)
		return
	}
	if _, ok := token.Claims.(jwt.Claims); !ok && !token.Valid {
		common.Debug("Token invalid: %s", err.Error())
		RespondWithError(w, http.StatusUnauthorized, ERRORTOKENINVALID)
		return
	}
	claims, ok := token.Claims.(jwt.MapClaims)
	if ok && token.Valid {
		refreshUuid, ok := claims["token_uuid"].(string)
		if !ok {
			common.Info("Token claim 'token_uuid' is invalid")
			RespondWithError(w, http.StatusUnprocessableEntity, ERRORTOKENINVALID)
			return
		}
		userUuid, ok := claims["user_uuid"].(string)
		if !ok {
			common.Info("Token claim 'user_uuid' is invalid")
			RespondWithError(w, http.StatusUnprocessableEntity, ERRORTOKENINVALID)
			return
		}
		//Delete the previous Refresh Token
		deleted, delErr := deleteTokenInCache(refreshUuid)
		if delErr != nil || deleted == 0 { //if any goes wrong
			common.Warn("Error deleting token in cache: %s", delErr.Error())
			RespondWithError(w, http.StatusUnauthorized, ERRORUNAUTHORIZED)
			return
		}
		permissions, err := getMemberPermissions(userUuid)
		if err != nil {
			common.Warn("Error getting permissions: %s", err.Error())
			RespondWithError(w, http.StatusInternalServerError, ERRORINTERNAL)
			return
		}
		//Create new pairs of refresh and access tokens
		ts, createErr := createToken(userUuid, permissions, common.GetConfigInt("jwt.access_ttl_minutes"), common.GetConfigInt("jwt.refresh_ttl_days"))
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

func ResetCredentialsToken(uuid string, ttl int) (string, error) {
	token, err := createToken(uuid, []string{ResetCredentialsPermission}, ttl, 0)
	return token.AccessToken, err
}

func ParticipateEventToken(uuid string, ttl int) (string, error) {
	token, err := createToken(uuid, []string{ParticipateEventPermission}, ttl, 0)
	return token.AccessToken, err
}

func getMemberPermissions(uuid string) ([]string, error) {
	member := model.Member{UUID: uuid}
	if err := member.Get(); err != nil {
		return []string{}, err
	}
	var permissions []string
	if member.Type == model.MemberTypeMember {
		permissions = append(permissions, model.MemberTypeMember)
	}
	if member.Type == model.MemberTypeAdmin {
		permissions = append(permissions, model.MemberTypeMember)
		permissions = append(permissions, model.MemberTypeAdmin)
	}
	return permissions, nil
}

func ForgotPassword(w http.ResponseWriter, r *http.Request) {
	var member model.Member
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&member); err != nil {
		common.Debug("Cannot decode forget password request: %s", err.Error())
		RespondWithError(w, http.StatusUnprocessableEntity, ERRORINVALIDPAYLOAD)
		return
	}
	err := member.GetByEmail()
	if err != nil {
		common.Warn("Cannot get user by email: %s", err.Error())
		RespondWithError(w, http.StatusInternalServerError, ERRORGETMEMBER)
		return
	}
	if err == nil {
		n := model.Notification{NotificationType: model.TypeForgotPassword, ObjectUUID: member.UUID, SendDate: int(time.Now().Unix())}
		if err := n.CreateNotification(); err != nil {
			common.Warn("Error creating notification: %s", err.Error())
			RespondWithError(w, http.StatusInternalServerError, ERRORNOTIFICATION)
			return
		}
	}
	RespondWithJSON(w, http.StatusAccepted, "")
}
