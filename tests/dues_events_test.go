package tests

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/vilisseranen/castellers/controller"
	"github.com/vilisseranen/castellers/model"
)

func (test *TestHelper) setPresence(t *testing.T, adminToken, eventUUID, memberUUID, presence string) {
	t.Helper()
	code, _ := test.duesRequest("POST", "/events/"+eventUUID+"/members/"+memberUUID, adminToken,
		[]byte(fmt.Sprintf(`{"presence":"%s"}`, presence)))
	if err := test.checkResponseCode(http.StatusCreated, code); err != nil {
		t.Fatal(err)
	}
}

func TestEventDuesCollection(t *testing.T) {
	h.clearTables()
	adminToken := h.addAnAdmin()
	uuids := h.createDuesSeason(t, adminToken)

	// An event at which the winter dues are collected.
	payload := fmt.Sprintf(`{"name":"pratique","startDate":1527894960,"endDate":1527902160,"type":"practice","duesCollectionUuid":"%s"}`, uuids["winter"])
	code, body := h.duesRequest("POST", "/events", adminToken, []byte(payload))
	if err := h.checkResponseCode(http.StatusCreated, code); err != nil {
		t.Fatal(err)
	}
	var event model.Event
	json.Unmarshal(body, &event)
	if event.DuesCollectionUUID != uuids["winter"] {
		t.Errorf("Expected the created event to point to the winter collection. Got '%s'", event.DuesCollectionUUID)
	}

	get := func() model.Event {
		code, body := h.duesRequest("GET", "/events/"+event.UUID, adminToken, nil)
		if err := h.checkResponseCode(http.StatusOK, code); err != nil {
			t.Fatal(err)
		}
		var e model.Event
		json.Unmarshal(body, &e)
		return e
	}
	if e := get(); e.DuesCollectionUUID != uuids["winter"] {
		t.Errorf("Expected the event to point to the winter collection. Got '%s'", e.DuesCollectionUUID)
	}

	// It shows in the list of events too.
	code, body = h.duesRequest("GET", "/events?page=-1", adminToken, nil)
	if err := h.checkResponseCode(http.StatusOK, code); err != nil {
		t.Fatal(err)
	}
	var events []model.Event
	json.Unmarshal(body, &events)
	if len(events) != 1 || events[0].DuesCollectionUUID != uuids["winter"] {
		t.Errorf("Expected one past event pointing to the winter collection. Got %+v", events)
	}

	// Changing the collection, then removing it.
	update := func(collectionUUID string) int {
		payload := fmt.Sprintf(`{"name":"pratique","startDate":1527894960,"endDate":1527902160,"type":"practice","duesCollectionUuid":"%s"}`, collectionUUID)
		code, _ := h.duesRequest("PUT", "/events/"+event.UUID, adminToken, []byte(payload))
		return code
	}
	if err := h.checkResponseCode(http.StatusOK, update(uuids["spring"])); err != nil {
		t.Fatal(err)
	}
	if e := get(); e.DuesCollectionUUID != uuids["spring"] {
		t.Errorf("Expected the event to point to the spring collection. Got '%s'", e.DuesCollectionUUID)
	}
	if err := h.checkResponseCode(http.StatusOK, update("")); err != nil {
		t.Fatal(err)
	}
	if e := get(); e.DuesCollectionUUID != "" {
		t.Errorf("Expected the event to point to no collection. Got '%s'", e.DuesCollectionUUID)
	}

	// An unknown collection is refused, on update and on creation.
	if err := h.checkResponseCode(http.StatusBadRequest, update("0123")); err != nil {
		t.Error(err)
	}
	code, _ = h.duesRequest("POST", "/events", adminToken,
		[]byte(`{"name":"pratique","startDate":1527894960,"endDate":1527902160,"type":"practice","duesCollectionUuid":"0123"}`))
	if err := h.checkResponseCode(http.StatusBadRequest, code); err != nil {
		t.Error(err)
	}

	// An event without collection still works.
	code, body = h.duesRequest("POST", "/events", adminToken,
		[]byte(`{"name":"diada","startDate":1527894960,"endDate":1527902160,"type":"presentation"}`))
	if err := h.checkResponseCode(http.StatusCreated, code); err != nil {
		t.Fatal(err)
	}
	json.Unmarshal(body, &event)
	if event.DuesCollectionUUID != "" {
		t.Errorf("Expected no collection. Got '%s'", event.DuesCollectionUUID)
	}
}

