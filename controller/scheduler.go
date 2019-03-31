package controller

import (
	"database/sql"
	"fmt"
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
	s.cron.AddFunc("@every 10m", checkAndSendNotification)

	// Look for upcoming events and generate notifications
	s.cron.AddFunc("@every 10m", generateEventsNotifications)

	s.cron.Start()
}

func checkAndSendNotification() {
	// Get notifications
	n := model.Notification{}
	notificationsToSend, err := n.GetNotificationsReady()
	if err != nil {
		fmt.Printf("%v\n", err)
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
			if common.GetConfigBool("debug") == false { // Don't send email in debug
				loginLink := common.GetConfigString("domain") + "/#/login?" +
					"m=" + m.UUID +
					"&c=" + m.Code
				profileLink := loginLink + "&next=memberEdit/" + m.UUID
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
				fmt.Printf("%v\n", err)
				notification.Delivered = model.NotificationDeliveryFailure
				notification.UpdateNotificationStatus()
				continue
			}
			if event.StartDate < uint(time.Now().Unix()) {
				// Event has begun or is finished, we don't send the notification
				fmt.Printf("Event %v has already started.\n", event.UUID)
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
				fmt.Printf("%v\n", err)
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
						fmt.Printf("%v\n", err)
						failures += 1
						continue
					}
				}
				// Send the email
				if common.GetConfigBool("debug") == false { // Don't send email in debug
					loginLink := common.GetConfigString("domain") + "/#/login?" +
						"m=" + member.UUID +
						"&c=" + member.Code
					profileLink := loginLink + "&next=memberEdit/" + member.UUID
					participationLink := loginLink + "&next=practices" +
						"&action=participateEvent" +
						"&objectUUID=" + event.UUID +
						"&payload="
					answer := "false"
					if p.Answer == common.AnswerYes || p.Answer == common.AnswerNo {
						answer = "true"
					}
					var location, err = time.LoadLocation("America/Montreal")
					if err != nil {
						fmt.Printf("%v\n", err)
						failures += 1
						continue
					}
					eventDate := time.Unix(int64(event.StartDate), 0).In(location).Format("02-01-2006")
					// get eventDate as a string
					if err := common.SendReminderEmail(member.Email, member.FirstName, member.Language, participationLink, profileLink, answer, p.Answer, event.Name, eventDate); err != nil {
						fmt.Printf("%v\n", err)
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
		}
	}
}

func generateEventsNotifications() {
	e := model.Event{}
	events, err := e.GetUpcomingEventsWithoutNotification()
	if err != nil {
		fmt.Println("Error generating event notifications.")
		return
	}
	n := model.Notification{NotificationType: model.TypeUpcomingEvent}
	for _, event := range events {
		if (event.StartDate - uint(time.Now().Unix())) < uint(common.GetConfigInt("notification_time_before_event")) {
			n.AuthorUUID = "0"
			n.ObjectUUID = event.UUID
			n.SendDate = int(time.Now().Unix())
			err = n.CreateNotification()
			if err != nil {
				fmt.Printf("Error creating event notification for event: %v.", event.UUID)
			}
		} else {
			continue
		}
	}
}
