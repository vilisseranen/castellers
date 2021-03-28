package controller

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"net/http"
	"time"

	"github.com/gorilla/mux"

	"github.com/vilisseranen/castellers/common"
	"github.com/vilisseranen/castellers/mail"
	"github.com/vilisseranen/castellers/model"
)

func GetMember(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	UUID := vars["member_uuid"]
	m := model.Member{UUID: UUID}
	if err := m.Get(); err != nil {
		switch err {
		case sql.ErrNoRows:
			RespondWithError(w, http.StatusNotFound, "Member not found")
		default:
			RespondWithError(w, http.StatusInternalServerError, err.Error())
		}
		return
	}
	au, err := ExtractToken(r)
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, UnauthorizedMessage)
		return
	}
	if !common.StringInSlice(model.MemberTypeAdmin, au.Permissions) {
		m.Roles = []string{}
		m.Extra = ""

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
	if !emailAvailable(m) {
		RespondWithError(w, http.StatusBadRequest, EmailUnavailableMessage)
		return
	}
	if missingRequiredFields(m) {
		RespondWithError(w, http.StatusBadRequest, "Invalid request payload: missing required fields")
		return
	}
	if err := model.ValidNumberOrEmpty(m.Height); err != nil {
		common.Info("Error validating Height: " + m.Height)
		RespondWithError(w, http.StatusBadRequest, "Error validating Height: "+err.Error())
		return
	}
	if err := model.ValidNumberOrEmpty(m.Weight); err != nil {
		common.Info("Error validating Weight: " + m.Weight)
		RespondWithError(w, http.StatusBadRequest, "Error validating Weight: "+err.Error())
		return
	}
	if err := model.ValidateRoles(m.Roles); err != nil {
		common.Error("Error validating roles: " + err.Error())
		RespondWithError(w, http.StatusBadRequest, err.Error())
		return
	}
	if err := model.ValidateLanguage(m.Language); err != nil {
		common.Error("Error validating language: " + err.Error())
		RespondWithError(w, http.StatusBadRequest, err.Error())
		return
	}
	m.UUID = common.GenerateUUID()
	m.Code = common.GenerateCode()
	// We will need admin info later for the email
	vars := mux.Vars(r)
	UUID := vars["admin_uuid"]
	a := model.Member{UUID: UUID}
	if err := a.Get(); err != nil {
		common.Error("Failed to get admin for CreateMember.")
		RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}
	// Create the Member now
	if err := m.CreateMember(); err != nil {
		RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}
	payload := mail.EmailRegisterPayload{Member: m, Author: a}
	payloadBytes := new(bytes.Buffer)
	json.NewEncoder(payloadBytes).Encode(payload)
	n := model.Notification{NotificationType: model.TypeMemberRegistration, AuthorUUID: m.UUID, ObjectUUID: m.UUID, SendDate: int(time.Now().Unix()), Payload: payloadBytes.Bytes()}
	if err := n.CreateNotification(); err != nil {
		RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
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
	// If email is changing, we need to check if it is used
	currentMember := model.Member{UUID: m.UUID}
	err := currentMember.Get()
	if err != nil {
		common.Info("Member cannt be found")
		RespondWithError(w, http.StatusBadRequest, ErrorGetMemberMessage)
		return
	}
	if currentMember.Email != m.Email && !emailAvailable(m) {
		common.Debug(currentMember.Email)
		common.Debug(m.Email)
		common.Debug("%s", emailAvailable(m))
		common.Info("Email %s is not available", m.Email)
		RespondWithError(w, http.StatusBadRequest, EmailUnavailableMessage)
		return
	}
	if missingRequiredFields(m) {
		RespondWithError(w, http.StatusBadRequest, "Invalid request payload: missing required fields")
		return
	}
	if err := model.ValidNumberOrEmpty(m.Height); err != nil {
		common.Error("Error validating Height: " + m.Height)
		RespondWithError(w, http.StatusBadRequest, "Error validating Height: "+err.Error())
		return
	}
	if err := model.ValidNumberOrEmpty(m.Weight); err != nil {
		common.Error("Error validating Weight: " + m.Weight)
		RespondWithError(w, http.StatusBadRequest, "Error validating Weight: "+err.Error())
		return
	}
	if err := model.ValidateRoles(m.Roles); err != nil {
		common.Error("Error validating roles: " + err.Error())
		RespondWithError(w, http.StatusBadRequest, err.Error())
		return
	}
	if err := model.ValidateLanguage(m.Language); err != nil {
		common.Error("Error validating language: " + err.Error())
		RespondWithError(w, http.StatusBadRequest, err.Error())
		return
	}
	// Make sure we are not changing the profile of somebody else
	vars := mux.Vars(r)
	UUID := vars["member_uuid"]
	m.UUID = UUID

	// Check if we can change role
	// If caller is admin, we can change the role
	// If caller is member, we cannot change he role
	au, err := ExtractToken(r)
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, UnauthorizedMessage)
	}
	if !common.StringInSlice(model.MemberTypeAdmin, au.Permissions) {
		// get current user and use existing values for roles, extra and type
		existingMember := model.Member{UUID: UUID}
		if err := existingMember.Get(); err != nil {
			switch err {
			case sql.ErrNoRows:
				RespondWithError(w, http.StatusNotFound, "Member not found")
			default:
				RespondWithError(w, http.StatusInternalServerError, err.Error())
			}
			return
		}
		if existingMember.Type != m.Type {
			RespondWithError(w, http.StatusForbidden, UnauthorizedMessage)
			return
		}
		m.Roles = existingMember.Roles
		m.Extra = existingMember.Extra

	}
	if err := m.EditMember(); err != nil {
		RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}
	RespondWithJSON(w, http.StatusAccepted, m)
}

