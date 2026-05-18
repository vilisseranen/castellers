package controller

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"time"

	"github.com/gorilla/mux"

	"github.com/vilisseranen/castellers/common"
	"github.com/vilisseranen/castellers/mail"
	"github.com/vilisseranen/castellers/model"
)

const (
	ERRORQUEUEEVENTREMINDER   = "error queueing event reminders"
	ERROREVENTALREADYSTARTED  = "event has already started"
	ERRORINVALIDREMINDERAUDIENCE = "invalid reminder audience"
	ERRORREMINDERMEMBERSREQUIRED = "memberUuids required for members audience"
)

type sendEventRemindersRequest struct {
	Audience    string   `json:"audience"`
	MemberUUIDs []string `json:"memberUuids"`
}

type sendEventRemindersResponse struct {
	NotificationID int    `json:"notificationId"`
	Message        string `json:"message"`
}

func SendEventReminders(w http.ResponseWriter, r *http.Request) {
	ctx, span := tracer.Start(r.Context(), "SendEventReminders")
	defer span.End()

	vars := mux.Vars(r)
	eventUUID := vars["event_uuid"]

	var body sendEventRemindersRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		RespondWithError(w, http.StatusBadRequest, ERRORINVALIDPAYLOAD)
		return
	}

	if !isValidReminderAudience(body.Audience) {
		RespondWithError(w, http.StatusBadRequest, ERRORINVALIDREMINDERAUDIENCE)
		return
	}
	if body.Audience == model.ManualReminderAudienceMembers && len(body.MemberUUIDs) == 0 {
		RespondWithError(w, http.StatusBadRequest, ERRORREMINDERMEMBERSREQUIRED)
		return
	}

	event := model.Event{UUID: eventUUID}
	if err := event.Get(ctx); err != nil {
		switch err {
		case sql.ErrNoRows:
			RespondWithError(w, http.StatusNotFound, ERROREVENTNOTFOUND)
		default:
			RespondWithError(w, http.StatusInternalServerError, ERRORGETEVENT)
		}
		return
	}
	if event.StartDate < uint(time.Now().Unix()) {
		RespondWithError(w, http.StatusBadRequest, ERROREVENTALREADYSTARTED)
		return
	}

	tokenAuth, err := ExtractToken(ctx, r)
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, ERRORAUTHENTICATION)
		return
	}

	payload := model.ManualReminderPayload{
		Audience:    body.Audience,
		MemberUUIDs: body.MemberUUIDs,
		RequestedBy: tokenAuth.UserId,
	}
	payloadBytes := new(bytes.Buffer)
	if err := json.NewEncoder(payloadBytes).Encode(payload); err != nil {
		RespondWithError(w, http.StatusInternalServerError, ERRORQUEUEEVENTREMINDER)
		return
	}

	n := model.Notification{
		NotificationType: model.TypeManualEventReminder,
		ObjectUUID:       event.UUID,
		SendDate:         int(time.Now().Unix()),
		Payload:          payloadBytes.Bytes(),
	}
	if err := n.CreateNotification(ctx); err != nil {
		RespondWithError(w, http.StatusInternalServerError, ERRORNOTIFICATION)
		return
	}

	RespondWithJSON(w, http.StatusAccepted, sendEventRemindersResponse{
		NotificationID: n.ID,
		Message:        "reminders queued",
	})
}

func isValidReminderAudience(audience string) bool {
	switch audience {
	case model.ManualReminderAudienceDefault,
		model.ManualReminderAudienceNoAnswerActive,
		model.ManualReminderAudienceNoAnswerActivePaused,
		model.ManualReminderAudienceMembers:
		return true
	default:
		return false
	}
}

