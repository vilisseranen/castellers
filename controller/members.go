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

const (
	ERRORGETMEMBER         = "Error getting member"
	ERRORGETMEMBERS        = "Error getting members"
	ERRORCREATEMEMBER      = "Error creating member"
	ERRORMEMBERNOTFOUND    = "Member not found"
	ERRORMEMBERHEIGHT      = "Error with provided height"
	ERRORMEMBERWEIGHT      = "Error with the provided weight"
	ERRORMEMBERROLES       = "Error with the provided roles"
	ERRORMEMBERLANGUAGE    = "Error with the provided language"
	ERRORUPDATEMEMBER      = "Error updating member"
	ERRORDELETEMEMBER      = "Error deleting member"
	ERRORREGISTRATIONEMAIL = "Error sending the registration email"
	ERRORRESETCREDENTIALS  = "Error resetting credentials"
	ERROREMAILUNAVAILABLE  = "This email is already used by another member."
)

func GetMember(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	UUID := vars["member_uuid"]

	// if member, request can only be about themselves
	// if admin can be for anyone

	tokenAuth, err := ExtractToken(r)
	if err != nil {
		common.Warn("Error reading token: %s", err.Error())
		RespondWithError(w, http.StatusInternalServerError, ERRORAUTHENTICATION)
		return
	}
	if common.StringInSlice(model.MEMBERSTYPEADMIN, tokenAuth.Permissions) || UUID == tokenAuth.UserId {
		m := model.Member{UUID: UUID}
		if err := m.Get(); err != nil {
			switch err {
			case sql.ErrNoRows:
				common.Debug("Member not found: %s", err.Error())
				RespondWithError(w, http.StatusNotFound, ERRORMEMBERNOTFOUND)
			default:
				common.Warn("Error getting member: %s", err.Error())
				RespondWithError(w, http.StatusInternalServerError, ERRORMEMBERNOTFOUND)
			}
			return
		}
		if !common.StringInSlice(model.MEMBERSTYPEADMIN, tokenAuth.Permissions) {
			m.Roles = []string{}
			m.Extra = ""
		}
		RespondWithJSON(w, http.StatusOK, m)
		return

	}
	common.Info("Permissions: %s", tokenAuth.Permissions)
	RespondWithError(w, http.StatusUnauthorized, ERRORUNAUTHORIZED)
	return
}

func GetMembers(w http.ResponseWriter, r *http.Request) {
	m := model.Member{}
	members, err := m.GetAll()
	if err != nil {
		common.Warn("Error getting members: %s", err.Error())
		RespondWithError(w, http.StatusInternalServerError, ERRORGETMEMBERS)
		return
	}
	RespondWithJSON(w, http.StatusOK, members)
}

func CreateMember(w http.ResponseWriter, r *http.Request) {
	// Decode info to create member
	var m model.Member
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&m); err != nil {
		common.Debug("Invalid request payload: %s", err.Error())
		RespondWithError(w, http.StatusBadRequest, ERRORINVALIDPAYLOAD)
		return
	}
	defer r.Body.Close()
	if !emailAvailable(m) {
		common.Info("Email not available: %s", m.Email)
		RespondWithError(w, http.StatusBadRequest, ERROREMAILUNAVAILABLE)
		return
	}
	if missingRequiredFields(m) {
		common.Info("Missing fields in request payload")
		RespondWithError(w, http.StatusBadRequest, ERRORMISSINGFIELDS)
		return
	}
	if err := model.ValidNumberOrEmpty(m.Height); err != nil {
		common.Info("Error validating Height: " + m.Height)
		RespondWithError(w, http.StatusBadRequest, ERRORMEMBERHEIGHT)
		return
	}
	if err := model.ValidNumberOrEmpty(m.Weight); err != nil {
		common.Info("Error validating Weight: " + m.Weight)
		RespondWithError(w, http.StatusBadRequest, ERRORMEMBERWEIGHT)
		return
	}
	if err := model.ValidateRoles(m.Roles); err != nil {
		common.Info("Error validating roles: " + err.Error())
		RespondWithError(w, http.StatusBadRequest, ERRORMEMBERROLES)
		return
	}
	if err := model.ValidateLanguage(m.Language); err != nil {
		common.Info("Error validating language: " + err.Error())
		RespondWithError(w, http.StatusBadRequest, ERRORMEMBERLANGUAGE)
		return
	}
	m.UUID = common.GenerateUUID()
	m.Code = common.GenerateCode()
	// We will need admin info later for the email
	tokenAuth, err := ExtractToken(r)
	if err != nil {
		common.Warn("Error reading token: %s", err.Error())
		RespondWithError(w, http.StatusInternalServerError, ERRORAUTHENTICATION)
		return
	}
	a := model.Member{UUID: tokenAuth.UserId}
	if err := a.Get(); err != nil {
		common.Warn("Failed to get admin for CreateMember: %s", err.Error())
		RespondWithError(w, http.StatusInternalServerError, ERRORCREATEMEMBER)
		return
	}
	// Create the Member now
	if err := m.CreateMember(); err != nil {
		common.Warn("Error creating member: %s", err.Error())
		RespondWithError(w, http.StatusInternalServerError, ERRORCREATEMEMBER)
		return
	}
	payload := mail.EmailRegisterPayload{Member: m, Author: a}
	payloadBytes := new(bytes.Buffer)
	json.NewEncoder(payloadBytes).Encode(payload)
	n := model.Notification{NotificationType: model.TypeMemberRegistration, ObjectUUID: m.UUID, SendDate: int(time.Now().Unix()), Payload: payloadBytes.Bytes()}
	if err := n.CreateNotification(); err != nil {
		common.Warn("Error creating notification: %s", err.Error())
		RespondWithError(w, http.StatusInternalServerError, ERRORNOTIFICATION)
		return
	}
	RespondWithJSON(w, http.StatusCreated, m)
}

