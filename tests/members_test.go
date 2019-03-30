package tests

import (
	"bytes"
	"encoding/json"
	"net/http"
	"sort"
	"testing"

	"github.com/vilisseranen/castellers/model"
)

func TestCreateMember(t *testing.T) {
	h.clearTables()
	h.addAnAdmin()

	payload := []byte(`{
		"firstName":"Clément",
		"lastName": "Contini",
		"height": "180",
		"weight": "70",
		"extra":"Santi",
		"roles": ["segon","baix","terç"],
		"type": "member",
		"email": "vilisseranen@gmail.com",
		"language": "fr"}`)

	req, _ := http.NewRequest("POST", "/api/admins/deadfeed/members", bytes.NewBuffer(payload))
	req.Header.Add("X-Member-Code", "tutu")
	response := h.executeRequest(req)

	h.checkResponseCode(t, http.StatusCreated, response.Code)

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
	h.clearTables()
	h.addAnAdmin()

	payload := []byte(`{
		"firstName":"Clément",
		"lastName": "Contini",
		"extra":"Santi",
		"roles": "segond,toto,baix,terç",
		"type": "member",
		"email": "vilisseranen@gmail.com"}`)

	req, _ := http.NewRequest("POST", "/api/admins/deadfeed/members", bytes.NewBuffer(payload))
	req.Header.Add("X-Member-Code", "tutu")
	response := h.executeRequest(req)

	h.checkResponseCode(t, http.StatusBadRequest, response.Code)
}

func TestCreateMemberNoExtra(t *testing.T) {
	h.clearTables()
	h.addAnAdmin()

	payload := []byte(`{
		"firstName":"Clément",
		"lastName": "Contini",
		"type": "member",
		"email": "vilisseranen@gmail.com",
		"language": "cat"}`)

	req, _ := http.NewRequest("POST", "/api/admins/deadfeed/members", bytes.NewBuffer(payload))
	req.Header.Add("X-Member-Code", "tutu")
	response := h.executeRequest(req)

	h.checkResponseCode(t, http.StatusCreated, response.Code)

	var m map[string]interface{}
	json.Unmarshal(response.Body.Bytes(), &m)

	if m["firstName"] != "Clément" {
		t.Errorf("Expected member name to be 'Clément'. Got '%v'", m["firstName"])
	}

	if m["extra"] != "" {
		t.Errorf("Expected extra to be ''. Got '%v'", m["extra"])
	}

	if m["roles"] != nil {
		t.Errorf("Expected roles to be nil. Got '%v'", m["roles"])
	}

	var member_uuid string
	member_uuid = m["uuid"].(string)

	req, _ = http.NewRequest("GET", "/api/admins/deadfeed/members/"+member_uuid, nil)
	req.Header.Add("X-Member-Code", "tutu")
	response = h.executeRequest(req)
	h.checkResponseCode(t, http.StatusOK, response.Code)
}

func TestUpdateMember(t *testing.T) {
	h.clearTables()
	h.addAnAdmin()
	h.addAMember()

	req, _ := http.NewRequest("GET", "/api/admins/deadfeed/members/deadbeef", nil)
	req.Header.Add("X-Member-Code", "tutu")
	response := h.executeRequest(req)
	h.checkResponseCode(t, http.StatusOK, response.Code)

	var m map[string]interface{}
	json.Unmarshal(response.Body.Bytes(), &m)

	m["extra"] = "Cap de pinya"
	payload, error := json.Marshal(m)
	if error != nil {
		t.Errorf(error.Error())
	}

	req, _ = http.NewRequest("PUT", "/api/admins/deadfeed/members/deadbeef", bytes.NewBuffer(payload))
	req.Header.Add("X-Member-Code", "tutu")
	response = h.executeRequest(req)

	h.checkResponseCode(t, http.StatusAccepted, response.Code)

	json.Unmarshal(response.Body.Bytes(), &m)

	if m["extra"] != "Cap de pinya" {
		t.Errorf("Expected extra to be 'Cap de pinya'. Got '%v'", m["extra"])
	}
}

