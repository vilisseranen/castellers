package controller

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"regexp"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	"go.elastic.co/apm"

	"github.com/vilisseranen/castellers/common"
	"github.com/vilisseranen/castellers/mail"
	"github.com/vilisseranen/castellers/model"
)

const (
	ERROREVENTNOTFOUND        = "Event not found"
	ERRORGETEVENT             = "Error getting event"
	ERRORGETEVENTS            = "Error getting events"
	ERRORGETPRESENCE          = "Error getting presence"
	ERRORGETATTENDANCE        = "Error getting attendance"
	ERRORCREATERECURRINGEVENT = "Error creating recurring event"
	ERRORGETTINGTIMEZONE      = "Error getting timezone"
	ERRORUPDATEEVENT          = "Error updating event"
	ERRORDELETEEVENT          = "Error deleting event"
)

// Regex to match any positive number followed by w (week) or d (days)
var intervalRegex = regexp.MustCompile(`^([1-9]\d*)(w|d)$`)

const intervalDaySecond = 60 * 60 * 24
const intervalWeekSecond = 60 * 60 * 24 * 7
const DEFAULT_LIMIT = 10
const MAX_LIMIT = 100

func GetEvent(w http.ResponseWriter, r *http.Request) {
	span, ctx := apm.StartSpan(r.Context(), "GetEvent", APM_SPAN_TYPE_REQUEST)
	defer span.End()

	vars := mux.Vars(r)
	UUID := vars["uuid"]
	e := model.Event{UUID: UUID}
	if err := e.Get(ctx); err != nil {
		switch err {
		case sql.ErrNoRows:
			common.Debug("Event not found: %s", err.Error())
			RespondWithError(w, http.StatusNotFound, ERROREVENTNOTFOUND)
		default:
			common.Warn("Error getting event: %s", err.Error())
			RespondWithError(w, http.StatusInternalServerError, ERRORGETEVENT)
		}
		return
	}
	if requestHasAuthorizationToken(ctx, r) {
		common.Debug("Request has authorization token")
		tokenAuth, err := ExtractToken(ctx, r)
		if err != nil {
			common.Warn("Error reading token: %s", err.Error())
			RespondWithError(w, http.StatusUnauthorized, ERRORTOKENEXPIRED)
			return
		}
		p := model.Participation{EventUUID: e.UUID, MemberUUID: tokenAuth.UserId}
		common.Debug("Getting participation for %s", p)
		if err := p.GetParticipation(r.Context()); err != nil {
			// the sql.ErrNoRows error is OK, it means the member has not yet given an answer for this event
			if err != sql.ErrNoRows {
				common.Warn("Error checking participation of member %s to event %s", tokenAuth.UserId, e.UUID)
				RespondWithError(w, http.StatusInternalServerError, ERRORGETPARTICIPATION)
				return
			}
		}
		e.Participation = p.Answer
		if common.StringInSlice(model.MEMBERSTYPEADMIN, tokenAuth.Permissions) {
			if err := e.GetAttendance(ctx); err != nil {
				common.Warn("Error counting the number of people registered or the event: %s", err.Error())
				RespondWithError(w, http.StatusInternalServerError, ERRORGETPRESENCE)
				return
			}
		}
	}
	RespondWithJSON(w, http.StatusOK, e)
}

func GetEvents(w http.ResponseWriter, r *http.Request) {
	span, ctx := apm.StartSpan(r.Context(), "GetEvents", APM_SPAN_TYPE_REQUEST)
	defer span.End()

	limit, _ := strconv.Atoi(r.FormValue("limit"))
	page, _ := strconv.Atoi(r.FormValue("page"))
	pastEvents := false
	if limit < 1 {
		limit = DEFAULT_LIMIT
	} else if limit > MAX_LIMIT {
		limit = MAX_LIMIT
	}
	if page < 0 {
		page = (page + 1) * -1
		pastEvents = true
	}
	e := model.Event{}
	events, err := e.GetAll(ctx, page, limit, pastEvents)
	if err != nil {
		common.Warn("Error getting events: %s", err.Error())
		RespondWithError(w, http.StatusInternalServerError, ERRORGETEVENTS)
		return
	}
	// if request is authenticated
	if requestHasAuthorizationToken(ctx, r) {
		tokenAuth, err := ExtractToken(ctx, r)
		if err != nil && err.Error() == "Token is expired" {
			common.Debug("Token expired, cannot get participation: %s", err.Error())
		} else if err != nil {
			common.Warn("Error reading token: %s", err.Error())
			RespondWithError(w, http.StatusInternalServerError, ERRORAUTHENTICATION)
			return
		} else {
			for index, event := range events {
				p := model.Participation{EventUUID: event.UUID, MemberUUID: tokenAuth.UserId}
				if err := p.GetParticipation(ctx); err != nil {
					switch err {
					case sql.ErrNoRows:
						common.Debug("No participation for member %s for event %s", tokenAuth.UserId, event.UUID)
						continue
					default:
						common.Warn("Error getting participation: %s", err.Error())
						RespondWithError(w, http.StatusInternalServerError, ERRORGETPARTICIPATION)
					}
				}
				events[index].Participation = p.Answer
			}
			// if token contain permission admin
			if common.StringInSlice(model.MEMBERSTYPEADMIN, tokenAuth.Permissions) {
				for index, event := range events {
					if err := event.GetAttendance(ctx); err != nil {
						common.Warn("Error getting attendance: %s", err.Error())
						RespondWithError(w, http.StatusInternalServerError, ERRORGETATTENDANCE)
						return
					}
					events[index].Attendance = event.Attendance
				}
			}
		}
	}
	RespondWithJSON(w, http.StatusOK, events)
}

