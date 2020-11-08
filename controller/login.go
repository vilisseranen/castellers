package controller

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/vilisseranen/castellers/common"
	"github.com/vilisseranen/castellers/model"
)

type Credentials struct {
	ID   string `json:"id"`
	Code string `json:"code"`
}

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

const createCredentialsPermission = "create_credentials"
const resetCredentialsPermission = "reset_credentials"

func Login(w http.ResponseWriter, r *http.Request) {
	var credentialsInRequest model.Credentials
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&credentialsInRequest); err != nil {
		RespondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}
	credentialsInDB := model.Credentials{Username: credentialsInRequest.Username}
	if err := credentialsInDB.GetCredentials(); err != nil {
		switch err {
		case sql.ErrNoRows:
			RespondWithError(w, http.StatusUnauthorized, UnauthorizedMessage)
			return
		default:
			RespondWithError(w, http.StatusInternalServerError, err.Error())
			return
		}
	}
	err := common.CompareHashAndPassword(credentialsInDB.PasswordHashed, credentialsInRequest.Password)
	if err != nil {
		RespondWithError(w, http.StatusUnauthorized, UnauthorizedMessage)
		return
	}
	member := model.Member{UUID: credentialsInDB.UUID}
	if err := member.Get(); err != nil {
		RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}
	var permissions []string
	if member.Type == model.MemberTypeMember {
		permissions = append(permissions, model.MemberTypeMember)
	}
	if member.Type == model.MemberTypeAdmin {
		permissions = append(permissions, model.MemberTypeMember)
		permissions = append(permissions, model.MemberTypeAdmin)
	}
	token, err := createToken(member.UUID, permissions)
	if err != nil {
		RespondWithError(w, http.StatusUnprocessableEntity, err.Error())
		return
	}
	tokens := map[string]string{
		"access_token":  token.AccessToken,
		"refresh_token": token.RefreshToken,
	}
	RespondWithJSON(w, http.StatusOK, tokens)
}

func Logout(w http.ResponseWriter, r *http.Request) {
	au, err := ExtractToken(r)
	if err != nil {
		RespondWithError(w, http.StatusBadRequest, UnauthorizedMessage)
		return
	}
	deleted, delErr := deleteTokenInCache(au.RefreshUuid)
	if delErr != nil || deleted == 0 {
		RespondWithError(w, http.StatusBadRequest, UnauthorizedMessage)
		return
	}
	deleted, delErr = deleteTokenInCache(au.TokenUuid)
	if delErr != nil || deleted == 0 {
		RespondWithError(w, http.StatusBadRequest, UnauthorizedMessage)
		return
	}
	RespondWithJSON(w, http.StatusAccepted, "Successfully logged out")
}

func createToken(uuid string, permissions []string) (*TokenDetails, error) {
	td := &TokenDetails{}
	td.AtExpires = time.Now().Add(time.Minute * 15).Unix()
	td.AccessUuid = common.GenerateUUID()

	td.RtExpires = time.Now().Add(time.Hour * 24 * 7).Unix()
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

func verifyToken(r *http.Request) (*jwt.Token, error) {
	tokenString := extractTokenString(r)
	common.Debug("tokenString: " + tokenString)
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			common.Debug("Incorrect signing method")
			return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(common.GetConfigString("jwt.access_secret")), nil
	})
	if err != nil {
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
	token, err := verifyToken(r)
	if err != nil {
		common.Debug(err.Error())
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
	at := time.Unix(td.AtExpires, 0)
	rt := time.Unix(td.RtExpires, 0)
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

func Test(w http.ResponseWriter, r *http.Request) {
	token, err := createToken("123", []string{createCredentialsPermission})
	if err != nil {
		RespondWithError(w, http.StatusBadRequest, "Cannot create create_credentials token")
		return
	}
	RespondWithJSON(w, http.StatusOK, token.AccessToken)
}
