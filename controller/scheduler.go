package controller

import (
	"context"
	"database/sql"
	"encoding/json"
	"sort"
	"time"

	"github.com/robfig/cron"

	"github.com/vilisseranen/castellers/common"
	"github.com/vilisseranen/castellers/mail"
	"github.com/vilisseranen/castellers/model"
)

type Scheduler struct {
	cron *cron.Cron
}

func (s *Scheduler) Start() {
	s.cron = cron.New()

	// Send notifications when they are ready
	s.cron.AddFunc("@every 10s", checkAndSendNotification)

	// Look for upcoming events and generate reminder notifications
	s.cron.AddFunc("@every 10s", generateEventsNotificationsReminder)

	// Look for upcoming events and generate summary notifications
	s.cron.AddFunc("@every 10m", generateEventsNotificationsSummary)

	// Change status of member who have not participated in some time
	s.cron.AddFunc("@every 10m", pauseAbsentMembers)

	s.cron.Start()
}

// RunNotificationDeliveryOnce processes all pending notifications (used by tests and cron).
func RunNotificationDeliveryOnce() {
	checkAndSendNotification()
}

// RunPauseAbsentMembersOnce runs the inactivity pause scan once (used by tests and cron).
func RunPauseAbsentMembersOnce() {
	pauseAbsentMembers()
}

