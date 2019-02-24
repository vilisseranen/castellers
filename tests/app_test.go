package tests

import (
	"bytes"
	"encoding/json"
	"net/http"
	"testing"
)

func TestInitialize(t *testing.T) {
	h.clearTables()
	payload := []byte(`{
		"firstName":"Chimo",
		"lastName":"Anaïs",
		"extra":"Cap de colla",
		"roles": ["second"],
		"email": "vilisseranen@gmail.com"}`)

	req, _ := http.NewRequest("POST", "/api/initialize", bytes.NewBuffer(payload))
	response := h.executeRequest(req)

	// First admin should succeed
	h.checkResponseCode(t, http.StatusCreated, response.Code)

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
	response = h.executeRequest(req)

	// First admin should succeed
	h.checkResponseCode(t, http.StatusUnauthorized, response.Code)
}

func TestNotInitialized(t *testing.T) {
	h.clearTables()

	req, _ := http.NewRequest("GET", "/api/initialize", nil)
	response := h.executeRequest(req)

	// First admin should succeed
	h.checkResponseCode(t, http.StatusNoContent, response.Code)
}
