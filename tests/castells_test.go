package tests

import (
	"bytes"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/vilisseranen/castellers/model"
)

func TestCreateCastell(t *testing.T) {
	h.clearTables()
	access_token := h.addAnAdmin()
	payload := []byte(`{"name":"torre de la festa major","positions":[{"name":"enxaneta","column":1,"cordon":3,"part":"pom"},{"name":"dos","column":1,"cordon":1,"part":"pom"},{"name":"dos","column":2,"cordon":1,"part":"pom"},{"name":"baix","column":1,"cordon":1,"part":"tronc"},{"name":"baix","column":2,"cordon":1,"part":"tronc"},{"name":"segon","column":1,"cordon":2,"part":"tronc"},{"name":"segon","column":2,"cordon":2,"part":"tronc"},{"name":"acotxador","column":2,"cordon":2,"part":"pom"}],"type":"2d5","position_members":[{"member_uuid":"deadfeed","position":{"cordon":1,"column":1,"part":"tronc","name":"baix"}}]}`)
	req, _ := http.NewRequest("POST", "/api/v1/castells/models", bytes.NewBuffer(payload))

	req.Header.Add("Authorization", "Bearer "+access_token)
	response := h.executeRequest(req)

	if err := h.checkResponseCode(http.StatusCreated, response.Code); err != nil {
		t.Error(err)
	}
	var m model.CastellModel
	json.Unmarshal(response.Body.Bytes(), &m)

	if m.Name != "torre de la festa major" {
		t.Errorf("Expected name to be 'torre de la festa major'. Got '%v'", m.Name)
	}

	if len(m.PositionMembers) != 1 {
		t.Errorf("Expected to have 1 member. Got '%v'", len(m.PositionMembers))
	} else if m.PositionMembers[0].MemberUUID != "deadfeed" ||
		m.PositionMembers[0].Position.Name != "baix" {
		t.Errorf("Expected member %s at position %s. Got '%v' at '%v'", "deadfeed", "baix", m.PositionMembers[0].MemberUUID, m.PositionMembers[0].Position.Name)
	}
}
