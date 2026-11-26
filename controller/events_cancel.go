package controller

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/gorilla/mux"

	"github.com/vilisseranen/castellers/common"
	"github.com/vilisseranen/castellers/mail"
	"github.com/vilisseranen/castellers/model"
)

const (
	ERRORCANCELEVENT         = "error cancelling event"
	ERRORRESTOREEVENT        = "error restoring event"
	ERRORINVALIDCANCELREASON = "invalid cancel reason"
	ERROREVENTCANCELLED      = "event is cancelled"
	ERROREVENTNOTCANCELLED   = "event is not cancelled"
)

// MAXCANCELNOTELENGTH limits the free text given with the reason.
const MAXCANCELNOTELENGTH = 500

type cancelEventRequest struct {
	Reason string `json:"reason"`
	Note   string `json:"note"`
}

// getEventOrRespond loads the event of the request. When it cannot, it answers
// the request and returns false.
func getEventOrRespond(ctx context.Context, w http.ResponseWriter, e *model.Event, errorMessage string) bool {
	if err := e.Get(ctx); err != nil {
		switch err {
		case sql.ErrNoRows:
			RespondWithError(w, http.StatusNotFound, ERROREVENTNOTFOUND)
		default:
			common.Warn("Error getting event %s: %s", e.UUID, err.Error())
			RespondWithError(w, http.StatusInternalServerError, errorMessage)
		}
		return false
	}
	return true
}

// CancelEvent cancels an event: it stays listed, shown as cancelled with the
// reason, but nobody can answer anymore and no reminder is sent. The members
// are told by email, unless the admin asks not to: ?notify=false.
// Cancelling an event that is already cancelled changes the reason and the
// note, without telling the members again.
func CancelEvent(w http.ResponseWriter, r *http.Request) {
	ctx, span := tracer.Start(r.Context(), "CancelEvent")
	defer span.End()

	e := model.Event{UUID: mux.Vars(r)["uuid"]}
	if !getEventOrRespond(ctx, w, &e, ERRORCANCELEVENT) {
		return
	}
	var body cancelEventRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		common.Debug("Invalid request payload: %s", err.Error())
		RespondWithError(w, http.StatusBadRequest, ERRORINVALIDPAYLOAD)
		return
	}
	defer r.Body.Close()
	body.Note = strings.TrimSpace(body.Note)
	if !common.StringInSlice(body.Reason, model.ValidCancelReasons) {
		RespondWithError(w, http.StatusBadRequest, ERRORINVALIDCANCELREASON)
		return
	}
	if len([]rune(body.Note)) > MAXCANCELNOTELENGTH {
		RespondWithError(w, http.StatusBadRequest, ERRORINVALIDPAYLOAD)
		return
	}

	wasCancelled := e.Cancelled == 1
	e.CancelReason = body.Reason
	e.CancelNote = body.Note
	if err := e.CancelEvent(ctx); err != nil {
		common.Warn("Error cancelling event %s: %s", e.UUID, err.Error())
		RespondWithError(w, http.StatusInternalServerError, ERRORCANCELEVENT)
		return
	}
	if wasCancelled {
		common.Debug("Event %s was already cancelled, not sending the notification again", e.UUID)
	} else if !queueEventCancellationNotification(ctx, r, e, model.TypeEventCancelled) {
		RespondWithError(w, http.StatusInternalServerError, ERRORNOTIFICATION)
		return
	}
	RespondWithJSON(w, http.StatusOK, e)
}

// RestoreEvent puts a cancelled event back. The members are told by email,
// unless the admin asks not to: ?notify=false.
func RestoreEvent(w http.ResponseWriter, r *http.Request) {
	ctx, span := tracer.Start(r.Context(), "RestoreEvent")
	defer span.End()

	e := model.Event{UUID: mux.Vars(r)["uuid"]}
	if !getEventOrRespond(ctx, w, &e, ERRORRESTOREEVENT) {
		return
	}
	if e.Cancelled != 1 {
		RespondWithError(w, http.StatusBadRequest, ERROREVENTNOTCANCELLED)
		return
	}
	if err := e.RestoreEvent(ctx); err != nil {
		common.Warn("Error restoring event %s: %s", e.UUID, err.Error())
		RespondWithError(w, http.StatusInternalServerError, ERRORRESTOREEVENT)
		return
	}
	if !queueEventCancellationNotification(ctx, r, e, model.TypeEventRestored) {
		RespondWithError(w, http.StatusInternalServerError, ERRORNOTIFICATION)
		return
	}
	RespondWithJSON(w, http.StatusOK, e)
}

