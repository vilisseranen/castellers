package controller

import (
	"database/sql"
	"sort"
	"time"

	"github.com/robfig/cron"

	"github.com/vilisseranen/castellers/common"
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
	s.cron.AddFunc("@every 10m", generateEventsNotificationsReminder)

	// Look for upcoming events and generate summary notifications
	s.cron.AddFunc("@every 10m", generateEventsNotificationsSummary)

	s.cron.Start()
}

func checkAndSendNotification() {
	// Get notifications
	n := model.Notification{}
	notificationsToSend, err := n.GetNotificationsReady()
	if err != nil {
		common.Error("%v\n", err)
	}
	// Check all notifications that are ready
	for _, notification := range notificationsToSend {
		notification.Delivered = model.NotificationDeliveryInProgress
		notification.UpdateNotificationStatus()
		switch notificationType := notification.NotificationType; notificationType {
		case model.TypeMemberRegistration:
			// This is a user registration
			m := model.Member{UUID: notification.ObjectUUID}
			a := model.Member{UUID: notification.AuthorUUID}
			err := m.Get()
			if err != nil {
				notification.Delivered = model.NotificationDeliveryFailure
				notification.UpdateNotificationStatus()
				continue
			}
			err = a.Get()
			if err != nil {
				notification.Delivered = model.NotificationDeliveryFailure
				notification.UpdateNotificationStatus()
				continue
			}
			// Send the email
			if common.GetConfigBool("smtp_enabled") {
				// Get a token to create credentials
				resetCredentialsToken, err := ResetCredentialsToken(m.UUID, 1440)
				if err != nil {
					notification.Delivered = model.NotificationDeliveryFailure
					notification.UpdateNotificationStatus()
					continue
				}
				loginLink := common.GetConfigString("domain") + "/reset?" +
					"t=" + resetCredentialsToken +
					"&a=activation"
				profileLink := common.GetConfigString("domain") + "/memberEdit/" + m.UUID
				if err := common.SendRegistrationEmail(m.Email, m.FirstName, m.Language, a.FirstName, a.Extra, loginLink, profileLink); err != nil {
					notification.Delivered = model.NotificationDeliveryFailure
					notification.UpdateNotificationStatus()
					continue
				}
			}
			notification.Delivered = model.NotificationDeliverySuccess
			notification.UpdateNotificationStatus()
		case model.TypeUpcomingEvent:
			// This is a reminder for an upcoming event
			event := model.Event{UUID: notification.ObjectUUID}
			err := event.Get()
			if err != nil {
				// Cannot get the event, complete failure
				common.Error("%v\n", err)
				notification.Delivered = model.NotificationDeliveryFailure
				notification.UpdateNotificationStatus()
				continue
			}
			if event.StartDate < uint(time.Now().Unix()) {
				// Event has begun or is finished, we don't send the notification
				common.Info("Event %v has already started.\n", event.UUID)
				notification.Delivered = model.NotificationTooLate
				notification.UpdateNotificationStatus()
				continue
			}
			// Get All members
			m := model.Member{}
			p := model.Participation{}
			members, err := m.GetAll()
			if err != nil {
				// Cannot get the members, complete failure
				common.Error("%v\n", err)
				notification.Delivered = model.NotificationDeliveryFailure
				notification.UpdateNotificationStatus()
				continue
			}
			failures := 0
			for _, member := range members {
				p.MemberUUID = member.UUID
				p.EventUUID = event.UUID
				err = p.GetParticipation()
				if err != nil {
					switch err {
					case sql.ErrNoRows:
						p.Answer = ""
					default:
						common.Error("%v\n", err)
						failures += 1
						continue
					}
				}
				// Send the email
				if member.Subscribed == 1 {
					loginLink := common.GetConfigString("domain") + "/login?" +
						"m=" + member.UUID +
						"&c=" + member.Code
					profileLink := loginLink + "&next=memberEdit/" + member.UUID
					participationLink := loginLink + "&next=events" +
						"&action=participateEvent" +
						"&objectUUID=" + event.UUID +
						"&payload="
					answer := "false"
					if p.Answer == common.AnswerYes || p.Answer == common.AnswerNo {
						answer = "true"
					}
					var location, err = time.LoadLocation("America/Montreal")
					if err != nil {
						common.Error("%v\n", err)
						failures += 1
						continue
					}
					eventDate := time.Unix(int64(event.StartDate), 0).In(location).Format("02-01-2006")
					// get eventDate as a string
					if err := common.SendReminderEmail(member.Email, member.FirstName, member.Language, participationLink, profileLink, answer, p.Answer, event.Name, eventDate); err != nil {
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
			notification.UpdateNotificationStatus()
		case model.TypeSummaryEvent:
			event := model.Event{UUID: notification.ObjectUUID}
			err := event.Get()
			if err != nil {
				// Cannot get the event, complete failure
				common.Error("%v\n", err)
				notification.Delivered = model.NotificationDeliveryFailure
				notification.UpdateNotificationStatus()
				continue
			}
			if event.StartDate < uint(time.Now().Unix()) {
				// Event has begun or is finished, we don't send the notification
				common.Info("Event %v has already started.\n", event.UUID)
				notification.Delivered = model.NotificationTooLate
				notification.UpdateNotificationStatus()
				continue
			}
			m := model.Member{}
			members, err := m.GetAll()
			if err != nil {
				// Cannot get the members, complete failure
				common.Error("%v\n", err)
				notification.Delivered = model.NotificationDeliveryFailure
				notification.UpdateNotificationStatus()
				continue
			}
			failures := 0
			// Get participation for all members
			for index, member := range members {
				p := model.Participation{EventUUID: notification.ObjectUUID, MemberUUID: member.UUID}
				if err := p.GetParticipation(); err != nil {
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
			}
			// Sort by FirstName then by Participation
			sort.Slice(members, func(i, j int) bool { return members[i].FirstName < members[j].FirstName })
			sort.Slice(members, func(i, j int) bool { return members[i].Participation > members[j].Participation })
			// Send email to all admins
			for _, member := range members {
				if member.Type == model.MemberTypeAdmin {
					// Send the email
					if member.Subscribed == 1 {
						loginLink := common.GetConfigString("domain") + "/login?" +
							"m=" + member.UUID +
							"&c=" + member.Code
						profileLink := loginLink + "&next=memberEdit/" + member.UUID
						var location, err = time.LoadLocation("America/Montreal")
						if err != nil {
							common.Error("%v\n", err)
							failures += 1
							continue
						}
						eventDate := time.Unix(int64(event.StartDate), 0).In(location).Format("02-01-2006")
						// get eventDate as a string
						// TO FIX
						if err := common.SendSummaryEmail(member.Email, member.FirstName, member.Language,
							profileLink, event.Name, eventDate, ""); err != nil {
							common.Error("%v\n", err)
							failures += 1
							continue
						}
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
			notification.UpdateNotificationStatus()
		case model.TypeForgotPassword:
			m := model.Member{UUID: notification.ObjectUUID}
			err := m.Get()
			if err != nil {
				notification.Delivered = model.NotificationDeliveryFailure
				notification.UpdateNotificationStatus()
				continue
			}
			if common.GetConfigBool("smtp_enabled") {
				// Get a token to create credentials
				resetCredentialsToken, err := ResetCredentialsToken(m.UUID, 60)
				if err != nil {
					notification.Delivered = model.NotificationDeliveryFailure
					notification.UpdateNotificationStatus()
					continue
				}
				credentials := model.Credentials{UUID: m.UUID}
				err = credentials.GetCredentialsByUUID()
				if err != nil {
					notification.Delivered = model.NotificationDeliveryFailure
					notification.UpdateNotificationStatus()
					continue
				}
				resetLink := common.GetConfigString("domain") + "/reset?" +
					"t=" + resetCredentialsToken +
					"&a=reset&u=" + credentials.Username
				profileLink := common.GetConfigString("domain") + "/memberEdit/" + m.UUID
				if err := common.SendForgotPasswordEmail(m.Email, m.FirstName, m.Language, resetLink, profileLink); err != nil {
					notification.Delivered = model.NotificationDeliveryFailure
					notification.UpdateNotificationStatus()
					continue
				}
			}
		}
	}
}

func generateEventsNotificationsReminder() {
	e := model.Event{}
	events, err := e.GetUpcomingEventsWithoutNotification(model.TypeUpcomingEvent)
	if err != nil {
		common.Error("Error generating event notifications.")
		return
	}
	n := model.Notification{NotificationType: model.TypeUpcomingEvent}
	for _, event := range events {
		if (event.StartDate - uint(time.Now().Unix())) < uint(common.GetConfigInt("reminder_time_before_event")) {
			n.AuthorUUID = "0"
			n.ObjectUUID = event.UUID
			n.SendDate = int(time.Now().Unix())
			err = n.CreateNotification()
			if err != nil {
				common.Error("Error creating event notification for event: %v.", event.UUID)
			}
		} else {
			continue
		}
	}
}

func generateEventsNotificationsSummary() {
	e := model.Event{}
	events, err := e.GetUpcomingEventsWithoutNotification(model.TypeSummaryEvent)
	if err != nil {
		common.Error("Error generating event notifications.")
		return
	}
	n := model.Notification{NotificationType: model.TypeSummaryEvent}
	for _, event := range events {
		if (event.StartDate - uint(time.Now().Unix())) < uint(common.GetConfigInt("summary_time_before_event")) {
			n.AuthorUUID = "0"
			n.ObjectUUID = event.UUID
			n.SendDate = int(time.Now().Unix())
			err = n.CreateNotification()
			if err != nil {
				common.Error("Error creating event notification for event: %v.", event.UUID)
			}
		} else {
			continue
		}
	}
}
