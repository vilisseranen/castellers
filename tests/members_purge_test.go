package tests

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"github.com/vilisseranen/castellers/controller"
	"github.com/vilisseranen/castellers/model"
)

const purgeDelayForTests = int64(90 * 24 * 3600)

// setMemberDeletedAt marks a member as deleted at the given time (Unix seconds).
func (test *TestHelper) setMemberDeletedAt(uuid string, deletedAt int64) {
	db, err := sql.Open("sqlite3", testDbName)
	if err != nil {
		tFatal(err)
	}
	defer db.Close()
	if _, err = db.Exec("UPDATE members SET status = ?, deleted_at = ? WHERE uuid = ?", model.MEMBERSSTATUSDELETED, deletedAt, uuid); err != nil {
		tFatal(err)
	}
}

func (test *TestHelper) countRows(query string, args ...interface{}) int {
	db, err := sql.Open("sqlite3", testDbName)
	if err != nil {
		tFatal(err)
	}
	defer db.Close()
	var count int
	if err = db.QueryRow(query, args...).Scan(&count); err != nil {
		tFatal(err)
	}
	return count
}

// listMembers returns the members of GET /members, by UUID.
func (test *TestHelper) listMembers(t *testing.T, token, query string) map[string]map[string]interface{} {
	t.Helper()
	code, body := test.duesRequest("GET", "/members"+query, token, nil)
	if err := test.checkResponseCode(http.StatusOK, code); err != nil {
		t.Fatal(err)
	}
	var members []map[string]interface{}
	if err := json.Unmarshal(body, &members); err != nil {
		t.Fatal(err)
	}
	byUUID := map[string]map[string]interface{}{}
	for _, member := range members {
		byUUID[member["uuid"].(string)] = member
	}
	return byUUID
}

// A deleted member is only listed when asked for, with the day of their purge.
func TestDeletedMemberListedOnRequest(t *testing.T) {
	h.clearTables()
	adminToken := h.addAnAdmin()
	h.addAMember()

	before := time.Now().Unix()
	code, _ := h.duesRequest("DELETE", "/members/deadbeef", adminToken, nil)
	if err := h.checkResponseCode(http.StatusOK, code); err != nil {
		t.Fatal(err)
	}

	if _, listed := h.listMembers(t, adminToken, "")["deadbeef"]; listed {
		t.Error("Expected the deleted member not to be listed by default")
	}
	if _, listed := h.listMembers(t, adminToken, "?status=active,created,paused")["deadbeef"]; listed {
		t.Error("Expected the deleted member not to be listed with the other statuses")
	}

	deleted, listed := h.listMembers(t, adminToken, "?status=deleted")["deadbeef"]
	if !listed {
		t.Fatal("Expected the deleted member to be listed when asked for")
	}
	if deleted["firstName"] != "Ramon" || deleted["status"] != model.MEMBERSSTATUSDELETED {
		t.Errorf("Expected Ramon, deleted. Got '%v', '%v'", deleted["firstName"], deleted["status"])
	}
	deletedAt := int64(deleted["deletedAt"].(float64))
	purgeAt := int64(deleted["purgeAt"].(float64))
	if deletedAt < before || deletedAt > time.Now().Unix() {
		t.Errorf("Expected deletedAt to be now. Got %d", deletedAt)
	}
	if purgeAt != deletedAt+purgeDelayForTests {
		t.Errorf("Expected the purge 90 days after the deletion. Got %d seconds after", purgeAt-deletedAt)
	}
	if _, listed := h.listMembers(t, adminToken, "?status=deleted")["deadfeed"]; listed {
		t.Error("Expected only deleted members when asking for them")
	}
}

