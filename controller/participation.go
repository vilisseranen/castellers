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
	event_uuid := vars["event_uuid"]
	member_uuid := vars["member_uuid"]
	event := model.Event{UUID: event_uuid}
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
	if p.Answer != common.ANSWER_YES &&
		p.Answer != common.ANSWER_NO &&
		p.Answer != common.ANSWER_MAYBE {
		RespondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}
	defer r.Body.Close()

	p.EventUUID = event_uuid
	p.MemberUUID = member_uuid

	if err := p.Participate(); err != nil {
		RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}
	RespondWithJSON(w, http.StatusCreated, p)
}

func PresenceEvent(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	event_uuid := vars["event_uuid"]
	member_uuid := vars["member_uuid"]
	event := model.Event{UUID: event_uuid}
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
	if p.Presence != common.ANSWER_YES && p.Presence != common.ANSWER_NO {
		RespondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}
	defer r.Body.Close()

	p.EventUUID = event_uuid
	p.MemberUUID = member_uuid

	if err := p.Present(); err != nil {
		RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}
	RespondWithJSON(w, http.StatusCreated, p)
}

func GetEventParticipation(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	event_uuid := vars["event_uuid"]
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
		p := model.Participation{EventUUID: event_uuid, MemberUUID: member.UUID}
		if err := p.GetParticipation(); err != nil {
			switch err {
			case sql.ErrNoRows:
				continue
			default:
				RespondWithError(w, http.StatusInternalServerError, err.Error())
			}
		}
		members[index].Participation = p.Answer
	}
	RespondWithJSON(w, http.StatusOK, members)
}
