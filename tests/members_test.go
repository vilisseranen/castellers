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
	access_token := h.addAnAdmin()

	payload := []byte(`{
		"firstName":"Clément",
		"lastName": "Contini",
		"height": "180",
		"weight": "70",
		"extra":"Santi",
		"roles": ["segon","baix","terç"],
		"type": "member",
		"email": "vilisseranen@gmail.com",
		"contact": "514-111-1111",
		"language": "fr",
		"subscribed": 0}`)

	req, _ := http.NewRequest("POST", "/api/v1/members", bytes.NewBuffer(payload))
	req.Header.Add("Authorization", "Bearer "+access_token)
	response := h.executeRequest(req)

	if err := h.checkResponseCode(http.StatusCreated, response.Code); err != nil {
		t.Error(err)
	}

	var m map[string]interface{}
	json.Unmarshal(response.Body.Bytes(), &m)

	if m["firstName"] != "Clément" {
		t.Errorf("Expected member first name to be 'Clément'. Got '%v'", m["firstName"])
	}
	if m["extra"] != "Santi" {
		t.Errorf("Expected extra to be 'Santi'. Got '%v'", m["extra"])
	}
	if m["contact"] != "514-111-1111" {
		t.Errorf("Expected contact to be '514-111-1111'. Got '%v'", m["contact"])
	}
}

func TestCreateMemberInvalidRole(t *testing.T) {
	h.clearTables()
	access_token := h.addAnAdmin()

	payload := []byte(`{
		"firstName":"Clément",
		"lastName": "Contini",
		"extra":"Santi",
		"roles": "segond,toto,baix,terç",
		"type": "member",
		"language": "fr",
		"email": "vilisseranen@gmail.com"}`)

	req, _ := http.NewRequest("POST", "/api/v1/members", bytes.NewBuffer(payload))
	req.Header.Add("Authorization", "Bearer "+access_token)
	response := h.executeRequest(req)

	if err := h.checkResponseCode(http.StatusBadRequest, response.Code); err != nil {
		t.Error(err)
	}
}

func TestCreateMemberInvalidType(t *testing.T) {
	h.clearTables()
	access_token := h.addAnAdmin()

	payload := []byte(`{
		"firstName":"Clément",
		"lastName": "Contini",
		"extra":"Santi",
		"type": "toto",
		"language": "fr",
		"email": "vilisseranen@gmail.com"}`)

	req, _ := http.NewRequest("POST", "/api/v1/members", bytes.NewBuffer(payload))
	req.Header.Add("Authorization", "Bearer "+access_token)
	response := h.executeRequest(req)

	if err := h.checkResponseCode(http.StatusBadRequest, response.Code); err != nil {
		t.Error(err)
	}
}

func TestCreateMemberNoExtra(t *testing.T) {
	h.clearTables()
	access_token := h.addAnAdmin()

	payload := []byte(`{
		"firstName":"Clément",
		"lastName": "Contini",
		"type": "member",
		"email": "vilisseranen@gmail.com",
		"language": "cat"}`)

	req, _ := http.NewRequest("POST", "/api/v1/members", bytes.NewBuffer(payload))
	req.Header.Add("Authorization", "Bearer "+access_token)
	response := h.executeRequest(req)

	if err := h.checkResponseCode(http.StatusCreated, response.Code); err != nil {
		t.Error(err)
	}

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

	var memberUUID string
	memberUUID = m["uuid"].(string)

	req, _ = http.NewRequest("GET", "/api/v1/members/"+memberUUID, nil)
	req.Header.Add("Authorization", "Bearer "+access_token)
	response = h.executeRequest(req)
	if err := h.checkResponseCode(http.StatusOK, response.Code); err != nil {
		t.Error(err)
	}
}

func TestUpdateMember(t *testing.T) {
	h.clearTables()
	access_token := h.addAnAdmin()
	h.addAMember()

	req, _ := http.NewRequest("GET", "/api/v1/members/deadbeef", nil)
	req.Header.Add("Authorization", "Bearer "+access_token)
	response := h.executeRequest(req)
	if err := h.checkResponseCode(http.StatusOK, response.Code); err != nil {
		t.Error(err)
	}

	var m map[string]interface{}
	json.Unmarshal(response.Body.Bytes(), &m)

	m["extra"] = "Cap de pinya"
	m["subscribed"] = 1
	m["contact"] = "514-111-1111"
	payload, error := json.Marshal(m)
	if error != nil {
		t.Errorf(error.Error())
	}

	req, _ = http.NewRequest("PUT", "/api/v1/members/deadbeef", bytes.NewBuffer(payload))
	req.Header.Add("Authorization", "Bearer "+access_token)
	response = h.executeRequest(req)

	if err := h.checkResponseCode(http.StatusAccepted, response.Code); err != nil {
		t.Error(err)
	}

	req, _ = http.NewRequest("GET", "/api/v1/members/deadbeef", bytes.NewBuffer(payload))
	req.Header.Add("Authorization", "Bearer "+access_token)
	response = h.executeRequest(req)

	json.Unmarshal(response.Body.Bytes(), &m)

	if m["extra"] != "Cap de pinya" {
		t.Errorf("Expected extra to be 'Cap de pinya'. Got '%v'", m["extra"])
	}
	if m["subscribed"] != 1.0 {
		t.Errorf("Expected subscribed to be '1'. Got '%v'", m["subscribed"])
	}
	if m["contact"] != "514-111-1111" {
		t.Errorf("Expected contact to be '514-111-1111'. Got '%v'", m["contact"])
	}
}

