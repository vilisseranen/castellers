package tests

import (
	"bytes"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/vilisseranen/castellers/model"
)

func TestCreateEvent(t *testing.T) {
	h.clearTables()
	access_token := h.addAnAdmin()

	payload := []byte(`{"name":"diada","startDate":1527894960, "endDate":1528046040, "type":"presentation"}`)
	req, _ := http.NewRequest("POST", "/api/admins/deadfeed/events", bytes.NewBuffer(payload))

	req.Header.Add("Authorization", "Bearer "+access_token)
	response := h.executeRequest(req)

	if err := h.checkResponseCode(http.StatusCreated, response.Code); err != nil {
		t.Error(err)
	}

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
	if event.Type != "presentation" {
		t.Errorf("Expected event type to be 'presentation'. Got '%v'", event.Type)
	}

	req, _ = http.NewRequest("GET", "/api/events/"+event.UUID, nil)
	response = h.executeRequest(req)

	if err := h.checkResponseCode(http.StatusOK, response.Code); err != nil {
		t.Error(err)
	}

}

func TestCreateEventNoDate(t *testing.T) {
	h.clearTables()
	h.addAnAdmin()

	payload := []byte(`{"name":"diada","type":"presentation"}`)
	req, _ := http.NewRequest("POST", "/api/admins/deadfeed/events", bytes.NewBuffer(payload))

	req.Header.Add("X-Member-Code", "tutu")
	response := h.executeRequest(req)

	if err := h.checkResponseCode(http.StatusCreated, response.Code); err != nil {
		t.Error(err)
	}

	var event model.Event
	json.Unmarshal(response.Body.Bytes(), &event)

	if event.Name != "diada" {
		t.Errorf("Expected event name to be 'diada'. Got '%v'", event.Name)
	}
	if event.Type != "presentation" {
		t.Errorf("Expected event type to be 'presentation'. Got '%v'", event.Type)
	}
}

func TestGetNonExistentEvent(t *testing.T) {
	h.clearTables()

	req, _ := http.NewRequest("GET", "/api/events/deadbeef", nil)
	response := h.executeRequest(req)

	if err := h.checkResponseCode(http.StatusNotFound, response.Code); err != nil {
		t.Error(err)
	}
	var m map[string]string
	json.Unmarshal(response.Body.Bytes(), &m)
	if m["error"] != "Event not found" {
		t.Errorf("Expected the 'error' key of the response to be set to 'Event not found'. Got '%s'", m["error"])
	}
}

func TestCreateEventNonAdmin(t *testing.T) {
	h.clearTables()
	h.addAMember()

	payload := []byte(`{"name":"diada","startDate":"2018-06-01 23:16", "endDate":"2018-06-03 17:14"}`)

	req, _ := http.NewRequest("POST", "/api/admins/deadfeed/events", bytes.NewBuffer(payload))
	req.Header.Add("X-Member-Code", "tutu")
	response := h.executeRequest(req)

	if err := h.checkResponseCode(http.StatusUnauthorized, response.Code); err != nil {
		t.Error(err)
	}
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
	h.clearTables()
	h.addAnAdmin()

	payload := []byte(`{"name":"diada","startDate":1529016300, "endDate":1529027100, "recurring": {"interval": "1w", "until": 1532645100}, "type":"practice"}`)

	req, _ := http.NewRequest("POST", "/api/admins/deadfeed/events", bytes.NewBuffer(payload))
	req.Header.Add("X-Member-Code", "tutu")
	response := h.executeRequest(req)

	if err := h.checkResponseCode(http.StatusCreated, response.Code); err != nil {
		t.Error(err)
	}

	req, _ = http.NewRequest("GET", "/api/events?count=10&start=1", nil)
	response = h.executeRequest(req)

	if err := h.checkResponseCode(http.StatusOK, response.Code); err != nil {
		t.Error(err)
	}
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
		if event.Type != "practice" {
			t.Errorf("Expected event %v type to be '%v'. Got '%v'", i, "practice", event.Type)
		}
	}
}

