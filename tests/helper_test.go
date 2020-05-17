package tests

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/vilisseranen/castellers/app"
)

var a app.App
var h TestHelper

const testDbName = "test_database.db"

type TestHelper struct {
	app app.App
}

func TestMain(m *testing.M) {
	os.Chdir("..")
	h.app = app.App{}
	os.Setenv("APP_DB_NAME", "test_database.db")
	os.Setenv("APP_LOG_FILE", "castellers.log")
	os.Setenv("APP_SMTP_SERVER", "192.168.1.100:25")
	os.Setenv("APP_DEBUG", "true")
	os.Setenv("APP_KEY", "fsjKJWJIJIJndndokspfkshtgrfghggcf4q32324")
	os.Setenv("APP_KEY_SALT", "dtgftgft7hftgth")
	os.Setenv("APP_PASSWORD_PEPPER", "gkjsneisuefsi")

	h.removeExistingTables()
	h.app.Initialize()

	code := m.Run()

	h.clearTables()

	os.Exit(code)
}

func (test *TestHelper) executeRequest(req *http.Request) *httptest.ResponseRecorder {
	rr := httptest.NewRecorder()
	test.app.Router.ServeHTTP(rr, req)

	return rr
}

func (test *TestHelper) checkResponseCode(expected, actual int) error {
	if expected != actual {
		return fmt.Errorf("Expected response code %d. Got %d\n", expected, actual)
	}
	return nil
}

func (test *TestHelper) addEvent(uuid, name string, startDate, endDate int) {
	db, err := sql.Open("sqlite3", testDbName)
	if err != nil {
		log.Fatal(err)
	}
	tx, err := db.Begin()
	if err != nil {
		log.Fatal(err)
	}
	stmt, err := tx.Prepare("INSERT INTO events(uuid, name, startDate, endDate, type, description, locationName, lat, lng) VALUES(?, ?, ?, ?, ?, ?, ?, ?, ?);")
	if err != nil {
		log.Fatal(err)
	}
	defer stmt.Close()
	_, err = stmt.Exec(uuid, name, startDate, endDate, "presentation", "", "", 0.0, 0.0)
	if err != nil {
		log.Fatal(err)
	}
	tx.Commit()
}

func (test *TestHelper) addAMember() {
	test.addMember("deadbeef", "Ramon", "Gerard", "", "", "Cap de rengla", "segon,baix,terç", "member", "ramon@gerard.ca", "", "toto")
}

func (test *TestHelper) addAnAdmin() {
	test.addMember("deadfeed", "Romà", "Èric", "", "", "Cap de colla", "baix,second", "admin", "romà@eric.ca", "", "tutu")
}

func (test *TestHelper) addMember(uuid, firstName, lastName, height, weight, extra, roles, memberType, email, contact, code string) {
	db, err := sql.Open("sqlite3", testDbName)
	if err != nil {
		log.Fatal(err)
	}
	tx, err := db.Begin()

	if err != nil {
		log.Fatal(err)
	}
	stmt, err := tx.Prepare("INSERT INTO members(uuid, firstName, lastName, height, weight, roles, extra, type, email, contact, code) VALUES(?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)")
	if err != nil {
		log.Fatal(err)
	}
	defer stmt.Close()
	_, err = stmt.Exec(uuid, firstName, lastName, height, weight, roles, extra, memberType, email, contact, code)
	if err != nil {
		log.Fatal(err)
	}
	tx.Commit()
}

func (test *TestHelper) removeExistingTables() {
	db, err := sql.Open("sqlite3", testDbName)
	if err != nil {
		log.Fatal(err)
	}
	db.Exec("DROP TABLE IF EXISTS schema_version")
	db.Exec("DROP TABLE IF EXISTS events")
	db.Exec("DROP TABLE IF EXISTS members")
	db.Exec("DROP TABLE IF EXISTS participation")
	db.Exec("DROP TABLE IF EXISTS recurring_events")
	db.Exec("DROP TABLE IF EXISTS notifications")
}

func (test *TestHelper) clearTables() {
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
