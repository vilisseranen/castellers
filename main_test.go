package main_test

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/vilisseranen/castellers"
	"github.com/vilisseranen/castellers/model"
)

var a main.App

const TEST_DB_NAME = "test_database.db"

func TestMain(m *testing.M) {
	a = main.App{}
	a.Initialize(TEST_DB_NAME)

	ensureTablesExist()

	code := m.Run()

	clearTables()

	os.Exit(code)
}

func TestGetNonExistentEvent(t *testing.T) {
	clearTables()

	req, _ := http.NewRequest("GET", "/events/deadbeef", nil)
	response := executeRequest(req)

	checkResponseCode(t, http.StatusNotFound, response.Code)

	var m map[string]string
	json.Unmarshal(response.Body.Bytes(), &m)
	if m["error"] != "Event not found" {
		t.Errorf("Expected the 'error' key of the response to be set to 'Event not found'. Got '%s'", m["error"])
	}
}

func TestCreateEvent(t *testing.T) {
	clearTables()
	addMember("deadbeef", "ian", "admin")

	payload := []byte(`{"name":"diada","startDate":1527894960, "endDate":1528046040}`)

	req, _ := http.NewRequest("POST", "/admins/deadbeef/events", bytes.NewBuffer(payload))
	response := executeRequest(req)

	checkResponseCode(t, http.StatusCreated, response.Code)

	var event model.Event
	json.Unmarshal(response.Body.Bytes(), &event)

	if event.Name != "diada" {
		t.Errorf("Expected event name to be 'diada'. Got '%v'", event.Name)
	}

	if event.StartDate != 1527894960 {
		t.Errorf("Expected event start date to be '1527894960'. Got '%v'", event.StartDate)
	}

	if event.EndDate != 1528046040 {
		t.Errorf("Expected event end date to be '1528046040'. Got '%v'", event.EndDate)
	}
}

func TestCreateEventNonAdmin(t *testing.T) {
	clearTables()
	addMember("deadbeef", "ian", "admin")

	payload := []byte(`{"name":"diada","startDate":"2018-06-01 23:16", "endDate":"2018-06-03 17:14"}`)

	req, _ := http.NewRequest("POST", "/admins/4b1d/events", bytes.NewBuffer(payload))
	response := executeRequest(req)

	checkResponseCode(t, http.StatusUnauthorized, response.Code)

	var m map[string]interface{}
	json.Unmarshal(response.Body.Bytes(), &m)

	if m["name"] != nil {
		t.Errorf("Expected event name to be ''. Got '%v'", m["name"])
	}

	if m["startDate"] != nil {
		t.Errorf("Expected event start date to be ''. Got '%v'", m["date"])
	}

	if m["endDate"] != nil {
		t.Errorf("Expected event end date to be '2018-06-03 17:14'. Got '%v'", m["date"])
	}
}

func TestCreateWeeklyEvent(t *testing.T) {
	clearTables()
	addMember("deadbeef", "ian", "admin")

	payload := []byte(`{"name":"diada","startDate":1529016300, "endDate":1529027100, "recurring": {"interval": "1w", "until": 1532645100}}`)

	req, _ := http.NewRequest("POST", "/admins/deadbeef/events", bytes.NewBuffer(payload))
	response := executeRequest(req)

	checkResponseCode(t, http.StatusCreated, response.Code)

	req, _ = http.NewRequest("GET", "/events?count=10&start=0", nil)
	response = executeRequest(req)

	checkResponseCode(t, http.StatusOK, response.Code)

	var events = make([]model.Event, 0)
	json.Unmarshal(response.Body.Bytes(), &events)

	if len(events) != 7 {
		t.Errorf("Expected to have %v events. Got '%v'", 7, len(events))
	}

	for i, event := range events {
		correctTimestamp := uint(1529016300 + i*(60*60*24*7))
		if event.Name != "diada" {
			t.Errorf("Expected event name to be '%v'. Got '%v'", "diada", event.Name)
		}
		if event.StartDate != correctTimestamp {
			t.Errorf("Expected event %v start date to be '%v'. Got '%v'", i, correctTimestamp, event.StartDate)
		}
	}
}

func TestCreateDailyEvent(t *testing.T) {
	clearTables()
	addMember("deadbeef", "ian", "admin")

	payload := []byte(`{"name":"diada","startDate":1529157600, "endDate":1529193600, "recurring": {"interval": "1d", "until": 1529244000}}`)

	req, _ := http.NewRequest("POST", "/admins/deadbeef/events", bytes.NewBuffer(payload))
	response := executeRequest(req)

	checkResponseCode(t, http.StatusCreated, response.Code)

	req, _ = http.NewRequest("GET", "/events?count=10&start=0", nil)
	response = executeRequest(req)

	checkResponseCode(t, http.StatusOK, response.Code)

	var events = make([]model.Event, 0)
	json.Unmarshal(response.Body.Bytes(), &events)

	if len(events) != 2 {
		t.Errorf("Expected to have %v events. Got '%v'", 2, len(events))
	}

	for i, event := range events {
		correctTimestamp := uint(1529157600 + i*(60*60*24))
		if event.Name != "diada" {
			t.Errorf("Expected event name to be '%v'. Got '%v'", "diada", event.Name)
		}
		if event.StartDate != correctTimestamp {
			t.Errorf("Expected event %v start date to be '%v'. Got '%v'", i, correctTimestamp, event.StartDate)
		}
	}
}

