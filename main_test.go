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

const testDbName = "test_database.db"

func TestMain(m *testing.M) {
	a = main.App{}
	os.Setenv("APP_DB_NAME", "test_database.db")
	os.Setenv("APP_LOG_FILE", "castellers.log")
	os.Setenv("APP_SMTP_SERVER", "192.168.1.100:25")
	os.Setenv("APP_DEBUG", "true")
	a.Initialize()

	ensureTablesExist()

	code := m.Run()

	clearTables()

	os.Exit(code)
}

func TestInitialize(t *testing.T) {
	clearTables()
	payload := []byte(`{"firstName":"Chimo", "lastName":"Anaïs", "extra":"Cap de colla", "roles": "second", "email": "vilisseranen@gmail.com"}`)

	req, _ := http.NewRequest("POST", "/api/initialize", bytes.NewBuffer(payload))
	response := executeRequest(req)

	// First admin should succeed
	checkResponseCode(t, http.StatusCreated, response.Code)

	var m map[string]interface{}
	json.Unmarshal(response.Body.Bytes(), &m)

	if m["firstName"] != "Chimo" {
		t.Errorf("Expected member first name to be 'Chimo'. Got '%v'", m["firstName"])
	}

	if m["lastName"] != "Anaïs" {
		t.Errorf("Expected member last name to be 'Anaïs'. Got '%v'", m["lastName"])
	}

	if m["type"] != "admin" {
		t.Errorf("Expected type to be 'admin'. Got '%v'", m["type"])
	}

	if m["uuid"] == "" {
		t.Errorf("A uuid must be returned. Got '%v'", m["uuid"])
	}

	// Second admin should fail
	payload = []byte(`{"name":"Clément", "extra":"Cap de rengles"}`)
	req, _ = http.NewRequest("POST", "/api/initialize", bytes.NewBuffer(payload))
	response = executeRequest(req)

	// First admin should succeed
	checkResponseCode(t, http.StatusUnauthorized, response.Code)
}

func TestNotInitialized(t *testing.T) {
	clearTables()

	req, _ := http.NewRequest("GET", "/api/initialize", nil)
	response := executeRequest(req)

	// First admin should succeed
	checkResponseCode(t, http.StatusNoContent, response.Code)
}

func TestGetNonExistentEvent(t *testing.T) {
	clearTables()

	req, _ := http.NewRequest("GET", "/api/events/deadbeef", nil)
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
	addAnAdmin()

	payload := []byte(`{"name":"diada","startDate":1527894960, "endDate":1528046040}`)

	req, _ := http.NewRequest("POST", "/api/admins/deadfeed/events", bytes.NewBuffer(payload))
	req.Header.Add("X-Member-Code", "tutu")
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
	addAMember()

	payload := []byte(`{"name":"diada","startDate":"2018-06-01 23:16", "endDate":"2018-06-03 17:14"}`)

	req, _ := http.NewRequest("POST", "/api/admins/deadfeed/events", bytes.NewBuffer(payload))
	req.Header.Add("X-Member-Code", "tutu")
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
	addAnAdmin()

	payload := []byte(`{"name":"diada","startDate":1529016300, "endDate":1529027100, "recurring": {"interval": "1w", "until": 1532645100}}`)

	req, _ := http.NewRequest("POST", "/api/admins/deadfeed/events", bytes.NewBuffer(payload))
	req.Header.Add("X-Member-Code", "tutu")
	response := executeRequest(req)

	checkResponseCode(t, http.StatusCreated, response.Code)

	req, _ = http.NewRequest("GET", "/api/events?count=10&start=0", nil)
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
	addAnAdmin()

	payload := []byte(`{"name":"diada","startDate":1529157600, "endDate":1529193600, "recurring": {"interval": "1d", "until": 1529244000}}`)

	req, _ := http.NewRequest("POST", "/api/admins/deadfeed/events", bytes.NewBuffer(payload))
	req.Header.Add("X-Member-Code", "tutu")
	response := executeRequest(req)

	checkResponseCode(t, http.StatusCreated, response.Code)

	req, _ = http.NewRequest("GET", "/api/events?count=10&start=0", nil)
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
	addAnAdmin()

	payload := []byte(`{
		"firstName":"Clément",
		"lastName": "Contini",
		"extra":"Santi",
		"roles": "segond,baix,terç",
		"type": "member",
		"email": "vilisseranen@gmail.com"}`)

	req, _ := http.NewRequest("POST", "/api/admins/deadfeed/members", bytes.NewBuffer(payload))
	req.Header.Add("X-Member-Code", "tutu")
	response := executeRequest(req)

	checkResponseCode(t, http.StatusCreated, response.Code)

	var m map[string]interface{}
	json.Unmarshal(response.Body.Bytes(), &m)

	if m["firstName"] != "Clément" {
		t.Errorf("Expected member first name to be 'Clément'. Got '%v'", m["firstName"])
	}

	if m["extra"] != "Santi" {
		t.Errorf("Expected extra to be 'Santi'. Got '%v'", m["extra"])
	}
}

