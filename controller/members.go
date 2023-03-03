package controller

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gorilla/mux"

	"github.com/vilisseranen/castellers/common"
	"github.com/vilisseranen/castellers/mail"
	"github.com/vilisseranen/castellers/model"
)

const (
	ERRORGETMEMBER              = "error getting member"
	ERRORGETMEMBERS             = "error getting members"
	ERRORCREATEMEMBER           = "error creating member"
	ERRORMEMBERNOTFOUND         = "member not found"
	ERRORMEMBERHEIGHT           = "error with provided height"
	ERRORMEMBERWEIGHT           = "error with the provided weight"
	ERRORMEMBERROLES            = "error with the roles provided"
	ERRORMEMBERLANGUAGE         = "error with the language provided"
	ERRORMEMBERTYPE             = "error with the type provided"
	ERRORUPDATEMEMBER           = "error updating member"
	ERRORDELETEMEMBER           = "error deleting member"
	ERRORREGISTRATIONEMAIL      = "error sending the registration email"
	ERRORRESETCREDENTIALS       = "error resetting credentials"
	ERROREMAILUNAVAILABLE       = "this email is already used by another member."
	ERRORGUESTREGISTRATIONEMAIL = "guests cannot receive the registration email."
	ERRORUPDATEMEMBERTYPE       = "error changing the type of the member"
	ERRORACTIVATINGMEMBER       = "error setting the member as active"
	ERRORCHANGINGMEMBERSTATUS   = "error changing the status of the member"
)

