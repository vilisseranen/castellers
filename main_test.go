package main_test

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/satori/go.uuid"
	"github.com/vilisseranen/castellers"
)

var a main.App

func TestMain(m *testing.M) {
	a = main.App{}
	a.Initialize("test_database.db")

	ensureTablesExist()

	code := m.Run()

	clearTables()

	os.Exit(code)
}

func TestEmptyTable(t *testing.T) {
	clearTables()

	req, _ := http.NewRequest("GET", "/events", nil)
	response := executeRequest(req)

	checkResponseCode(t, http.StatusOK, response.Code)

	if body := response.Body.String(); body != "[]" {
		t.Errorf("Expected an empty array. Got %s", body)
	}
}

func TestGetNonExistentEvent(t *testing.T) {
	clearTables()

	req, _ := http.NewRequest("GET", "/event/deadbeef", nil)
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

	payload := []byte(`{"name":"diada","startDate":"2018-06-01 23:16", "endDate":"2018-06-03 17:14"}`)

	req, _ := http.NewRequest("POST", "/event", bytes.NewBuffer(payload))
	response := executeRequest(req)

	checkResponseCode(t, http.StatusCreated, response.Code)

	var m map[string]interface{}
	json.Unmarshal(response.Body.Bytes(), &m)

	if m["name"] != "diada" {
		t.Errorf("Expected event name to be 'diada'. Got '%v'", m["name"])
	}

	if m["startDate"] != "2018-06-01 23:16" {
		t.Errorf("Expected event date to be '2018-06-01 23:16'. Got '%v'", m["date"])
	}

	if m["endDate"] != "2018-06-03 17:14" {
		t.Errorf("Expected event date to be '2018-06-03 17:14'. Got '%v'", m["date"])
	}
}

func TestGetEvent(t *testing.T) {
	clearTables()
	addEvent("deadbeef", "An event", "2018-06-03 18:00", "2018-06-03 21:00")

	req, _ := http.NewRequest("GET", "/event/deadbeef", nil)
	response := executeRequest(req)

	checkResponseCode(t, http.StatusOK, response.Code)

	var m map[string]interface{}
	json.Unmarshal(response.Body.Bytes(), &m)

	if m["name"] != "An event" {
		t.Errorf("Expected event name to be 'An event'. Got '%v'", m["name"])
	}

	if m["startDate"] != "2018-06-03 18:00" {
		t.Errorf("Expected event start date to be '2018-06-01 23:16'. Got '%v'", m["date"])
	}

	if m["endDate"] != "2018-06-03 21:00" {
		t.Errorf("Expected event end date to be '2018-06-03 17:14'. Got '%v'", m["date"])
	}
}

func TestUpdateEvent(t *testing.T) {
	clearTables()
	addEvent("deadbeef", "An event", "2018-06-03 18:00", "2018-06-03 21:00")

	req, _ := http.NewRequest("GET", "/events/deadbeef", nil)

	response := executeRequest(req)
	var originalEvent map[string]interface{}
	json.Unmarshal(response.Body.Bytes(), &originalEvent)

	payload := []byte(`{"name":"test event - updated name","startDate":"2018-06-03 19:00", "endDate":"2018-06-03 22:00"}`)

	req, _ = http.NewRequest("PUT", "/event/deadbeef", bytes.NewBuffer(payload))
	response = executeRequest(req)

	checkResponseCode(t, http.StatusOK, response.Code)

	var m map[string]interface{}
	json.Unmarshal(response.Body.Bytes(), &m)

	if m["name"] == originalEvent["name"] {
		t.Errorf("Expected the name to change from '%v' to '%v'. Got '%v'", originalEvent["name"], "test event - updated name", m["name"])
	}

	if m["startDate"] == originalEvent["startDate"] {
		t.Errorf("Expected the price to change from '%v' to '%v'. Got '%v'", originalEvent["date"], "2018-06-03 19:00", m["startDate"])
	}
	if m["endDate"] == originalEvent["endDate"] {
		t.Errorf("Expected the price to change from '%v' to '%v'. Got '%v'", originalEvent["date"], "2018-06-03 22:00", m["enDate"])
	}
}

func TestDeleteEvent(t *testing.T) {
	clearTables()
	addEvent("deadbeef", "An event", "2018-06-03 18:00", "2018-06-03 21:00")

	req, _ := http.NewRequest("GET", "/event/deadbeef", nil)
	response := executeRequest(req)
	checkResponseCode(t, http.StatusOK, response.Code)

	req, _ = http.NewRequest("DELETE", "/event/deadbeef", nil)
	response = executeRequest(req)
	checkResponseCode(t, http.StatusOK, response.Code)

	/*
		req, _ = http.NewRequest("GET", "/event/deadbeef", nil)
		response = executeRequest(req)
		checkResponseCode(t, http.StatusNotFound, response.Code)
	*/
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

func addEvent(uuid, name, startDate, endDate string) {
	tx, err := a.DB.Begin()
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

func addAdmin() {
	tx, err := a.DB.Begin()

	if err != nil {
		log.Fatal(err)
	}
	stmt, err := tx.Prepare("INSERT INTO admins(uuid) VALUES(?)")
	if err != nil {
		log.Fatal(err)
	}
	defer stmt.Close()
	uuid, err := uuid.NewV4()
	if err != nil {
		log.Fatal(err)
	}
	_, err = stmt.Exec(uuid)
	if err != nil {
		log.Fatal(err)
	}
	tx.Commit()
}

func ensureTablesExist() {
	a.DB.Exec("DROP TABLE events")
	a.DB.Exec("DROP TABLE admins")
	a.DB.Exec(main.EventsTableCreationQuery)
	a.DB.Exec(main.AdminsTableCreationQuery)
}

func clearTables() {
	a.DB.Exec("DELETE FROM events")
	a.DB.Exec("UPDATE sqlite_sequence SET seq = 0 WHERE name = 'events'")
	a.DB.Exec("DELETE FROM admins")
	a.DB.Exec("UPDATE sqlite_sequence SET seq = 0 WHERE name = 'admins'")

}
