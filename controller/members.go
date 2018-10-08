package controller

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/vilisseranen/castellers/common"
	"github.com/vilisseranen/castellers/model"
)

func GetMember(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	uuid := vars["member_uuid"]
	m := model.Member{UUID: uuid}
	if err := m.Get(); err != nil {
		switch err {
		case sql.ErrNoRows:
			RespondWithError(w, http.StatusNotFound, "Member not found")
		default:
			RespondWithError(w, http.StatusInternalServerError, err.Error())
		}
		return
	}
	RespondWithJSON(w, http.StatusOK, m)
}

func GetMembers(w http.ResponseWriter, r *http.Request) {
	m := model.Member{}
	members, err := m.GetAll()
	if err != nil {
		switch err {
		default:
			RespondWithError(w, http.StatusInternalServerError, err.Error())
		}
		return
	}
	RespondWithJSON(w, http.StatusOK, members)
}

func CreateMember(w http.ResponseWriter, r *http.Request) {
	// Decode info to create member
	var m model.Member
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&m); err != nil {
		RespondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}
	defer r.Body.Close()
	if missingRequiredFields(m) {
		RespondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}
	if err := model.ValidateRoles(m.Roles); err != nil {
		fmt.Println("Error validating roles: " + err.Error())
		RespondWithError(w, http.StatusBadRequest, err.Error())
		return
	}
	if err := model.ValidateLanguage(m.Language); err != nil {
		fmt.Println("Error validating language: " + err.Error())
		RespondWithError(w, http.StatusBadRequest, err.Error())
		return
	}
	m.UUID = common.GenerateUUID()
	m.Code = common.GenerateCode()
	// We will need admin info later for the email
	vars := mux.Vars(r)
	uuid := vars["admin_uuid"]
	a := model.Member{UUID: uuid}
	if err := a.Get(); err != nil {
		fmt.Println("Failed to get admin for CreateMember.")
		RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}
	// Create the Member now
	if err := m.CreateMember(); err != nil {
		RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}
	// Send the email
	if common.GetConfigBool("debug") == false { // Don't send email in debug
		loginLink := common.GetConfigString("domain") + "/#/login?" +
			"m=" + m.UUID +
			"&c=" + m.Code
		profileLink := loginLink + "&next=memberEdit/" + m.UUID
		if err := common.SendRegistrationEmail(m.Email, m.FirstName, a.FirstName, a.Extra, loginLink, profileLink); err != nil {
			m.DeleteMember()
			RespondWithError(w, http.StatusInternalServerError, err.Error())
			return
		}
	}
	RespondWithJSON(w, http.StatusCreated, m)
}

func EditMember(w http.ResponseWriter, r *http.Request) {
	var m model.Member
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&m); err != nil {
		RespondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}
	defer r.Body.Close()
	if missingRequiredFields(m) {
		RespondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}
	if err := model.ValidateRoles(m.Roles); err != nil {
		fmt.Println("Error validating roles: " + err.Error())
		RespondWithError(w, http.StatusBadRequest, err.Error())
		return
	}
	if err := model.ValidateLanguage(m.Language); err != nil {
		fmt.Println("Error validating language: " + err.Error())
		RespondWithError(w, http.StatusBadRequest, err.Error())
		return
	}
	if err := m.EditMember(); err != nil {
		RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}
	RespondWithJSON(w, http.StatusAccepted, m)
}

func DeleteMember(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	uuid := vars["member_uuid"]
	m := model.Member{UUID: uuid}
	if err := m.DeleteMember(); err != nil {
		RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}
	RespondWithJSON(w, http.StatusOK, nil)
}

func GetRoles(w http.ResponseWriter, r *http.Request) {
	roles := model.ValidRoleList
	RespondWithJSON(w, http.StatusOK, roles)
}

func SendRegistrationEmail(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	uuid := vars["member_uuid"]
	m := model.Member{UUID: uuid}
	if err := m.Get(); err != nil {
		switch err {
		case sql.ErrNoRows:
			RespondWithError(w, http.StatusNotFound, "Member not found")
		default:
			RespondWithError(w, http.StatusInternalServerError, err.Error())
		}
		return
	}
	vars = mux.Vars(r)
	uuid = vars["admin_uuid"]
	a := model.Member{UUID: uuid}
	if err := a.Get(); err != nil {
		fmt.Println("Failed to get admin.")
		RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if common.GetConfigBool("debug") == false { // Don't send email in debug
		loginLink := common.GetConfigString("domain") + "/#/login?" +
			"m=" + m.UUID +
			"&c=" + m.Code
		profileLink := loginLink + "&next=memberEdit/" + m.UUID
		if err := common.SendRegistrationEmail(m.Email, m.FirstName, a.FirstName, a.Extra, loginLink, profileLink); err != nil {
			RespondWithError(w, http.StatusInternalServerError, err.Error())
			return
		}
	}
	RespondWithJSON(w, http.StatusOK, nil)

}

func missingRequiredFields(m model.Member) bool {
	return (m.FirstName == "" || m.LastName == "" || m.Type == "" || m.Email == "" || m.Language == "")
}