func TestPromoteSelf(t *testing.T) {
	h.clearTables()
	access_token := h.addAMember()

	req, _ := http.NewRequest("GET", "/api/v1/members/deadbeef", nil)
	req.Header.Add("Authorization", "Bearer "+access_token)
	response := h.executeRequest(req)
	if err := h.checkResponseCode(http.StatusOK, response.Code); err != nil {
		t.Error(err)
	}

	var m map[string]interface{}
	json.Unmarshal(response.Body.Bytes(), &m)

	m["type"] = "admin"
	payload, error := json.Marshal(m)
	if error != nil {
		t.Errorf(error.Error())
	}

	req, _ = http.NewRequest("PUT", "/api/v1/members/deadbeef", bytes.NewBuffer(payload))
	req.Header.Add("Authorization", "Bearer "+access_token)
	response = h.executeRequest(req)

	if err := h.checkResponseCode(http.StatusForbidden, response.Code); err != nil {
		t.Error(err)
	}
}

func TestPromoteByAdmin(t *testing.T) {
	h.clearTables()
	access_token := h.addAnAdmin()
	h.addAMember()

	req, _ := http.NewRequest("GET", "/api/v1/members/deadbeef", nil)
	req.Header.Add("Authorization", "Bearer "+access_token)
	response := h.executeRequest(req)
	if err := h.checkResponseCode(http.StatusOK, response.Code); err != nil {
		t.Error(err)
	}

	var m map[string]interface{}
	json.Unmarshal(response.Body.Bytes(), &m)

	m["type"] = "admin"
	payload, error := json.Marshal(m)
	if error != nil {
		t.Errorf(error.Error())
	}

	req, _ = http.NewRequest("PUT", "/api/v1/members/deadbeef", bytes.NewBuffer(payload))
	req.Header.Add("Authorization", "Bearer "+access_token)
	response = h.executeRequest(req)

	if err := h.checkResponseCode(http.StatusAccepted, response.Code); err != nil {
		t.Error(err)
	}
}

func TestDeleteMember(t *testing.T) {
	h.clearTables()
	access_token := h.addAnAdmin()
	h.addAMember()

	req, _ := http.NewRequest("GET", "/api/v1/members/deadbeef", nil)
	req.Header.Add("Authorization", "Bearer "+access_token)
	response := h.executeRequest(req)
	if err := h.checkResponseCode(http.StatusOK, response.Code); err != nil {
		t.Error(err)
	}
	req, _ = http.NewRequest("DELETE", "/api/v1/members/deadbeef", nil)
	req.Header.Add("Authorization", "Bearer "+access_token)
	response = h.executeRequest(req)
	if err := h.checkResponseCode(http.StatusOK, response.Code); err != nil {
		t.Error(err)
	}
	req, _ = http.NewRequest("GET", "/api/v1/members/deadbeef", nil)
	req.Header.Add("Authorization", "Bearer "+access_token)
	response = h.executeRequest(req)
	if err := h.checkResponseCode(http.StatusNotFound, response.Code); err != nil {
		t.Error(err)
	}
}

func TestDeleteSelfAdmin(t *testing.T) {
	h.clearTables()
	access_token := h.addAnAdmin()

	req, _ := http.NewRequest("DELETE", "/api/v1/members/deadfeed", nil)
	req.Header.Add("Authorization", "Bearer "+access_token)
	response := h.executeRequest(req)
	if err := h.checkResponseCode(http.StatusLocked, response.Code); err != nil {
		t.Error(err)
	}
}