func TestCreateMemberInvalidRole(t *testing.T) {
	clearTables()
	addAnAdmin()

	payload := []byte(`{
		"firstName":"Clément",
		"lastName": "Contini",
		"extra":"Santi",
		"roles": "segond,toto,baix,terç",
		"type": "member",
		"email": "vilisseranen@gmail.com"}`)

	req, _ := http.NewRequest("POST", "/api/admins/deadfeed/members", bytes.NewBuffer(payload))
	req.Header.Add("X-Member-Code", "tutu")
	response := executeRequest(req)

	checkResponseCode(t, http.StatusBadRequest, response.Code)
}

func TestCreateMemberNoExtra(t *testing.T) {
	clearTables()
	addAnAdmin()

	payload := []byte(`{
		"firstName":"Clément",
		"lastName": "Contini",
		"type": "member",
		"email": "vilisseranen@gmail.com"}`)

	req, _ := http.NewRequest("POST", "/api/admins/deadfeed/members", bytes.NewBuffer(payload))
	req.Header.Add("X-Member-Code", "tutu")
	response := executeRequest(req)

	checkResponseCode(t, http.StatusCreated, response.Code)

	var m map[string]interface{}
	json.Unmarshal(response.Body.Bytes(), &m)

	if m["firstName"] != "Clément" {
		t.Errorf("Expected member name to be 'Clément'. Got '%v'", m["firstName"])
	}

	if m["extra"] != "" {
		t.Errorf("Expected extra to be ''. Got '%v'", m["extra"])
	}

	if m["roles"] != "" {
		t.Errorf("Expected roles to be ''. Got '%v'", m["roles"])
	}
}

func TestGetEvent(t *testing.T) {
	clearTables()
	addEvent("deadbeef", "An event", 1527894960, 1528046040)

	req, _ := http.NewRequest("GET", "/api/events/deadbeef", nil)
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
	addAMember()

	req, _ := http.NewRequest("GET", "/api/members/deadbeef", nil)
	req.Header.Add("X-Member-Code", "toto")
	response := executeRequest(req)

	checkResponseCode(t, http.StatusOK, response.Code)

	var m map[string]interface{}
	json.Unmarshal(response.Body.Bytes(), &m)

	if m["firstName"] != "Ramon" {
		t.Errorf("Expected member name to be 'Ramon'. Got '%v'", m["firstName"])
	}
}

func TestGetEvents(t *testing.T) {
	clearTables()
	addEvent("deadbeef", "An event", 1527894960, 1528046040)
	addEvent("deadfeed", "Another event", 1527994960, 1527996960)

	req, _ := http.NewRequest("GET", "/api/events?count=2&start=0", nil)
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
	addAnAdmin()

	req, _ := http.NewRequest("GET", "/api/events/deadbeef", nil)
	response := executeRequest(req)

	var originalEvent model.Event
	json.Unmarshal(response.Body.Bytes(), &originalEvent)

	payload := []byte(`{"name":"test event - updated name","startDate":1528052400, "endDate":1528063200}`)

	req, _ = http.NewRequest("PUT", "/api/admins/deadfeed/events/deadbeef", bytes.NewBuffer(payload))
	req.Header.Add("X-Member-Code", "tutu")
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
	addAnAdmin()

	req, _ := http.NewRequest("GET", "/api/events/deadbeef", nil)
	response := executeRequest(req)
	checkResponseCode(t, http.StatusOK, response.Code)

	req, _ = http.NewRequest("DELETE", "/api/admins/deadfeed/events/deadbeef", nil)
	req.Header.Add("X-Member-Code", "tutu")
	response = executeRequest(req)
	checkResponseCode(t, http.StatusOK, response.Code)

	req, _ = http.NewRequest("GET", "/api/events/deadbeef", nil)
	response = executeRequest(req)
	checkResponseCode(t, http.StatusNotFound, response.Code)
}

