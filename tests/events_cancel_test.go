package tests

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"
	"testing"

	"github.com/vilisseranen/castellers/controller"
	"github.com/vilisseranen/castellers/model"
)

func (test *TestHelper) countNotifications(t *testing.T, notificationType string) int {
	db, err := sql.Open("sqlite3", testDbName)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	var count int
	if err := db.QueryRow("SELECT COUNT(*) FROM notifications WHERE notificationType = ?", notificationType).Scan(&count); err != nil {
		t.Fatal(err)
	}
	return count
}

func (test *TestHelper) post(url, token, body string) (int, []byte) {
	req, _ := http.NewRequest("POST", "/api/v1"+url, bytes.NewBufferString(body))
	if token != "" {
		req.Header.Add("Authorization", "Bearer "+token)
	}
	response := test.executeRequest(req)
	return response.Code, response.Body.Bytes()
}

func (test *TestHelper) getEvent(t *testing.T, uuid string) model.Event {
	req, _ := http.NewRequest("GET", "/api/v1/events/"+uuid, nil)
	response := test.executeRequest(req)
	if err := test.checkResponseCode(http.StatusOK, response.Code); err != nil {
		t.Fatal(err)
	}
	var event model.Event
	if err := json.Unmarshal(response.Body.Bytes(), &event); err != nil {
		t.Fatal(err)
	}
	return event
}

// A cancelled event is still there, shown as cancelled with the reason, and
// the members are told, unless the admin asks not to.
func TestCancelEvent(t *testing.T) {
	h.clearTables()
	admin := h.addAnAdmin()
	start := futureEventStart()
	h.addEvent("ca01", "Told", start, start+7200)
	h.addEvent("ca02", "Silent", start, start+7200)
	h.addEvent("ca03", "Past", 1528048800, 1528059600)

	code, body := h.post("/events/ca01/cancel", admin, `{"reason":"weather","note":" Storm at 6pm "}`)
	if err := h.checkResponseCode(http.StatusOK, code); err != nil {
		t.Fatal(err)
	}
	var answer model.Event
	json.Unmarshal(body, &answer)
	if answer.Cancelled != 1 || answer.CancelReason != "weather" || answer.CancelNote != "Storm at 6pm" {
		t.Errorf("Unexpected answer to the cancellation: %+v", answer)
	}
	event := h.getEvent(t, "ca01")
	if event.Cancelled != 1 || event.CancelReason != "weather" || event.CancelNote != "Storm at 6pm" {
		t.Errorf("Expected the event to be cancelled for the weather. Got %+v", event)
	}
	if count := h.countNotifications(t, model.TypeEventCancelled); count != 1 {
		t.Errorf("Expected 1 notification after a cancellation. Got %d", count)
	}

	// Changing the reason does not tell the members again.
	code, _ = h.post("/events/ca01/cancel", admin, `{"reason":"other"}`)
	if err := h.checkResponseCode(http.StatusOK, code); err != nil {
		t.Error(err)
	}
	event = h.getEvent(t, "ca01")
	if event.Cancelled != 1 || event.CancelReason != "other" || event.CancelNote != "" {
		t.Errorf("Expected the reason to change. Got %+v", event)
	}
	if count := h.countNotifications(t, model.TypeEventCancelled); count != 1 {
		t.Errorf("Expected no new notification when the reason changes. Got %d", count)
	}

	code, _ = h.post("/events/ca02/cancel?notify=false", admin, `{"reason":"technical"}`)
	if err := h.checkResponseCode(http.StatusOK, code); err != nil {
		t.Error(err)
	}
	code, _ = h.post("/events/ca03/cancel", admin, `{"reason":"other"}`)
	if err := h.checkResponseCode(http.StatusOK, code); err != nil {
		t.Error(err)
	}
	if count := h.countNotifications(t, model.TypeEventCancelled); count != 1 {
		t.Errorf("Expected no notification with notify=false nor for a past event. Got %d", count)
	}

	// The cancelled events are listed, as cancelled.
	req, _ := http.NewRequest("GET", "/api/v1/events", nil)
	response := h.executeRequest(req)
	var events []model.Event
	json.Unmarshal(response.Body.Bytes(), &events)
	if len(events) != 2 {
		t.Fatalf("Expected the 2 upcoming events to be listed. Got %d", len(events))
	}
	for _, listed := range events {
		if listed.Cancelled != 1 || listed.CancelReason == "" {
			t.Errorf("Expected %s to be listed as cancelled. Got %+v", listed.UUID, listed)
		}
	}
}

func TestCancelEventInvalid(t *testing.T) {
	h.clearTables()
	admin := h.addAnAdmin()
	member := h.addAMember()
	start := futureEventStart()
	h.addEvent("ca01", "An event", start, start+7200)

	cases := []struct {
		name, url, token, body string
		code                   int
	}{
		{"unknown reason", "/events/ca01/cancel", admin, `{"reason":"lazy"}`, http.StatusBadRequest},
		{"no reason", "/events/ca01/cancel", admin, `{"note":"why not"}`, http.StatusBadRequest},
		{"not an admin", "/events/ca01/cancel", member, `{"reason":"weather"}`, http.StatusUnauthorized},
		{"unknown event", "/events/ca99/cancel", admin, `{"reason":"weather"}`, http.StatusNotFound},
		{"restore an event that is not cancelled", "/events/ca01/restore", admin, ``, http.StatusBadRequest},
	}
	for _, c := range cases {
		code, _ := h.post(c.url, c.token, c.body)
		if err := h.checkResponseCode(c.code, code); err != nil {
			t.Errorf("%s: %v", c.name, err)
		}
	}
	if event := h.getEvent(t, "ca01"); event.Cancelled != 0 {
		t.Errorf("Expected the event not to be cancelled. Got %+v", event)
	}
	if count := h.countNotifications(t, model.TypeEventCancelled); count != 0 {
		t.Errorf("Expected no notification. Got %d", count)
	}
}

