package tests

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	"github.com/vilisseranen/castellers/model"
)

const duesTestSeason = 2026

type memberDuesResponse struct {
	Collection model.DuesCollection `json:"collection"`
	Status     string               `json:"status"`
	UpdatedAt  int64                `json:"updatedAt"`
}

// login returns an access token for existing credentials.
func (test *TestHelper) login(username, password string) string {
	payload := []byte(fmt.Sprintf(`{"username":"%s","password":"%s"}`, username, password))
	req, _ := http.NewRequest("POST", "/api/v1/login", bytes.NewBuffer(payload))
	response := test.executeRequest(req)
	var t map[string]interface{}
	json.Unmarshal(response.Body.Bytes(), &t)
	return t["access_token"].(string)
}

func (test *TestHelper) duesRequest(method, url, token string, payload []byte) (int, []byte) {
	var req *http.Request
	if payload != nil {
		req, _ = http.NewRequest(method, "/api/v1"+url, bytes.NewBuffer(payload))
	} else {
		req, _ = http.NewRequest(method, "/api/v1"+url, nil)
	}
	req.Header.Add("Authorization", "Bearer "+token)
	response := test.executeRequest(req)
	return response.Code, response.Body.Bytes()
}

// createDuesSeason creates the four collections of the test season and returns
// their UUIDs by kind.
func (test *TestHelper) createDuesSeason(t *testing.T, adminToken string) map[string]string {
	t.Helper()
	payload := []byte(`{"collections":[
		{"kind":"summer","amount":1000,"dueDate":1783000000},
		{"kind":"cercle","amount":3500,"dueDate":1760000000},
		{"kind":"winter","amount":2000,"dueDate":1768000000},
		{"kind":"spring","amount":2000,"dueDate":1775000000}]}`)
	code, body := test.duesRequest("PUT", fmt.Sprintf("/dues/collections/%d", duesTestSeason), adminToken, payload)
	if err := test.checkResponseCode(http.StatusOK, code); err != nil {
		t.Fatal(err)
	}
	var collections []model.DuesCollection
	json.Unmarshal(body, &collections)
	uuids := map[string]string{}
	for _, c := range collections {
		uuids[c.Kind] = c.UUID
	}
	return uuids
}

func (test *TestHelper) recordDues(t *testing.T, adminToken, memberUUID, action, note string, collectionUUIDs ...string) (int, int) {
	t.Helper()
	payload, _ := json.Marshal(map[string]interface{}{
		"memberUuid":      memberUUID,
		"collectionUuids": collectionUUIDs,
		"action":          action,
		"note":            note,
	})
	code, body := test.duesRequest("POST", "/dues/entries", adminToken, payload)
	var result map[string]int
	json.Unmarshal(body, &result)
	return code, result["entries"]
}

func (test *TestHelper) memberDues(t *testing.T, token, memberUUID string) map[string]memberDuesResponse {
	t.Helper()
	code, body := test.duesRequest("GET", "/members/"+memberUUID+"/dues", token, nil)
	if err := test.checkResponseCode(http.StatusOK, code); err != nil {
		t.Fatal(err)
	}
	var dues []memberDuesResponse
	json.Unmarshal(body, &dues)
	byKind := map[string]memberDuesResponse{}
	for _, d := range dues {
		byKind[d.Collection.Kind] = d
	}
	return byKind
}

func (test *TestHelper) duesLog(t *testing.T, adminToken, query string) []model.DuesLogEntry {
	t.Helper()
	code, body := test.duesRequest("GET", "/dues/log"+query, adminToken, nil)
	if err := test.checkResponseCode(http.StatusOK, code); err != nil {
		t.Fatal(err)
	}
	var entries []model.DuesLogEntry
	json.Unmarshal(body, &entries)
	return entries
}

