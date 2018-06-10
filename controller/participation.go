package controller

import (
	"database/sql"
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/vilisseranen/castellers/model"
)

func ParticipateEvent(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	event_uuid := vars["event_uuid"]
	member_uuid := vars["member_uuid"]
	event := model.Event{UUID: event_uuid}
	member := model.Member{UUID: member_uuid}
	if err := member.Get(); err != nil {
		switch err {
		case sql.ErrNoRows:
			respondWithError(w, http.StatusUnauthorized, "You are not authorized to register for this event.")
		default:
			respondWithError(w, http.StatusInternalServerError, err.Error())
		}
		return
	}
	if err := event.Get(); err != nil {
		switch err {
		case sql.ErrNoRows:
			respondWithError(w, http.StatusUnauthorized, "You are not authorized to register for this event.")
		default:
			respondWithError(w, http.StatusInternalServerError, err.Error())
		}
		return
	}

	var p model.Participation
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&p); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}
	defer r.Body.Close()

	p.EventUUID = event_uuid
	p.MemberUUID = member_uuid

	if err := p.Participate(); err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondWithJSON(w, http.StatusCreated, p)
}