func DeleteMember(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	UUID := vars["member_uuid"]
	adminUUID := vars["admin_uuid"]
	m := model.Member{UUID: UUID}
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
	if adminUUID == UUID && m.Type == "admin" {
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
	UUID := vars["member_uuid"]
	m := model.Member{UUID: UUID}
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
	UUID = vars["admin_uuid"]
	a := model.Member{UUID: UUID}
	if err := a.Get(); err != nil {
		common.Error("Failed to get admin.")
		RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}
	payload := mail.EmailRegisterPayload{Member: m, Author: a}
	payloadBytes := new(bytes.Buffer)
	json.NewEncoder(payloadBytes).Encode(payload)
	n := model.Notification{NotificationType: model.TypeMemberRegistration, AuthorUUID: a.UUID, ObjectUUID: m.UUID, SendDate: int(time.Now().Unix()), Payload: payloadBytes.Bytes()}
	if err := n.CreateNotification(); err != nil {
		RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}
	RespondWithJSON(w, http.StatusOK, nil)
}

func missingRequiredFields(m model.Member) bool {
	return (m.FirstName == "" || m.LastName == "" || m.Type == "" || m.Email == "" || m.Language == "")
}

func emailAvailable(m model.Member) bool {
	err := m.GetByEmail()
	if err != nil && err.Error() == model.MemberEmailNotFoundMessage {
		common.Debug("Error getting by email: %s", err.Error())
		return true
	}
	common.Debug("Email %s is available", m.Email)
	return false
}

func ResetCredentials(w http.ResponseWriter, r *http.Request) {
	tokenAuth, err := ExtractToken(r)
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}
	c := model.Credentials{UUID: tokenAuth.UserId}
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&c); err != nil {
		RespondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}
	password, err := common.GenerateFromPassword(c.Password)
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}
	// if username is not provided, fetch it in DB
	if c.Username == "" {
		err := c.GetCredentialsByUUID()
		if err != nil {
			RespondWithError(w, http.StatusBadRequest, "Invalid request payload")
			return
		}
	}
	err = c.ResetCredentials(c.Username, password)
	if err != nil {
		RespondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}
	// resetCredentialsToken should only be used once
	if common.StringInSlice(ResetCredentialsPermission, tokenAuth.Permissions) {
		_, err = deleteTokenInCache(tokenAuth.TokenUuid)
		if err != nil {
			RespondWithError(w, http.StatusInternalServerError, err.Error())
			return
		}
	}
	RespondWithJSON(w, http.StatusOK, "")
}

// Returns true if it's valid, false otherwise
func validateChangeType(m model.Member, code string, adminUuid string) bool {
	// Make sure a user does not promote him or herself
	currentUser := model.Member{UUID: m.UUID}
	// If member does not exist can't do any action.
	if err := currentUser.Get(); err != nil {
		return false
	}

	if currentUser.Type == model.MemberTypeMember && m.Type == model.MemberTypeAdmin && adminUuid == "" {
		return false
	}
	return true
}
