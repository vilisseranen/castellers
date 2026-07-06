package tests

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/vilisseranen/castellers/common"
	"github.com/vilisseranen/castellers/model"
)

const (
	welcomeSeriesUUID = "000000000000000000000000000000005e100001"
	casalBadgeUUID    = "00000000000000000000000000000000bad00001"
	camisaBadgeUUID   = "00000000000000000000000000000000bad00002"
	amuntBadgeUUID    = "00000000000000000000000000000000bad00005"
)

func (test *TestHelper) memberHasBadge(t *testing.T, memberUUID, badgeUUID, token string) bool {
	t.Helper()
	req, _ := http.NewRequest("GET", "/api/v1/members/"+memberUUID+"/badges", nil)
	req.Header.Add("Authorization", "Bearer "+token)
	response := test.executeRequest(req)
	if err := test.checkResponseCode(http.StatusOK, response.Code); err != nil {
		t.Error(err)
	}
	var badges []model.MemberBadge
	json.Unmarshal(response.Body.Bytes(), &badges)
	for _, b := range badges {
		if b.BadgeUUID == badgeUUID {
			return true
		}
	}
	return false
}

func (test *TestHelper) clearMemberBadges() {
	db, err := sql.Open("sqlite3", testDbName)
	if err != nil {
		common.Fatal(err.Error())
	}
	db.Exec("DELETE FROM member_badges")
}

func TestGetBadges(t *testing.T) {
	h.clearTables()
	h.clearMemberBadges()
	accessToken := h.addAMember()

	req, _ := http.NewRequest("GET", "/api/v1/badges", nil)
	req.Header.Add("Authorization", "Bearer "+accessToken)
	response := h.executeRequest(req)

	if err := h.checkResponseCode(http.StatusOK, response.Code); err != nil {
		t.Error(err)
	}

	var series []model.BadgeSeries
	json.Unmarshal(response.Body.Bytes(), &series)

	if len(series) != 1 {
		t.Fatalf("Expected 1 badge series. Got %d", len(series))
	}
	if series[0].Code != "welcome" {
		t.Errorf("Expected series code 'welcome'. Got '%s'", series[0].Code)
	}
	if len(series[0].Badges) != 7 {
		t.Errorf("Expected 7 badges in the welcome series. Got %d", len(series[0].Badges))
	}
}

func TestAssignAndGetMemberBadges(t *testing.T) {
	h.clearTables()
	h.clearMemberBadges()
	adminToken := h.addAnAdmin()
	memberToken := h.addAMember()

	payload := []byte(`{"memberUuids":["deadbeef"]}`)
	req, _ := http.NewRequest("POST", "/api/v1/badges/"+casalBadgeUUID+"/members", bytes.NewBuffer(payload))
	req.Header.Add("Authorization", "Bearer "+adminToken)
	response := h.executeRequest(req)
	if err := h.checkResponseCode(http.StatusOK, response.Code); err != nil {
		t.Error(err)
	}

	// Assigning again is idempotent.
	req, _ = http.NewRequest("POST", "/api/v1/badges/"+casalBadgeUUID+"/members", bytes.NewBuffer([]byte(`{"memberUuids":["deadbeef"]}`)))
	req.Header.Add("Authorization", "Bearer "+adminToken)
	response = h.executeRequest(req)
	if err := h.checkResponseCode(http.StatusOK, response.Code); err != nil {
		t.Error(err)
	}

	// The member can read their own badges.
	req, _ = http.NewRequest("GET", "/api/v1/members/deadbeef/badges", nil)
	req.Header.Add("Authorization", "Bearer "+memberToken)
	response = h.executeRequest(req)
	if err := h.checkResponseCode(http.StatusOK, response.Code); err != nil {
		t.Error(err)
	}
	var badges []model.MemberBadge
	json.Unmarshal(response.Body.Bytes(), &badges)
	if len(badges) != 1 {
		t.Fatalf("Expected 1 unlocked badge. Got %d", len(badges))
	}
	if badges[0].BadgeUUID != casalBadgeUUID {
		t.Errorf("Expected badge %s. Got %s", casalBadgeUUID, badges[0].BadgeUUID)
	}
}

