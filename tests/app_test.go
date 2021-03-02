package tests

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
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
	if err := h.checkResponseCode(http.StatusCreated, response.Code); err != nil {
		t.Error(err)
	}

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

	if err := h.checkResponseCode(http.StatusUnauthorized, response.Code); err != nil {
		t.Error(err)
	}
}

func TestNotInitialized(t *testing.T) {
	h.clearTables()

	req, _ := http.NewRequest("GET", "/api/initialize", nil)
	response := h.executeRequest(req)

	if err := h.checkResponseCode(http.StatusNoContent, response.Code); err != nil {
		t.Error(err)
	}
}

func TestVersion(t *testing.T) {

	b, err := ioutil.ReadFile("VERSION")
	if err != nil {
		t.Error(err.Error())
	}
	correctVersion := string(b)

	type version struct {
		Version string `json:"version"`
	}

	req, _ := http.NewRequest("GET", "/api/version", nil)
	response := h.executeRequest(req)

	var v version
	json.Unmarshal(response.Body.Bytes(), &v)

	if err := h.checkResponseCode(http.StatusOK, response.Code); err != nil {
		t.Error(err)
	}

	if v.Version != correctVersion {
		t.Errorf("Expected version to be '%s'. Got '%s'", correctVersion, v.Version)
	}
}