func TestDuesCollections(t *testing.T) {
	h.clearTables()
	adminToken := h.addAnAdmin()
	memberToken := h.addAMember()

	uuids := h.createDuesSeason(t, adminToken)
	if len(uuids) != 4 {
		t.Fatalf("Expected 4 collections. Got %d", len(uuids))
	}

	// Any member can read them, in the order of the season.
	code, body := h.duesRequest("GET", fmt.Sprintf("/dues/collections?season=%d", duesTestSeason), memberToken, nil)
	if err := h.checkResponseCode(http.StatusOK, code); err != nil {
		t.Fatal(err)
	}
	var collections []model.DuesCollection
	json.Unmarshal(body, &collections)
	if len(collections) != 4 {
		t.Fatalf("Expected 4 collections. Got %d", len(collections))
	}
	expectedOrder := []string{"cercle", "winter", "spring", "summer"}
	for i, kind := range expectedOrder {
		if collections[i].Kind != kind {
			t.Errorf("Expected collection %d to be '%s'. Got '%s'", i, kind, collections[i].Kind)
		}
	}
	if collections[0].Amount != 3500 {
		t.Errorf("Expected the cercle amount to be 3500. Got %d", collections[0].Amount)
	}

	// Another season has no collection.
	code, body = h.duesRequest("GET", "/dues/collections?season=2030", memberToken, nil)
	if err := h.checkResponseCode(http.StatusOK, code); err != nil {
		t.Error(err)
	}
	json.Unmarshal(body, &collections)
	if len(collections) != 0 {
		t.Errorf("Expected no collection in 2030. Got %d", len(collections))
	}

	// Updating one kind keeps its UUID and leaves the others untouched.
	code, body = h.duesRequest("PUT", fmt.Sprintf("/dues/collections/%d", duesTestSeason), adminToken,
		[]byte(`{"collections":[{"kind":"winter","amount":2500,"dueDate":1768100000}]}`))
	if err := h.checkResponseCode(http.StatusOK, code); err != nil {
		t.Fatal(err)
	}
	json.Unmarshal(body, &collections)
	if len(collections) != 4 {
		t.Fatalf("Expected 4 collections after the update. Got %d", len(collections))
	}
	for _, c := range collections {
		if c.UUID != uuids[c.Kind] {
			t.Errorf("Expected '%s' to keep its uuid", c.Kind)
		}
		if c.Kind == "winter" && (c.Amount != 2500 || c.DueDate != 1768100000) {
			t.Errorf("Expected winter to be updated. Got amount %d, dueDate %d", c.Amount, c.DueDate)
		}
		if c.Kind == "spring" && c.Amount != 2000 {
			t.Errorf("Expected spring to be untouched. Got amount %d", c.Amount)
		}
	}
}

func TestDuesCollectionsValidation(t *testing.T) {
	h.clearTables()
	adminToken := h.addAnAdmin()
	memberToken := h.addAMember()
	url := fmt.Sprintf("/dues/collections/%d", duesTestSeason)
	valid := []byte(`{"collections":[{"kind":"winter","amount":2000,"dueDate":1768000000}]}`)

	// A member cannot define collections.
	code, _ := h.duesRequest("PUT", url, memberToken, valid)
	if err := h.checkResponseCode(http.StatusUnauthorized, code); err != nil {
		t.Error(err)
	}

	invalid := map[string]string{
		"unknown kind":    `{"collections":[{"kind":"autumn","amount":2000,"dueDate":1768000000}]}`,
		"duplicate kind":  `{"collections":[{"kind":"winter","amount":2000,"dueDate":1768000000},{"kind":"winter","amount":1000,"dueDate":1768000000}]}`,
		"negative amount": `{"collections":[{"kind":"winter","amount":-1,"dueDate":1768000000}]}`,
		"no due date":     `{"collections":[{"kind":"winter","amount":2000}]}`,
		"no collection":   `{"collections":[]}`,
		"not json":        `nope`,
	}
	for name, payload := range invalid {
		code, _ := h.duesRequest("PUT", url, adminToken, []byte(payload))
		if err := h.checkResponseCode(http.StatusBadRequest, code); err != nil {
			t.Errorf("%s: %v", name, err)
		}
	}

	code, _ = h.duesRequest("PUT", "/dues/collections/1900", adminToken, valid)
	if err := h.checkResponseCode(http.StatusBadRequest, code); err != nil {
		t.Errorf("season out of range: %v", err)
	}
	code, _ = h.duesRequest("GET", "/dues/collections?season=abc", adminToken, nil)
	if err := h.checkResponseCode(http.StatusBadRequest, code); err != nil {
		t.Errorf("season not a number: %v", err)
	}
}