func TestPromoteSelf(t *testing.T) {
	h.clearTables()
	h.addAMember()

	req, _ := http.NewRequest("GET", "/api/members/deadbeef", nil)
	req.Header.Add("X-Member-Code", "toto")
	response := h.executeRequest(req)
	h.checkResponseCode(t, http.StatusOK, response.Code)

	var m map[string]interface{}
	json.Unmarshal(response.Body.Bytes(), &m)

	m["type"] = "admin"
	payload, error := json.Marshal(m)
	if error != nil {
		t.Errorf(error.Error())
	}

	req, _ = http.NewRequest("PUT", "/api/members/deadbeef", bytes.NewBuffer(payload))
	req.Header.Add("X-Member-Code", "toto")
	response = h.executeRequest(req)

	h.checkResponseCode(t, http.StatusForbidden, response.Code)
}

func TestPromoteByAdmin(t *testing.T) {
	h.clearTables()
	h.addAnAdmin()
	h.addAMember()

	req, _ := http.NewRequest("GET", "/api/admins/deadfeed/members/deadbeef", nil)
	req.Header.Add("X-Member-Code", "tutu")
	response := h.executeRequest(req)
	h.checkResponseCode(t, http.StatusOK, response.Code)

	var m map[string]interface{}
	json.Unmarshal(response.Body.Bytes(), &m)

	m["type"] = "admin"
	payload, error := json.Marshal(m)
	if error != nil {
		t.Errorf(error.Error())
	}

	req, _ = http.NewRequest("PUT", "/api/admins/deadfeed/members/deadbeef", bytes.NewBuffer(payload))
	req.Header.Add("X-Member-Code", "tutu")
	response = h.executeRequest(req)

	h.checkResponseCode(t, http.StatusAccepted, response.Code)
}

func TestDeleteMember(t *testing.T) {
	h.clearTables()
	h.addAnAdmin()
	h.addAMember()

	req, _ := http.NewRequest("GET", "/api/admins/deadfeed/members/deadbeef", nil)
	req.Header.Add("X-Member-Code", "tutu")
	response := h.executeRequest(req)
	h.checkResponseCode(t, http.StatusOK, response.Code)

	req, _ = http.NewRequest("DELETE", "/api/admins/deadfeed/members/deadbeef", nil)
	req.Header.Add("X-Member-Code", "tutu")
	response = h.executeRequest(req)
	h.checkResponseCode(t, http.StatusOK, response.Code)

	req, _ = http.NewRequest("GET", "/api/admins/deadfeed/members/deadbeef", nil)
	req.Header.Add("X-Member-Code", "tutu")
	response = h.executeRequest(req)
	h.checkResponseCode(t, http.StatusNotFound, response.Code)
}

func TestDeleteSelfAdmin(t *testing.T) {
	h.clearTables()
	h.addAnAdmin()

	req, _ := http.NewRequest("DELETE", "/api/admins/deadfeed/members/deadfeed", nil)
	req.Header.Add("X-Member-Code", "tutu")
	response := h.executeRequest(req)
	h.checkResponseCode(t, http.StatusLocked, response.Code)
}