func CreateEvent(w http.ResponseWriter, r *http.Request) {
	span, ctx := apm.StartSpan(r.Context(), "CreateEvent", APM_SPAN_TYPE_REQUEST)
	defer span.End()

	// Decode the event
	var event model.Event
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&event); err != nil {
		common.Debug("Invalid request payload: %s", err.Error())
		RespondWithError(w, http.StatusBadRequest, ERRORINVALIDPAYLOAD)
		return
	}
	defer r.Body.Close()
	common.Debug("Creating event: %s", event)

	// Validation on events data
	if validEventData(ctx, event) == false {
		common.Debug("Invalid request payload: %s", event)
		RespondWithError(w, http.StatusBadRequest, ERRORINVALIDPAYLOAD)
		return
	}

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
				common.Debug("Invalid request payload: %s", err.Error())
				RespondWithError(w, http.StatusBadRequest, ERRORINVALIDPAYLOAD)
			}
			switch interval[2] {
			case "d":
				intervalSeconds *= intervalDaySecond
			case "w":
				intervalSeconds *= intervalWeekSecond
			}
			// Create the recurringEvent
			var recurringEvent model.RecurringEvent
			recurringEvent.UUID = common.GenerateUUID()
			recurringEvent.Name = event.Name
			recurringEvent.Description = event.Description
			recurringEvent.Interval = event.Recurring.Interval
			if err := recurringEvent.CreateRecurringEvent(ctx); err != nil {
				common.Warn("Error creating recurring event: %s", err.Error())
				RespondWithError(w, http.StatusInternalServerError, ERRORCREATERECURRINGEVENT)
				return
			}
			// Compute the list of events
			for date := event.StartDate; date <= event.Recurring.Until; date += intervalSeconds {
				var anEvent model.Event

				anEvent.UUID = common.GenerateUUID()
				anEvent.Name = recurringEvent.Name
				anEvent.Description = recurringEvent.Description
				anEvent.StartDate = date
				anEvent.EndDate = date + event.EndDate - event.StartDate
				anEvent.RecurringEvent = recurringEvent.UUID
				anEvent.Type = event.Type
				events = append(events, anEvent)

				// Adjust for Daylight Saving Time
				var location, err = time.LoadLocation("America/Montreal")
				if err != nil {
					common.Warn("Error getting timezone data: %s", err.Error())
					RespondWithError(w, http.StatusInternalServerError, ERRORGETTINGTIMEZONE)
					return
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
			common.Debug("Invalid request payload")
			RespondWithError(w, http.StatusBadRequest, ERRORINVALIDPAYLOAD)
			return
		}
	}

	// Create the events
	for _, event := range events {
		if err := event.CreateEvent(ctx); err != nil {
			common.Warn("Error creating recurring event: %s", err.Error())
			RespondWithError(w, http.StatusInternalServerError, ERRORCREATERECURRINGEVENT)
			return
		}
	}

	// Send notification
	if event.StartDate > uint(time.Now().Unix()) { // Do not send emails for events in the past

		// Encode payload
		event.UUID = events[0].UUID
		payload := mail.EmailCreateEventPayload{Event: event}
		payloadBytes := new(bytes.Buffer)
		json.NewEncoder(payloadBytes).Encode(payload)

		n := model.Notification{NotificationType: model.TypeEventCreated, SendDate: int(time.Now().Unix()), Payload: payloadBytes.Bytes()}
		if err := n.CreateNotification(ctx); err != nil {
			common.Warn("Error creating notification: %s", err.Error())
			RespondWithError(w, http.StatusInternalServerError, ERRORNOTIFICATION)
			return
		}
	}
	RespondWithJSON(w, http.StatusCreated, event)
}

