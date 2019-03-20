package controller

import (
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
		switch notificationType := notification.NotificationType; notificationType {
		// This is a user registration
		case model.TypeMemberRegistration:
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