func TestGetMember(t *testing.T) {
	h.clearTables()
	h.addAMember()

	req, _ := http.NewRequest("GET", "/api/members/deadbeef", nil)
	req.Header.Add("X-Member-Code", "toto")
	response := h.executeRequest(req)

	h.checkResponseCode(t, http.StatusOK, response.Code)

	var m map[string]interface{}
	json.Unmarshal(response.Body.Bytes(), &m)

	if m["firstName"] != "Ramon" {
		t.Errorf("Expected member name to be 'Ramon'. Got '%v'", m["firstName"])
	}
}
func TestGetMemberType(t *testing.T) {
	h.clearTables()
	h.addAMember()
	h.addAnAdmin()

	req, _ := http.NewRequest("GET", "/api/members/deadfeed", nil)
	req.Header.Add("X-Member-Code", "tutu")
	response := h.executeRequest(req)

	h.checkResponseCode(t, http.StatusOK, response.Code)

	var m map[string]interface{}
	json.Unmarshal(response.Body.Bytes(), &m)

	if m["type"] != "admin" {
		t.Errorf("Expected presence to be 'admin'. Got '%v'", m["type"])
	}

	req, _ = http.NewRequest("GET", "/api/members/deadbeef", nil)
	req.Header.Add("X-Member-Code", "toto")
	response = h.executeRequest(req)

	h.checkResponseCode(t, http.StatusOK, response.Code)

	json.Unmarshal(response.Body.Bytes(), &m)

	if m["type"] != "member" {
		t.Errorf("Expected type to be 'member'. Got '%v'", m["type"])
	}
}

func TestGetRoles(t *testing.T) {
	h.clearTables()

	req, _ := http.NewRequest("GET", "/api/roles", nil)
	response := h.executeRequest(req)

	h.checkResponseCode(t, http.StatusOK, response.Code)

	var m []string
	json.Unmarshal(response.Body.Bytes(), &m)

	// Check all roles
	validRoleList := model.ValidRoleList

	sort.Strings(validRoleList)
	sort.Strings(m)

	for i := range validRoleList {
		if validRoleList[i] != m[i] {
			t.Errorf("%v is not a valid role", m[i])
		}
	}
}

func TestCreateMemberWrongHeight(t *testing.T) {
	h.clearTables()
	h.addAnAdmin()

	payload := []byte(`{
		"firstName":"Clément",
		"lastName": "Contini",
		"height": "180,10",
		"extra":"Santi",
		"roles": ["segon","baix","terç"],
		"type": "member",
		"email": "vilisseranen@gmail.com",
		"language": "fr"}`)

	req, _ := http.NewRequest("POST", "/api/admins/deadfeed/members", bytes.NewBuffer(payload))
	req.Header.Add("X-Member-Code", "tutu")
	response := h.executeRequest(req)

	h.checkResponseCode(t, http.StatusBadRequest, response.Code)
}
func TestCreateMemberWrongWeight(t *testing.T) {
	h.clearTables()
	h.addAnAdmin()

	payload := []byte(`{
		"firstName":"Clément",
		"lastName": "Contini",
		"weight": "70.1260",
		"extra":"Santi",
		"roles": ["segon","baix","terç"],
		"type": "member",
		"email": "vilisseranen@gmail.com",
		"language": "fr"}`)

	req, _ := http.NewRequest("POST", "/api/admins/deadfeed/members", bytes.NewBuffer(payload))
	req.Header.Add("X-Member-Code", "tutu")
	response := h.executeRequest(req)

	h.checkResponseCode(t, http.StatusBadRequest, response.Code)
}

func TestUpdateSelf(t *testing.T) {
	h.clearTables()
	h.addAMember()

	req, _ := http.NewRequest("GET", "/api/members/deadbeef", nil)
	req.Header.Add("X-Member-Code", "toto")
	response := h.executeRequest(req)
	h.checkResponseCode(t, http.StatusOK, response.Code)

	var m map[string]interface{}
	json.Unmarshal(response.Body.Bytes(), &m)

	m["height"] = "180"
	payload, error := json.Marshal(m)
	if error != nil {
		t.Errorf(error.Error())
	}

	req, _ = http.NewRequest("PUT", "/api/members/deadbeef", bytes.NewBuffer(payload))
	req.Header.Add("X-Member-Code", "toto")
	response = h.executeRequest(req)

	h.checkResponseCode(t, http.StatusAccepted, response.Code)

	json.Unmarshal(response.Body.Bytes(), &m)

	if m["height"] != "180" {
		t.Errorf("Expected extra to be '180'. Got '%v'", m["height"])
	}
}
