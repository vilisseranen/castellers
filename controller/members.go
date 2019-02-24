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
		RespondWithError(w, http.StatusBadRequest, "Invalid request payload: missing required fields")
		return
	}
	if err := model.ValidNumberOrEmpty(m.Height); err != nil {
		fmt.Println("Error validating Height: " + m.Height)
		RespondWithError(w, http.StatusBadRequest, "Error validating Height: "+err.Error())
		return
	}
	if err := model.ValidNumberOrEmpty(m.Weight); err != nil {
		fmt.Println("Error validating Weight: " + m.Weight)
		RespondWithError(w, http.StatusBadRequest, "Error validating Weight: "+err.Error())
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
		if err := common.SendRegistrationEmail(m.Email, m.FirstName, m.Language, a.FirstName, a.Extra, loginLink, profileLink); err != nil {
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
		RespondWithError(w, http.StatusBadRequest, "Invalid request payload: missing required fields")
		return
	}
	if err := model.ValidNumberOrEmpty(m.Height); err != nil {
		fmt.Println("Error validating Height: " + m.Height)
		RespondWithError(w, http.StatusBadRequest, "Error validating Height: "+err.Error())
		return
	}
	if err := model.ValidNumberOrEmpty(m.Weight); err != nil {
		fmt.Println("Error validating Weight: " + m.Weight)
		RespondWithError(w, http.StatusBadRequest, "Error validating Weight: "+err.Error())
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
	code := r.Header.Get("X-Member-Code")
	if !validateChangeRole(m, code) {
		RespondWithError(w, http.StatusForbidden, "Cannot change role.")
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
	admin_uuid := vars["admin_uuid"]
	m := model.Member{UUID: uuid}
	// Cannot delete self if admin
	if err := m.Get(); err != nil {
		switch err {
		case sql.ErrNoRows:
			RespondWithError(w, http.StatusNotFound, "Member not found")
		default:
			RespondWithError(w, http.StatusInternalServerError, err.Error())
		}
		return
	}
	if admin_uuid == uuid && m.Type == "admin" {
		RespondWithError(w, http.StatusLocked, "Cannot remove yourself if admin")
		return
	}
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
		if err := common.SendRegistrationEmail(m.Email, m.FirstName, m.Language, a.FirstName, a.Extra, loginLink, profileLink); err != nil {
			RespondWithError(w, http.StatusInternalServerError, err.Error())
			return
		}
	}
	RespondWithJSON(w, http.StatusOK, nil)

}

func missingRequiredFields(m model.Member) bool {
	return (m.FirstName == "" || m.LastName == "" || m.Type == "" || m.Email == "" || m.Language == "")
}

// Returns true if it's valid, false otherwise
func validateChangeRole(m model.Member, code string) bool {
	// There are 2 cases when we cannot allow to change a role:
	// - a regular user wants to promote itself
	// - the last admin wants to demote itself
	current_user := model.Member{UUID: m.UUID}
	if err := current_user.Get(); err != nil {
		return false
	}
	// There are only problems when changes are made on ourselves
	if current_user.Code != code {
		return true
	}
	if current_user.Type == "member" {
		if m.Type == "admin" {
			return false
		}
	} else {
		all_users, err := m.GetAll()
		if err != nil {
			return false
		}
		var count_admins = 0
		for _, user := range all_users {
			if user.Type == "admin" {
				count_admins += 1
			}
		}
		if count_admins > 0 {
			return true
		}
	}
	return false
}