func TestRecordDuesAndMemberView(t *testing.T) {
	h.clearTables()
	adminToken := h.addAnAdmin()
	memberToken := h.addAMember()
	uuids := h.createDuesSeason(t, adminToken)

	// Nothing recorded yet: everything is unpaid.
	dues := h.memberDues(t, memberToken, "deadbeef")
	if len(dues) != 4 {
		t.Fatalf("Expected 4 dues. Got %d", len(dues))
	}
	for kind, d := range dues {
		if d.Status != model.DuesActionUnpaid || d.UpdatedAt != 0 {
			t.Errorf("Expected '%s' to be unpaid with no date. Got '%s', %d", kind, d.Status, d.UpdatedAt)
		}
	}

	// The whole year at once: one entry per collection.
	code, entries := h.recordDues(t, adminToken, "deadbeef", model.DuesActionPaid, "",
		uuids["winter"], uuids["spring"], uuids["summer"])
	if err := h.checkResponseCode(http.StatusOK, code); err != nil {
		t.Fatal(err)
	}
	if entries != 3 {
		t.Errorf("Expected 3 entries. Got %d", entries)
	}

	// Doing it again changes nothing and writes nothing.
	code, entries = h.recordDues(t, adminToken, "deadbeef", model.DuesActionPaid, "",
		uuids["winter"], uuids["spring"], uuids["summer"])
	if err := h.checkResponseCode(http.StatusOK, code); err != nil {
		t.Fatal(err)
	}
	if entries != 0 {
		t.Errorf("Expected no new entry. Got %d", entries)
	}
	if log := h.duesLog(t, adminToken, ""); len(log) != 3 {
		t.Errorf("Expected 3 entries in the journal. Got %d", len(log))
	}

	// The cercle is replaced by an exemption, winter is unticked.
	if code, _ := h.recordDues(t, adminToken, "deadbeef", model.DuesActionExempt, "", uuids["cercle"]); code != http.StatusOK {
		t.Fatalf("Expected 200 for the exemption. Got %d", code)
	}
	if code, _ := h.recordDues(t, adminToken, "deadbeef", model.DuesActionUnpaid, "", uuids["winter"]); code != http.StatusOK {
		t.Fatalf("Expected 200 for the unticking. Got %d", code)
	}

	// The member sees their own dues.
	dues = h.memberDues(t, memberToken, "deadbeef")
	expected := map[string]string{
		"cercle": model.DuesActionExempt,
		"winter": model.DuesActionUnpaid,
		"spring": model.DuesActionPaid,
		"summer": model.DuesActionPaid,
	}
	for kind, status := range expected {
		if dues[kind].Status != status {
			t.Errorf("Expected '%s' to be '%s'. Got '%s'", kind, status, dues[kind].Status)
		}
		if dues[kind].UpdatedAt == 0 {
			t.Errorf("Expected '%s' to have a date", kind)
		}
	}

	// An admin sees them too, and the states of everybody for the season.
	if adminView := h.memberDues(t, adminToken, "deadbeef"); adminView["spring"].Status != model.DuesActionPaid {
		t.Errorf("Expected the admin to see spring as paid. Got '%s'", adminView["spring"].Status)
	}
	code, body := h.duesRequest("GET", fmt.Sprintf("/dues/states?season=%d", duesTestSeason), adminToken, nil)
	if err := h.checkResponseCode(http.StatusOK, code); err != nil {
		t.Fatal(err)
	}
	var states []model.DuesState
	json.Unmarshal(body, &states)
	if len(states) != 4 {
		t.Fatalf("Expected 4 states. Got %d", len(states))
	}
	for _, s := range states {
		if s.MemberUUID != "deadbeef" {
			t.Errorf("Unexpected member in the states: %s", s.MemberUUID)
		}
		if s.CollectionUUID == uuids["winter"] && s.Status != model.DuesActionUnpaid {
			t.Errorf("Expected the state of winter to be unpaid. Got '%s'", s.Status)
		}
	}
	code, body = h.duesRequest("GET", "/dues/states?season=2030", adminToken, nil)
	if err := h.checkResponseCode(http.StatusOK, code); err != nil {
		t.Fatal(err)
	}
	json.Unmarshal(body, &states)
	if len(states) != 0 {
		t.Errorf("Expected no state in 2030. Got %d", len(states))
	}
}