func EditMember(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	UUID := vars["member_uuid"]

	// if member, request can only be about themselves
	// if admin can be for anyone

	tokenAuth, err := ExtractToken(r)
	if err != nil {
		common.Warn("Error reading token: %s", err.Error())
		RespondWithError(w, http.StatusInternalServerError, ERRORAUTHENTICATION)
		return
	}
	if common.StringInSlice(model.MEMBERSTYPEADMIN, tokenAuth.Permissions) || UUID == tokenAuth.UserId {
		var m model.Member
		decoder := json.NewDecoder(r.Body)
		if err := decoder.Decode(&m); err != nil {
			common.Debug("Invalid request payload: %s", err.Error())
			RespondWithError(w, http.StatusBadRequest, ERRORINVALIDPAYLOAD)
			return
		}
		defer r.Body.Close()
		// Make sure we are not changing the profile of somebody else
		m.UUID = UUID
		// If email is changing, we need to check if it is used
		currentMember := model.Member{UUID: m.UUID}
		err = currentMember.Get()
		if err != nil {
			common.Info("Member cannot be found: %s", err.Error())
			RespondWithError(w, http.StatusBadRequest, ERRORGETMEMBER)
			return
		}
		if currentMember.Email != m.Email && !emailAvailable(m) {
			common.Info("Email not available. Current: %s, requested: %s, emailAvailable: %s", currentMember.Email, m.Email, emailAvailable(m))
			RespondWithError(w, http.StatusBadRequest, ERROREMAILUNAVAILABLE)
			return
		}
		if missingRequiredFields(m) {
			common.Info("Missing fields in request payload")
			RespondWithError(w, http.StatusBadRequest, ERRORMISSINGFIELDS)
			return
		}
		if err := model.ValidNumberOrEmpty(m.Height); err != nil {
			common.Info("Error validating Height: " + m.Height)
			RespondWithError(w, http.StatusBadRequest, ERRORMEMBERHEIGHT)
			return
		}
		if err := model.ValidNumberOrEmpty(m.Weight); err != nil {
			common.Info("Error validating Weight: " + m.Weight)
			RespondWithError(w, http.StatusBadRequest, ERRORMEMBERWEIGHT)
			return
		}
		if err := model.ValidateRoles(m.Roles); err != nil {
			common.Info("Error validating roles: " + err.Error())
			RespondWithError(w, http.StatusBadRequest, err.Error())
			return
		}
		if err := model.ValidateLanguage(m.Language); err != nil {
			common.Info("Error validating language: " + err.Error())
			RespondWithError(w, http.StatusBadRequest, err.Error())
			return
		}

		// Check if we can change role
		// If caller is admin, we can change the role
		// If caller is member, we cannot change he role
		if !common.StringInSlice(model.MEMBERSTYPEADMIN, tokenAuth.Permissions) {
			// get current user and use existing values for roles, extra and type
			existingMember := model.Member{UUID: UUID}
			if err := existingMember.Get(); err != nil {
				switch err {
				case sql.ErrNoRows:
					common.Debug("Member not found: %s", err.Error())
					RespondWithError(w, http.StatusNotFound, ERRORMEMBERNOTFOUND)
				default:
					common.Warn("Error getting member: %s", err.Error())
					RespondWithError(w, http.StatusInternalServerError, ERRORUPDATEMEMBER)
				}
				return
			}
			if existingMember.Type != m.Type {
				common.Info("Member tries to change their type from %s to %s", existingMember.Type, m.Type)
				RespondWithError(w, http.StatusForbidden, ERRORUNAUTHORIZED)
				return
			}
			m.Roles = existingMember.Roles
			m.Extra = existingMember.Extra

		}
		if err := m.EditMember(); err != nil {
			common.Warn("Error updating member: %s", err.Error())
			RespondWithError(w, http.StatusInternalServerError, ERRORUPDATEMEMBER)
			return
		}
		RespondWithJSON(w, http.StatusAccepted, m)
		return
	}
	RespondWithError(w, http.StatusUnauthorized, ERRORUNAUTHORIZED)
}

