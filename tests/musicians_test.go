package tests

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/vilisseranen/castellers/common"
	"github.com/vilisseranen/castellers/controller"
	"github.com/vilisseranen/castellers/model"
)

func (test *TestHelper) setMemberMusician(uuid string, musician int) {
	db, err := sql.Open("sqlite3", testDbName)
	if err != nil {
		common.Fatal(err.Error())
	}
	defer db.Close()
	if _, err = db.Exec("UPDATE members SET musician = ? WHERE uuid = ?", musician, uuid); err != nil {
		common.Fatal(err.Error())
	}
}

func (test *TestHelper) setEventType(uuid, eventType string) {
	db, err := sql.Open("sqlite3", testDbName)
	if err != nil {
		common.Fatal(err.Error())
	}
	defer db.Close()
	if _, err = db.Exec("UPDATE events SET type = ? WHERE uuid = ?", eventType, uuid); err != nil {
		common.Fatal(err.Error())
	}
}

func (test *TestHelper) getMemberAsMap(t *testing.T, token, uuid string) map[string]interface{} {
	req, _ := http.NewRequest("GET", "/api/v1/members/"+uuid, nil)
	req.Header.Add("Authorization", "Bearer "+token)
	response := test.executeRequest(req)
	if err := test.checkResponseCode(http.StatusOK, response.Code); err != nil {
		t.Fatal(err)
	}
	var m map[string]interface{}
	json.Unmarshal(response.Body.Bytes(), &m)
	return m
}

func (test *TestHelper) putMember(t *testing.T, token, uuid string, m map[string]interface{}) {
	payload, err := json.Marshal(m)
	if err != nil {
		t.Fatal(err)
	}
	req, _ := http.NewRequest("PUT", "/api/v1/members/"+uuid, bytes.NewBuffer(payload))
	req.Header.Add("Authorization", "Bearer "+token)
	response := test.executeRequest(req)
	if err := test.checkResponseCode(http.StatusAccepted, response.Code); err != nil {
		t.Fatal(err)
	}
}

func TestCreateMusicEvent(t *testing.T) {
	h.clearTables()
	accessToken := h.addAnAdmin()

	payload := []byte(`{"name":"Répétition","startDate":1527894960, "endDate":1527898560, "type":"music"}`)
	req, _ := http.NewRequest("POST", "/api/v1/events", bytes.NewBuffer(payload))
	req.Header.Add("Authorization", "Bearer "+accessToken)
	response := h.executeRequest(req)
	if err := h.checkResponseCode(http.StatusCreated, response.Code); err != nil {
		t.Fatal(err)
	}
	var event model.Event
	json.Unmarshal(response.Body.Bytes(), &event)
	if event.Type != model.EVENTSTYPEMUSIC {
		t.Errorf("Expected event type to be 'music'. Got '%v'", event.Type)
	}
}

func TestGetEventsFilteredByType(t *testing.T) {
	h.clearTables()
	start := futureEventStart()
	h.addEvent("deadbeef", "diada", start, start+3600)
	h.addEvent("deadfeed", "entrainement", start+10, start+3600)
	h.setEventType("deadfeed", "practice")
	h.addEvent("deadface", "repetition", start+20, start+3600)
	h.setEventType("deadface", "music")

	names := func(query string) []string {
		req, _ := http.NewRequest("GET", "/api/v1/events"+query, nil)
		response := h.executeRequest(req)
		if err := h.checkResponseCode(http.StatusOK, response.Code); err != nil {
			t.Fatal(err)
		}
		var events []model.Event
		json.Unmarshal(response.Body.Bytes(), &events)
		list := []string{}
		for _, event := range events {
			list = append(list, event.Name)
		}
		return list
	}
	check := func(query string, expected ...string) {
		got := names(query)
		if len(got) != len(expected) {
			t.Errorf("%s: expected %v, got %v", query, expected, got)
			return
		}
		for i := range expected {
			if got[i] != expected[i] {
				t.Errorf("%s: expected %v, got %v", query, expected, got)
				return
			}
		}
	}

	// Without the parameter, nothing changes: all the events are listed.
	check("", "diada", "entrainement", "repetition")
	check("?type=presentation,social,practice", "diada", "entrainement")
	check("?type=presentation,social,music", "diada", "repetition")
	check("?type=music", "repetition")
	check("?type=social")
	// Unknown types are ignored.
	check("?type=unknown", "diada", "entrainement", "repetition")
	check("?type=unknown,music", "repetition")
	// The filter also applies to the pages.
	check("?type=practice,music&limit=1&page=1", "repetition")
}

