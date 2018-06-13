package controller

import (
	"encoding/json"
	"net/http"

	"github.com/vilisseranen/castellers/model"
)

func Initialize(w http.ResponseWriter, r *http.Request) {

	// Only execute if it is the first member
	var m model.Member
	members, err := m.GetAll(0, 1)
	if err != nil {
		switch err {
		default:
			RespondWithError(w, http.StatusInternalServerError, err.Error())
		}
		return
	}
	if len(members) > 0 {
		RespondWithError(w, http.StatusUnauthorized, "The app is already initialized")
		return
	}

	// Create the first admin
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&m); err != nil {
		RespondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}
	m.Type = model.MEMBER_TYPE_ADMIN // Make sure it's an admin
	defer r.Body.Close()
	if err := m.CreateMember(); err != nil {
		RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}
	RespondWithJSON(w, http.StatusCreated, m)
}