func TestCreateMember(t *testing.T) {
	clearTables()
	addMember("deadbeef", "ian", "admin")

	payload := []byte(`{"name":"clement", "extra":"Santi"}`)

	req, _ := http.NewRequest("POST", "/admins/deadbeef/members", bytes.NewBuffer(payload))
	response := executeRequest(req)

	checkResponseCode(t, http.StatusCreated, response.Code)

	var m map[string]interface{}
	json.Unmarshal(response.Body.Bytes(), &m)

	if m["name"] != "clement" {
		t.Errorf("Expected member name to be 'clement'. Got '%v'", m["name"])
	}

	if m["extra"] != "Santi" {
		t.Errorf("Expected extra to be 'Santi'. Got '%v'", m["extra"])
	}
}

func TestCreateMemberNoExtra(t *testing.T) {
	clearTables()
	addMember("deadbeef", "ian", "admin")

	payload := []byte(`{"name":"clement","roles": ["baix", "second"]}`)

	req, _ := http.NewRequest("POST", "/admins/deadbeef/members", bytes.NewBuffer(payload))
	response := executeRequest(req)

	checkResponseCode(t, http.StatusCreated, response.Code)

	var m map[string]interface{}
	json.Unmarshal(response.Body.Bytes(), &m)

	if m["name"] != "clement" {
		t.Errorf("Expected member name to be 'clement'. Got '%v'", m["name"])
	}

	if m["extra"] != "" {
		t.Errorf("Expected extra to be ''. Got '%v'", m["extra"])
	}
}

func TestGetEvent(t *testing.T) {
	clearTables()
	addEvent("deadbeef", "An event", 1527894960, 1528046040)

	req, _ := http.NewRequest("GET", "/events/deadbeef", nil)
	response := executeRequest(req)

	checkResponseCode(t, http.StatusOK, response.Code)

	var m model.Event
	json.Unmarshal(response.Body.Bytes(), &m)

	if m.Name != "An event" {
		t.Errorf("Expected event name to be 'An event'. Got '%v'", m.Name)
	}

	if m.StartDate != 1527894960 {
		t.Errorf("Expected event start date to be '1527894960'. Got '%v'", m.StartDate)
	}

	if m.EndDate != 1528046040 {
		t.Errorf("Expected event end date to be '1528046040'. Got '%v'", m.EndDate)
	}
}

func TestGetMember(t *testing.T) {
	clearTables()
	addMember("deadbeef", "Clément", "member")

	req, _ := http.NewRequest("GET", "/members/deadbeef", nil)
	response := executeRequest(req)

	checkResponseCode(t, http.StatusOK, response.Code)

	var m map[string]interface{}
	json.Unmarshal(response.Body.Bytes(), &m)

	if m["name"] != "Clément" {
		t.Errorf("Expected member name to be 'Clément'. Got '%v'", m["name"])
	}
}

func TestGetEvents(t *testing.T) {
	clearTables()
	addEvent("deadbeef", "An event", 1527894960, 1528046040)
	addEvent("deadfeed", "Another event", 1527994960, 1527996960)

	req, _ := http.NewRequest("GET", "/events?count=2&start=0", nil)
	response := executeRequest(req)

	checkResponseCode(t, http.StatusOK, response.Code)

	var m [2]model.Event
	json.Unmarshal(response.Body.Bytes(), &m)

	if m[0].Name != "An event" {
		t.Errorf("Expected event name to be 'An event'. Got '%v'", m[0].Name)
	}

	if m[0].StartDate != 1527894960 {
		t.Errorf("Expected event start date to be '1527894960'. Got '%v'", m[0].StartDate)
	}

	if m[0].EndDate != 1528046040 {
		t.Errorf("Expected event end date to be '1528046040'. Got '%v'", m[0].EndDate)
	}

	if m[1].Name != "Another event" {
		t.Errorf("Expected event name to be 'Another event'. Got '%v'", m[1].Name)
	}

	if m[1].StartDate != 1527994960 {
		t.Errorf("Expected event start date to be '1527994960'. Got '%v'", m[1].StartDate)
	}

	if m[1].EndDate != 1527996960 {
		t.Errorf("Expected event end date to be '1527996960'. Got '%v'", m[1].EndDate)
	}
}