func TestGetBadgeMembers(t *testing.T) {
	h.clearTables()
	h.clearMemberBadges()
	adminToken := h.addAnAdmin()
	h.addAMember() // creates member "deadbeef"

	payload := []byte(`{"memberUuids":["deadbeef"]}`)
	req, _ := http.NewRequest("POST", "/api/v1/badges/"+casalBadgeUUID+"/members", bytes.NewBuffer(payload))
	req.Header.Add("Authorization", "Bearer "+adminToken)
	if err := h.checkResponseCode(http.StatusOK, h.executeRequest(req).Code); err != nil {
		t.Error(err)
	}

	req, _ = http.NewRequest("GET", "/api/v1/badges/"+casalBadgeUUID+"/members", nil)
	req.Header.Add("Authorization", "Bearer "+adminToken)
	response := h.executeRequest(req)
	if err := h.checkResponseCode(http.StatusOK, response.Code); err != nil {
		t.Error(err)
	}
	var memberUUIDs []string
	json.Unmarshal(response.Body.Bytes(), &memberUUIDs)
	if len(memberUUIDs) != 1 || memberUUIDs[0] != "deadbeef" {
		t.Errorf("Expected [deadbeef]. Got %v", memberUUIDs)
	}
}

func TestRemoveBadge(t *testing.T) {
	h.clearTables()
	h.clearMemberBadges()
	adminToken := h.addAnAdmin()

	assign := []byte(`{"memberUuids":["deadfeed"]}`)
	req, _ := http.NewRequest("POST", "/api/v1/badges/"+camisaBadgeUUID+"/members", bytes.NewBuffer(assign))
	req.Header.Add("Authorization", "Bearer "+adminToken)
	if err := h.checkResponseCode(http.StatusOK, h.executeRequest(req).Code); err != nil {
		t.Error(err)
	}

	remove := []byte(`{"memberUuids":["deadfeed"]}`)
	req, _ = http.NewRequest("DELETE", "/api/v1/badges/"+camisaBadgeUUID+"/members", bytes.NewBuffer(remove))
	req.Header.Add("Authorization", "Bearer "+adminToken)
	if err := h.checkResponseCode(http.StatusOK, h.executeRequest(req).Code); err != nil {
		t.Error(err)
	}

	req, _ = http.NewRequest("GET", "/api/v1/members/deadfeed/badges", nil)
	req.Header.Add("Authorization", "Bearer "+adminToken)
	response := h.executeRequest(req)
	var badges []model.MemberBadge
	json.Unmarshal(response.Body.Bytes(), &badges)
	if len(badges) != 0 {
		t.Errorf("Expected no badge after removal. Got %d", len(badges))
	}
}

func TestAssignBadgeForbiddenForMember(t *testing.T) {
	h.clearTables()
	h.clearMemberBadges()
	memberToken := h.addAMember()

	payload := []byte(`{"memberUuids":["deadbeef"]}`)
	req, _ := http.NewRequest("POST", "/api/v1/badges/"+casalBadgeUUID+"/members", bytes.NewBuffer(payload))
	req.Header.Add("Authorization", "Bearer "+memberToken)
	response := h.executeRequest(req)
	if err := h.checkResponseCode(http.StatusUnauthorized, response.Code); err != nil {
		t.Error(err)
	}
}

func TestAssignBadgeNotFound(t *testing.T) {
	h.clearTables()
	h.clearMemberBadges()
	adminToken := h.addAnAdmin()

	payload := []byte(`{"memberUuids":["deadfeed"]}`)
	req, _ := http.NewRequest("POST", "/api/v1/badges/00000000000000000000000000000000deadbad0/members", bytes.NewBuffer(payload))
	req.Header.Add("Authorization", "Bearer "+adminToken)
	response := h.executeRequest(req)
	if err := h.checkResponseCode(http.StatusNotFound, response.Code); err != nil {
		t.Error(err)
	}
}