func TestParticipateEvent(t *testing.T) {
	clearTables()
	addAMember()
	addEvent("deadbeef", "diada", 1528048800, 1528059600)

	payload := []byte(`{"answer":"maybe"}`)

	req, _ := http.NewRequest("POST", "/api/events/deadbeef/members/deadbeef", bytes.NewBuffer(payload))
	req.Header.Add("X-Member-Code", "toto")
	response := executeRequest(req)

	checkResponseCode(t, http.StatusCreated, response.Code)

	var m map[string]interface{}
	json.Unmarshal(response.Body.Bytes(), &m)

	if m["answer"] != "maybe" {
		t.Errorf("Expected answer to be 'maybe'. Got '%v'", m["answer"])
	}
}

func TestPresenceEvent(t *testing.T) {
	clearTables()
	addAMember()
	addAnAdmin()
	addEvent("deadbeef", "diada", 1528048800, 1528059600)

	payload := []byte(`{"presence":"yes"}`)

	req, _ := http.NewRequest("POST", "/api/admins/deadfeed/events/deadbeef/members/baada55", bytes.NewBuffer(payload))
	req.Header.Add("X-Member-Code", "tutu")
	response := executeRequest(req)

	checkResponseCode(t, http.StatusCreated, response.Code)

	var m map[string]interface{}
	json.Unmarshal(response.Body.Bytes(), &m)

	if m["presence"] != "yes" {
		t.Errorf("Expected presence to be 'yes'. Got '%v'", m["presence"])
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
	db, err := sql.Open("sqlite3", testDbName)
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

func addAMember() {
	addMember("deadbeef", "Ramon", "Gerard", "Cap de rengla", "segond, baix, terç", "member", "ramon@gerard.ca", "toto")
}

func addAnAdmin() {
	addMember("deadfeed", "Romà", "Èric", "Cap de colla", "baix, second", "admin", "romà@eric.ca", "tutu")
}

func addMember(uuid, firstName, lastName, extra, roles, member_type, email, code string) {
	db, err := sql.Open("sqlite3", testDbName)
	if err != nil {
		log.Fatal(err)
	}
	tx, err := db.Begin()

	if err != nil {
		log.Fatal(err)
	}
	stmt, err := tx.Prepare("INSERT INTO members(uuid, firstName, lastName, roles, extra, type, email, code) VALUES(?, ?, ?, ?, ?, ?, ?, ?)")
	if err != nil {
		log.Fatal(err)
	}
	defer stmt.Close()
	_, err = stmt.Exec(uuid, firstName, lastName, roles, extra, member_type, email, code)
	if err != nil {
		log.Fatal(err)
	}
	tx.Commit()
}

func ensureTablesExist() {
	db, err := sql.Open("sqlite3", testDbName)
	if err != nil {
		log.Fatal(err)
	}
	db.Exec("DROP TABLE events")
	db.Exec("DROP TABLE members")
	db.Exec("DROP TABLE participation")
	db.Exec(model.EventsTableCreationQuery)
	db.Exec(model.MembersTableCreationQuery)
	db.Exec(model.ParticipationTableCreationQuery)
}

func clearTables() {
	db, err := sql.Open("sqlite3", testDbName)
	if err != nil {
		log.Fatal(err)
	}
	db.Exec("DELETE FROM events")
	db.Exec("UPDATE sqlite_sequence SET seq = 0 WHERE name = 'events'")
	db.Exec("DELETE FROM members")
	db.Exec("UPDATE sqlite_sequence SET seq = 0 WHERE name = 'members'")
	db.Exec("DELETE FROM participation")
}