// Dues are collected once: only the first event of a series points to the
// collection.
func TestRecurringEventDuesCollection(t *testing.T) {
	h.clearTables()
	adminToken := h.addAnAdmin()
	uuids := h.createDuesSeason(t, adminToken)

	start := futureEventStart()
	payload := fmt.Sprintf(
		`{"name":"pratique","startDate":%d,"endDate":%d,"type":"practice","duesCollectionUuid":"%s","recurring":{"interval":"1w","until":%d}}`,
		start, start+7200, uuids["winter"], start+15*86400)
	code, _ := h.duesRequest("POST", "/events", adminToken, []byte(payload))
	if err := h.checkResponseCode(http.StatusCreated, code); err != nil {
		t.Fatal(err)
	}

	code, body := h.duesRequest("GET", "/events", adminToken, nil)
	if err := h.checkResponseCode(http.StatusOK, code); err != nil {
		t.Fatal(err)
	}
	var events []model.Event
	json.Unmarshal(body, &events)
	if len(events) != 3 {
		t.Fatalf("Expected 3 events. Got %d", len(events))
	}
	if events[0].DuesCollectionUUID != uuids["winter"] {
		t.Errorf("Expected the first event to point to the winter collection. Got '%s'", events[0].DuesCollectionUUID)
	}
	for _, e := range events[1:] {
		if e.DuesCollectionUUID != "" {
			t.Errorf("Expected the following events to point to no collection. Got '%s'", e.DuesCollectionUUID)
		}
	}
}

func TestDuesSettings(t *testing.T) {
	h.clearTables()
	adminToken := h.addAnAdmin()
	memberToken := h.addAMember()

	read := func(token string) string {
		code, body := h.duesRequest("GET", "/dues/settings", token, nil)
		if err := h.checkResponseCode(http.StatusOK, code); err != nil {
			t.Fatal(err)
		}
		var settings map[string]string
		json.Unmarshal(body, &settings)
		return settings["interacEmail"]
	}
	if email := read(memberToken); email != "" {
		t.Errorf("Expected no address at first. Got '%s'", email)
	}

	// A member cannot change the settings.
	code, _ := h.duesRequest("PUT", "/dues/settings", memberToken, []byte(`{"interacEmail":"tresorerie@colla.ca"}`))
	if err := h.checkResponseCode(http.StatusUnauthorized, code); err != nil {
		t.Error(err)
	}

	code, _ = h.duesRequest("PUT", "/dues/settings", adminToken, []byte(`{"interacEmail":" tresorerie@colla.ca "}`))
	if err := h.checkResponseCode(http.StatusOK, code); err != nil {
		t.Fatal(err)
	}
	if email := read(memberToken); email != "tresorerie@colla.ca" {
		t.Errorf("Expected 'tresorerie@colla.ca'. Got '%s'", email)
	}

	for _, invalid := range []string{"pas une adresse", "Trésorerie <tresorerie@colla.ca>", "a@b.ca, c@d.ca"} {
		payload, _ := json.Marshal(map[string]string{"interacEmail": invalid})
		code, _ := h.duesRequest("PUT", "/dues/settings", adminToken, payload)
		if err := h.checkResponseCode(http.StatusBadRequest, code); err != nil {
			t.Errorf("'%s': %v", invalid, err)
		}
	}
	if email := read(adminToken); email != "tresorerie@colla.ca" {
		t.Errorf("Expected the address to be kept. Got '%s'", email)
	}

	// An empty address removes it.
	code, _ = h.duesRequest("PUT", "/dues/settings", adminToken, []byte(`{"interacEmail":""}`))
	if err := h.checkResponseCode(http.StatusOK, code); err != nil {
		t.Fatal(err)
	}
	if email := read(memberToken); email != "" {
		t.Errorf("Expected the address to be removed. Got '%s'", email)
	}
}

