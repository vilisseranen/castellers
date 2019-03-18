package model

import (
	"fmt"
	"log"
	"time"
)

const notificationsTable = "notifications"

const NotificationsTableCreationQuery = `CREATE TABLE IF NOT EXISTS notifications
(
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	notificationType TEXT NOT NULL,
	authorUUID TEXT NOT NULL,
	objectUUID TEXT NOT NULL, 
	sendDate INTEGER NOT NULL,
	delivered INTEGER NOT NULL DEFAULT 0
);`

const TypeMemberRegistration = "memberRegistration"

const NotificationNotDelivered = 0
const NotificationDeliverySuccess = 1
const NotificationDeliveryFailure = 2
const NotificationDeliveryPartialFailure = 3
const NotificationDeliveryInProgress = 99

type Notification struct {
	ID               int
	NotificationType string
	AuthorUUID       string
	ObjectUUID       string
	SendDate         int
	Delivered        int
}

func (n *Notification) CreateNotification() error {
	tx, err := db.Begin()
	if err != nil {
		fmt.Printf("%v\n", n)
		return err
	}
	stmt, err := tx.Prepare(fmt.Sprintf(
		"INSERT INTO %s (notificationType, authorUUID, objectUUID, sendDate) VALUES (?, ?, ?, ?)",
		notificationsTable))
	if err != nil {
		fmt.Printf("%v\n", n)
		return err
	}
	defer stmt.Close()
	_, err = stmt.Exec(
		stringOrNull(n.NotificationType),
		stringOrNull(n.AuthorUUID),
		stringOrNull(n.ObjectUUID),
		n.SendDate)
	if err != nil {
		fmt.Printf("%v\n", n)
		return err
	}
	tx.Commit()
	return err
}

func (n *Notification) GetNotificationsReady() ([]Notification, error) {
	now := time.Now().Unix()
	rows, err := db.Query(fmt.Sprintf(
		"SELECT id, notificationType, authorUUID, objectUUID, sendDate FROM %s WHERE sendDate <= ? AND delivered=0",
		notificationsTable), now)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()
	notifications := []Notification{}
	for rows.Next() {
		var n Notification
		if err = rows.Scan(&n.ID, &n.NotificationType, &n.AuthorUUID, &n.ObjectUUID, &n.SendDate); err != nil {
			return nil, err
		}
		notifications = append(notifications, n)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	return notifications, nil
}

func (n *Notification) UpdateNotificationStatus() error {
	tx, err := db.Begin()
	if err != nil {
		fmt.Printf("%v\n", n)
		return err
	}
	stmt, err := tx.Prepare(fmt.Sprintf(
		"UPDATE %s SET delivered = ? WHERE id = ?",
		notificationsTable))
	if err != nil {
		fmt.Printf("%v\n", n)
		return err
	}
	defer stmt.Close()
	_, err = stmt.Exec(
		n.Delivered,
		n.ID)
	if err != nil {
		fmt.Printf("%v\n", n)
		return err
	}
	tx.Commit()
	return err
}