func TestDuesPermissions(t *testing.T) {
	h.clearTables()
	adminToken := h.addAnAdmin()
	memberToken := h.addAMember()
	h.addMember("cafe", "Clément", "Other", "", "", "", "", "member", "clement@other.ca", "")
	h.addCredentials("cafe", "other", "other")
	otherToken := h.login("other", "other")
	uuids := h.createDuesSeason(t, adminToken)
	h.recordDues(t, adminToken, "deadbeef", model.DuesActionPaid, "", uuids["winter"])

	// A member cannot read the dues of somebody else.
	code, _ := h.duesRequest("GET", "/members/deadbeef/dues", otherToken, nil)
	if err := h.checkResponseCode(http.StatusForbidden, code); err != nil {
		t.Error(err)
	}
	// Nor without a token.
	req, _ := http.NewRequest("GET", "/api/v1/members/deadbeef/dues", nil)
	if err := h.checkResponseCode(http.StatusUnauthorized, h.executeRequest(req).Code); err != nil {
		t.Error(err)
	}
	// Unknown member.
	code, _ = h.duesRequest("GET", "/members/0123/dues", adminToken, nil)
	if err := h.checkResponseCode(http.StatusNotFound, code); err != nil {
		t.Error(err)
	}

	// Recording, the states and the journal are for admins only.
	code, _ = h.recordDues(t, memberToken, "deadbeef", model.DuesActionPaid, "", uuids["spring"])
	if err := h.checkResponseCode(http.StatusUnauthorized, code); err != nil {
		t.Error(err)
	}
	code, _ = h.duesRequest("GET", "/dues/states", memberToken, nil)
	if err := h.checkResponseCode(http.StatusUnauthorized, code); err != nil {
		t.Error(err)
	}
	code, _ = h.duesRequest("GET", "/dues/log", memberToken, nil)
	if err := h.checkResponseCode(http.StatusUnauthorized, code); err != nil {
		t.Error(err)
	}
	if dues := h.memberDues(t, memberToken, "deadbeef"); dues["spring"].Status != model.DuesActionUnpaid {
		t.Errorf("Expected spring to stay unpaid. Got '%s'", dues["spring"].Status)
	}
}

func TestRecordDuesValidation(t *testing.T) {
	h.clearTables()
	adminToken := h.addAnAdmin()
	h.addAMember()
	h.addMember("babe", "Petit", "Enxaneta", "", "", "", "", "canalla", "", "")
	h.addMember("beef", "Una", "Convidada", "", "", "", "", "guest", "", "")
	uuids := h.createDuesSeason(t, adminToken)

	// The canalla and guests do not pay dues.
	for _, uuid := range []string{"babe", "beef"} {
		code, _ := h.recordDues(t, adminToken, uuid, model.DuesActionPaid, "", uuids["winter"])
		if err := h.checkResponseCode(http.StatusBadRequest, code); err != nil {
			t.Errorf("member %s: %v", uuid, err)
		}
	}

	code, _ := h.recordDues(t, adminToken, "0123", model.DuesActionPaid, "", uuids["winter"])
	if err := h.checkResponseCode(http.StatusNotFound, code); err != nil {
		t.Errorf("unknown member: %v", err)
	}
	code, _ = h.recordDues(t, adminToken, "deadbeef", "refunded", "", uuids["winter"])
	if err := h.checkResponseCode(http.StatusBadRequest, code); err != nil {
		t.Errorf("unknown action: %v", err)
	}
	code, _ = h.recordDues(t, adminToken, "deadbeef", model.DuesActionPaid, "")
	if err := h.checkResponseCode(http.StatusBadRequest, code); err != nil {
		t.Errorf("no collection: %v", err)
	}
	code, _ = h.recordDues(t, adminToken, "", model.DuesActionPaid, "", uuids["winter"])
	if err := h.checkResponseCode(http.StatusBadRequest, code); err != nil {
		t.Errorf("no member: %v", err)
	}
	longNote := string(bytes.Repeat([]byte("a"), model.DuesNoteMaxLength+1))
	code, _ = h.recordDues(t, adminToken, "deadbeef", model.DuesActionPaid, longNote, uuids["winter"])
	if err := h.checkResponseCode(http.StatusBadRequest, code); err != nil {
		t.Errorf("note too long: %v", err)
	}

	// An unknown collection cancels the whole request.
	code, _ = h.recordDues(t, adminToken, "deadbeef", model.DuesActionPaid, "", uuids["winter"], "0123")
	if err := h.checkResponseCode(http.StatusNotFound, code); err != nil {
		t.Errorf("unknown collection: %v", err)
	}
	if log := h.duesLog(t, adminToken, ""); len(log) != 0 {
		t.Errorf("Expected an empty journal. Got %d entries", len(log))
	}
}