// queueEventCancellationNotification queues the email telling the members that
// an event is cancelled, or restored. Nothing is queued when the admin asked
// not to tell the members (?notify=false) or when the event is in the past.
// It returns false when the notification could not be queued.
func queueEventCancellationNotification(ctx context.Context, r *http.Request, e model.Event, notificationType string) bool {
	if r.FormValue("notify") == "false" {
		common.Debug("Event %s: no %s notification, as requested", e.UUID, notificationType)
		return true
	}
	if e.StartDate <= uint(time.Now().Unix()) { // Do not send emails for events in the past
		common.Debug("Event %s is in the past, not sending the notification", e.UUID)
		return true
	}
	payloadBytes := new(bytes.Buffer)
	if err := json.NewEncoder(payloadBytes).Encode(mail.EmailCancelledEventPayload{Event: e}); err != nil {
		common.Warn("Error encoding notification: %s", err.Error())
		return false
	}
	n := model.Notification{NotificationType: notificationType, ObjectUUID: e.UUID, SendDate: int(time.Now().Unix()), Payload: payloadBytes.Bytes()}
	if err := n.CreateNotification(ctx); err != nil {
		common.Warn("Error creating notification: %s", err.Error())
		return false
	}
	return true
}

// deliverEventCancellationNotification sends the email of a cancelled, or
// restored, event to the members concerned by this event.
func deliverEventCancellationNotification(ctx context.Context, notification *model.Notification) {
	var payload mail.EmailCancelledEventPayload
	if err := json.Unmarshal(notification.Payload, &payload); err != nil {
		common.Error("%v\n", err)
		notification.Delivered = model.NotificationDeliveryFailure
		notification.UpdateNotificationStatus(ctx)
		return
	}
	// The admin may have changed their mind since: tell the members what is
	// true now, and nothing when the email no longer matches the event.
	event := model.Event{UUID: payload.Event.UUID}
	if err := event.Get(ctx); err != nil {
		common.Error("%v\n", err)
		notification.Delivered = model.NotificationDeliveryFailure
		notification.UpdateNotificationStatus(ctx)
		return
	}
	restored := notification.NotificationType == model.TypeEventRestored
	if restored == (event.Cancelled == 1) || event.StartDate < uint(time.Now().Unix()) {
		common.Info("Event %v: the %s notification is outdated.\n", event.UUID, notification.NotificationType)
		notification.Delivered = model.NotificationTooLate
		notification.UpdateNotificationStatus(ctx)
		return
	}
	m := model.Member{}
	members, err := m.GetAll(ctx, []string{model.MEMBERSSTATUSACTIVATED, model.MEMBERSSTATUSPAUSED}, []string{})
	if err != nil {
		common.Error("Error getting members: %v\n", err)
		notification.Delivered = model.NotificationDeliveryFailure
		notification.UpdateNotificationStatus(ctx)
		return
	}
	members = EventAudience(event, members)
	failures := 0
	for _, member := range members {
		if member.Subscribed != 1 {
			continue
		}
		memberPayload := mail.EmailCancelledEventPayload{Member: member, Event: event, Restored: restored}
		if err := mail.SendCancelledEventEmail(ctx, memberPayload); err != nil {
			common.Error("%v\n", err)
			failures += 1
		}
	}
	if failures == 0 {
		notification.Delivered = model.NotificationDeliverySuccess
	} else if failures == len(members) {
		notification.Delivered = model.NotificationDeliveryFailure
	} else {
		notification.Delivered = model.NotificationDeliveryPartialFailure
	}
	notification.UpdateNotificationStatus(ctx)
}
