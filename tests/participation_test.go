package tests

import (
	"bytes"
	"encoding/json"
	"net/http"
	"testing"
)

func TestParticipateEvent(t *testing.T) {
	h.clearTables()
	h.addAMember()
	h.addEvent("deadbeef", "diada", 1528048800, 1528059600)

	payload := []byte(`{"answer":"maybe"}`)

	req, _ := http.NewRequest("POST", "/api/events/deadbeef/members/deadbeef", bytes.NewBuffer(payload))
	req.Header.Add("X-Member-Code", "toto")
	response := h.executeRequest(req)

	h.checkResponseCode(t, http.StatusCreated, response.Code)

	var m map[string]interface{}
	json.Unmarshal(response.Body.Bytes(), &m)

	if m["answer"] != "maybe" {
		t.Errorf("Expected answer to be 'maybe'. Got '%v'", m["answer"])
	}
}

func TestPresenceEvent(t *testing.T) {
	h.clearTables()
	h.addAMember()
	h.addAnAdmin()
	h.addEvent("deadbeef", "diada", 1528048800, 1528059600)

	payload := []byte(`{"presence":"yes"}`)

	req, _ := http.NewRequest("POST", "/api/admins/deadfeed/events/deadbeef/members/baada55", bytes.NewBuffer(payload))
	req.Header.Add("X-Member-Code", "tutu")
	response := h.executeRequest(req)

	h.checkResponseCode(t, http.StatusCreated, response.Code)

	var m map[string]interface{}
	json.Unmarshal(response.Body.Bytes(), &m)

	if m["presence"] != "yes" {
		t.Errorf("Expected presence to be 'yes'. Got '%v'", m["presence"])
	}
}

func TestGetParticipation(t *testing.T) {
	h.clearTables()
	h.addAMember()
	h.addEvent("deadbeef", "diada", 1528048800, 1528059600)

	payload := []byte(`{"answer":"yes"}`)

	req, _ := http.NewRequest("POST", "/api/events/deadbeef/members/deadbeef", bytes.NewBuffer(payload))
	req.Header.Add("X-Member-Code", "toto")
	response := h.executeRequest(req)

	req, _ = http.NewRequest("GET", "/api/events/deadbeef/members/deadbeef", nil)
	req.Header.Add("X-Member-Code", "toto")
	response = h.executeRequest(req)

	h.checkResponseCode(t, http.StatusOK, response.Code)

	var m map[string]interface{}
	json.Unmarshal(response.Body.Bytes(), &m)

	if m["answer"] != "yes" {
		t.Errorf("Expected answer to be 'yes'. Got '%v'", m["answer"])
	}
}

func TestGetNoParticipation(t *testing.T) {
	h.clearTables()
	h.addAMember()
	h.addEvent("deadbeef", "diada", 1528048800, 1528059600)

	req, _ := http.NewRequest("GET", "/api/events/deadbeef/members/deadbeef", nil)
	req.Header.Add("X-Member-Code", "toto")
	response := h.executeRequest(req)

	h.checkResponseCode(t, 204, response.Code)
}
