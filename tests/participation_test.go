package tests

import (
	"bytes"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/vilisseranen/castellers/model"
)

func TestParticipateEvent(t *testing.T) {
	h.clearTables()
	h.addAMember()
	h.addEvent("deadbeef", "diada", 1528048800, 1528059600)

	payload := []byte(`{"answer":"maybe"}`)

	req, _ := http.NewRequest("POST", "/api/events/deadbeef/members/deadbeef", bytes.NewBuffer(payload))
	req.Header.Add("X-Member-Code", "toto")
	response := h.executeRequest(req)

	if err := h.checkResponseCode(http.StatusCreated, response.Code); err != nil {
		t.Error(err)
	}

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

	req, _ := http.NewRequest("POST", "/api/admins/deadfeed/events/deadbeef/members/deadbeef", bytes.NewBuffer(payload))
	req.Header.Add("X-Member-Code", "tutu")
	response := h.executeRequest(req)

	if err := h.checkResponseCode(http.StatusCreated, response.Code); err != nil {
		t.Error(err)
	}

	var m map[string]interface{}
	json.Unmarshal(response.Body.Bytes(), &m)

	if m["presence"] != "yes" {
		t.Errorf("Expected presence to be 'yes'. Got '%v'", m["presence"])
	}
}

func TestGetParticipants(t *testing.T) {
	h.clearTables()
	h.addAnAdmin()
	h.addAMember()
	h.addEvent("deadbeef", "diada", 1528048800, 1528059600)

	payload := []byte(`{"answer":"yes"}`)

	req, _ := http.NewRequest("POST", "/api/events/deadbeef/members/deadbeef", bytes.NewBuffer(payload))
	req.Header.Add("X-Member-Code", "toto")
	response := h.executeRequest(req)

	if err := h.checkResponseCode(http.StatusCreated, response.Code); err != nil {
		t.Error(err)
	}

	payload = []byte(`{"answer":"no"}`)

	req, _ = http.NewRequest("POST", "/api/events/deadbeef/members/deadfeed", bytes.NewBuffer(payload))
	req.Header.Add("X-Member-Code", "tutu")
	response = h.executeRequest(req)

	if err := h.checkResponseCode(http.StatusCreated, response.Code); err != nil {
		t.Error(err)
	}

	req, _ = http.NewRequest("GET", "/api/admins/deadfeed/events/deadbeef/members", nil)
	req.Header.Add("X-Member-Code", "tutu")
	response = h.executeRequest(req)

	if err := h.checkResponseCode(http.StatusOK, response.Code); err != nil {
		t.Error(err)
	}

	var members = make([]model.Member, 0)
	json.Unmarshal(response.Body.Bytes(), &members)

	if len(members) != 2 {
		t.Errorf("Expected to have %v members. Got '%v'", 2, len(members))
	}

	for _, member := range members {
		if member.UUID == "deadbeef" && member.Participation != "yes" {
			t.Errorf("Expected member participation to be '%v'. Got '%v'", "yes", member.Participation)
		}
		if member.UUID == "deadfeed" && member.Participation != "no" {
			t.Errorf("Expected member participation to be '%v'. Got '%v'", "no", member.Participation)
		}
	}
}

func TestGetPresence(t *testing.T) {
	h.clearTables()
	h.addAnAdmin()
	h.addAMember()
	h.addEvent("deadbeef", "diada", 1528048800, 1528059600)

	payload := []byte(`{"presence":"yes"}`)

	req, _ := http.NewRequest("POST", "/api/admins/deadfeed/events/deadbeef/members/deadbeef", bytes.NewBuffer(payload))
	req.Header.Add("X-Member-Code", "tutu")
	response := h.executeRequest(req)

	if err := h.checkResponseCode(http.StatusCreated, response.Code); err != nil {
		t.Error(err)
	}

	req, _ = http.NewRequest("GET", "/api/admins/deadfeed/events/deadbeef/members", nil)
	req.Header.Add("X-Member-Code", "tutu")
	response = h.executeRequest(req)

	if err := h.checkResponseCode(http.StatusOK, response.Code); err != nil {
		t.Error(err)
	}

	var members = make([]model.Member, 0)
	json.Unmarshal(response.Body.Bytes(), &members)

	for _, member := range members {
		if member.UUID == "deadbeef" && member.Presence != "yes" {
			t.Errorf("Expected member presence to be '%v'. Got '%v'", "yes", member.Presence)
		}
	}
}

func TestPresenceWrongEvent(t *testing.T) {
	h.clearTables()
	h.addAnAdmin()
	h.addAMember()
	h.addEvent("deadbeef", "diada", 1528048800, 1528059600)

	payload := []byte(`{"presence":"yes"}`)

	req, _ := http.NewRequest("POST", "/api/admins/deadfeed/events/123/members/deadbeef", bytes.NewBuffer(payload))
	req.Header.Add("X-Member-Code", "tutu")
	response := h.executeRequest(req)

	if err := h.checkResponseCode(http.StatusBadRequest, response.Code); err != nil {
		t.Error(err)
	}

}

func TestPresenceWrongMember(t *testing.T) {
	h.clearTables()
	h.addAnAdmin()
	h.addAMember()
	h.addEvent("deadbeef", "diada", 1528048800, 1528059600)

	payload := []byte(`{"presence":"yes"}`)

	req, _ := http.NewRequest("POST", "/api/admins/deadfeed/events/deadbeef/members/123", bytes.NewBuffer(payload))
	req.Header.Add("X-Member-Code", "tutu")
	response := h.executeRequest(req)

	if err := h.checkResponseCode(http.StatusBadRequest, response.Code); err != nil {
		t.Error(err)
	}

}