func checkAndSendNotification() {

	ctx, span := tracer.Start(context.Background(), "checkAndSendNotification")
	defer span.End()

	// Get notifications
	n := model.Notification{}
	notificationsToSend, err := n.GetNotificationsReady(ctx)
	if err != nil {
		common.Error("%v\n", err)
	}
	// Check all notifications that are ready
	for _, notification := range notificationsToSend {
		notification.Delivered = model.NotificationDeliveryInProgress
		notification.UpdateNotificationStatus(ctx)
		switch notificationType := notification.NotificationType; notificationType {
		case model.TypeMemberRegistration:
			// Send the email
			if common.GetConfigBool("smtp.enabled") {
				var payload mail.EmailRegisterPayload
				if err := json.Unmarshal(notification.Payload, &payload); err != nil {
					common.Error("%v\n", err)
					notification.Delivered = model.NotificationDeliveryFailure
					notification.UpdateNotificationStatus(ctx)
					continue
				}
				// Get a token to create credentials
				resetCredentialsToken, err := ResetCredentialsToken(ctx, payload.Member.UUID, payload.Member.Email, common.GetConfigInt("jwt.registration_ttl_minutes"))
				if err != nil {
					notification.Delivered = model.NotificationDeliveryFailure
					notification.UpdateNotificationStatus(ctx)
					continue
				}
				payload.Token = resetCredentialsToken
				if err := mail.SendRegistrationEmail(ctx, payload); err != nil {
					notification.Delivered = model.NotificationDeliveryFailure
					notification.UpdateNotificationStatus(ctx)
					continue
				}
			}
			notification.Delivered = model.NotificationDeliverySuccess
			notification.UpdateNotificationStatus(ctx)
		case model.TypeUpcomingEvent:
			m := model.Member{}
			members, err := m.GetAll(ctx, []string{model.MEMBERSSTATUSACTIVATED, model.MEMBERSSTATUSPAUSED}, []string{})
			if err != nil {
				common.Error("%v\n", err)
				notification.Delivered = model.NotificationDeliveryFailure
				notification.UpdateNotificationStatus(ctx)
				continue
			}
			deliverEventReminderNotification(ctx, &notification, members)
		case model.TypeManualEventReminder:
			var payload model.ManualReminderPayload
			if err := json.Unmarshal(notification.Payload, &payload); err != nil {
				common.Error("%v\n", err)
				notification.Delivered = model.NotificationDeliveryFailure
				notification.UpdateNotificationStatus(ctx)
				continue
			}
			event := model.Event{UUID: notification.ObjectUUID}
			members, err := resolveManualReminderRecipients(ctx, event, payload)
			if err != nil {
				common.Error("%v\n", err)
				notification.Delivered = model.NotificationDeliveryFailure
				notification.UpdateNotificationStatus(ctx)
				continue
			}
			deliverEventReminderNotification(ctx, &notification, members)
		case model.TypeSummaryEvent:
			event := model.Event{UUID: notification.ObjectUUID}
			err := event.Get(ctx)
			if err != nil {
				// Cannot get the event, complete failure
				common.Error("%v\n", err)
				notification.Delivered = model.NotificationDeliveryFailure
				notification.UpdateNotificationStatus(ctx)
				continue
			}
			if event.StartDate < uint(time.Now().Unix()) {
				// Event has begun or is finished, we don't send the notification
				common.Info("Event %v has already started.\n", event.UUID)
				notification.Delivered = model.NotificationTooLate
				notification.UpdateNotificationStatus(ctx)
				continue
			}
			m := model.Member{}
			members, err := m.GetAll(ctx, []string{}, []string{})
			if err != nil {
				// Cannot get the members, complete failure
				common.Error("%v\n", err)
				notification.Delivered = model.NotificationDeliveryFailure
				notification.UpdateNotificationStatus(ctx)
				continue
			}
			failures := 0
			// Get participation for all members
			for index, member := range members {
				p := model.Participation{EventUUID: notification.ObjectUUID, MemberUUID: member.UUID}
				if err := p.GetParticipation(ctx); err != nil {
					switch err {
					case sql.ErrNoRows:
						members[index].Participation = ""
					default:
						// Cannot get participation for user
						failures += 1
						continue
					}
				}
				members[index].Participation = p.Answer
				members[index].Presence = p.Presence
			}
			// Sort by FirstName then by Participation
			sort.Slice(members, func(i, j int) bool { return members[i].FirstName < members[j].FirstName })
			sort.Slice(members, func(i, j int) bool { return members[i].Participation > members[j].Participation })
			// Only show active members or members who have given a participation answer.
			// Long-paused members who never answered would otherwise clutter the summary.
			participantsForEmail := make([]model.Member, 0, len(members))
			for _, member := range members {
				if member.Status == model.MEMBERSSTATUSACTIVATED || member.Participation != "" {
					participantsForEmail = append(participantsForEmail, member)
				}
			}
			// Send email to all admins
			for _, member := range members {
				if member.Type == model.MEMBERSTYPEADMIN && member.Subscribed == 1 { // Send the email
					// get eventDate as a string
					payload := mail.EmailSummaryPayload{Member: member, Event: event, Participants: participantsForEmail}
					if err := mail.SendSummaryEmail(ctx, payload); err != nil {
						common.Error("%v\n", err)
						failures += 1
						continue
					}
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
		case model.TypeForgotPassword:
			m := model.Member{UUID: notification.ObjectUUID}
			err := m.Get(ctx)
			if err != nil {
				common.Debug("Error getting member for reset password: %s", err.Error())
				notification.Delivered = model.NotificationDeliveryFailure
				notification.UpdateNotificationStatus(ctx)
				continue
			}
			if common.GetConfigBool("smtp.enabled") {
				// Get a token to create credentials
				resetCredentialsToken, err := ResetCredentialsToken(ctx, m.UUID, m.Email, common.GetConfigInt("jwt.reset_ttl_minutes"))
				if err != nil {
					common.Debug("Error creating token for reset password: %s", err.Error())
					notification.Delivered = model.NotificationDeliveryFailure
					notification.UpdateNotificationStatus(ctx)
					continue
				}
				credentials := model.Credentials{UUID: m.UUID}
				err = credentials.GetCredentialsByUUID(ctx)
				if err != nil && err != sql.ErrNoRows {
					common.Debug("Error getting current credentials for reset password: %s", err.Error())
					notification.Delivered = model.NotificationDeliveryFailure
					notification.UpdateNotificationStatus(ctx)
					continue
				}
				payload := mail.EmailForgotPasswordPayload{Member: m, Token: resetCredentialsToken, Credentials: credentials}
				if err := mail.SendForgotPasswordEmail(ctx, payload); err != nil {
					common.Debug("Error sending email for reset password: %s", err.Error())
					notification.Delivered = model.NotificationDeliveryFailure
					notification.UpdateNotificationStatus(ctx)
					continue
				}
			}
			notification.Delivered = model.NotificationDeliverySuccess
			notification.UpdateNotificationStatus(ctx)
		case model.TypeEventDeleted:
			// Get All members
			m := model.Member{}
			members, err := m.GetAll(ctx, []string{model.MEMBERSSTATUSACTIVATED, model.MEMBERSSTATUSPAUSED}, []string{})
			if err != nil {
				// Cannot get the members, complete failure
				common.Error("Error getting members: %v\n", err)
				notification.Delivered = model.NotificationDeliveryFailure
				notification.UpdateNotificationStatus(ctx)
				continue
			}
			failures := 0
			for _, member := range members {
				// Send the email
				if member.Subscribed == 1 {

					var payload mail.EmailDeletedEventPayload
					if err := json.Unmarshal(notification.Payload, &payload); err != nil {
						common.Error("%v\n", err)
						failures += 1
						continue
					}
					payload.Member = member

					if err := mail.SendDeletedEventEmail(ctx, payload); err != nil {
						common.Error("%v\n", err)
						failures += 1
						continue
					}
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
		case model.TypeEventModified:
			// Get All members
			m := model.Member{}
			members, err := m.GetAll(ctx, []string{model.MEMBERSSTATUSACTIVATED, model.MEMBERSSTATUSPAUSED}, []string{})
			if err != nil {
				// Cannot get the members, complete failure
				common.Error("%v\n", err)
				notification.Delivered = model.NotificationDeliveryFailure
				notification.UpdateNotificationStatus(ctx)
				continue
			}
			failures := 0
			for _, member := range members {
				// Send the email
				if member.Subscribed == 1 {
					var payload mail.EmailModifiedPayload
					if err := json.Unmarshal(notification.Payload, &payload); err != nil {
						common.Error("%v\n", err)
						failures += 1
						continue
					}
					payload.Member = member
					if err := mail.SendModifiedEventEmail(ctx, payload); err != nil {
						common.Error("%v\n", err)
						failures += 1
						continue
					}
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
		case model.TypeEventCreated:
			// Get All members
			m := model.Member{}
			members, err := m.GetAll(ctx, []string{model.MEMBERSSTATUSACTIVATED, model.MEMBERSSTATUSPAUSED}, []string{})
			if err != nil {
				// Cannot get the members, complete failure
				common.Error("%v\n", err)
				notification.Delivered = model.NotificationDeliveryFailure
				notification.UpdateNotificationStatus(ctx)
				continue
			}
			failures := 0
			for _, member := range members {
				// Send the email
				if member.Subscribed == 1 {
					var payload mail.EmailCreateEventPayload
					if err := json.Unmarshal(notification.Payload, &payload); err != nil {
						common.Error("%v\n", err)
						failures += 1
						continue
					}
					payload.Member = member
					if err := mail.SendCreateEventEmail(ctx, payload); err != nil {
						common.Error("%v\n", err)
						failures += 1
						continue
					}
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
	}
}

func generateEventsNotificationsReminder() {

	ctx, span := tracer.Start(context.Background(), "generateEventsNotificationsReminder")
	defer span.End()

	e := model.Event{}
	events, err := e.GetUpcomingEventsWithoutNotification(ctx, model.TypeUpcomingEvent)
	if err != nil {
		common.Error("Error generating event notifications.")
		return
	}
	n := model.Notification{NotificationType: model.TypeUpcomingEvent}
	for _, event := range events {
		if (event.StartDate - uint(time.Now().Unix())) < uint(common.GetConfigInt("reminder_time_before_event")) {
			n.ObjectUUID = event.UUID
			n.SendDate = int(time.Now().Unix())
			err = n.CreateNotification(ctx)
			if err != nil {
				common.Error("Error creating event notification for event: %v.", event.UUID)
			}
		} else {
			continue
		}
	}
}

func generateEventsNotificationsSummary() {

	ctx, span := tracer.Start(context.Background(), "generateEventsNotificationsSummary")
	defer span.End()

	e := model.Event{}
	events, err := e.GetUpcomingEventsWithoutNotification(ctx, model.TypeSummaryEvent)
	if err != nil {
		common.Error("Error generating event notifications.")
		return
	}
	n := model.Notification{NotificationType: model.TypeSummaryEvent}
	for _, event := range events {
		if (event.StartDate - uint(time.Now().Unix())) < uint(common.GetConfigInt("summary_time_before_event")) {
			n.ObjectUUID = event.UUID
			n.SendDate = int(time.Now().Unix())
			err = n.CreateNotification(ctx)
			if err != nil {
				common.Error("Error creating event notification for event: %v.", event.UUID)
			}
		} else {
			continue
		}
	}
}

func pauseAbsentMembers() {
	ctx, span := tracer.Start(context.Background(), "pauseMembers")
	defer span.End()

	m := model.Member{}
	members, err := m.GetAll(ctx, []string{model.MEMBERSSTATUSACTIVATED}, []string{})
	if err != nil {
		common.Error("%v\n", err)
	}
	for _, member := range members {
		// Get last participation
		lastEvent, err := member.GetMemberLastParticipation(ctx)
		if err != nil {
			common.Error("%v\n", err)
		}
		// A manual reactivation by an admin resets the inactivity counter, so we
		// keep the most recent of the last participated event and last_activity_date.
		lastActivityDate, err := member.GetLastActivityDate(ctx)
		if err != nil {
			common.Error("%v\n", err)
		}
		referenceDate := int64(lastEvent.StartDate)
		if lastActivityDate > referenceDate {
			referenceDate = lastActivityDate
		}
		if time.Now().Unix()-referenceDate > int64(common.GetConfigInt("inactive_delay_days"))*24*3600 {
			common.Debug("Setting member %v as %s", member, model.MEMBERSSTATUSPAUSED)
			member.SetStatus(ctx, model.MEMBERSSTATUSPAUSED)
		}
	}
}