func TestOnlyAdminSetsMusician(t *testing.T) {
	h.clearTables()
	memberToken := h.addAMember()
	adminToken := h.addAnAdmin()

	// A member cannot make themselves a musician.
	m := h.getMemberAsMap(t, memberToken, "deadbeef")
	if m["musician"] != float64(0) {
		t.Fatalf("Expected musician to be 0. Got '%v'", m["musician"])
	}
	m["musician"] = 1
	h.putMember(t, memberToken, "deadbeef", m)
	m = h.getMemberAsMap(t, memberToken, "deadbeef")
	if m["musician"] != float64(0) {
		t.Errorf("A member changed their musician flag: got '%v'", m["musician"])
	}

	// An admin can.
	m = h.getMemberAsMap(t, adminToken, "deadbeef")
	m["musician"] = 1
	h.putMember(t, adminToken, "deadbeef", m)
	m = h.getMemberAsMap(t, memberToken, "deadbeef")
	if m["musician"] != float64(1) {
		t.Errorf("Expected musician to be 1 after the admin set it. Got '%v'", m["musician"])
	}

	// The member keeps it when they update their own profile, whatever they send.
	m["musician"] = 0
	m["height"] = "180"
	h.putMember(t, memberToken, "deadbeef", m)
	m = h.getMemberAsMap(t, memberToken, "deadbeef")
	if m["musician"] != float64(1) {
		t.Errorf("A member removed their musician flag: got '%v'", m["musician"])
	}

	// The list of members tells the admins who is a musician.
	req, _ := http.NewRequest("GET", "/api/v1/members", nil)
	req.Header.Add("Authorization", "Bearer "+adminToken)
	response := h.executeRequest(req)
	var members []model.Member
	json.Unmarshal(response.Body.Bytes(), &members)
	found := false
	for _, member := range members {
		if member.UUID == "deadbeef" {
			found = true
			if member.Musician != 1 {
				t.Errorf("Expected the list to show the member as a musician. Got '%v'", member.Musician)
			}
		}
	}
	if !found {
		t.Error("Member not found in the list")
	}
}

func TestEventAudience(t *testing.T) {
	members := []model.Member{
		{UUID: "a", Musician: 1},
		{UUID: "b"},
		{UUID: "c", Musician: 1},
	}
	uuids := func(list []model.Member) string {
		s := ""
		for _, member := range list {
			s += member.UUID
		}
		return s
	}
	practice := model.Event{Type: "practice"}
	music := model.Event{Type: model.EVENTSTYPEMUSIC}

	if got := uuids(controller.EventAudience(practice, members)); got != "abc" {
		t.Errorf("A practice concerns everybody. Got '%s'", got)
	}
	if got := uuids(controller.EventAudience(music, members)); got != "ac" {
		t.Errorf("A rehearsal only concerns the musicians. Got '%s'", got)
	}
	if got := uuids(controller.ModifiedEventAudience(music, music, members)); got != "ac" {
		t.Errorf("A modified rehearsal only concerns the musicians. Got '%s'", got)
	}
	// An event that changes type concerned, or now concerns, everybody.
	if got := uuids(controller.ModifiedEventAudience(practice, music, members)); got != "abc" {
		t.Errorf("A practice turned into a rehearsal concerns everybody. Got '%s'", got)
	}
	if got := uuids(controller.ModifiedEventAudience(music, practice, members)); got != "abc" {
		t.Errorf("A rehearsal turned into a practice concerns everybody. Got '%s'", got)
	}
}

func TestMusicEventReminderDelivered(t *testing.T) {
	h.clearTables()
	accessToken := h.addAnAdmin()
	h.addMember("aabbccdd", "Ada", "Lovelace", "", "", "", "baix", "member", "ada@test.ca", "")
	h.setMemberStatus("aabbccdd", model.MEMBERSSTATUSACTIVATED)
	h.setMemberMusician("aabbccdd", 1)
	h.addMember("bbccddee", "Bob", "Builder", "", "", "", "baix", "member", "bob@test.ca", "")
	h.setMemberStatus("bbccddee", model.MEMBERSSTATUSACTIVATED)
	start := futureEventStart()
	h.addEvent("deadbeef", "repetition", start, start+3600)
	h.setEventType("deadbeef", "music")

	payload := []byte(`{"audience":"default"}`)
	req, _ := http.NewRequest("POST", "/api/v1/events/deadbeef/reminders", bytes.NewBuffer(payload))
	req.Header.Add("Authorization", "Bearer "+accessToken)
	response := h.executeRequest(req)
	if err := h.checkResponseCode(http.StatusAccepted, response.Code); err != nil {
		t.Fatal(err)
	}

	controller.RunNotificationDeliveryOnce()
	if delivered := h.getLatestNotificationDelivered(); delivered != model.NotificationDeliverySuccess {
		t.Errorf("Expected delivery success, got %d", delivered)
	}
}
