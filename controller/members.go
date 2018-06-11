package controller

import (
	"database/sql"
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/vilisseranen/castellers/model"
)

func GetMember(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	uuid := vars["uuid"]
	m := model.Member{UUID: uuid}
	if err := m.Get(); err != nil {
		switch err {
		case sql.ErrNoRows:
			respondWithError(w, http.StatusNotFound, "Member not found")
		default:
			respondWithError(w, http.StatusInternalServerError, err.Error())
		}
		return
	}
	respondWithJSON(w, http.StatusOK, m)
}

func CreateMember(w http.ResponseWriter, r *http.Request) {
	// Check if admin exists
	vars := mux.Vars(r)
	uuid := vars["uuid"]
	admin, err := isAdmin(uuid)
	if err == sql.ErrNoRows || admin == false {
		respondWithError(w, http.StatusUnauthorized, "This admin is not authorized to create members.")
		return
	} else if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}
	// Create the event
	var m model.Member
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&m); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}
	defer r.Body.Close()
	if err := m.CreateMember(); err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondWithJSON(w, http.StatusCreated, m)
}