func TestUpdateEvent(t *testing.T) {
	clearTables()
	addEvent("deadbeef", "An event", 1528048800, 1528059600)
	addMember("deadbeef", "ian", "admin")

	req, _ := http.NewRequest("GET", "/events/deadbeef", nil)

	response := executeRequest(req)
	var originalEvent model.Event
	json.Unmarshal(response.Body.Bytes(), &originalEvent)

	payload := []byte(`{"name":"test event - updated name","startDate":1528052400, "endDate":1528063200}`)

	req, _ = http.NewRequest("PUT", "/admins/deadbeef/events/deadbeef", bytes.NewBuffer(payload))
	response = executeRequest(req)

	checkResponseCode(t, http.StatusOK, response.Code)

	var m model.Event
	json.Unmarshal(response.Body.Bytes(), &m)

	if m.Name == originalEvent.Name {
		t.Errorf("Expected the name to change from '%v' to '%v'. Got '%v'", originalEvent.Name, "test event - updated name", m.Name)
	}

	if m.StartDate == originalEvent.StartDate {
		t.Errorf("Expected the price to change from '%v' to '%v'. Got '%v'", originalEvent.StartDate, "2018-06-03 19:00", m.StartDate)
	}
	if m.EndDate == originalEvent.EndDate {
		t.Errorf("Expected the price to change from '%v' to '%v'. Got '%v'", originalEvent.EndDate, "2018-06-03 22:00", m.EndDate)
	}
}

func TestDeleteEvent(t *testing.T) {
	clearTables()
	addEvent("deadbeef", "An event", 1528048800, 1528059600)
	addMember("deadbeef", "ian", "admin")

	req, _ := http.NewRequest("GET", "/events/deadbeef", nil)
	response := executeRequest(req)
	checkResponseCode(t, http.StatusOK, response.Code)

	req, _ = http.NewRequest("DELETE", "/admins/deadbeef/events/deadbeef", nil)
	response = executeRequest(req)
	checkResponseCode(t, http.StatusOK, response.Code)

	req, _ = http.NewRequest("GET", "/events/deadbeef", nil)
	response = executeRequest(req)
	checkResponseCode(t, http.StatusNotFound, response.Code)
}

func TestParticipateEvent(t *testing.T) {
	clearTables()
	addMember("deadbeef", "toto", "member")
	addEvent("deadbeef", "diada", 1528048800, 1528059600)

	payload := []byte(`{"answer":"yes"}`)

	req, _ := http.NewRequest("POST", "/events/deadbeef/members/deadbeef", bytes.NewBuffer(payload))
	response := executeRequest(req)

	checkResponseCode(t, http.StatusCreated, response.Code)

	var m map[string]interface{}
	json.Unmarshal(response.Body.Bytes(), &m)

	if m["answer"] != "yes" {
		t.Errorf("Expected answer to be 'yes'. Got '%v'", m["name"])
	}
}

func executeRequest(req *http.Request) *httptest.ResponseRecorder {
	rr := httptest.NewRecorder()
	a.Router.ServeHTTP(rr, req)

	return rr
}

func checkResponseCode(t *testing.T, expected, actual int) {
	if expected != actual {
		t.Errorf("Expected response code %d. Got %d\n", expected, actual)
	}
}

func addEvent(uuid, name string, startDate, endDate int) {
	db, err := sql.Open("sqlite3", TEST_DB_NAME)
	if err != nil {
		log.Fatal(err)
	}
	tx, err := db.Begin()
	if err != nil {
		log.Fatal(err)
	}
	stmt, err := tx.Prepare("INSERT INTO events(uuid, name, startDate, endDate) VALUES(?, ?, ?, ?);")
	if err != nil {
		log.Fatal(err)
	}
	defer stmt.Close()
	_, err = stmt.Exec(uuid, name, startDate, endDate)
	if err != nil {
		log.Fatal(err)
	}
	tx.Commit()
}

func addMember(uuid, name, member_type string) {
	db, err := sql.Open("sqlite3", TEST_DB_NAME)
	if err != nil {
		log.Fatal(err)
	}
	tx, err := db.Begin()

	if err != nil {
		log.Fatal(err)
	}
	stmt, err := tx.Prepare("INSERT INTO members(uuid, name, extra, type) VALUES(?, ?, ?, ?)")
	if err != nil {
		log.Fatal(err)
	}
	defer stmt.Close()
	_, err = stmt.Exec(uuid, name, "", member_type)
	if err != nil {
		log.Fatal(err)
	}
	tx.Commit()
}

func ensureTablesExist() {
	db, err := sql.Open("sqlite3", TEST_DB_NAME)
	if err != nil {
		log.Fatal(err)
	}
	db.Exec("DROP TABLE events")
	db.Exec("DROP TABLE members")
	db.Exec("DROP TABLE presences")
	db.Exec(model.EventsTableCreationQuery)
	db.Exec(model.MembersTableCreationQuery)
	db.Exec(model.ParticipationTableCreationQuery)
}

func clearTables() {
	db, err := sql.Open("sqlite3", TEST_DB_NAME)
	if err != nil {
		log.Fatal(err)
	}
	db.Exec("DELETE FROM events")
	db.Exec("UPDATE sqlite_sequence SET seq = 0 WHERE name = 'events'")
	db.Exec("DELETE FROM members")
	db.Exec("UPDATE sqlite_sequence SET seq = 0 WHERE name = 'members'")
	db.Exec("DELETE FROM participation")
}
