package controller

import (
	"encoding/json"
	"io/ioutil"
	"net/http"

	"github.com/go-redis/redis"
	"github.com/vilisseranen/castellers/common"
)

const UnauthorizedMessage = "You are not authorized to perform this action."
const EmailUnavailableMessage = "This email is already used by another member."
const ErrorGetMemberMessage = "Error while getting member."

var RedisClient *redis.Client
var version string

func RespondWithError(w http.ResponseWriter, code int, message string) {
	RespondWithJSON(w, code, map[string]string{"error": message})
}

func RespondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	response, _ := json.Marshal(payload)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(response)
}

func InitializeRedis() {
	//Initializing redis
	dsn := common.GetConfigString("redis_dsn")
	RedisClient = redis.NewClient(&redis.Options{
		Addr: dsn, //redis port
	})
	_, err := RedisClient.Ping().Result()
	if err != nil {
		common.Fatal(err.Error())
	}
}

func Version(w http.ResponseWriter, r *http.Request) {
	b, err := ioutil.ReadFile("VERSION")
	if err != nil {
		common.Fatal(err.Error())
	}

	type version struct {
		Version string `json:"version"`
	}

	v := version{Version: string(b)}

	RespondWithJSON(w, http.StatusOK, v)
}