func GetMember(w http.ResponseWriter, r *http.Request) {
	ctx, span := tracer.Start(r.Context(), "GetMember")
	defer span.End()

	vars := mux.Vars(r)
	UUID := vars["member_uuid"]

	// if member, request can only be about themselves
	// if admin can be for anyone

	tokenAuth, err := ExtractToken(r.Context(), r)
	if err != nil {
		common.Warn("Error reading token: %s", err.Error())
		RespondWithError(w, http.StatusInternalServerError, ERRORAUTHENTICATION)
		return
	}
	if common.StringInSlice(model.MEMBERSTYPEADMIN, tokenAuth.Permissions) || UUID == tokenAuth.UserId {
		m := model.Member{UUID: UUID}
		if err := m.Get(ctx); err != nil {
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
}

func GetMembers(w http.ResponseWriter, r *http.Request) {
	ctx, span := tracer.Start(r.Context(), "GetMembers")
	defer span.End()

	memberStatusList := memberStatusListFromQuery(r.FormValue("status"))
	memberTypeList := memberTypeListFromQuery(r.FormValue("type"))

	m := model.Member{}
	members, err := m.GetAll(ctx, memberStatusList, memberTypeList)
	if err != nil {
		common.Warn("Error getting members: %s", err.Error())
		RespondWithError(w, http.StatusInternalServerError, ERRORGETMEMBERS)
		return
	}
	RespondWithJSON(w, http.StatusOK, members)
}

func CreateMember(w http.ResponseWriter, r *http.Request) {
	ctx, span := tracer.Start(r.Context(), "CreateMember")
	defer span.End()
	// Decode info to create member
	var m model.Member
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&m); err != nil {
		common.Debug("Invalid request payload: %s", err.Error())
		RespondWithError(w, http.StatusBadRequest, ERRORINVALIDPAYLOAD)
		return
	}
	defer r.Body.Close()
	if err := model.ValidateType(m.Type); err != nil {
		common.Info("Error validating language: " + err.Error())
		RespondWithError(w, http.StatusBadRequest, ERRORMEMBERTYPE)
		return
	}
	if m.Type != model.MEMBERSTYPEGUEST && m.Type != model.MEMBERSTYPECANALLA {
		if !emailAvailable(ctx, m) {
			common.Info("Email not available: %s", m.Email)
			RespondWithError(w, http.StatusBadRequest, ERROREMAILUNAVAILABLE)
			return
		}
	} else {
		m.Email = ""
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
	// We will need admin info later for the email
	tokenAuth, err := ExtractToken(r.Context(), r)
	if err != nil {
		common.Warn("Error reading token: %s", err.Error())
		RespondWithError(w, http.StatusInternalServerError, ERRORAUTHENTICATION)
		return
	}
	a := model.Member{UUID: tokenAuth.UserId}
	if err := a.Get(ctx); err != nil {
		common.Warn("Failed to get admin for CreateMember: %s", err.Error())
		RespondWithError(w, http.StatusInternalServerError, ERRORCREATEMEMBER)
		return
	}
	// Create the Member now
	if err := m.CreateMember(ctx); err != nil {
		common.Warn("Error creating member: %s", err.Error())
		RespondWithError(w, http.StatusInternalServerError, ERRORCREATEMEMBER)
		return
	}
	// When a guest is converted to a regular, we need to set the status to created
	if m.Type == model.MEMBERSTYPEGUEST || m.Type == model.MEMBERSTYPECANALLA {
		err := m.SetStatus(ctx, model.MEMBERSSTATUSACTIVATED)
		if err != nil {
			common.Error(fmt.Sprintf("Error changing member status to %s", model.MEMBERSSTATUSCREATED))
			RespondWithError(w, http.StatusInternalServerError, ERRORCHANGINGMEMBERSTATUS)
			return
		}
	} else {
		payload := mail.EmailRegisterPayload{Member: m, Author: a}
		payloadBytes := new(bytes.Buffer)
		json.NewEncoder(payloadBytes).Encode(payload)
		n := model.Notification{NotificationType: model.TypeMemberRegistration, ObjectUUID: m.UUID, SendDate: int(time.Now().Unix()), Payload: payloadBytes.Bytes()}
		if err := n.CreateNotification(ctx); err != nil {
			common.Warn("Error creating notification: %s", err.Error())
			RespondWithError(w, http.StatusInternalServerError, ERRORNOTIFICATION)
			return
		}
	}
	RespondWithJSON(w, http.StatusCreated, m)
}

func EditMember(w http.ResponseWriter, r *http.Request) {
	ctx, span := tracer.Start(r.Context(), "EditMember")
	defer span.End()

	vars := mux.Vars(r)
	UUID := vars["member_uuid"]

	// if member, request can only be about themselves
	// if admin can be for anyone

	tokenAuth, err := ExtractToken(ctx, r)
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
		err = currentMember.Get(ctx)
		if err != nil {
			common.Info("Member cannot be found: %s", err.Error())
			RespondWithError(w, http.StatusBadRequest, ERRORGETMEMBER)
			return
		}
		// A regular cannot be converted to a guest
		if currentMember.Type != model.MEMBERSTYPEGUEST && m.Type == model.MEMBERSTYPEGUEST {
			common.Info("Cannot change a regular member into a guest. Current: %s, requested: %s", currentMember.Email, m.Email)
			RespondWithError(w, http.StatusBadRequest, ERRORUPDATEMEMBERTYPE)
			return
		}
		if currentMember.Email != m.Email && !emailAvailable(ctx, m) {
			common.Info("Email not available. Current: %s, requested: %s, emailAvailable: %s", currentMember.Email, m.Email, emailAvailable(ctx, m))
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
		// If caller is member, we cannot change the role
		if !common.StringInSlice(model.MEMBERSTYPEADMIN, tokenAuth.Permissions) {
			// get current user and use existing values for roles, extra and type
			existingMember := model.Member{UUID: UUID}
			if err := existingMember.Get(ctx); err != nil {
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
		if err := m.EditMember(ctx); err != nil {
			common.Warn("Error updating member: %s", err.Error())
			RespondWithError(w, http.StatusInternalServerError, ERRORUPDATEMEMBER)
			return
		}
		// When a guest is converted to a regular, we need to set the status to created
		// Does not apply to canalla, they will stay activated and won't receive the welcome email
		if currentMember.Type == model.MEMBERSTYPEGUEST && m.Type != model.MEMBERSTYPEGUEST && m.Type != model.MEMBERSTYPECANALLA {
			err := m.SetStatus(ctx, model.MEMBERSSTATUSCREATED)
			if err != nil {
				common.Error(fmt.Sprintf("Error changing member status to %s", model.MEMBERSSTATUSCREATED))
				RespondWithError(w, http.StatusInternalServerError, ERRORCHANGINGMEMBERSTATUS)
				return
			}
		}
		RespondWithJSON(w, http.StatusAccepted, m)
		return
	}
	RespondWithError(w, http.StatusUnauthorized, ERRORUNAUTHORIZED)
}

func DeleteMember(w http.ResponseWriter, r *http.Request) {
	ctx, span := tracer.Start(r.Context(), "DeleteMember")
	defer span.End()

	vars := mux.Vars(r)
	UUID := vars["member_uuid"]
	m := model.Member{UUID: UUID}
	// Cannot delete self if admin
	if err := m.Get(ctx); err != nil {
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
	tokenAuth, err := ExtractToken(r.Context(), r)
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
	if err := m.DeleteMember(ctx); err != nil {
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
	ctx, span := tracer.Start(r.Context(), "SendRegistrationEmail")
	defer span.End()
	vars := mux.Vars(r)
	UUID := vars["member_uuid"]
	m := model.Member{UUID: UUID}
	if err := m.Get(ctx); err != nil {
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
	if m.Type == model.MEMBERSTYPEGUEST || m.Type == model.MEMBERSTYPECANALLA {
		common.Warn("Cannot send a registration email to a %s (%s)", m.Type, m.UUID)
		RespondWithError(w, http.StatusForbidden, ERRORGUESTREGISTRATIONEMAIL)
		return
	}
	// We will need admin info later for the email
	tokenAuth, err := ExtractToken(r.Context(), r)
	if err != nil {
		common.Warn("Error reading token: %s", err.Error())
		RespondWithError(w, http.StatusInternalServerError, ERRORAUTHENTICATION)
		return
	}
	a := model.Member{UUID: tokenAuth.UserId}
	if err := a.Get(ctx); err != nil {
		common.Warn("Failed to get admin: %s", err.Error())
		RespondWithError(w, http.StatusInternalServerError, ERRORREGISTRATIONEMAIL)
		return
	}
	payload := mail.EmailRegisterPayload{Member: m, Author: a}
	payloadBytes := new(bytes.Buffer)
	json.NewEncoder(payloadBytes).Encode(payload)
	n := model.Notification{NotificationType: model.TypeMemberRegistration, ObjectUUID: m.UUID, SendDate: int(time.Now().Unix()), Payload: payloadBytes.Bytes()}
	if err := n.CreateNotification(ctx); err != nil {
		common.Warn("Error creating notification: %s", err.Error())
		RespondWithError(w, http.StatusInternalServerError, ERRORNOTIFICATION)
		return
	}
	RespondWithJSON(w, http.StatusOK, nil)
}

func missingRequiredFields(m model.Member) bool {
	missingFields := false
	if m.Type == model.MEMBERSTYPEGUEST || m.Type == model.MEMBERSTYPECANALLA { // Guests don't have an email
		missingFields = (m.FirstName == "" || m.LastName == "" || m.Type == "" || m.Language == "")
	} else {
		missingFields = (m.FirstName == "" || m.LastName == "" || m.Type == "" || m.Email == "" || m.Language == "")
	}
	return missingFields
}

func emailAvailable(ctx context.Context, m model.Member) bool {
	ctx, span := tracer.Start(ctx, "emailAvailable")
	defer span.End()
	err := m.GetByEmail(ctx)
	if err != nil && err.Error() == model.MEMBERSEMAILNOTFOUNDMESSAGE {
		common.Debug("Error getting by email: %s", err.Error())
		return true
	}
	common.Debug("Email %s is available", m.Email)
	return false
}

func ResetCredentials(w http.ResponseWriter, r *http.Request) {
	ctx, span := tracer.Start(r.Context(), "ResetCredentials")
	defer span.End()
	tokenAuth, err := ExtractToken(r.Context(), r)
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
		err := c.GetCredentialsByUUID(ctx)
		if err != nil {
			common.Debug("Invalid request payload: %s", err.Error())
			RespondWithError(w, http.StatusBadRequest, ERRORINVALIDPAYLOAD)
			return
		}
	}
	err = c.ResetCredentials(ctx, c.Username, password)
	if err != nil {
		common.Debug("Invalid request payload: %s", err.Error())
		RespondWithError(w, http.StatusBadRequest, ERRORINVALIDPAYLOAD)
		return
	}
	// resetCredentialsToken should only be used once
	if common.StringInSlice(ResetCredentialsPermission, tokenAuth.Permissions) {
		_, err = deleteTokenInCache(r.Context(), tokenAuth.TokenUuid)
		if err != nil {
			common.Warn("Error deleting token in cache: %s", err.Error())
			RespondWithError(w, http.StatusInternalServerError, ERRORINTERNAL)
			return
		}
	}
	RespondWithJSON(w, http.StatusOK, "")
}

func memberStatusListFromQuery(queryParam string) []string {
	memberStatusList := []string{}
	for _, status := range strings.Split(queryParam, ",") {
		if status != "" && common.StringInSlice(status, []string{
			model.MEMBERSSTATUSACTIVATED,
			model.MEMBERSSTATUSCREATED,
			model.MEMBERSSTATUSDELETED,
			model.MEMBERSSTATUSPAUSED,
			model.MEMBERSSTATUSPURGED}) {
			memberStatusList = append(memberStatusList, status)
		}
	}
	return memberStatusList
}

func memberTypeListFromQuery(queryParam string) []string {
	memberTypeList := []string{}
	for _, mType := range strings.Split(queryParam, ",") {
		if mType != "" && common.StringInSlice(mType, []string{
			model.MEMBERSTYPEADMIN,
			model.MEMBERSTYPEGUEST,
			model.MEMBERSTYPEREGULAR,
			model.MEMBERSTYPECANALLA}) {
			memberTypeList = append(memberTypeList, mType)
		}
	}
	return memberTypeList
}
