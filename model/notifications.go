package model

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/vilisseranen/castellers/common"
)

const notificationsTable = "notifications"

const TypeMemberRegistration = "memberRegistration"
const TypeUpcomingEvent = "upcomingEvent"
const TypeSummaryEvent = "summaryEvent"
const TypeForgotPassword = "forgotPassword"
const TypeEventDeleted = "eventDeleted"
const TypeEventModified = "eventModified"
const TypeEventCreated = "eventCreated"

const NotificationNotDelivered = 0
const NotificationDeliverySuccess = 1
const NotificationDeliveryFailure = 2
const NotificationDeliveryPartialFailure = 3
const NotificationTooLate = 98
const NotificationDeliveryInProgress = 99

type Notification struct {
	ID               int
	NotificationType string
	ObjectUUID       string // TODO: remove this field
	SendDate         int
	Delivered        int
	Payload          []byte
}

func (n *Notification) CreateNotification() error {
	stmt, err := db.Prepare(fmt.Sprintf(
		"INSERT INTO %s (notificationType, objectUUID, sendDate, payload) VALUES (?, ?, ?, ?)",
		notificationsTable))
	defer stmt.Close()
	if err != nil {
		common.Error(err.Error())
		common.Error("%v\n", n)
		return err
	}
	_, err = stmt.Exec(
		stringOrNull(n.NotificationType),
		stringOrNull(n.ObjectUUID),
		n.SendDate,
		n.Payload)
	if err != nil {
		common.Error(err.Error())
		common.Error("%v\n", n)
		return err
	}
	return err
}

func (n *Notification) GetNotificationsReady() ([]Notification, error) {
	now := time.Now().Unix()
	rows, err := db.Query(fmt.Sprintf(
		"SELECT id, notificationType, objectUUID, sendDate, payload FROM %s WHERE sendDate <= ? AND delivered=0",
		notificationsTable), now)
	defer rows.Close()
	if err != nil {
		common.Fatal(err.Error())
	}
	notifications := []Notification{}
	for rows.Next() {
		var n Notification
		var objectUUID sql.NullString // to manage possible NULL fields
		if err = rows.Scan(&n.ID, &n.NotificationType, &objectUUID, &n.SendDate, &n.Payload); err != nil {
			return nil, err
		}
		n.ObjectUUID = nullToEmptyString(objectUUID)
		notifications = append(notifications, n)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	return notifications, nil
}

func (n *Notification) UpdateNotificationStatus() error {
	stmt, err := db.Prepare(fmt.Sprintf(
		"UPDATE %s SET delivered = ? WHERE id = ?",
		notificationsTable))
	defer stmt.Close()
	if err != nil {
		common.Error(err.Error())
		common.Error("%v\n", n)
		return err
	}
	_, err = stmt.Exec(
		n.Delivered,
		n.ID)
	if err != nil {
		common.Error(err.Error())
		common.Error("%v\n", n)
		return err
	}
	return err
}
