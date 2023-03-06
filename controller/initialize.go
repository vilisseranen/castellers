package controller

import (
	"bytes"
	"encoding/json"
	"net/http"
	"time"

	"github.com/vilisseranen/castellers/common"
	"github.com/vilisseranen/castellers/mail"
	"github.com/vilisseranen/castellers/model"
)

const (
	ERRORAPPALREADYINITIALIZED = "the app is already initialized"
)

func Initialize(w http.ResponseWriter, r *http.Request) {
	ctx, span := tracer.Start(r.Context(), "Initialize")
	defer span.End()

	// Only execute if it is the first member
	var m model.Member
	members, err := m.GetAll(ctx, []string{}, []string{})
	if err != nil {
		common.Warn("Error getting all members: %s", err.Error())
		RespondWithError(w, http.StatusInternalServerError, ERRORGETMEMBER)
		return
	}
	if len(members) > 0 {
		common.Debug("There is already a member: cannot initialize")
		RespondWithError(w, http.StatusUnauthorized, ERRORAPPALREADYINITIALIZED)
		return
	}

	// Create the first admin
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&m); err != nil {
		common.Debug("Error decoding member: %s", err.Error())
		RespondWithError(w, http.StatusBadRequest, ERRORINVALIDPAYLOAD)
		return
	}
	m.Type = model.MEMBERSTYPEADMIN // Make sure it's an admin
	defer r.Body.Close()
	m.UUID = common.GenerateUUID()

	// Create the Member now
	if err := m.CreateMember(ctx); err != nil {
		common.Warn("Error creating first admin: %s", err.Error())
		RespondWithError(w, http.StatusInternalServerError, ERRORCREATEMEMBER)
		return
	}
	payload := mail.EmailRegisterPayload{Member: m, Author: m}
	payloadBytes := new(bytes.Buffer)
	json.NewEncoder(payloadBytes).Encode(payload)
	n := model.Notification{NotificationType: model.TypeMemberRegistration, ObjectUUID: m.UUID, SendDate: int(time.Now().Unix()), Payload: payloadBytes.Bytes()}
	if err := n.CreateNotification(ctx); err != nil {
		common.Warn("Error creating notification: %s", err.Error())
		RespondWithError(w, http.StatusInternalServerError, ERRORNOTIFICATION)
		return
	}

	RespondWithJSON(w, http.StatusCreated, m)
}

func IsInitialized(w http.ResponseWriter, r *http.Request) {
	ctx, span := tracer.Start(r.Context(), "IsInitialized")
	defer span.End()

	var m model.Member
	members, err := m.GetAll(ctx, []string{}, []string{})
	if err != nil {
		common.Warn("Error getting members: %s", err.Error())
		RespondWithError(w, http.StatusInternalServerError, ERRORGETMEMBER)
		return
	}
	if len(members) > 0 {
		RespondWithJSON(w, http.StatusOK, nil)
		return
	}
	common.Info("The app has no member")
	RespondWithJSON(w, http.StatusNoContent, nil)
}
