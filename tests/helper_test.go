package tests

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/alicebob/miniredis"
	"github.com/vilisseranen/castellers/app"
	"github.com/vilisseranen/castellers/common"
)

var a app.App
var h TestHelper

const testDbName = "test_database.db"

type TestHelper struct {
	app app.App
}

func TestMain(m *testing.M) {
	redis := miniredis.NewMiniRedis()
	err := redis.StartAddr(":6380")
	if err != nil {
		log.Fatalf("Cannot create mock redis: %v", err)
	}
	os.Chdir("..")
	h.app = app.App{}
	os.Setenv("APP_DB_NAME", "test_database.db")
	os.Setenv("APP_LOG_FILE", "castellers.log")
	os.Setenv("APP_SMTP_SERVER", "192.168.1.100:25")
	os.Setenv("APP_DEBUG", "true")
	os.Setenv("APP_KEY", "fsjKJWJIJIJndndokspfkshtgrfghggcf4q32324")
	os.Setenv("APP_KEY_SALT", "dtgftgft7hftgth")
	os.Setenv("APP_PASSWORD_PEPPER", "gkjsneisuefsi")
	os.Setenv("APP_REDIS_DSN", "localhost:6380")
	os.Setenv("APP_ACCESS_SECRET", "sefsefsefsefhftgdfs")
	os.Setenv("APP_REFRESHSECRET", "zsgrxdrgzdrgsfefsef")

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
		return fmt.Errorf("Expected response code %d. Got %d", expected, actual)
	}
	return nil
}

func (test *TestHelper) addEvent(uuid, name string, startDate, endDate int) {
	db, err := sql.Open("sqlite3", testDbName)
	if err != nil {
		common.Fatal(err.Error())
	}
	tx, err := db.Begin()
	if err != nil {
		common.Fatal(err.Error())
	}
	stmt, err := tx.Prepare("INSERT INTO events(uuid, name, startDate, endDate, type, description, locationName, lat, lng) VALUES(?, ?, ?, ?, ?, ?, ?, ?, ?);")
	if err != nil {
		common.Fatal(err.Error())
	}
	defer stmt.Close()
	_, err = stmt.Exec(uuid, name, startDate, endDate, "presentation", "", "", 0.0, 0.0)
	if err != nil {
		common.Fatal(err.Error())
	}
	tx.Commit()
}

func (test *TestHelper) addAMember() string {
	test.addMember("deadbeef", "Ramon", "Gerard", "", "", "Cap de rengla", "segon,baix,terç", "member", "ramon@gerard.ca", "", "toto")
	test.addCredentials("deadbeef", "member", "member")
	payload := []byte(`{
		"username":"member",
		"password":"member"}`)

	req, _ := http.NewRequest("POST", "/api/login", bytes.NewBuffer(payload))
	response := h.executeRequest(req)
	var t map[string]interface{}
	json.Unmarshal(response.Body.Bytes(), &t)
	return t["access_token"].(string)
}

func (test *TestHelper) addAnAdmin() string {
	test.addMember("deadfeed", "Romà", "Èric", "", "", "Cap de colla", "baix,second", "admin", "romà@eric.ca", "", "tutu")
	test.addCredentials("deadfeed", "admin", "admin")
	payload := []byte(`{
		"username":"admin",
		"password":"admin"}`)

	req, _ := http.NewRequest("POST", "/api/login", bytes.NewBuffer(payload))
	response := h.executeRequest(req)
	var t map[string]interface{}
	json.Unmarshal(response.Body.Bytes(), &t)
	return t["access_token"].(string)
}

func (test *TestHelper) addMember(uuid, firstName, lastName, height, weight, extra, roles, memberType, email, contact, code string) {
	db, err := sql.Open("sqlite3", testDbName)
	if err != nil {
		common.Fatal(err.Error())
	}
	tx, err := db.Begin()

	if err != nil {
		common.Fatal(err.Error())
	}
	stmt, err := tx.Prepare("INSERT INTO members(uuid, firstName, lastName, height, weight, roles, extra, type, email, contact, code) VALUES(?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)")
	if err != nil {
		common.Fatal(err.Error())
	}
	defer stmt.Close()
	_, err = stmt.Exec(
		uuid,
		common.Encrypt(firstName),
		common.Encrypt(lastName),
		common.Encrypt(height),
		common.Encrypt(weight),
		common.Encrypt(roles),
		common.Encrypt(extra),
		common.Encrypt(memberType),
		common.Encrypt(email),
		common.Encrypt(contact),
		code)
	if err != nil {
		common.Fatal(err.Error())
	}
	tx.Commit()
}

func (test *TestHelper) addCredentials(uuid, username, password string) {
	db, err := sql.Open("sqlite3", testDbName)
	if err != nil {
		common.Fatal(err.Error())
	}
	tx, err := db.Begin()

	if err != nil {
		common.Fatal(err.Error())
	}
	stmt, err := tx.Prepare("INSERT INTO members_credentials(uuid, username, password) VALUES(?, ?, ?)")
	if err != nil {
		common.Fatal(err.Error())
	}
	defer stmt.Close()
	encryptedPassword, err := common.GenerateFromPassword(password)
	if err != nil {
		common.Fatal(err.Error())
	}
	_, err = stmt.Exec(
		uuid,
		username,
		encryptedPassword)
	if err != nil {
		common.Fatal(err.Error())
	}
	tx.Commit()
}

func (test *TestHelper) removeExistingTables() {
	db, err := sql.Open("sqlite3", testDbName)
	if err != nil {
		common.Fatal(err.Error())
	}
	db.Exec("DROP TABLE IF EXISTS schema_version")
	db.Exec("DROP TABLE IF EXISTS events")
	db.Exec("DROP TABLE IF EXISTS members")
	db.Exec("DROP TABLE IF EXISTS participation")
	db.Exec("DROP TABLE IF EXISTS recurring_events")
	db.Exec("DROP TABLE IF EXISTS notifications")
	db.Exec("DROP TABLE IF EXISTS members_credentials")
}

func (test *TestHelper) clearTables() {
	db, err := sql.Open("sqlite3", testDbName)
	if err != nil {
		common.Fatal(err.Error())
	}
	db.Exec("DELETE FROM events")
	db.Exec("UPDATE sqlite_sequence SET seq = 0 WHERE name = 'events'")
	db.Exec("DELETE FROM members")
	db.Exec("DELETE FROM members_credentials")
	db.Exec("UPDATE sqlite_sequence SET seq = 0 WHERE name = 'members'")
	db.Exec("DELETE FROM participation")
}