func TestDuesLog(t *testing.T) {
	h.clearTables()
	adminToken := h.addAnAdmin()
	h.addAMember()
	h.addMember("cafe", "Clément", "Other", "", "", "", "", "member", "clement@other.ca", "")
	h.addMember("feed", "Second", "Admin", "", "", "", "", "admin", "second@admin.ca", "")
	h.addCredentials("feed", "second", "second")
	secondAdminToken := h.login("second", "second")
	uuids := h.createDuesSeason(t, adminToken)

	h.recordDues(t, adminToken, "deadbeef", model.DuesActionPaid, "virement reçu le 3 février", uuids["winter"], uuids["spring"])
	h.recordDues(t, secondAdminToken, "cafe", model.DuesActionPaid, "", uuids["winter"])
	h.recordDues(t, secondAdminToken, "deadbeef", model.DuesActionUnpaid, "coché par erreur", uuids["spring"])

	// Most recent first, with names, action and note.
	log := h.duesLog(t, adminToken, "")
	if len(log) != 4 {
		t.Fatalf("Expected 4 entries. Got %d", len(log))
	}
	last := log[0]
	if last.MemberUUID != "deadbeef" || last.Action != model.DuesActionUnpaid || last.Kind != "spring" || last.Season != duesTestSeason {
		t.Errorf("Unexpected last entry: %+v", last)
	}
	if last.MemberName != "Ramon Gerard" {
		t.Errorf("Expected member name 'Ramon Gerard'. Got '%s'", last.MemberName)
	}
	if last.ActorUUID != "feed" || last.ActorName != "Second Admin" {
		t.Errorf("Expected the author to be 'Second Admin'. Got '%s' (%s)", last.ActorName, last.ActorUUID)
	}
	if last.Note != "coché par erreur" {
		t.Errorf("Expected note 'coché par erreur'. Got '%s'", last.Note)
	}
	if last.CreatedAt == 0 {
		t.Error("Expected the entry to have a date")
	}
	first := log[3]
	if first.ActorName != "Romà Èric" || first.Note != "virement reçu le 3 février" || first.Action != model.DuesActionPaid {
		t.Errorf("Unexpected first entry: %+v", first)
	}

	// The note is not stored in clear text.
	db, err := sql.Open("sqlite3", testDbName)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	var clearNotes int
	if err := db.QueryRow("SELECT COUNT(id) FROM dues_log WHERE CAST(note AS TEXT) LIKE '%virement%'").Scan(&clearNotes); err != nil {
		t.Fatal(err)
	}
	if clearNotes != 0 {
		t.Errorf("Expected notes to be encrypted. Found %d in clear text", clearNotes)
	}

	// Filters.
	filters := map[string]int{
		"?member=deadbeef":               3,
		"?member=cafe":                   1,
		"?actor=feed":                    2,
		"?collection=" + uuids["winter"]: 2,
		"?member=deadbeef&collection=" + uuids["spring"]: 2,
		fmt.Sprintf("?season=%d", duesTestSeason):        4,
		"?season=2030":      0,
		"?limit=1":          1,
		"?limit=2&offset=3": 1,
	}
	for query, expected := range filters {
		if entries := h.duesLog(t, adminToken, query); len(entries) != expected {
			t.Errorf("%s: expected %d entries. Got %d", query, expected, len(entries))
		}
	}

	// The journal keeps the name of an admin whose account was deleted.
	if _, err := db.Exec("UPDATE members SET status = 'deleted' WHERE uuid = 'feed'"); err != nil {
		t.Fatal(err)
	}
	log = h.duesLog(t, adminToken, "?actor=feed")
	if len(log) != 2 {
		t.Fatalf("Expected 2 entries by the deleted admin. Got %d", len(log))
	}
	if log[0].ActorName != "Second Admin" {
		t.Errorf("Expected the name of the deleted admin to be kept. Got '%s'", log[0].ActorName)
	}
}
