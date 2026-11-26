package tests

import (
	"bytes"
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"github.com/vilisseranen/castellers/model"
)

func TestValidateShirt(t *testing.T) {
	now := time.Date(2026, time.November, 26, 18, 0, 0, 0, time.UTC)
	cases := []struct {
		size  int
		since string
		valid bool
	}{
		{0, "", true},
		{28, "", true},
		{36, "2019-06-24", true},
		{44, "2026-11-26", true},
		{0, "2019-06-24", true},
		{26, "", false},
		{46, "", false},
		{35, "", false},
		{-36, "", false},
		{36, "2027-01-01", false},
		{36, "24/06/2019", false},
		{36, "2019-02-30", false},
		{36, "2019-06", false},
	}
	for _, c := range cases {
		err := model.ValidateShirt(c.size, c.since, now)
		if c.valid && err != nil {
			t.Errorf("Expected size %d since %q to be valid. Got %v", c.size, c.since, err)
		}
		if !c.valid && err == nil {
			t.Errorf("Expected size %d since %q to be invalid", c.size, c.since)
		}
	}
}

// The admins note the size of the shirt of a member and the day they got it.
// The member sees them, but cannot change them.
func TestOnlyAdminSetsShirt(t *testing.T) {
	h.clearTables()
	memberToken := h.addAMember()
	adminToken := h.addAnAdmin()

	// Not everybody has a shirt.
	m := h.getMemberAsMap(t, memberToken, "deadbeef")
	if m["shirtSize"] != float64(0) || m["shirtSince"] != "" {
		t.Fatalf("Expected no shirt. Got size '%v' since '%v'", m["shirtSize"], m["shirtSince"])
	}

	// A member cannot note their own shirt.
	m["shirtSize"] = 36
	m["shirtSince"] = "2019-06-24"
	h.putMember(t, memberToken, "deadbeef", m)
	m = h.getMemberAsMap(t, memberToken, "deadbeef")
	if m["shirtSize"] != float64(0) || m["shirtSince"] != "" {
		t.Errorf("A member noted their own shirt: size '%v' since '%v'", m["shirtSize"], m["shirtSince"])
	}

	// An admin can.
	m = h.getMemberAsMap(t, adminToken, "deadbeef")
	m["shirtSize"] = 36
	m["shirtSince"] = "2019-06-24"
	h.putMember(t, adminToken, "deadbeef", m)
	m = h.getMemberAsMap(t, memberToken, "deadbeef")
	if m["shirtSize"] != float64(36) || m["shirtSince"] != "2019-06-24" {
		t.Errorf("Expected size 36 since 2019-06-24. Got size '%v' since '%v'", m["shirtSize"], m["shirtSince"])
	}

	// The member keeps their shirt when they update their own profile,
	// whatever they send, even something invalid.
	m["shirtSize"] = 35
	m["shirtSince"] = "yesterday"
	m["height"] = "180"
	h.putMember(t, memberToken, "deadbeef", m)
	m = h.getMemberAsMap(t, memberToken, "deadbeef")
	if m["shirtSize"] != float64(36) || m["shirtSince"] != "2019-06-24" || m["height"] != "180" {
		t.Errorf("A member changed their shirt: size '%v' since '%v'", m["shirtSize"], m["shirtSince"])
	}

	// The list of members gives the shirts to the admins.
	req, _ := http.NewRequest("GET", "/api/v1/members", nil)
	req.Header.Add("Authorization", "Bearer "+adminToken)
	response := h.executeRequest(req)
	var members []model.Member
	json.Unmarshal(response.Body.Bytes(), &members)
	found := false
	for _, member := range members {
		if member.UUID == "deadbeef" {
			found = true
			if member.ShirtSize != 36 || member.ShirtSince != "2019-06-24" {
				t.Errorf("Expected the list to give the shirt. Got size %d since %q", member.ShirtSize, member.ShirtSince)
			}
		} else if member.ShirtSize != 0 || member.ShirtSince != "" {
			t.Errorf("Expected %s to have no shirt. Got size %d since %q", member.UUID, member.ShirtSize, member.ShirtSince)
		}
	}
	if !found {
		t.Error("Expected the member to be listed")
	}

	// An admin cannot note a size that does not exist.
	for _, change := range []map[string]interface{}{
		{"shirtSize": 35},
		{"shirtSize": 46},
		{"shirtSince": "2019-02-30"},
	} {
		m = h.getMemberAsMap(t, adminToken, "deadbeef")
		for key, value := range change {
			m[key] = value
		}
		payload, _ := json.Marshal(m)
		req, _ := http.NewRequest("PUT", "/api/v1/members/deadbeef", bytes.NewBuffer(payload))
		req.Header.Add("Authorization", "Bearer "+adminToken)
		if err := h.checkResponseCode(http.StatusBadRequest, h.executeRequest(req).Code); err != nil {
			t.Errorf("%v: %v", change, err)
		}
	}

	// The admin removes the shirt.
	m = h.getMemberAsMap(t, adminToken, "deadbeef")
	m["shirtSize"] = 0
	m["shirtSince"] = ""
	h.putMember(t, adminToken, "deadbeef", m)
	m = h.getMemberAsMap(t, adminToken, "deadbeef")
	if m["shirtSize"] != float64(0) || m["shirtSince"] != "" {
		t.Errorf("Expected the shirt to be removed. Got size '%v' since '%v'", m["shirtSize"], m["shirtSince"])
	}
}

// A new member can be created with their shirt.
func TestCreateMemberWithShirt(t *testing.T) {
	h.clearTables()
	adminToken := h.addAnAdmin()

	create := func(shirt string) (int, model.Member) {
		payload := []byte(`{"firstName":"Clément","lastName":"Contini","roles":["segon"],"type":"member",` +
			`"email":"shirt@example.com","language":"fr","subscribed":0,` + shirt + `}`)
		req, _ := http.NewRequest("POST", "/api/v1/members", bytes.NewBuffer(payload))
		req.Header.Add("Authorization", "Bearer "+adminToken)
		response := h.executeRequest(req)
		var created model.Member
		json.Unmarshal(response.Body.Bytes(), &created)
		return response.Code, created
	}

	code, created := create(`"shirtSize":40,"shirtSince":"2024-09-11"`)
	if err := h.checkResponseCode(http.StatusCreated, code); err != nil {
		t.Fatal(err)
	}
	m := h.getMemberAsMap(t, adminToken, created.UUID)
	if m["shirtSize"] != float64(40) || m["shirtSince"] != "2024-09-11" {
		t.Errorf("Expected size 40 since 2024-09-11. Got size '%v' since '%v'", m["shirtSize"], m["shirtSince"])
	}

	code, _ = create(`"shirtSize":41`)
	if err := h.checkResponseCode(http.StatusBadRequest, code); err != nil {
		t.Error(err)
	}
}
