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
	m.Type = model.MemberTypeAdmin // Make sure it's an admin
	defer r.Body.Close()
	m.UUID = common.GenerateUUID()
	m.Code = common.GenerateCode()

	// Create the Member now
	if err := m.CreateMember(); err != nil {
		RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}
	payload := mail.EmailRegisterPayload{Member: m, Author: m}
	payloadBytes := new(bytes.Buffer)
	json.NewEncoder(payloadBytes).Encode(payload)
	n := model.Notification{NotificationType: model.TypeMemberRegistration, ObjectUUID: m.UUID, SendDate: int(time.Now().Unix()), Payload: payloadBytes.Bytes()}
	if err := n.CreateNotification(); err != nil {
		RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
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
	}
	RespondWithJSON(w, http.StatusNoContent, nil)
	return

}