func TestRestoreEvent(t *testing.T) {
	h.clearTables()
	admin := h.addAnAdmin()
	start := futureEventStart()
	h.addEvent("ca01", "An event", start, start+7200)

	h.post("/events/ca01/cancel", admin, `{"reason":"weather","note":"Storm"}`)
	code, _ := h.post("/events/ca01/restore", admin, ``)
	if err := h.checkResponseCode(http.StatusOK, code); err != nil {
		t.Fatal(err)
	}
	event := h.getEvent(t, "ca01")
	if event.Cancelled != 0 || event.CancelReason != "" || event.CancelNote != "" {
		t.Errorf("Expected the event to be restored. Got %+v", event)
	}
	if count := h.countNotifications(t, model.TypeEventRestored); count != 1 {
		t.Errorf("Expected 1 notification after restoring. Got %d", count)
	}

	h.post("/events/ca01/cancel?notify=false", admin, `{"reason":"weather"}`)
	code, _ = h.post("/events/ca01/restore?notify=false", admin, ``)
	if err := h.checkResponseCode(http.StatusOK, code); err != nil {
		t.Error(err)
	}
	if count := h.countNotifications(t, model.TypeEventRestored); count != 1 {
		t.Errorf("Expected no new notification with notify=false. Got %d", count)
	}
}

// Nobody answers for a cancelled event, and no reminder is sent for it.
func TestCancelledEventIsFrozen(t *testing.T) {
	h.clearTables()
	admin := h.addAnAdmin()
	member := h.addAMember()
	start := futureEventStart()
	h.addEvent("ca01", "Cancelled", start, start+7200)
	h.addEvent("ca02", "Maintained", start, start+7200)
	h.post("/events/ca01/cancel?notify=false", admin, `{"reason":"weather"}`)

	code, _ := h.post("/members/events/ca01", member, `{"answer":"yes"}`)
	if err := h.checkResponseCode(http.StatusBadRequest, code); err != nil {
		t.Errorf("answer: %v", err)
	}
	code, _ = h.post("/events/ca01/reminders", admin, `{"audience":"default"}`)
	if err := h.checkResponseCode(http.StatusBadRequest, code); err != nil {
		t.Errorf("manual reminder: %v", err)
	}

	// Neither the automatic reminder nor the summary is generated.
	for _, notificationType := range []string{model.TypeUpcomingEvent, model.TypeSummaryEvent} {
		e := model.Event{}
		events, err := e.GetUpcomingEventsWithoutNotification(context.Background(), notificationType)
		if err != nil {
			t.Fatal(err)
		}
		if len(events) != 1 || events[0].UUID != "ca02" {
			t.Errorf("%s: expected only the maintained event. Got %+v", notificationType, events)
		}
	}

	// A modification of a cancelled event keeps it cancelled, and is not announced.
	req, _ := http.NewRequest("PUT", "/api/v1/events/ca01", bytes.NewBufferString(
		`{"name":"Renamed","startDate":`+strconv.Itoa(start+60)+`,"endDate":`+strconv.Itoa(start+7200)+`,"type":"presentation"}`))
	req.Header.Add("Authorization", "Bearer "+admin)
	if err := h.checkResponseCode(http.StatusOK, h.executeRequest(req).Code); err != nil {
		t.Fatal(err)
	}
	event := h.getEvent(t, "ca01")
	if event.Name != "Renamed" || event.Cancelled != 1 || event.CancelReason != "weather" {
		t.Errorf("Expected the renamed event to stay cancelled. Got %+v", event)
	}
	if count := h.countNotifications(t, model.TypeEventModified); count != 0 {
		t.Errorf("Expected no notification for the change of a cancelled event. Got %d", count)
	}

	// Once restored, members answer again.
	h.post("/events/ca01/restore?notify=false", admin, ``)
	code, _ = h.post("/members/events/ca01", member, `{"answer":"yes"}`)
	if err := h.checkResponseCode(http.StatusCreated, code); err != nil {
		t.Errorf("answer after restoring: %v", err)
	}
}

// A notification that waits while the admin changes their mind is not sent.
func TestOutdatedCancellationNotification(t *testing.T) {
	h.clearTables()
	admin := h.addAnAdmin()
	start := futureEventStart()
	h.addEvent("ca01", "An event", start, start+7200)

	h.post("/events/ca01/cancel", admin, `{"reason":"weather"}`)
	h.post("/events/ca01/restore?notify=false", admin, ``)
	controller.RunNotificationDeliveryOnce()
	if delivered := h.getLatestNotificationDelivered(); delivered != model.NotificationTooLate {
		t.Errorf("Expected the outdated notification to be dropped. Got status %d", delivered)
	}
}