func TestGetMember(t *testing.T) {
	h.clearTables()
	access_token := h.addAMember()

	req, _ := http.NewRequest("GET", "/api/v1/members/deadbeef", nil)
	req.Header.Add("Authorization", "Bearer "+access_token)
	response := h.executeRequest(req)

	if err := h.checkResponseCode(http.StatusOK, response.Code); err != nil {
		t.Error(err)
	}

	var m map[string]interface{}
	json.Unmarshal(response.Body.Bytes(), &m)

	if m["firstName"] != "Ramon" {
		t.Errorf("Expected member name to be 'Ramon'. Got '%v'", m["firstName"])
	}
}
func TestGetMemberType(t *testing.T) {
	h.clearTables()
	access_token_member := h.addAMember()
	access_token_admin := h.addAnAdmin()

	req, _ := http.NewRequest("GET", "/api/v1/members/deadfeed", nil)
	req.Header.Add("Authorization", "Bearer "+access_token_admin)
	response := h.executeRequest(req)

	if err := h.checkResponseCode(http.StatusOK, response.Code); err != nil {
		t.Error(err)
	}

	var m map[string]interface{}
	json.Unmarshal(response.Body.Bytes(), &m)

	if m["type"] != "admin" {
		t.Errorf("Expected presence to be 'admin'. Got '%v'", m["type"])
	}

	req, _ = http.NewRequest("GET", "/api/v1/members/deadbeef", nil)
	req.Header.Add("Authorization", "Bearer "+access_token_member)
	response = h.executeRequest(req)

	if err := h.checkResponseCode(http.StatusOK, response.Code); err != nil {
		t.Error(err)
	}

	json.Unmarshal(response.Body.Bytes(), &m)

	if m["type"] != "member" {
		t.Errorf("Expected type to be 'member'. Got '%v'", m["type"])
	}
}

func TestGetMemberByEmailSuccess(t *testing.T) {
	h.clearTables()
	h.addAMember()

	memberWithExistingEmail := model.Member{Email: "ramon@gerard.ca"}
	memberWithExistingEmail.GetByEmail()

	if memberWithExistingEmail.UUID != "deadbeef" {
		t.Errorf("Expected member with email %s to retrieved with UUID %s but got UUID %s.", "ramon@gerard.ca", "deadbeef", memberWithExistingEmail.UUID)
	}
}

func TestGetMemberByEmailFail(t *testing.T) {
	h.clearTables()
	h.addAMember()

	memberWithExistingEmail := model.Member{Email: "toto@tutu.ca"}
	err := memberWithExistingEmail.GetByEmail()

	if err.Error() != model.MEMBERSEMAILNOTFOUNDMESSAGE {
		t.Errorf("Expected GetByEmail fail with error '%s' but got '%s'", model.MEMBERSEMAILNOTFOUNDMESSAGE, err.Error())
	}
}

func TestGetRoles(t *testing.T) {
	h.clearTables()

	req, _ := http.NewRequest("GET", "/api/v1/members/roles", nil)
	response := h.executeRequest(req)

	if err := h.checkResponseCode(http.StatusOK, response.Code); err != nil {
		t.Error(err)
	}

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
	access_token := h.addAnAdmin()

	payload := []byte(`{
		"firstName":"Clément",
		"lastName": "Contini",
		"height": "180,10",
		"extra":"Santi",
		"roles": ["segon","baix","terç"],
		"type": "member",
		"email": "vilisseranen@gmail.com",
		"language": "fr"}`)

	req, _ := http.NewRequest("POST", "/api/v1/members", bytes.NewBuffer(payload))
	req.Header.Add("Authorization", "Bearer "+access_token)
	response := h.executeRequest(req)

	if err := h.checkResponseCode(http.StatusBadRequest, response.Code); err != nil {
		t.Error(err)
	}
}
func TestCreateMemberWrongWeight(t *testing.T) {
	h.clearTables()
	access_token := h.addAnAdmin()

	payload := []byte(`{
		"firstName":"Clément",
		"lastName": "Contini",
		"weight": "70.1260",
		"extra":"Santi",
		"roles": ["segon","baix","terç"],
		"type": "member",
		"email": "vilisseranen@gmail.com",
		"language": "fr"}`)

	req, _ := http.NewRequest("POST", "/api/v1/members", bytes.NewBuffer(payload))
	req.Header.Add("Authorization", "Bearer "+access_token)
	response := h.executeRequest(req)

	if err := h.checkResponseCode(http.StatusBadRequest, response.Code); err != nil {
		t.Error(err)
	}
}

func TestUpdateSelf(t *testing.T) {
	h.clearTables()
	access_token := h.addAMember()

	req, _ := http.NewRequest("GET", "/api/v1/members/deadbeef", nil)
	req.Header.Add("Authorization", "Bearer "+access_token)
	response := h.executeRequest(req)
	h.checkResponseCode(http.StatusOK, response.Code)

	var m map[string]interface{}
	json.Unmarshal(response.Body.Bytes(), &m)

	m["height"] = "180"
	payload, error := json.Marshal(m)
	if error != nil {
		t.Errorf(error.Error())
	}

	req, _ = http.NewRequest("PUT", "/api/v1/members/deadbeef", bytes.NewBuffer(payload))
	req.Header.Add("Authorization", "Bearer "+access_token)
	response = h.executeRequest(req)

	h.checkResponseCode(http.StatusAccepted, response.Code)

	json.Unmarshal(response.Body.Bytes(), &m)

	if m["height"] != "180" {
		t.Errorf("Expected extra to be '180'. Got '%v'", m["height"])
	}
}