func TestAttendance(t *testing.T) {
	h.clearTables()
	adminToken := h.addAnAdmin()
	memberToken := h.addAMember()
	h.addMember("cafe", "Clément", "Other", "", "", "", "", "member", "clement@other.ca", "")
	h.addCredentials("cafe", "other", "other")
	otherToken := h.login("other", "other")
	h.addEvent("e1", "pratique 1", 1700000000, 1700007200)
	h.addEvent("e2", "pratique 2", 1700600000, 1700607200)
	h.addEvent("e3", "pratique 3", 1701200000, 1701207200)

	read := func(token, memberUUID string) model.Attendance {
		code, body := h.duesRequest("GET", "/members/"+memberUUID+"/attendance", token, nil)
		if err := h.checkResponseCode(http.StatusOK, code); err != nil {
			t.Fatal(err)
		}
		var a model.Attendance
		json.Unmarshal(body, &a)
		return a
	}

	// Never came.
	if a := read(memberToken, "deadbeef"); a.Count != 0 || a.FirstVisit != 0 || a.MemberUUID != "deadbeef" {
		t.Errorf("Expected no attendance. Got %+v", a)
	}

	// Present at the second and third events, absent from the first one. An
	// answer without confirmed presence does not count.
	h.setPresence(t, adminToken, "e1", "deadbeef", "no")
	h.setPresence(t, adminToken, "e3", "deadbeef", "yes")
	h.setPresence(t, adminToken, "e2", "deadbeef", "yes")
	h.setPresence(t, adminToken, "e3", "cafe", "yes")
	h.duesRequest("POST", "/members/events/e1", otherToken, []byte(`{"answer":"yes"}`))

	if a := read(memberToken, "deadbeef"); a.Count != 2 || a.FirstVisit != 1700600000 {
		t.Errorf("Expected 2 events since 1700600000. Got %+v", a)
	}
	if a := read(adminToken, "cafe"); a.Count != 1 || a.FirstVisit != 1701200000 {
		t.Errorf("Expected 1 event since 1701200000. Got %+v", a)
	}

	// A member cannot read the attendance of somebody else.
	code, _ := h.duesRequest("GET", "/members/deadbeef/attendance", otherToken, nil)
	if err := h.checkResponseCode(http.StatusForbidden, code); err != nil {
		t.Error(err)
	}

	// Admins get everybody who came at least once.
	code, _ = h.duesRequest("GET", "/dues/attendance", memberToken, nil)
	if err := h.checkResponseCode(http.StatusUnauthorized, code); err != nil {
		t.Error(err)
	}
	code, body := h.duesRequest("GET", "/dues/attendance", adminToken, nil)
	if err := h.checkResponseCode(http.StatusOK, code); err != nil {
		t.Fatal(err)
	}
	var all []model.Attendance
	json.Unmarshal(body, &all)
	counts := map[string]int{}
	for _, a := range all {
		counts[a.MemberUUID] = a.Count
	}
	if len(all) != 2 || counts["deadbeef"] != 2 || counts["cafe"] != 1 {
		t.Errorf("Expected deadbeef: 2 and cafe: 1. Got %+v", all)
	}
}

// The reminder of an event mentions the dues only to those who still have to
// settle them.
func TestDuesReminderFor(t *testing.T) {
	h.clearTables()
	adminToken := h.addAnAdmin()
	h.addAMember()
	uuids := h.createDuesSeason(t, adminToken)
	ctx := context.Background()
	start := uint(time.Now().Unix()) + 86400

	member := model.Member{UUID: "deadbeef", Type: model.MEMBERSTYPEREGULAR}
	event := model.Event{UUID: "e1", StartDate: start, DuesCollectionUUID: uuids["summer"]}

	// Not settled: the reminder gives the kind and the amount.
	dues, err := controller.DuesReminderFor(ctx, event, member)
	if err != nil {
		t.Fatal(err)
	}
	if dues == nil {
		t.Fatal("Expected dues to remind")
	}
	if dues.Kind != "summer" || dues.Amount != 1000 || dues.InteracEmail != "" {
		t.Errorf("Unexpected dues: %+v", dues)
	}

	// With the Interac address once it is set.
	if code, _ := h.duesRequest("PUT", "/dues/settings", adminToken, []byte(`{"interacEmail":"tresorerie@colla.ca"}`)); code != http.StatusOK {
		t.Fatalf("Expected 200. Got %d", code)
	}
	if dues, _ := controller.DuesReminderFor(ctx, event, member); dues == nil || dues.InteracEmail != "tresorerie@colla.ca" {
		t.Errorf("Expected the Interac address. Got %+v", dues)
	}

	// Nothing for the canalla and guests, nor when no dues are collected.
	for _, memberType := range []string{model.MEMBERSTYPECANALLA, model.MEMBERSTYPEGUEST} {
		other := model.Member{UUID: "deadbeef", Type: memberType}
		if dues, _ := controller.DuesReminderFor(ctx, event, other); dues != nil {
			t.Errorf("Expected nothing for a member of type %s. Got %+v", memberType, dues)
		}
	}
	if dues, _ := controller.DuesReminderFor(ctx, model.Event{UUID: "e2", StartDate: start}, member); dues != nil {
		t.Errorf("Expected nothing for an event without collection. Got %+v", dues)
	}
	// Nor when the collection does not exist any more.
	if dues, err := controller.DuesReminderFor(ctx, model.Event{UUID: "e3", DuesCollectionUUID: "0123"}, member); dues != nil || err != nil {
		t.Errorf("Expected nothing for an unknown collection. Got %+v, %v", dues, err)
	}

	// Paid or exempt: nothing. Unticked: reminded again.
	for _, step := range []struct {
		action string
		remind bool
	}{
		{model.DuesActionPaid, false},
		{model.DuesActionUnpaid, true},
		{model.DuesActionExempt, false},
	} {
		if code, _ := h.recordDues(t, adminToken, "deadbeef", step.action, "", uuids["summer"]); code != http.StatusOK {
			t.Fatalf("Expected 200. Got %d", code)
		}
		dues, err := controller.DuesReminderFor(ctx, event, member)
		if err != nil {
			t.Fatal(err)
		}
		if (dues != nil) != step.remind {
			t.Errorf("After '%s', expected remind=%v. Got %+v", step.action, step.remind, dues)
		}
	}
}
