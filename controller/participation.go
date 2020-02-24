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
	member := model.Member{UUID: memberUUID}
	if err := event.Get(); err != nil {
		switch err {
		case sql.ErrNoRows:
			RespondWithError(w, http.StatusUnauthorized, "You are not authorized to register for this event.")
		default:
			RespondWithError(w, http.StatusInternalServerError, err.Error())
		}
		return
	}
	if err := member.Get(); err != nil {
		switch err {
		case sql.ErrNoRows:
			RespondWithError(w, http.StatusBadRequest, "This member does not exist.")
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
	member := model.Member{UUID: memberUUID}
	if err := event.Get(); err != nil {
		switch err {
		case sql.ErrNoRows:
			RespondWithError(w, http.StatusBadRequest, "This event does not exist.")
		default:
			RespondWithError(w, http.StatusInternalServerError, err.Error())
		}
		return
	}
	if err := member.Get(); err != nil {
		switch err {
		case sql.ErrNoRows:
			RespondWithError(w, http.StatusBadRequest, "This member does not exist.")
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
	if p.Presence != common.AnswerYes && p.Presence != common.AnswerNo && p.Presence != "" {
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
	m := model.Member{}
	members, err := m.GetAll()
	if err != nil {
		switch err {
		default:
			RespondWithError(w, http.StatusInternalServerError, err.Error())
		}
		return
	}
	for index, member := range members {
		p := model.Participation{EventUUID: eventUUID, MemberUUID: member.UUID}
		if err := p.GetParticipation(); err != nil {
			switch err {
			case sql.ErrNoRows:
				continue
			default:
				RespondWithError(w, http.StatusInternalServerError, err.Error())
			}
		}
		members[index].Participation = p.Answer
		members[index].Presence = p.Presence
	}
	RespondWithJSON(w, http.StatusOK, members)
}
