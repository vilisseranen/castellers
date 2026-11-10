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

// The pilar de 4 is the only castell with pinya positions:
// agulla, contrafort, 2 crosses and 8 rengles of 2 cordons.
func TestPinyaPositionsOnlyInPilarDe4(t *testing.T) {
	pinyaPositions := func(castellType string) map[string]int {
		req, _ := http.NewRequest("GET", "/api/v1/castells/types/"+castellType, nil)
		response := h.executeRequest(req)
		if err := h.checkResponseCode(http.StatusOK, response.Code); err != nil {
			t.Fatal(err)
		}
		var c model.CastellType
		json.Unmarshal(response.Body.Bytes(), &c)
		count := map[string]int{}
		for _, p := range c.Positions {
			if p.Part == "pinya" {
				count[p.Name]++
			}
		}
		return count
	}

	expected := map[string]int{"agulla": 1, "contrafort": 1, "crossa": 2, "mans": 4, "vent": 4, "lateral": 8}
	got := pinyaPositions("1d4")
	if len(got) != len(expected) {
		t.Errorf("Expected pinya positions %v in 1d4. Got %v", expected, got)
	}
	for name, n := range expected {
		if got[name] != n {
			t.Errorf("Expected %d %s in the pinya of 1d4. Got %d", n, name, got[name])
		}
	}

	for _, castellType := range []string{"1d5", "2d5", "3d6", "4d6"} {
		if got := pinyaPositions(castellType); len(got) != 0 {
			t.Errorf("Expected no pinya position in %s. Got %v", castellType, got)
		}
	}
}

func TestCreateCastellWithPinya(t *testing.T) {
	h.clearTables()
	access_token := h.addAnAdmin()
	payload := []byte(`{"name":"pilar de la diada","type":"1d4","position_members":[{"member_uuid":"deadfeed","position":{"cordon":0,"column":1,"part":"pinya","name":"agulla"}}]}`)
	req, _ := http.NewRequest("POST", "/api/v1/castells/models", bytes.NewBuffer(payload))
	req.Header.Add("Authorization", "Bearer "+access_token)
	response := h.executeRequest(req)
	if err := h.checkResponseCode(http.StatusCreated, response.Code); err != nil {
		t.Fatal(err)
	}
	var created model.CastellModel
	json.Unmarshal(response.Body.Bytes(), &created)

	// Read it back: the position must have been stored, cordon 0 included
	req, _ = http.NewRequest("GET", "/api/v1/castells/models/"+created.UUID, nil)
	req.Header.Add("Authorization", "Bearer "+access_token)
	response = h.executeRequest(req)
	if err := h.checkResponseCode(http.StatusOK, response.Code); err != nil {
		t.Fatal(err)
	}
	var m model.CastellModel
	json.Unmarshal(response.Body.Bytes(), &m)
	if len(m.PositionMembers) != 1 {
		t.Fatalf("Expected to have 1 member. Got '%v'", len(m.PositionMembers))
	}
	p := m.PositionMembers[0]
	if p.MemberUUID != "deadfeed" || p.Position.Name != "agulla" || p.Position.Part != "pinya" || p.Position.Cordon != 0 || p.Position.Column != 1 {
		t.Errorf("Expected member deadfeed as agulla of the pinya. Got %+v", p)
	}
}
