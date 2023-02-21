package controller

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"net/http"

	"github.com/redis/go-redis/extra/redisotel/v9"
	"github.com/redis/go-redis/v9"
	"go.opentelemetry.io/otel"

	"github.com/vilisseranen/castellers/common"
)

const (
	ERRORINVALIDPAYLOAD = "Invalid request payload"
	ERRORINTERNAL       = "Internal error"
	ERRORNOTIFICATION   = "Error creating notification"
	ERRORAUTHENTICATION = "Error with the provided token"
	ERRORMISSINGFIELDS  = "Missing fields in request payload"
	ERRORUNAUTHORIZED   = "You are not authorized to perform this action."
)

var RedisClient *redis.Client
var version string
var tracer = otel.Tracer("castellers")

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
	_, err := RedisClient.Ping(context.Background()).Result()
	if err != nil {
		common.Fatal(err.Error())
	}
	if err := redisotel.InstrumentTracing(RedisClient); err != nil {
		panic(err)
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
