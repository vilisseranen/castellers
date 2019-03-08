package controller

import (
	"database/sql"
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"

	"github.com/vilisseranen/castellers/common"
	"github.com/vilisseranen/castellers/model"
)

func ParticipateEvent(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	eventUUID := vars["event_uuid"]
	memberUUID := vars["member_uuid"]
	event := model.Event{UUID: eventUUID}
	if err := event.Get(); err != nil {
		switch err {
		case sql.ErrNoRows:
			RespondWithError(w, http.StatusUnauthorized, "You are not authorized to register for this event.")
		default:
			RespondWithError(w, http.StatusInternalServerError, err.Error())
		}
		return
	}
	var p model.Participation
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&p); err != nil {
		RespondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}
	if p.Answer != common.AnswerYes &&
		p.Answer != common.AnswerNo &&
		p.Answer != common.AnswerMaybe {
		RespondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}
	defer r.Body.Close()

	p.EventUUID = eventUUID
	p.MemberUUID = memberUUID

	if err := p.Participate(); err != nil {
		RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}
	RespondWithJSON(w, http.StatusCreated, p)
}

func PresenceEvent(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	eventUUID := vars["event_uuid"]
	memberUUID := vars["member_uuid"]
	event := model.Event{UUID: eventUUID}
	if err := event.Get(); err != nil {
		switch err {
		case sql.ErrNoRows:
			RespondWithError(w, http.StatusUnauthorized, "This event does not exist.")
		default:
			RespondWithError(w, http.StatusInternalServerError, err.Error())
		}
		return
	}
	var p model.Participation
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&p); err != nil {
		RespondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}
	if p.Presence != common.AnswerYes && p.Presence != common.AnswerNo {
		RespondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}
	defer r.Body.Close()

	p.EventUUID = eventUUID
	p.MemberUUID = memberUUID

	if err := p.Present(); err != nil {
		RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}
	RespondWithJSON(w, http.StatusCreated, p)
}

func GetEventParticipation(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	eventUUID := vars["event_uuid"]
	memberUUID := vars["member_uuid"]
	p := model.Participation{EventUUID: eventUUID, MemberUUID: memberUUID}
	if err := p.GetParticipation(); err != nil {
		switch err {
		case sql.ErrNoRows:
			w.WriteHeader(204)
		default:
			RespondWithError(w, http.StatusInternalServerError, err.Error())
		}
		return
	}
	RespondWithJSON(w, http.StatusOK, p)
}
