package controller

import (
	"database/sql"
	"encoding/json"
	"github.com/vilisseranen/castellers/model"
	"net/http"
)

func respondWithError(w http.ResponseWriter, code int, message string) {
	respondWithJSON(w, code, map[string]string{"error": message})
}

func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	response, _ := json.Marshal(payload)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(response)
}

func isAdmin(uuid string) (bool, error) {
	member := model.Member{UUID: uuid}
	if err := member.Get(); err != nil {
		switch err {
		case sql.ErrNoRows:
			return false, err
		default:
			return false, err
		}
		if member.Type != model.MEMBER_TYPE_ADMIN {
			return false, nil
		}
	}
	return true, nil
}

func isMember(uuid string) (bool, error) {
	member := model.Member{UUID: uuid}
	if err := member.Get(); err != nil {
		switch err {
		case sql.ErrNoRows:
			return false, err
		default:
			return false, err
		}
	}
	return true, nil
}
