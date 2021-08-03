package controller

import (
	"database/sql"
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/vilisseranen/castellers/common"
	"github.com/vilisseranen/castellers/model"
)

const (
	ERRORPARTICIPATEEVENT = "Error setting participation to event"
	ERRORPRESENCEEVENT    = "Error setting presence to event"
	ERRORGETPARTICIPATION = "Error getting participation"
)

func ParticipateEvent(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	tokenAuth, err := ExtractToken(r)
	if err != nil {
		common.Warn("Error reading token: %s", err.Error())
		RespondWithError(w, http.StatusInternalServerError, ERRORAUTHENTICATION)
		return
	}
	eventUUID := vars["event_uuid"]
	memberUUID := tokenAuth.UserId
	event := model.Event{UUID: eventUUID}
	member := model.Member{UUID: memberUUID}
	if err := event.Get(); err != nil {
		switch err {
		case sql.ErrNoRows:
			common.Debug("Event not found: %s", err.Error())
			RespondWithError(w, http.StatusNotFound, ERROREVENTNOTFOUND)
		default:
			common.Warn("Error getting Event: %s", err.Error())
			RespondWithError(w, http.StatusInternalServerError, ERRORPARTICIPATEEVENT)
		}
		return
	}
	if err := member.Get(); err != nil {
		common.Warn("Error getting Member: %s", err.Error())
		RespondWithError(w, http.StatusInternalServerError, ERRORPARTICIPATEEVENT)
		return
	}
	var p model.Participation
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&p); err != nil {
		common.Debug("Invalid request payload: %s", err.Error())
		RespondWithError(w, http.StatusBadRequest, ERRORINVALIDPAYLOAD)
		return
	}
	if p.Answer != common.AnswerYes &&
		p.Answer != common.AnswerNo &&
		p.Answer != common.AnswerMaybe {
		common.Debug("Invalid request payload: %s", err.Error())
		RespondWithError(w, http.StatusBadRequest, ERRORINVALIDPAYLOAD)
		return
	}
	defer r.Body.Close()

	p.EventUUID = eventUUID
	p.MemberUUID = memberUUID

	if err := p.Participate(); err != nil {
		common.Warn("Error participating event: %s", err.Error())
		RespondWithError(w, http.StatusInternalServerError, ERRORPARTICIPATEEVENT)
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
			common.Debug("Event not found: %s", err.Error())
			RespondWithError(w, http.StatusBadRequest, ERROREVENTNOTFOUND)
		default:
			common.Warn("Error getting Event: %s", err.Error())
			RespondWithError(w, http.StatusInternalServerError, ERRORPRESENCEEVENT)
		}
		return
	}
	if err := member.Get(); err != nil {
		switch err {
		case sql.ErrNoRows:
			common.Debug("Member not found: %s", err.Error())
			RespondWithError(w, http.StatusBadRequest, ERRORPRESENCEEVENT)
			return
		default:
			common.Warn("Error getting Member: %s", err.Error())
			RespondWithError(w, http.StatusInternalServerError, ERRORPRESENCEEVENT)
		}
		return
	}
	var p model.Participation
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&p); err != nil {
		common.Debug("Invalid request payload: %s", err.Error())
		RespondWithError(w, http.StatusBadRequest, ERRORINVALIDPAYLOAD)
		return
	}
	if p.Presence != common.AnswerYes && p.Presence != common.AnswerNo && p.Presence != "" {
		common.Debug("Invalid request payload")
		RespondWithError(w, http.StatusBadRequest, ERRORINVALIDPAYLOAD)
		return
	}
	defer r.Body.Close()

	p.EventUUID = eventUUID
	p.MemberUUID = memberUUID

	if err := p.Present(); err != nil {
		common.Warn("Error setting presence to event: %s", err.Error())
		RespondWithError(w, http.StatusInternalServerError, ERRORPRESENCEEVENT)
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
		common.Warn("Error getting members: %s", err.Error())
		RespondWithError(w, http.StatusInternalServerError, ERRORGETPARTICIPATION)
		return
	}
	for index, member := range members {
		p := model.Participation{EventUUID: eventUUID, MemberUUID: member.UUID}
		if err := p.GetParticipation(); err != nil {
			switch err {
			case sql.ErrNoRows:
				continue
			default:
				common.Warn("Error getting participation: %s", err.Error())
				RespondWithError(w, http.StatusInternalServerError, ERRORGETPARTICIPATION)
				return
			}
		}
		members[index].Participation = p.Answer
		members[index].Presence = p.Presence
	}
	RespondWithJSON(w, http.StatusOK, members)
}
