package controller

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"regexp"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	_ "github.com/mattn/go-sqlite3"
	"github.com/vilisseranen/castellers/common"
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
		count = 100
	}
	if start < 1 {
		start = int(time.Now().Unix())
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
	vars := mux.Vars(r)
	var member_uuid string
	if vars["member_uuid"] != "" {
		member_uuid = vars["member_uuid"]
	} else if vars["admin_uuid"] != "" {
		member_uuid = vars["admin_uuid"]
	}
	if member_uuid != "" {
		for index, event := range events {
			p := model.Participation{EventUUID: event.UUID, MemberUUID: member_uuid}
			if err := p.GetParticipation(); err != nil {
				switch err {
				case sql.ErrNoRows:
					continue
				default:
					RespondWithError(w, http.StatusInternalServerError, err.Error())
				}
			}
			events[index].Participation = p.Answer
		}
	}
	if admin_uuid := vars["admin_uuid"]; admin_uuid != "" {
		for index, event := range events {
			if err := event.GetAttendance(); err != nil {
				switch err {
				default:
					RespondWithError(w, http.StatusInternalServerError, err.Error())
				}
			}
			events[index].Attendance = event.Attendance
		}
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
		event.UUID = common.GenerateUUID()
		events = append(events, event)
	} else {
		interval := intervalRegex.FindStringSubmatch(event.Recurring.Interval)
		if len(interval) != 0 && event.Recurring.Until >= event.StartDate {
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
			// Create the recurringEvent
			var recurringEvent model.RecurringEvent
			recurringEvent.UUID = common.GenerateUUID()
			recurringEvent.Name = event.Name
			recurringEvent.Description = event.Description
			recurringEvent.Interval = event.Recurring.Interval
			if err := recurringEvent.CreateRecurringEvent(); err != nil {
				RespondWithError(w, http.StatusInternalServerError, err.Error())
				return
			}
			// Compute the list of events
			for date := event.StartDate; date <= event.Recurring.Until; date += intervalSeconds {
				var anEvent model.Event

				anEvent.UUID = common.GenerateUUID()
				if event.UUID == "" {
					event.UUID = anEvent.UUID
				}
				anEvent.Name = recurringEvent.Name
				anEvent.Description = recurringEvent.Description
				anEvent.StartDate = date
				anEvent.EndDate = date + event.EndDate - event.StartDate
				anEvent.RecurringEvent = recurringEvent.UUID
				events = append(events, anEvent)

				// Adjust for Daylight Saving Time
				var location, err = time.LoadLocation("America/Montreal")
				if err != nil {
					RespondWithError(w, http.StatusInternalServerError, err.Error())
				}
				// This gives the offset of the current Zone in Montreal
				// In daylight Saving Time or Standard time accord to the time of year
				_, thisEventZoneOffset := time.Unix(int64(date), 0).In(location).Zone()
				_, nextEventZoneOffset := time.Unix(int64(date+intervalSeconds), 0).In(location).Zone()
				// If the event switch between EST and EDT, offset will adjust the time
				// So that the end user see always the event at the same time of day
				offset := thisEventZoneOffset - nextEventZoneOffset
				date = uint(int(date) + offset)
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