// Another member only learns the name of a former member and that they left.
func TestFormerMemberSeenByMember(t *testing.T) {
	h.clearTables()
	memberToken := h.addAMember()
	h.addMember("abc123", "Joan", "Marti", "170", "70", "note", "baix", "member", "joan@marti.ca", "514")
	h.setMemberDeletedAt("abc123", time.Now().Unix())

	members := h.listMembers(t, memberToken, "?status=active,created,deleted")
	former, listed := members["abc123"]
	if !listed {
		t.Fatal("Expected the deleted member to be listed when asked for")
	}
	if former["firstName"] != "Joan" || former["status"] != model.MEMBERSSTATUSDELETED {
		t.Errorf("Expected Joan, deleted. Got '%v', '%v'", former["firstName"], former["status"])
	}
	if former["email"] != "" || former["height"] != "" || former["contact"] != "" || former["extra"] != "" {
		t.Errorf("Expected no private data of the former member. Got %v", former)
	}
	if self, listed := members["deadbeef"]; !listed || self["status"] != "" {
		t.Errorf("Expected the status of a current member to stay private. Got %v", self)
	}
}

func TestRestoreMember(t *testing.T) {
	h.clearTables()
	adminToken := h.addAnAdmin()
	memberToken := h.addAMember()
	h.addMember("abc123", "Joan", "Marti", "", "", "", "", "member", "joan@marti.ca", "")
	h.addMember("123", "Pol", "Petit", "", "", "", "", "canalla", "", "")

	// Only a deleted member can be restored
	if code, _ := h.post("/members/abc123/restore", adminToken, ""); code != http.StatusNotFound {
		t.Errorf("Expected 404 when restoring a member who is not deleted. Got %d", code)
	}

	for _, uuid := range []string{"deadbeef", "abc123", "123"} {
		if code, _ := h.duesRequest("DELETE", "/members/"+uuid, adminToken, nil); code != http.StatusOK {
			t.Fatalf("Expected to delete member %s. Got %d", uuid, code)
		}
	}

	// Only by an admin. The session of the deleted member has ended anyway.
	if code, _ := h.post("/members/abc123/restore", memberToken, ""); code != http.StatusUnauthorized {
		t.Errorf("Expected 401 when a member restores somebody. Got %d", code)
	}
	if code, _ := h.post("/members/abc123/restore", "", ""); code != http.StatusUnauthorized {
		t.Errorf("Expected 401 when restoring without a token. Got %d", code)
	}

	// With a password: active again. Never chose one: created. Canalla: active.
	expected := map[string]string{
		"deadbeef": model.MEMBERSSTATUSACTIVATED,
		"abc123":   model.MEMBERSSTATUSCREATED,
		"123":      model.MEMBERSSTATUSACTIVATED,
	}
	for uuid, status := range expected {
		code, body := h.post("/members/"+uuid+"/restore", adminToken, "")
		if err := h.checkResponseCode(http.StatusOK, code); err != nil {
			t.Fatal(err)
		}
		var restored map[string]interface{}
		json.Unmarshal(body, &restored)
		if restored["status"] != status || restored["deletedAt"] != float64(0) || restored["purgeAt"] != float64(0) {
			t.Errorf("Expected member %s restored as %s. Got %v", uuid, status, restored)
		}
		if got := h.getMemberStatus(uuid); got != status {
			t.Errorf("Expected member %s to be %s. Got %s", uuid, status, got)
		}
		if h.countRows("SELECT COUNT(*) FROM members WHERE uuid = ? AND deleted_at IS NULL", uuid) != 1 {
			t.Errorf("Expected deleted_at of member %s to be emptied", uuid)
		}
	}
	if code, _ := h.duesRequest("GET", "/members/deadbeef", adminToken, nil); code != http.StatusOK {
		t.Errorf("Expected the restored member to be visible again. Got %d", code)
	}
	// They can log in again, with the same password
	if token := h.login("member", "member"); token == "" {
		t.Error("Expected the restored member to log in")
	}

	// Already restored
	if code, _ := h.post("/members/deadbeef/restore", adminToken, ""); code != http.StatusNotFound {
		t.Errorf("Expected 404 when restoring twice. Got %d", code)
	}
}