func TestAmuntBadgeAwardedOnSelfParticipation(t *testing.T) {
	h.clearTables()
	h.clearMemberBadges()
	memberToken := h.addAMember() // member "deadbeef"
	h.addEvent("beef0001", "diada", 1528048800, 1528059600)

	// The member confirms their own participation.
	payload := []byte(`{"answer":"yes"}`)
	req, _ := http.NewRequest("POST", "/api/v1/members/events/beef0001", bytes.NewBuffer(payload))
	req.Header.Add("Authorization", "Bearer "+memberToken)
	response := h.executeRequest(req)
	if err := h.checkResponseCode(http.StatusCreated, response.Code); err != nil {
		t.Error(err)
	}

	if !h.memberHasBadge(t, "deadbeef", amuntBadgeUUID, memberToken) {
		t.Error("Expected the Amunt badge to be auto-awarded after self participation")
	}
}

func TestAmuntBadgeAwardedForAnyAnswer(t *testing.T) {
	h.clearTables()
	h.clearMemberBadges()
	memberToken := h.addAMember() // member "deadbeef"
	h.addEvent("beef0002", "diada", 1528048800, 1528059600)

	// Answering "no" still counts: the member logged in and used the app.
	payload := []byte(`{"answer":"no"}`)
	req, _ := http.NewRequest("POST", "/api/v1/members/events/beef0002", bytes.NewBuffer(payload))
	req.Header.Add("Authorization", "Bearer "+memberToken)
	response := h.executeRequest(req)
	if err := h.checkResponseCode(http.StatusCreated, response.Code); err != nil {
		t.Error(err)
	}

	if !h.memberHasBadge(t, "deadbeef", amuntBadgeUUID, memberToken) {
		t.Error("Expected the Amunt badge to be awarded for any participation answer")
	}
}

func TestAmuntBadgeNotAwardedWhenAdminAnswersForAnother(t *testing.T) {
	h.clearTables()
	h.clearMemberBadges()
	h.addAMember() // member "deadbeef"
	adminToken := h.addAnAdmin() // admin "deadfeed"
	h.addEvent("beef0003", "diada", 1528048800, 1528059600)

	// The admin sets the participation on behalf of an unrelated member.
	payload := []byte(`{"answer":"yes"}`)
	req, _ := http.NewRequest("POST", "/api/v1/members/deadbeef/events/beef0003", bytes.NewBuffer(payload))
	req.Header.Add("Authorization", "Bearer "+adminToken)
	response := h.executeRequest(req)
	if err := h.checkResponseCode(http.StatusCreated, response.Code); err != nil {
		t.Error(err)
	}

	if h.memberHasBadge(t, "deadbeef", amuntBadgeUUID, adminToken) {
		t.Error("Expected NO Amunt badge when an admin answers on behalf of another member")
	}
}

func TestMemberCanViewAnotherProfile(t *testing.T) {
	h.clearTables()
	h.clearMemberBadges()
	memberToken := h.addAMember()
	h.addAnAdmin() // creates member "deadfeed"

	req, _ := http.NewRequest("GET", "/api/v1/members/deadfeed", nil)
	req.Header.Add("Authorization", "Bearer "+memberToken)
	response := h.executeRequest(req)
	if err := h.checkResponseCode(http.StatusOK, response.Code); err != nil {
		t.Error(err)
	}

	var m map[string]interface{}
	json.Unmarshal(response.Body.Bytes(), &m)
	if m["firstName"] != "Romà" {
		t.Errorf("Expected to see first name 'Romà'. Got '%v'", m["firstName"])
	}
	// Sensitive fields must be stripped when viewing another member.
	if m["email"] != "" {
		t.Errorf("Expected email to be hidden. Got '%v'", m["email"])
	}
	if m["type"] != "" {
		t.Errorf("Expected type to be hidden. Got '%v'", m["type"])
	}
}
