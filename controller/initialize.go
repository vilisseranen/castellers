package controller

import (
	"encoding/json"
	"net/http"

	"github.com/vilisseranen/castellers/common"
	"github.com/vilisseranen/castellers/model"
)

func Initialize(w http.ResponseWriter, r *http.Request) {

	// Only execute if it is the first member
	var m model.Member
	members, err := m.GetAll()
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
	m.UUID = common.GenerateUUID()
	m.Code = common.GenerateCode()

	if err := m.CreateMember(); err != nil {
		RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if common.GetConfigBool("debug") == false { // Don't send email in debug
		to := []string{m.Email}
		if err := common.SendMail("initialization", "Salut "+m.FirstName+". Ton code est: "+m.Code, to); err != nil {
			RespondWithError(w, http.StatusInternalServerError, err.Error())
			return
		}
	}
	RespondWithJSON(w, http.StatusCreated, m)
}

func IsInitialized(w http.ResponseWriter, r *http.Request) {
	var m model.Member
	members, err := m.GetAll()
	if err != nil {
		switch err {
		default:
			RespondWithError(w, http.StatusInternalServerError, err.Error())
		}
		return
	}
	if len(members) > 0 {
		RespondWithJSON(w, http.StatusOK, nil)
		return
	} else {
		RespondWithJSON(w, http.StatusNoContent, nil)
		return
	}
}