// The email of a deleted member can be given to somebody else in the meantime.
func TestRestoreMemberEmailTaken(t *testing.T) {
	h.clearTables()
	adminToken := h.addAnAdmin()
	h.addAMember()
	h.setMemberDeletedAt("deadbeef", time.Now().Unix())
	h.addMember("abc123", "Ramon", "Nou", "", "", "", "", "member", "Ramon@Gerard.ca", "")

	if code, _ := h.post("/members/deadbeef/restore", adminToken, ""); code != http.StatusConflict {
		t.Errorf("Expected 409 when the email is used by another member. Got %d", code)
	}
	if status := h.getMemberStatus("deadbeef"); status != model.MEMBERSSTATUSDELETED {
		t.Errorf("Expected the member to stay deleted. Got '%s'", status)
	}
}

// After the delay, nothing personal is left of a deleted member.
func TestPurgeDeletedMembers(t *testing.T) {
	h.clearTables()
	adminToken := h.addAnAdmin()
	h.addMember("deadbeef", "Ramon", "Gerard", "180", "80", "Cap de rengla", "segon,baix", "member", "ramon@gerard.ca", "514 555-1234")
	h.addCredentials("deadbeef", "member", "member")
	h.addMember("abc123", "Joan", "Marti", "", "", "", "", "member", "joan@marti.ca", "")
	h.addMember("123", "Pol", "Petit", "", "", "", "", "canalla", "", "")
	if code, _ := h.post("/members/deadbeef/dependents/123", adminToken, ""); code != http.StatusCreated {
		t.Fatalf("Expected to add a dependent. Got %d", code)
	}
	collections := h.createDuesSeason(t, adminToken)
	if code, _ := h.recordDues(t, adminToken, "deadbeef", model.DuesActionPaid, "paid by his sister", collections[model.DuesKindWinter]); code != http.StatusOK {
		t.Fatalf("Expected to record dues. Got %d", code)
	}
	now := time.Now().Unix()

	// Deleted yesterday, and deleted 89 days ago: too early
	h.setMemberDeletedAt("deadbeef", now-purgeDelayForTests+24*3600)
	h.setMemberDeletedAt("abc123", now-24*3600)
	controller.RunPurgeDeletedMembersOnce()
	for _, uuid := range []string{"deadbeef", "abc123"} {
		if status := h.getMemberStatus(uuid); status != model.MEMBERSSTATUSDELETED {
			t.Errorf("Expected member %s to stay deleted before the delay. Got '%s'", uuid, status)
		}
	}
	// The dependent of a deleted member is still theirs, in case they come back
	if h.countRows("SELECT COUNT(*) FROM members_dependent WHERE responsible_uuid = 'deadbeef'") != 1 {
		t.Error("Expected the dependent to be kept before the purge")
	}

	// 90 days ago: purged. The other deleted member and the others are untouched.
	h.setMemberDeletedAt("deadbeef", now-purgeDelayForTests-60)
	controller.RunPurgeDeletedMembersOnce()
	if status := h.getMemberStatus("deadbeef"); status != model.MEMBERSSTATUSPURGED {
		t.Fatalf("Expected the member to be purged. Got '%s'", status)
	}
	if status := h.getMemberStatus("abc123"); status != model.MEMBERSSTATUSDELETED {
		t.Errorf("Expected the member deleted yesterday to stay deleted. Got '%s'", status)
	}
	if status := h.getMemberStatus("deadfeed"); status == model.MEMBERSSTATUSPURGED || status == model.MEMBERSSTATUSDELETED {
		t.Errorf("Expected the admin to be untouched. Got '%s'", status)
	}
	if h.countRows("SELECT COUNT(*) FROM members_credentials WHERE uuid = 'deadbeef'") != 0 {
		t.Error("Expected the credentials of the purged member to be removed")
	}
	if h.countRows("SELECT COUNT(*) FROM members_credentials WHERE uuid = 'deadfeed'") != 1 {
		t.Error("Expected the credentials of the admin to be kept")
	}
	if h.countRows("SELECT COUNT(*) FROM members_dependent WHERE responsible_uuid = 'deadbeef' OR dependent_uuid = 'deadbeef'") != 0 {
		t.Error("Expected the links with dependents of the purged member to be removed")
	}
	if h.countRows("SELECT COUNT(*) FROM dues_log WHERE member_uuid = 'deadbeef'") != 1 {
		t.Error("Expected the dues of the purged member to be kept")
	}
	if h.countRows("SELECT COUNT(*) FROM dues_log WHERE member_uuid = 'deadbeef' AND note IS NOT NULL") != 0 {
		t.Error("Expected the notes about the dues of the purged member to be removed")
	}
	if h.countRows("SELECT COUNT(*) FROM members WHERE uuid = 'deadbeef' AND shirt_size IS NULL AND shirt_since IS NULL AND musician = 0 AND subscribed = 0") != 1 {
		t.Error("Expected the shirt, musician and subscription of the purged member to be emptied")
	}

	// Listed when asked for, without anything personal
	purged, listed := h.listMembers(t, adminToken, "?status=deleted,purged")["deadbeef"]
	if !listed {
		t.Fatal("Expected the purged member to be listed when asked for")
	}
	if purged["status"] != model.MEMBERSSTATUSPURGED || purged["type"] != "member" {
		t.Errorf("Expected a purged member. Got %v", purged)
	}
	for _, field := range []string{"firstName", "lastName", "height", "weight", "extra", "email", "contact"} {
		if purged[field] != "" {
			t.Errorf("Expected %s of the purged member to be empty. Got '%v'", field, purged[field])
		}
	}
	if roles, ok := purged["roles"].([]interface{}); !ok || len(roles) != 0 {
		t.Errorf("Expected no role for the purged member. Got %v", purged["roles"])
	}
	if _, listed := h.listMembers(t, adminToken, "")["deadbeef"]; listed {
		t.Error("Expected the purged member not to be listed by default")
	}
	if code, _ := h.duesRequest("GET", "/members/deadbeef", adminToken, nil); code != http.StatusNotFound {
		t.Errorf("Expected 404 for the profile of a purged member. Got %d", code)
	}
	// The journal of the dues keeps the entry, without a name
	log := h.duesLog(t, adminToken, "?member=deadbeef")
	if len(log) != 1 || log[0].MemberName != "" || log[0].Note != "" {
		t.Errorf("Expected 1 anonymous entry in the journal. Got %v", log)
	}

	// Too late to restore
	if code, _ := h.post("/members/deadbeef/restore", adminToken, ""); code != http.StatusNotFound {
		t.Errorf("Expected 404 when restoring a purged member. Got %d", code)
	}
	// Running again changes nothing
	controller.RunPurgeDeletedMembersOnce()
	if status := h.getMemberStatus("deadbeef"); status != model.MEMBERSSTATUSPURGED {
		t.Errorf("Expected the member to stay purged. Got '%s'", status)
	}
}

// A deleted dependent does not break the list of dependents of their parent.
func TestDeletedDependentIsSkipped(t *testing.T) {
	h.clearTables()
	adminToken := h.addAnAdmin()
	h.addMember("123", "Pol", "Petit", "", "", "", "", "canalla", "", "")
	h.addMember("456", "Nil", "Petit", "", "", "", "", "canalla", "", "")
	for _, uuid := range []string{"123", "456"} {
		if code, _ := h.post("/members/deadfeed/dependents/"+uuid, adminToken, ""); code != http.StatusCreated {
			t.Fatalf("Expected to add dependent %s. Got %d", uuid, code)
		}
	}
	h.setMemberDeletedAt("123", time.Now().Unix())

	dependents, err := (&model.Member{UUID: "deadfeed"}).GetDependents(t.Context())
	if err != nil {
		t.Fatalf("Expected the dependents despite a deleted one. Got error %v", err)
	}
	if len(dependents) != 1 || dependents[0].UUID != "456" {
		t.Errorf("Expected only the dependent who is not deleted. Got %v", dependents)
	}
}
