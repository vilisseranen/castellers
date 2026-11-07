package tests

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"testing"

	"github.com/vilisseranen/castellers/controller"
	"github.com/vilisseranen/castellers/model"
)

func lastRegistrationPayload(t *testing.T) []byte {
	t.Helper()
	db, err := sql.Open("sqlite3", testDbName)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	var payload []byte
	if err := db.QueryRow("SELECT payload FROM notifications WHERE notificationType = ? ORDER BY id DESC LIMIT 1", model.TypeMemberRegistration).Scan(&payload); err != nil {
		t.Fatal(err)
	}
	return payload
}

// Personal data must not be copied in clear text into the notifications table.
func TestRegistrationNotificationStoresOnlyUUIDs(t *testing.T) {
	h.clearTables()
	accessToken := h.addAnAdmin()

	payload := []byte(`{
		"firstName":"Clément", "lastName":"Contini", "height":"180", "weight":"70",
		"extra":"Santi", "roles":["segon"], "type":"member",
		"email":"clement@example.com", "contact":"514-111-1111", "language":"fr", "subscribed":1}`)
	req, _ := http.NewRequest("POST", "/api/v1/members", bytes.NewBuffer(payload))
	req.Header.Add("Authorization", "Bearer "+accessToken)
	response := h.executeRequest(req)
	if err := h.checkResponseCode(http.StatusCreated, response.Code); err != nil {
		t.Fatal(err)
	}
	var created map[string]interface{}
	json.Unmarshal(response.Body.Bytes(), &created)

	stored := lastRegistrationPayload(t)
	for _, secret := range []string{"Clément", "Contini", "clement@example.com", "514-111-1111", "Santi", "Romà", "eric.ca"} {
		if bytes.Contains(stored, []byte(secret)) {
			t.Errorf("notification payload contains %q: %s", secret, stored)
		}
	}
	var p model.MemberRegistrationPayload
	if err := json.Unmarshal(stored, &p); err != nil {
		t.Fatal(err)
	}
	if p.MemberUUID != created["uuid"] || p.AuthorUUID != "deadfeed" {
		t.Errorf("unexpected payload %+v", p)
	}
}

// The scheduler loads the member and the author from the database when sending.
func TestRegistrationEmailPayloadLoadsMembers(t *testing.T) {
	h.clearTables()
	h.addAnAdmin()
	h.addMember("deadbeef", "Ramon", "Gerard", "", "", "", "segon", "member", "ramon@gerard.ca", "")

	n, err := model.NewMemberRegistrationNotification("deadbeef", "deadfeed")
	if err != nil {
		t.Fatal(err)
	}
	payload, err := controller.RegistrationEmailPayload(context.Background(), n)
	if err != nil {
		t.Fatal(err)
	}
	if payload.Member.Email != "ramon@gerard.ca" || payload.Member.FirstName != "Ramon" || payload.Author.FirstName != "Romà" {
		t.Errorf("members not loaded: member=%s %s, author=%s", payload.Member.FirstName, payload.Member.Email, payload.Author.FirstName)
	}

	missing, _ := model.NewMemberRegistrationNotification("0000", "deadfeed")
	if _, err := controller.RegistrationEmailPayload(context.Background(), missing); err == nil {
		t.Error("expected an error for an unknown member")
	}
}

// Migration 0.24.3 rewrites existing registration notifications to keep only UUIDs.
func TestMigrationStripsPersonalDataFromNotifications(t *testing.T) {
	h.clearTables()
	db, err := sql.Open("sqlite3", testDbName)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	legacy := `{"member":{"uuid":"m1","firstName":"Ramon","email":"ramon@gerard.ca","contact":"514"},"author":{"uuid":"a1","firstName":"Romà"},"token":""}`
	insert := "INSERT INTO notifications (notificationType, objectUUID, sendDate, delivered, payload) VALUES (?, ?, 0, ?, ?)"
	db.Exec(insert, model.TypeMemberRegistration, "m1", model.NotificationDeliverySuccess, []byte(legacy))
	db.Exec(insert, model.TypeMemberRegistration, "m2", model.NotificationNotDelivered, []byte("not json"))
	db.Exec(insert, model.TypeBadgeAwarded, "b1", model.NotificationNotDelivered, []byte(`{"badgeCode":"x","memberUuids":["m1"]}`))

	migration, err := os.ReadFile("sql/0.24.3.sql")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(string(migration)); err != nil {
		t.Fatal(err)
	}

	rows, err := db.Query("SELECT objectUUID, payload FROM notifications ORDER BY id")
	if err != nil {
		t.Fatal(err)
	}
	defer rows.Close()
	got := map[string]string{}
	for rows.Next() {
		var uuid string
		var payload []byte
		rows.Scan(&uuid, &payload)
		got[uuid] = string(payload)
	}
	want := map[string]string{
		"m1": `{"memberUuid":"m1","authorUuid":"a1"}`,
		"m2": "",
		"b1": `{"badgeCode":"x","memberUuids":["m1"]}`,
	}
	for uuid, w := range want {
		if got[uuid] != w {
			t.Errorf("%s: got %q, want %q", uuid, got[uuid], w)
		}
	}
}

// Formatting a member or a notification (e.g. in a log line) must not reveal personal data.
func TestMemberAndNotificationFormattingHidesPersonalData(t *testing.T) {
	m := model.Member{UUID: "deadbeef", FirstName: "Ramon", Email: "ramon@gerard.ca", Contact: "514"}
	n := model.Notification{ID: 1, NotificationType: model.TypeMemberRegistration, ObjectUUID: "deadbeef", Payload: []byte(`{"email":"ramon@gerard.ca"}`)}
	for _, s := range []string{fmt.Sprintf("%v", m), fmt.Sprintf("%+v", m), fmt.Sprintf("%s", &m), fmt.Sprintf("%v", n), fmt.Sprintf("%v", &n)} {
		if strings.Contains(s, "Ramon") || strings.Contains(s, "ramon@gerard.ca") || strings.Contains(s, "514") {
			t.Errorf("formatted value leaks personal data: %s", s)
		}
		if !strings.Contains(s, "deadbeef") {
			t.Errorf("formatted value should still show the UUID: %s", s)
		}
	}
}