func TestCreateDailyEvent(t *testing.T) {
	h.clearTables()
	h.addAnAdmin()

	payload := []byte(`{"name":"diada","startDate":1529157600, "endDate":1529193600, "recurring": {"interval": "1d", "until": 1529244000}, "type":"practice"}`)

	req, _ := http.NewRequest("POST", "/api/admins/deadfeed/events", bytes.NewBuffer(payload))
	req.Header.Add("X-Member-Code", "tutu")
	response := h.executeRequest(req)

	if err := h.checkResponseCode(http.StatusCreated, response.Code); err != nil {
		t.Error(err)
	}

	req, _ = http.NewRequest("GET", "/api/events?count=10&start=1", nil)
	response = h.executeRequest(req)

	if err := h.checkResponseCode(http.StatusOK, response.Code); err != nil {
		t.Error(err)
	}

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

func TestGetEvent(t *testing.T) {
	h.clearTables()
	h.addEvent("deadbeef", "An event", 1527894960, 1528046040)

	req, _ := http.NewRequest("GET", "/api/events/deadbeef", nil)
	response := h.executeRequest(req)

	if err := h.checkResponseCode(http.StatusOK, response.Code); err != nil {
		t.Error(err)
	}

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
	if m.Type != "presentation" {
		t.Errorf("Expected event type to be 'presentation'. Got '%v'", m.Type)
	}
}

func TestGetEvents(t *testing.T) {
	h.clearTables()
	h.addEvent("deadbeef", "An event", 1527894960, 1528046040)
	h.addEvent("deadfeed", "Another event", 1527994960, 1527996960)

	req, _ := http.NewRequest("GET", "/api/events?count=2&start=1", nil)
	response := h.executeRequest(req)

	if err := h.checkResponseCode(http.StatusOK, response.Code); err != nil {
		t.Error(err)
	}

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

	if m[0].Type != "presentation" {
		t.Errorf("Expected event type to be 'presentation'. Got '%v'", m[0].Type)
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

	if m[1].Type != "presentation" {
		t.Errorf("Expected event type to be 'presentation'. Got '%v'", m[1].Type)
	}
}

func TestUpdateEvent(t *testing.T) {
	h.clearTables()
	h.addEvent("deadbeef", "An event", 1528048800, 1528059600)
	h.addAnAdmin()

	req, _ := http.NewRequest("GET", "/api/events/deadbeef", nil)
	response := h.executeRequest(req)

	var originalEvent model.Event
	json.Unmarshal(response.Body.Bytes(), &originalEvent)

	payload := []byte(`{"name": "test event - updated name", "startDate":1579218314,"endDate":1579228214,"recurring":{"interval":"1w","until":0},"type":"practice","location":{"lat":45.50073714334654,"lng":-73.6241186484186},"locationName":"Brébeuf","description":"new description"}`)

	req, _ = http.NewRequest("PUT", "/api/admins/deadfeed/events/deadbeef", bytes.NewBuffer(payload))
	req.Header.Add("X-Member-Code", "tutu")
	response = h.executeRequest(req)

	// Make sure the update request is successful
	if err := h.checkResponseCode(http.StatusOK, response.Code); err != nil {
		t.Error(err)
	}

	// Make sure the event is returned properly
	req, _ = http.NewRequest("GET", "/api/events/deadbeef", nil)
	response = h.executeRequest(req)

	var m model.Event
	json.Unmarshal(response.Body.Bytes(), &m)

	if m.Name != "test event - updated name" {
		t.Errorf("Expected the name to change from '%v' to '%v'. Got '%v'", originalEvent.Name, "test event - updated name", m.Name)
	}
	if m.StartDate != 1579218314 {
		t.Errorf("Expected the price to change from '%v' to '%v'. Got '%v'", originalEvent.StartDate, "1579218314", m.StartDate)
	}
	if m.EndDate != 1579228214 {
		t.Errorf("Expected the price to change from '%v' to '%v'. Got '%v'", originalEvent.EndDate, "1579228214", m.EndDate)
	}
	if m.Type != "practice" {
		t.Errorf("Expected the type to change from '%v' to '%v'. Got '%v'", originalEvent.Type, "practice", m.Type)
	}
	if m.Description != "new description" {
		t.Errorf("Expected the description to change from '%v' to '%v'. Got '%v'", originalEvent.Description, "new description", m.Description)
	}
}

func TestDeleteEvent(t *testing.T) {
	h.clearTables()
	h.addEvent("deadbeef", "An event", 1528048800, 1528059600)
	h.addAnAdmin()

	req, _ := http.NewRequest("GET", "/api/events/deadbeef", nil)
	response := h.executeRequest(req)
	if err := h.checkResponseCode(http.StatusOK, response.Code); err != nil {
		t.Error(err)
	}

	req, _ = http.NewRequest("DELETE", "/api/admins/deadfeed/events/deadbeef", nil)
	req.Header.Add("X-Member-Code", "tutu")
	response = h.executeRequest(req)
	if err := h.checkResponseCode(http.StatusOK, response.Code); err != nil {
		t.Error(err)
	}

	req, _ = http.NewRequest("GET", "/api/events/deadbeef", nil)
	response = h.executeRequest(req)
	if err := h.checkResponseCode(http.StatusNotFound, response.Code); err != nil {
		t.Error(err)
	}
}

func TestCreateEventEndBeforeBeginning(t *testing.T) {
	h.clearTables()
	h.addAnAdmin()

	payload := []byte(`{"name":"diada","startDate":1528046040, "endDate":1527894960}`)
	req, _ := http.NewRequest("POST", "/api/admins/deadfeed/events", bytes.NewBuffer(payload))

	req.Header.Add("X-Member-Code", "tutu")
	response := h.executeRequest(req)

	if err := h.checkResponseCode(http.StatusBadRequest, response.Code); err != nil {
		t.Error(err)
	}
}

func TestCreateEventEmptyName(t *testing.T) {
	h.clearTables()
	h.addAnAdmin()

	payload := []byte(`{"name":"","startDate":1527894960, "endDate":1528046040}`)
	req, _ := http.NewRequest("POST", "/api/admins/deadfeed/events", bytes.NewBuffer(payload))

	req.Header.Add("X-Member-Code", "tutu")
	response := h.executeRequest(req)

	if err := h.checkResponseCode(http.StatusBadRequest, response.Code); err != nil {
		t.Error(err)
	}
}

func TestUpdateEventEndBeforeBeginning(t *testing.T) {
	h.clearTables()
	h.addEvent("deadbeef", "An event", 1528048800, 1528059600)
	h.addAnAdmin()

	payload := []byte(`{"name":"An event","startDate":1528052400, "endDate":1518063200}`)

	req, _ := http.NewRequest("PUT", "/api/admins/deadfeed/events/deadbeef", bytes.NewBuffer(payload))
	req.Header.Add("X-Member-Code", "tutu")
	response := h.executeRequest(req)

	if err := h.checkResponseCode(http.StatusBadRequest, response.Code); err != nil {
		t.Error(err)
	}
}

func TestCreateEventWithLocationAndDescription(t *testing.T) {
	h.clearTables()
	h.addAnAdmin()

	payload := []byte(`{"name":"diada", "type":"presentation", "locationName": "Brébeuf", "location": {"lat": 45.50073714334654, "lng": -73.6241186484186}, "description": "First event description"}`)
	req, _ := http.NewRequest("POST", "/api/admins/deadfeed/events", bytes.NewBuffer(payload))

	req.Header.Add("X-Member-Code", "tutu")
	response := h.executeRequest(req)

	if err := h.checkResponseCode(http.StatusCreated, response.Code); err != nil {
		t.Error(err)
	}

	var event model.Event
	json.Unmarshal(response.Body.Bytes(), &event)

	if event.Location.Lat != 45.50073714334654 {
		t.Errorf("Expected lat to be '45.50073714334654'. Got '%v'", event.Location.Lat)
	}

	if event.Location.Lng != -73.6241186484186 {
		t.Errorf("Expected lng to be '-73.6241186484186'. Got '%v'", event.Location.Lng)
	}

	if event.LocationName != "Brébeuf" {
		t.Errorf("Expected event description to be 'Brébeuf'. Got '%v'", event.LocationName)
	}

	if event.Description != "First event description" {
		t.Errorf("Expected description to be 'First event description'. Got '%v'", event.Description)
	}

	req, _ = http.NewRequest("GET", "/api/events/"+event.UUID, nil)
	response = h.executeRequest(req)

	if err := h.checkResponseCode(http.StatusOK, response.Code); err != nil {
		t.Error(err)
	}

}