func resolveManualReminderRecipients(ctx context.Context, event model.Event, payload model.ManualReminderPayload) ([]model.Member, error) {
	switch payload.Audience {
	case model.ManualReminderAudienceDefault:
		m := model.Member{}
		return m.GetAll(ctx, []string{model.MEMBERSSTATUSACTIVATED, model.MEMBERSSTATUSPAUSED}, []string{})
	case model.ManualReminderAudienceNoAnswerActive:
		return membersWithNoAnswer(ctx, event.UUID, []string{model.MEMBERSSTATUSACTIVATED})
	case model.ManualReminderAudienceNoAnswerActivePaused:
		return membersWithNoAnswer(ctx, event.UUID, []string{model.MEMBERSSTATUSACTIVATED, model.MEMBERSSTATUSPAUSED})
	case model.ManualReminderAudienceMembers:
		members := []model.Member{}
		for _, uuid := range payload.MemberUUIDs {
			m := model.Member{UUID: uuid}
			if err := m.Get(ctx); err != nil {
				if err == sql.ErrNoRows {
					continue
				}
				return nil, err
			}
			if m.Status == model.MEMBERSSTATUSDELETED || m.Status == model.MEMBERSSTATUSPURGED {
				continue
			}
			members = append(members, m)
		}
		return members, nil
	default:
		return nil, nil
	}
}

func membersWithNoAnswer(ctx context.Context, eventUUID string, statuses []string) ([]model.Member, error) {
	m := model.Member{}
	all, err := m.GetAll(ctx, statuses, []string{})
	if err != nil {
		return nil, err
	}
	filtered := []model.Member{}
	p := model.Participation{EventUUID: eventUUID}
	for _, member := range all {
		p.MemberUUID = member.UUID
		if err := p.GetParticipation(ctx); err != nil {
			switch err {
			case sql.ErrNoRows:
				filtered = append(filtered, member)
			default:
				return nil, err
			}
			continue
		}
		if p.Answer == "" {
			filtered = append(filtered, member)
		}
	}
	return filtered, nil
}

func sendReminderEmailsToMembers(ctx context.Context, event model.Event, members []model.Member) int {
	failures := 0
	p := model.Participation{EventUUID: event.UUID}
	for _, member := range members {
		p.MemberUUID = member.UUID
		err := p.GetParticipation(ctx)
		if err != nil {
			switch err {
			case sql.ErrNoRows:
				p.Answer = ""
			default:
				common.Error("%v\n", err)
				failures++
				continue
			}
		}
		if member.Subscribed != 1 {
			continue
		}
		token, err := ParticipateEventToken(ctx, member.UUID, common.GetConfigInt("jwt.participation_ttl_minutes"))
		if err != nil {
			common.Error("%v\n", err)
			failures++
			continue
		}
		dependents, err := member.GetDependents(ctx)
		if err != nil {
			common.Error("%v\n", err)
			failures++
			continue
		}
		emailPayload := mail.EmailReminderPayload{
			Member:        member,
			Event:         event,
			Participation: p,
			Token:         token,
			Dependents:    dependents,
		}
		if err := mail.SendReminderEmail(ctx, emailPayload); err != nil {
			common.Error("%v\n", err)
			failures++
		}
	}
	return failures
}

func setReminderDeliveryStatus(notification *model.Notification, failures, memberCount int) {
	if failures == 0 {
		notification.Delivered = model.NotificationDeliverySuccess
	} else if failures == memberCount {
		notification.Delivered = model.NotificationDeliveryFailure
	} else {
		notification.Delivered = model.NotificationDeliveryPartialFailure
	}
}

func deliverEventReminderNotification(ctx context.Context, notification *model.Notification, members []model.Member) {
	event := model.Event{UUID: notification.ObjectUUID}
	if err := event.Get(ctx); err != nil {
		common.Error("%v\n", err)
		notification.Delivered = model.NotificationDeliveryFailure
		notification.UpdateNotificationStatus(ctx)
		return
	}
	if event.StartDate < uint(time.Now().Unix()) {
		common.Info("Event %v has already started.\n", event.UUID)
		notification.Delivered = model.NotificationTooLate
		notification.UpdateNotificationStatus(ctx)
		return
	}
	failures := sendReminderEmailsToMembers(ctx, event, members)
	setReminderDeliveryStatus(notification, failures, len(members))
	notification.UpdateNotificationStatus(ctx)
}
