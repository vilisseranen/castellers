package tests

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"github.com/vilisseranen/castellers/controller"
	"github.com/vilisseranen/castellers/model"
)

func futureEventStart() int {
	return int(time.Now().Unix()) + 86400*7
}

func TestSendEventRemindersDefaultAudience(t *testing.T) {
	h.clearTables()
	accessToken := h.addAnAdmin()
	start := futureEventStart()
	h.addEvent("deadbeef", "diada", start, start+3600)

	payload := []byte(`{"audience":"default"}`)
	req, _ := http.NewRequest("POST", "/api/v1/events/deadbeef/reminders", bytes.NewBuffer(payload))
	req.Header.Add("Authorization", "Bearer "+accessToken)
	response := h.executeRequest(req)

	if err := h.checkResponseCode(http.StatusAccepted, response.Code); err != nil {
		t.Error(err)
	}

	var body map[string]interface{}
	json.Unmarshal(response.Body.Bytes(), &body)
	if body["message"] != "reminders queued" {
		t.Errorf("Expected message 'reminders queued', got %v", body["message"])
	}

	nType, _ := h.getLatestNotificationType()
	if nType != model.TypeManualEventReminder {
		t.Errorf("Expected notification type %s, got %s", model.TypeManualEventReminder, nType)
	}
}

func TestSendEventRemindersMembersAudience(t *testing.T) {
	h.clearTables()
	accessToken := h.addAnAdmin()
	h.addMember("aabbccdd", "Ada", "Lovelace", "", "", "", "baix", "member", "ada@test.ca", "")
	h.setMemberStatus("aabbccdd", model.MEMBERSSTATUSACTIVATED)
	h.setMemberSubscribed("aabbccdd", 1)
	start := futureEventStart()
	h.addEvent("deadbeef", "diada", start, start+3600)

	payload := []byte(`{"audience":"members","memberUuids":["aabbccdd"]}`)
	req, _ := http.NewRequest("POST", "/api/v1/events/deadbeef/reminders", bytes.NewBuffer(payload))
	req.Header.Add("Authorization", "Bearer "+accessToken)
	response := h.executeRequest(req)

	if err := h.checkResponseCode(http.StatusAccepted, response.Code); err != nil {
		t.Error(err)
	}
}

func TestSendEventRemindersMembersEmpty(t *testing.T) {
	h.clearTables()
	accessToken := h.addAnAdmin()
	start := futureEventStart()
	h.addEvent("deadbeef", "diada", start, start+3600)

	payload := []byte(`{"audience":"members","memberUuids":[]}`)
	req, _ := http.NewRequest("POST", "/api/v1/events/deadbeef/reminders", bytes.NewBuffer(payload))
	req.Header.Add("Authorization", "Bearer "+accessToken)
	response := h.executeRequest(req)

	if err := h.checkResponseCode(http.StatusBadRequest, response.Code); err != nil {
		t.Error(err)
	}
}

func TestSendEventRemindersPastEvent(t *testing.T) {
	h.clearTables()
	accessToken := h.addAnAdmin()
	past := int(time.Now().Unix()) - 3600
	h.addEvent("deadbeef", "diada", past, past+1800)

	payload := []byte(`{"audience":"default"}`)
	req, _ := http.NewRequest("POST", "/api/v1/events/deadbeef/reminders", bytes.NewBuffer(payload))
	req.Header.Add("Authorization", "Bearer "+accessToken)
	response := h.executeRequest(req)

	if err := h.checkResponseCode(http.StatusBadRequest, response.Code); err != nil {
		t.Error(err)
	}
}

func TestSendEventRemindersNonAdmin(t *testing.T) {
	h.clearTables()
	accessToken := h.addAMember()
	start := futureEventStart()
	h.addEvent("deadbeef", "diada", start, start+3600)

	payload := []byte(`{"audience":"default"}`)
	req, _ := http.NewRequest("POST", "/api/v1/events/deadbeef/reminders", bytes.NewBuffer(payload))
	req.Header.Add("Authorization", "Bearer "+accessToken)
	response := h.executeRequest(req)

	if err := h.checkResponseCode(http.StatusUnauthorized, response.Code); err != nil {
		t.Error(err)
	}
}

func TestSendEventRemindersNoAnswerActive(t *testing.T) {
	h.clearTables()
	accessToken := h.addAnAdmin()
	h.addMember("aabbccdd", "Ada", "Lovelace", "", "", "", "baix", "member", "ada@test.ca", "")
	h.setMemberStatus("aabbccdd", model.MEMBERSSTATUSACTIVATED)
	h.addMember("bbccddee", "Bob", "Builder", "", "", "", "baix", "member", "bob@test.ca", "")
	h.setMemberStatus("bbccddee", model.MEMBERSSTATUSACTIVATED)
	start := futureEventStart()
	h.addEvent("deadbeef", "diada", start, start+3600)
	h.addParticipation("bbccddee", "deadbeef", "yes")

	payload := []byte(`{"audience":"no_answer_active"}`)
	req, _ := http.NewRequest("POST", "/api/v1/events/deadbeef/reminders", bytes.NewBuffer(payload))
	req.Header.Add("Authorization", "Bearer "+accessToken)
	response := h.executeRequest(req)

	if err := h.checkResponseCode(http.StatusAccepted, response.Code); err != nil {
		t.Error(err)
	}

	controller.RunNotificationDeliveryOnce()
	delivered := h.getLatestNotificationDelivered()
	if delivered != model.NotificationDeliverySuccess {
		t.Errorf("Expected delivery success, got %d", delivered)
	}
}

func (test *TestHelper) setMemberStatus(uuid, status string) {
	db, err := sql.Open("sqlite3", testDbName)
	if err != nil {
		tFatal(err)
	}
	defer db.Close()
	_, err = db.Exec("UPDATE members SET status = ? WHERE uuid = ?", status, uuid)
	if err != nil {
		tFatal(err)
	}
}

func (test *TestHelper) setMemberSubscribed(uuid string, subscribed int) {
	db, err := sql.Open("sqlite3", testDbName)
	if err != nil {
		tFatal(err)
	}
	defer db.Close()
	_, err = db.Exec("UPDATE members SET subscribed = ? WHERE uuid = ?", subscribed, uuid)
	if err != nil {
		tFatal(err)
	}
}

func (test *TestHelper) addParticipation(memberUUID, eventUUID, answer string) {
	db, err := sql.Open("sqlite3", testDbName)
	if err != nil {
		tFatal(err)
	}
	defer db.Close()
	_, err = db.Exec(
		"INSERT INTO participation(member_uuid, event_uuid, answer, presence) VALUES(?, ?, ?, ?)",
		memberUUID, eventUUID, answer, "")
	if err != nil {
		tFatal(err)
	}
}

func (test *TestHelper) getLatestNotificationType() (string, error) {
	db, err := sql.Open("sqlite3", testDbName)
	if err != nil {
		return "", err
	}
	defer db.Close()
	var nType string
	err = db.QueryRow(
		"SELECT notificationType FROM notifications ORDER BY id DESC LIMIT 1",
	).Scan(&nType)
	return nType, err
}

func (test *TestHelper) getLatestNotificationDelivered() int {
	db, err := sql.Open("sqlite3", testDbName)
	if err != nil {
		tFatal(err)
	}
	defer db.Close()
	var delivered int
	err = db.QueryRow(
		"SELECT delivered FROM notifications ORDER BY id DESC LIMIT 1",
	).Scan(&delivered)
	if err != nil {
		tFatal(err)
	}
	return delivered
}

func tFatal(err error) {
	if err != nil {
		panic(err)
	}
}
