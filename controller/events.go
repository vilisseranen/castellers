package controller

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"regexp"
	"strconv"

	"github.com/gorilla/mux"
	_ "github.com/mattn/go-sqlite3"
	"github.com/vilisseranen/castellers/model"
)

// Regex to match any positive number followed by w (week) or d (days)
var intervalRegex = regexp.MustCompile(`^([1-9]\d*)(w|d)$`)

const INTERVAL_DAY_SECOND = 60 * 60 * 24
const INTERVAL_WEEK_SECOND = 60 * 60 * 24 * 7

func GetEvent(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	uuid := vars["uuid"]
	e := model.Event{UUID: uuid}
	if err := e.Get(); err != nil {
		switch err {
		case sql.ErrNoRows:
			RespondWithError(w, http.StatusNotFound, "Event not found")
		default:
			RespondWithError(w, http.StatusInternalServerError, err.Error())
		}
		return
	}
	RespondWithJSON(w, http.StatusOK, e)
}

func GetEvents(w http.ResponseWriter, r *http.Request) {
	count, _ := strconv.Atoi(r.FormValue("count"))
	start, _ := strconv.Atoi(r.FormValue("start"))
	if count < 1 {
		count = 10
	}
	e := model.Event{}
	events, err := e.GetAll(start, count)
	if err != nil {
		switch err {
		default:
			RespondWithError(w, http.StatusInternalServerError, err.Error())
		}
		return
	}
	RespondWithJSON(w, http.StatusOK, events)
}

func CreateEvent(w http.ResponseWriter, r *http.Request) {

	// Decode the event
	var event model.Event
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&event); err != nil {
		RespondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}
	defer r.Body.Close()

	// Compute all events
	var events = make([]model.Event, 0)
	if event.Recurring.Interval == "" || event.Recurring.Until == 0 {
		events = append(events, event)
	} else {
		interval := intervalRegex.FindStringSubmatch(event.Recurring.Interval)
		if len(interval) != 0 && event.Recurring.Until > event.StartDate {
			inter, err := strconv.ParseUint(interval[1], 10, 32)
			intervalSeconds := uint(inter)
			if err != nil {
				RespondWithError(w, http.StatusBadRequest, "Invalid request payload")
			}
			switch interval[2] {
			case "d":
				intervalSeconds *= INTERVAL_DAY_SECOND
			case "w":
				intervalSeconds *= INTERVAL_WEEK_SECOND
			}
			for date := event.StartDate; date <= event.Recurring.Until; date += intervalSeconds {
				var anEvent model.Event
				anEvent.Name = event.Name
				anEvent.StartDate = date
				anEvent.EndDate = date + event.EndDate - event.StartDate
				events = append(events, anEvent)
			}
		} else {
			RespondWithError(w, http.StatusBadRequest, "Invalid request payload")
			return
		}
	}

	// Create the events
	for _, event := range events {
		if err := event.CreateEvent(); err != nil {
			RespondWithError(w, http.StatusInternalServerError, err.Error())
			return
		}
	}
	RespondWithJSON(w, http.StatusCreated, event)
}

func UpdateEvent(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	uuid := vars["uuid"]
	var e model.Event
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&e); err != nil {
		RespondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}
	defer r.Body.Close()
	e.UUID = uuid
	if err := e.UpdateEvent(); err != nil {
		RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}
	RespondWithJSON(w, http.StatusOK, e)
}

func DeleteEvent(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	uuid := vars["uuid"]
	e := model.Event{UUID: uuid}
	if err := e.DeleteEvent(); err != nil {
		RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}
	RespondWithJSON(w, http.StatusOK, nil)
}