func UpdateEvent(w http.ResponseWriter, r *http.Request) {
	span, ctx := apm.StartSpan(r.Context(), "UpdateEvent", APM_SPAN_TYPE_REQUEST)
	defer span.End()

	vars := mux.Vars(r)
	UUID := vars["uuid"]
	var e model.Event
	eventBeforeUpdate := model.Event{UUID: UUID}
	if err := eventBeforeUpdate.Get(ctx); err != nil {
		common.Warn("Error getting event before update: %s", err.Error())
		RespondWithError(w, http.StatusInternalServerError, ERRORUPDATEEVENT)
		return
	}
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&e); err != nil {
		common.Debug("Invalid request payload: %s", err.Error())
		RespondWithError(w, http.StatusBadRequest, ERRORINVALIDPAYLOAD)
		return
	}
	defer r.Body.Close()

	// Validation on events data
	if validEventData(ctx, e) == false {
		common.Debug("Invalid request payload: %s", e)
		RespondWithError(w, http.StatusBadRequest, ERRORINVALIDPAYLOAD)
		return
	}
	e.UUID = UUID

	if err := e.UpdateEvent(ctx); err != nil {
		common.Warn("Error updating event: %s", err.Error())
		RespondWithError(w, http.StatusInternalServerError, ERRORUPDATEEVENT)
		return
	}

	// Send notification
	if e.StartDate > uint(time.Now().Unix()) { // Do not send emails for events in the past

		// Encode payload
		payload := mail.EmailModifiedPayload{EventBeforeUpdate: eventBeforeUpdate, EventAfterUpdate: e}
		payloadBytes := new(bytes.Buffer)
		json.NewEncoder(payloadBytes).Encode(payload)

		n := model.Notification{NotificationType: model.TypeEventModified, SendDate: int(time.Now().Unix()), Payload: payloadBytes.Bytes()}
		if err := n.CreateNotification(ctx); err != nil {
			common.Warn("Error creating notification: %s", err.Error())
			RespondWithError(w, http.StatusInternalServerError, ERRORNOTIFICATION)
			return
		}
	} else {
		common.Debug("Event %s is in the past, not sending the notification", e.UUID)
	}
	RespondWithJSON(w, http.StatusOK, e)
}

func DeleteEvent(w http.ResponseWriter, r *http.Request) {
	span, ctx := apm.StartSpan(r.Context(), "DeleteEvent", APM_SPAN_TYPE_REQUEST)
	defer span.End()

	vars := mux.Vars(r)
	UUID := vars["uuid"]
	e := model.Event{UUID: UUID}
	if err := e.Get(ctx); err != nil {
		common.Warn("Error getting event to delete: %s")
		RespondWithError(w, http.StatusInternalServerError, ERRORDELETEEVENT)
		return
	}
	if err := e.DeleteEvent(ctx); err != nil {
		common.Warn("Error deleting event to delete: %s")
		RespondWithError(w, http.StatusInternalServerError, ERRORDELETEEVENT)
		return
	}
	if e.StartDate > uint(time.Now().Unix()) { // Do not send emails for events in the past
		payload := mail.EmailDeletedEventPayload{EventDeleted: e}
		payloadBytes := new(bytes.Buffer)
		json.NewEncoder(payloadBytes).Encode(payload)
		n := model.Notification{NotificationType: model.TypeEventDeleted, SendDate: int(time.Now().Unix()), Payload: payloadBytes.Bytes()}
		if err := n.CreateNotification(ctx); err != nil {
			common.Warn("Error creating notification: %s", err.Error())
			RespondWithError(w, http.StatusInternalServerError, ERRORNOTIFICATION)
			return
		}
	} else {
		common.Debug("Event %s is in the past, not sending the notification", e.UUID)
	}
	RespondWithJSON(w, http.StatusOK, nil)
}

func validEventData(ctx context.Context, event model.Event) bool {
	span, ctx := apm.StartSpan(ctx, "validEventData", APM_SPAN_TYPE_REQUEST)
	defer span.End()

	var valid = true
	var validType = false
	if event.StartDate > event.EndDate ||
		event.Name == "" {
		valid = false
	}
	for _, eventType := range model.ValidEventTypes {
		if event.Type == eventType {
			validType = true
		}
	}
	return (valid && validType)
}
