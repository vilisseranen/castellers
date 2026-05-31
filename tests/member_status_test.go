package tests

import (
	"bytes"
	"database/sql"
	"net/http"
	"testing"
	"time"

	"github.com/vilisseranen/castellers/controller"
	"github.com/vilisseranen/castellers/model"
)

func (test *TestHelper) getMemberStatus(uuid string) string {
	db, err := sql.Open("sqlite3", testDbName)
	if err != nil {
		tFatal(err)
	}
	defer db.Close()
	var status string
	if err = db.QueryRow("SELECT status FROM members WHERE uuid = ?", uuid).Scan(&status); err != nil {
		tFatal(err)
	}
	return status
}

func (test *TestHelper) setMemberLastActivityDate(uuid string, date int64) {
	db, err := sql.Open("sqlite3", testDbName)
	if err != nil {
		tFatal(err)
	}
	defer db.Close()
	if _, err = db.Exec("UPDATE members SET last_activity_date = ? WHERE uuid = ?", date, uuid); err != nil {
		tFatal(err)
	}
}

func TestSetMemberStatusByAdmin(t *testing.T) {
	h.clearTables()
	access_token := h.addAnAdmin()
	h.addAMember()
	h.setMemberStatus("deadbeef", model.MEMBERSSTATUSACTIVATED)

	// Pause the member
	payload := []byte(`{"status":"paused"}`)
	req, _ := http.NewRequest("PUT", "/api/v1/members/deadbeef/status", bytes.NewBuffer(payload))
	req.Header.Add("Authorization", "Bearer "+access_token)
	response := h.executeRequest(req)
	if err := h.checkResponseCode(http.StatusAccepted, response.Code); err != nil {
		t.Error(err)
	}
	if status := h.getMemberStatus("deadbeef"); status != model.MEMBERSSTATUSPAUSED {
		t.Errorf("Expected status to be 'paused'. Got '%s'", status)
	}

	// Reactivate the member
	payload = []byte(`{"status":"active"}`)
	req, _ = http.NewRequest("PUT", "/api/v1/members/deadbeef/status", bytes.NewBuffer(payload))
	req.Header.Add("Authorization", "Bearer "+access_token)
	response = h.executeRequest(req)
	if err := h.checkResponseCode(http.StatusAccepted, response.Code); err != nil {
		t.Error(err)
	}
	if status := h.getMemberStatus("deadbeef"); status != model.MEMBERSSTATUSACTIVATED {
		t.Errorf("Expected status to be 'active'. Got '%s'", status)
	}
}

func TestSetMemberStatusNonAdmin(t *testing.T) {
	h.clearTables()
	access_token := h.addAMember()

	payload := []byte(`{"status":"paused"}`)
	req, _ := http.NewRequest("PUT", "/api/v1/members/deadbeef/status", bytes.NewBuffer(payload))
	req.Header.Add("Authorization", "Bearer "+access_token)
	response := h.executeRequest(req)
	if err := h.checkResponseCode(http.StatusUnauthorized, response.Code); err != nil {
		t.Error(err)
	}
}

func TestSetMemberStatusInvalid(t *testing.T) {
	h.clearTables()
	access_token := h.addAnAdmin()
	h.addAMember()

	payload := []byte(`{"status":"deleted"}`)
	req, _ := http.NewRequest("PUT", "/api/v1/members/deadbeef/status", bytes.NewBuffer(payload))
	req.Header.Add("Authorization", "Bearer "+access_token)
	response := h.executeRequest(req)
	if err := h.checkResponseCode(http.StatusBadRequest, response.Code); err != nil {
		t.Error(err)
	}
}

// A member reactivated manually (recent last_activity_date) must not be paused
// again by the inactivity scan, even without any participation.
func TestManualReactivationSurvivesAutoPause(t *testing.T) {
	h.clearTables()
	h.addAMember()
	h.setMemberStatus("deadbeef", model.MEMBERSSTATUSACTIVATED)
	h.setMemberLastActivityDate("deadbeef", time.Now().Unix())

	controller.RunPauseAbsentMembersOnce()

	if status := h.getMemberStatus("deadbeef"); status != model.MEMBERSSTATUSACTIVATED {
		t.Errorf("Expected manually reactivated member to stay 'active'. Got '%s'", status)
	}
}

// An active member with no recent activity and no participation should still be
// paused by the inactivity scan (existing behaviour preserved).
func TestInactiveMemberStillPaused(t *testing.T) {
	h.clearTables()
	h.addAMember()
	h.setMemberStatus("deadbeef", model.MEMBERSSTATUSACTIVATED)
	h.setMemberLastActivityDate("deadbeef", 0)

	controller.RunPauseAbsentMembersOnce()

	if status := h.getMemberStatus("deadbeef"); status != model.MEMBERSSTATUSPAUSED {
		t.Errorf("Expected inactive member to be 'paused'. Got '%s'", status)
	}
}
