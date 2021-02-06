package controller

import (
	"encoding/json"
	"net/http"

	"github.com/go-redis/redis"
	"github.com/vilisseranen/castellers/common"
)

const UnauthorizedMessage = "You are not authorized to perform this action."

var RedisClient *redis.Client

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
