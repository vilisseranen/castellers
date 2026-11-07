package model

import (
	"context"
	"database/sql"
	"encoding/json"
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
const TypeManualEventReminder = "manualEventReminder"
const TypeBadgeAwarded = "badgeAwarded"

// MemberRegistrationPayload is stored on memberRegistration notifications.
// It only holds UUIDs: the members are loaded when the email is sent, so no
// personal data is stored in clear text in the notifications table.
type MemberRegistrationPayload struct {
	MemberUUID string `json:"memberUuid"`
	AuthorUUID string `json:"authorUuid"`
}

// NewMemberRegistrationNotification returns a notification, ready to be created,
// that sends the registration email to memberUUID on behalf of authorUUID.
func NewMemberRegistrationNotification(memberUUID, authorUUID string) (Notification, error) {
	payload, err := json.Marshal(MemberRegistrationPayload{MemberUUID: memberUUID, AuthorUUID: authorUUID})
	if err != nil {
		return Notification{}, err
	}
	return Notification{
		NotificationType: TypeMemberRegistration,
		ObjectUUID:       memberUUID,
		SendDate:         int(time.Now().Unix()),
		Payload:          payload,
	}, nil
}

// BadgeAwardedPayload is stored on badgeAwarded notifications.
type BadgeAwardedPayload struct {
	BadgeCode   string   `json:"badgeCode"`
	MemberUUIDs []string `json:"memberUuids"`
}

const ManualReminderAudienceDefault = "default"
const ManualReminderAudienceNoAnswerActive = "no_answer_active"
const ManualReminderAudienceNoAnswerActivePaused = "no_answer_active_paused"
const ManualReminderAudienceMembers = "members"

type ManualReminderPayload struct {
	Audience    string   `json:"audience"`
	MemberUUIDs []string `json:"memberUuids,omitempty"`
	RequestedBy string   `json:"requestedBy"`
}

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

// String keeps the payload, which may hold personal data, out of logs.
func (n Notification) String() string {
	return fmt.Sprintf("Notification{%d %s %s}", n.ID, n.NotificationType, n.ObjectUUID)
}

func (n *Notification) CreateNotification(ctx context.Context) error {
	ctx, span := tracer.Start(ctx, "Notification.CreateNotification")
	defer span.End()

	stmt, err := db.PrepareContext(ctx, fmt.Sprintf(
		"INSERT INTO %s (notificationType, objectUUID, sendDate, payload) VALUES (?, ?, ?, ?)",
		notificationsTable))
	defer stmt.Close()
	if err != nil {
		common.Error(err.Error())
		common.Error("Notification %d (%s, object %s)", n.ID, n.NotificationType, n.ObjectUUID)
		return err
	}
	result, err := stmt.ExecContext(ctx,
		stringOrNull(n.NotificationType),
		stringOrNull(n.ObjectUUID),
		n.SendDate,
		n.Payload)
	if err != nil {
		common.Error(err.Error())
		common.Error("Notification %d (%s, object %s)", n.ID, n.NotificationType, n.ObjectUUID)
		return err
	}
	id, err := result.LastInsertId()
	if err != nil {
		common.Error(err.Error())
		return err
	}
	n.ID = int(id)
	return nil
}

func (n *Notification) GetNotificationsReady(ctx context.Context) ([]Notification, error) {
	ctx, span := tracer.Start(ctx, "Notification.GetNotificationsReady")
	defer span.End()

	now := time.Now().Unix()
	rows, err := db.QueryContext(ctx, fmt.Sprintf(
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

func (n *Notification) UpdateNotificationStatus(ctx context.Context) error {
	ctx, span := tracer.Start(ctx, "Notification.UpdateNotificationStatus")
	defer span.End()

	stmt, err := db.PrepareContext(ctx, fmt.Sprintf(
		"UPDATE %s SET delivered = ? WHERE id = ?",
		notificationsTable))
	defer stmt.Close()
	if err != nil {
		common.Error(err.Error())
		common.Error("Notification %d (%s, object %s)", n.ID, n.NotificationType, n.ObjectUUID)
		return err
	}
	_, err = stmt.ExecContext(ctx,
		n.Delivered,
		n.ID)
	if err != nil {
		common.Error(err.Error())
		common.Error("Notification %d (%s, object %s)", n.ID, n.NotificationType, n.ObjectUUID)
		return err
	}
	return err
}