func DeleteMember(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	UUID := vars["member_uuid"]
	m := model.Member{UUID: UUID}
	// Cannot delete self if admin
	if err := m.Get(); err != nil {
		switch err {
		case sql.ErrNoRows:
			common.Debug("Member not found: %s", err.Error())
			RespondWithError(w, http.StatusNotFound, ERRORMEMBERNOTFOUND)
		default:
			common.Warn("Error getting member: %s", err.Error())
			RespondWithError(w, http.StatusInternalServerError, ERRORDELETEMEMBER)
		}
		return
	}
	tokenAuth, err := ExtractToken(r)
	if err != nil {
		common.Warn("Error reading token: %s", err.Error())
		RespondWithError(w, http.StatusInternalServerError, ERRORAUTHENTICATION)
		return
	}
	// TODO: admin should be allowed to delete their profile
	if tokenAuth.UserId == UUID && m.Type == "admin" {
		common.Info("Cannot remove yourself if admin")
		RespondWithError(w, http.StatusLocked, ERRORUNAUTHORIZED)
		return
	}
	if err := m.DeleteMember(); err != nil {
		common.Warn("Error deleting member: %s", err.Error())
		RespondWithError(w, http.StatusInternalServerError, ERRORDELETEMEMBER)
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
			common.Debug("Member not found: %s", err.Error())
			RespondWithError(w, http.StatusNotFound, ERRORMEMBERNOTFOUND)
		default:
			common.Warn("Error getting member: %s", err.Error())
			RespondWithError(w, http.StatusInternalServerError, ERRORREGISTRATIONEMAIL)
		}
		return
	}
	vars = mux.Vars(r)
	UUID = vars["admin_uuid"]
	a := model.Member{UUID: UUID}
	if err := a.Get(); err != nil {
		common.Warn("Failed to get admin: %s", err.Error())
		RespondWithError(w, http.StatusInternalServerError, ERRORREGISTRATIONEMAIL)
		return
	}
	payload := mail.EmailRegisterPayload{Member: m, Author: a}
	payloadBytes := new(bytes.Buffer)
	json.NewEncoder(payloadBytes).Encode(payload)
	n := model.Notification{NotificationType: model.TypeMemberRegistration, ObjectUUID: m.UUID, SendDate: int(time.Now().Unix()), Payload: payloadBytes.Bytes()}
	if err := n.CreateNotification(); err != nil {
		common.Warn("Error creating notification: %s", err.Error())
		RespondWithError(w, http.StatusInternalServerError, ERRORNOTIFICATION)
		return
	}
	RespondWithJSON(w, http.StatusOK, nil)
}

func missingRequiredFields(m model.Member) bool {
	return (m.FirstName == "" || m.LastName == "" || m.Type == "" || m.Email == "" || m.Language == "")
}

func emailAvailable(m model.Member) bool {
	err := m.GetByEmail()
	if err != nil && err.Error() == model.MEMBERSEMAILNOTFOUNDMESSAGE {
		common.Debug("Error getting by email: %s", err.Error())
		return true
	}
	common.Debug("Email %s is available", m.Email)
	return false
}

func ResetCredentials(w http.ResponseWriter, r *http.Request) {
	tokenAuth, err := ExtractToken(r)
	if err != nil {
		common.Warn("Error reading token: %s", err.Error())
		RespondWithError(w, http.StatusInternalServerError, ERRORAUTHENTICATION)
		return
	}
	// This function always use the UUID from the token
	// so we cannot change the password for somebody else
	c := model.Credentials{UUID: tokenAuth.UserId}
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&c); err != nil {
		common.Debug("Invalid request payload: %s", err.Error())
		RespondWithError(w, http.StatusBadRequest, ERRORINVALIDPAYLOAD)
		return
	}
	password, err := common.GenerateFromPassword(c.Password)
	if err != nil {
		common.Error("Error generating hashed password: %s", err.Error())
		RespondWithError(w, http.StatusInternalServerError, ERRORRESETCREDENTIALS)
		return
	}
	// if username is not provided, fetch it in DB
	if c.Username == "" {
		err := c.GetCredentialsByUUID()
		if err != nil {
			common.Debug("Invalid request payload: %s", err.Error())
			RespondWithError(w, http.StatusBadRequest, ERRORINVALIDPAYLOAD)
			return
		}
	}
	err = c.ResetCredentials(c.Username, password)
	if err != nil {
		common.Debug("Invalid request payload: %s", err.Error())
		RespondWithError(w, http.StatusBadRequest, ERRORINVALIDPAYLOAD)
		return
	}
	// resetCredentialsToken should only be used once
	if common.StringInSlice(ResetCredentialsPermission, tokenAuth.Permissions) {
		_, err = deleteTokenInCache(tokenAuth.TokenUuid)
		if err != nil {
			common.Warn("Error deleting token in cache: %s", err.Error())
			RespondWithError(w, http.StatusInternalServerError, ERRORINTERNAL)
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

	if currentUser.Type == model.MEMBERSTYPEREGULAR && m.Type == model.MEMBERSTYPEADMIN && adminUuid == "" {
		return false
	}
	return true
}
