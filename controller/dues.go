package controller

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	netmail "net/mail"
	"strconv"
	"strings"

	"github.com/gorilla/mux"

	"github.com/vilisseranen/castellers/common"
	"github.com/vilisseranen/castellers/mail"
	"github.com/vilisseranen/castellers/model"
)

const (
	ERRORGETDUES                = "error getting dues"
	ERRORSETDUESCOLLECTIONS     = "error saving dues collections"
	ERRORRECORDDUES             = "error recording dues"
	ERRORDUESSEASON             = "invalid season"
	ERRORDUESKIND               = "invalid or duplicate collection kind"
	ERRORDUESAMOUNT             = "invalid collection amount or due date"
	ERRORDUESACTION             = "invalid action"
	ERRORDUESCOLLECTIONS        = "collectionUuids required"
	ERRORDUESCOLLECTIONNOTFOUND = "collection not found"
	ERRORDUESNOTETOOLONG        = "note is too long"
	ERRORDUESMEMBERTYPE         = "dues do not apply to this type of member"
	ERRORDUESINTERACEMAIL       = "invalid email address"
	ERRORSETDUESSETTINGS        = "error saving dues settings"
)

// Seasons accepted by the API: wide enough for any real use, narrow enough to
// reject a typo.
const (
	duesSeasonMin = 2000
	duesSeasonMax = 2100
)

type duesCollectionsRequest struct {
	Collections []model.DuesCollection `json:"collections"`
}

type duesEntriesRequest struct {
	MemberUUID      string   `json:"memberUuid"`
	CollectionUUIDs []string `json:"collectionUuids"`
	Action          string   `json:"action"`
	Note            string   `json:"note"`
}

// memberDues is the state of one member for one collection, as shown to that
// member: it does not say who recorded it.
type memberDues struct {
	Collection model.DuesCollection `json:"collection"`
	Status     string               `json:"status"`
	UpdatedAt  int64                `json:"updatedAt"`
}

// seasonFromQuery reads the optional "season" query parameter. It returns 0
// when it is absent and false when it is not a valid season.
func seasonFromQuery(r *http.Request) (int, bool) {
	value := r.FormValue("season")
	if value == "" {
		return 0, true
	}
	season, err := strconv.Atoi(value)
	if err != nil || season < duesSeasonMin || season > duesSeasonMax {
		return 0, false
	}
	return season, true
}

// duesApplyTo tells whether a member pays dues: regular members and admins do,
// the canalla and guests do not.
func duesApplyTo(m model.Member) bool {
	return m.Type == model.MEMBERSTYPEREGULAR || m.Type == model.MEMBERSTYPEADMIN
}

// GetDuesCollections returns the collections (amounts and dates), for one
// season or for all of them. Any member can read them.
func GetDuesCollections(w http.ResponseWriter, r *http.Request) {
	ctx, span := tracer.Start(r.Context(), "GetDuesCollections")
	defer span.End()

	season, ok := seasonFromQuery(r)
	if !ok {
		RespondWithError(w, http.StatusBadRequest, ERRORDUESSEASON)
		return
	}
	collections, err := model.GetDuesCollections(ctx, season)
	if err != nil {
		common.Warn("Error getting dues collections: %s", err.Error())
		RespondWithError(w, http.StatusInternalServerError, ERRORGETDUES)
		return
	}
	RespondWithJSON(w, http.StatusOK, collections)
}

// SetDuesCollections creates or updates the collections of a season (admin
// only). Kinds that are not in the payload are left as they are.
func SetDuesCollections(w http.ResponseWriter, r *http.Request) {
	ctx, span := tracer.Start(r.Context(), "SetDuesCollections")
	defer span.End()

	season, err := strconv.Atoi(mux.Vars(r)["season"])
	if err != nil || season < duesSeasonMin || season > duesSeasonMax {
		RespondWithError(w, http.StatusBadRequest, ERRORDUESSEASON)
		return
	}

	var body duesCollectionsRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		RespondWithError(w, http.StatusBadRequest, ERRORINVALIDPAYLOAD)
		return
	}
	if len(body.Collections) == 0 {
		RespondWithError(w, http.StatusBadRequest, ERRORMISSINGFIELDS)
		return
	}
	kinds := map[string]bool{}
	for _, c := range body.Collections {
		if !model.ValidDuesKind(c.Kind) || kinds[c.Kind] {
			RespondWithError(w, http.StatusBadRequest, ERRORDUESKIND)
			return
		}
		kinds[c.Kind] = true
		if c.Amount < 0 || c.DueDate <= 0 {
			RespondWithError(w, http.StatusBadRequest, ERRORDUESAMOUNT)
			return
		}
	}

	if err := model.SetDuesCollections(ctx, season, body.Collections); err != nil {
		common.Warn("Error saving dues collections: %s", err.Error())
		RespondWithError(w, http.StatusInternalServerError, ERRORSETDUESCOLLECTIONS)
		return
	}
	collections, err := model.GetDuesCollections(ctx, season)
	if err != nil {
		common.Warn("Error getting dues collections: %s", err.Error())
		RespondWithError(w, http.StatusInternalServerError, ERRORGETDUES)
		return
	}
	RespondWithJSON(w, http.StatusOK, collections)
}

// GetMemberDues returns the dues of a member: every collection with the state
// of the member for it. A member can only read their own, an admin can read
// anyone's.
func GetMemberDues(w http.ResponseWriter, r *http.Request) {
	ctx, span := tracer.Start(r.Context(), "GetMemberDues")
	defer span.End()

	memberUUID := mux.Vars(r)["member_uuid"]

	tokenAuth, err := ExtractToken(ctx, r)
	if err != nil {
		RespondWithTokenError(w, err)
		return
	}
	if !common.StringInSlice(model.MEMBERSTYPEADMIN, tokenAuth.Permissions) && memberUUID != tokenAuth.UserId {
		RespondWithError(w, http.StatusForbidden, ERRORUNAUTHORIZED)
		return
	}

	m := model.Member{UUID: memberUUID}
	if err := m.Get(ctx); err != nil {
		switch err {
		case sql.ErrNoRows:
			RespondWithError(w, http.StatusNotFound, ERRORMEMBERNOTFOUND)
		default:
			common.Warn("Error getting member: %s", err.Error())
			RespondWithError(w, http.StatusInternalServerError, ERRORGETDUES)
		}
		return
	}

	collections, err := model.GetDuesCollections(ctx, 0)
	if err != nil {
		common.Warn("Error getting dues collections: %s", err.Error())
		RespondWithError(w, http.StatusInternalServerError, ERRORGETDUES)
		return
	}
	states, err := model.GetDuesStates(ctx, 0, memberUUID)
	if err != nil {
		common.Warn("Error getting dues states: %s", err.Error())
		RespondWithError(w, http.StatusInternalServerError, ERRORGETDUES)
		return
	}
	stateByCollection := map[string]model.DuesState{}
	for _, s := range states {
		stateByCollection[s.CollectionUUID] = s
	}

	dues := make([]memberDues, 0, len(collections))
	for _, c := range collections {
		d := memberDues{Collection: c, Status: model.DuesActionUnpaid}
		if s, ok := stateByCollection[c.UUID]; ok {
			d.Status = s.Status
			d.UpdatedAt = s.UpdatedAt
		}
		dues = append(dues, d)
	}
	RespondWithJSON(w, http.StatusOK, dues)
}

// GetDuesStates returns the current state of every member for the collections
// of a season, or of all seasons (admin only). Only the pairs that were
// recorded at least once are returned: the others are unpaid.
func GetDuesStates(w http.ResponseWriter, r *http.Request) {
	ctx, span := tracer.Start(r.Context(), "GetDuesStates")
	defer span.End()

	season, ok := seasonFromQuery(r)
	if !ok {
		RespondWithError(w, http.StatusBadRequest, ERRORDUESSEASON)
		return
	}
	states, err := model.GetDuesStates(ctx, season, "")
	if err != nil {
		common.Warn("Error getting dues states: %s", err.Error())
		RespondWithError(w, http.StatusInternalServerError, ERRORGETDUES)
		return
	}
	RespondWithJSON(w, http.StatusOK, states)
}

// RecordDues marks one or several collections as paid, unpaid or exempt for a
// member (admin only). Each change is written in the journal with its author.
func RecordDues(w http.ResponseWriter, r *http.Request) {
	ctx, span := tracer.Start(r.Context(), "RecordDues")
	defer span.End()

	var body duesEntriesRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		RespondWithError(w, http.StatusBadRequest, ERRORINVALIDPAYLOAD)
		return
	}
	if body.MemberUUID == "" {
		RespondWithError(w, http.StatusBadRequest, ERRORMISSINGFIELDS)
		return
	}
	if len(body.CollectionUUIDs) == 0 {
		RespondWithError(w, http.StatusBadRequest, ERRORDUESCOLLECTIONS)
		return
	}
	if !model.ValidDuesAction(body.Action) {
		RespondWithError(w, http.StatusBadRequest, ERRORDUESACTION)
		return
	}
	if len([]rune(body.Note)) > model.DuesNoteMaxLength {
		RespondWithError(w, http.StatusBadRequest, ERRORDUESNOTETOOLONG)
		return
	}

	m := model.Member{UUID: body.MemberUUID}
	if err := m.Get(ctx); err != nil {
		switch err {
		case sql.ErrNoRows:
			RespondWithError(w, http.StatusNotFound, ERRORMEMBERNOTFOUND)
		default:
			common.Warn("Error getting member: %s", err.Error())
			RespondWithError(w, http.StatusInternalServerError, ERRORRECORDDUES)
		}
		return
	}
	if !duesApplyTo(m) {
		RespondWithError(w, http.StatusBadRequest, ERRORDUESMEMBERTYPE)
		return
	}

	tokenAuth, err := ExtractToken(ctx, r)
	if err != nil {
		RespondWithTokenError(w, err)
		return
	}

	written, err := model.RecordDues(ctx, body.MemberUUID, body.CollectionUUIDs, body.Action, tokenAuth.UserId, body.Note)
	if err != nil {
		if err == model.ErrDuesCollectionNotFound {
			RespondWithError(w, http.StatusNotFound, ERRORDUESCOLLECTIONNOTFOUND)
			return
		}
		common.Warn("Error recording dues: %s", err.Error())
		RespondWithError(w, http.StatusInternalServerError, ERRORRECORDDUES)
		return
	}
	RespondWithJSON(w, http.StatusOK, map[string]int{"entries": written})
}

// GetDuesLog returns the journal of the dues (admin only), most recent entry
// first. Optional filters: member, collection, actor, season; paging with
// limit and offset.
func GetDuesLog(w http.ResponseWriter, r *http.Request) {
	ctx, span := tracer.Start(r.Context(), "GetDuesLog")
	defer span.End()

	season, ok := seasonFromQuery(r)
	if !ok {
		RespondWithError(w, http.StatusBadRequest, ERRORDUESSEASON)
		return
	}
	limit, _ := strconv.Atoi(r.FormValue("limit"))
	offset, _ := strconv.Atoi(r.FormValue("offset"))

	entries, err := model.GetDuesLog(ctx, model.DuesLogFilter{
		MemberUUID:     r.FormValue("member"),
		CollectionUUID: r.FormValue("collection"),
		ActorUUID:      r.FormValue("actor"),
		Season:         season,
		Limit:          limit,
		Offset:         offset,
	})
	if err != nil {
		common.Warn("Error getting dues log: %s", err.Error())
		RespondWithError(w, http.StatusInternalServerError, ERRORGETDUES)
		return
	}
	RespondWithJSON(w, http.StatusOK, entries)
}

type duesSettings struct {
	InteracEmail string `json:"interacEmail"`
}

// validDuesCollection tells whether an event can point to this collection: no
// collection at all is fine, an unknown one is not.
func validDuesCollection(ctx context.Context, collectionUUID string) bool {
	if collectionUUID == "" {
		return true
	}
	_, err := model.GetDuesCollection(ctx, collectionUUID)
	return err == nil
}

// DuesReminderFor returns what the reminder of an event should tell a member
// about dues, or nil when there is nothing to say: no dues are collected at the
// event, dues do not apply to the member, or the member has settled them or is
// exempt.
func DuesReminderFor(ctx context.Context, event model.Event, member model.Member) (*mail.DuesReminder, error) {
	if event.DuesCollectionUUID == "" || !duesApplyTo(member) {
		return nil, nil
	}
	collection, err := model.GetDuesCollection(ctx, event.DuesCollectionUUID)
	if err == model.ErrDuesCollectionNotFound {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	status, err := model.GetDuesStatus(ctx, member.UUID, collection.UUID)
	if err != nil {
		return nil, err
	}
	if status != model.DuesActionUnpaid {
		return nil, nil
	}
	interacEmail, err := model.GetDuesSetting(ctx, model.DuesSettingInteracEmail)
	if err != nil {
		return nil, err
	}
	return &mail.DuesReminder{
		Kind:         collection.Kind,
		Amount:       collection.Amount,
		InteracEmail: interacEmail,
	}, nil
}

// GetDuesSettings returns what members need to know to pay their dues: for
// now the address for Interac transfers. Any member can read it.
func GetDuesSettings(w http.ResponseWriter, r *http.Request) {
	ctx, span := tracer.Start(r.Context(), "GetDuesSettings")
	defer span.End()

	interacEmail, err := model.GetDuesSetting(ctx, model.DuesSettingInteracEmail)
	if err != nil {
		common.Warn("Error getting dues settings: %s", err.Error())
		RespondWithError(w, http.StatusInternalServerError, ERRORGETDUES)
		return
	}
	RespondWithJSON(w, http.StatusOK, duesSettings{InteracEmail: interacEmail})
}

// SetDuesSettings saves the settings of the dues (admin only). An empty
// address removes it.
func SetDuesSettings(w http.ResponseWriter, r *http.Request) {
	ctx, span := tracer.Start(r.Context(), "SetDuesSettings")
	defer span.End()

	var body duesSettings
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		RespondWithError(w, http.StatusBadRequest, ERRORINVALIDPAYLOAD)
		return
	}
	interacEmail := strings.TrimSpace(body.InteracEmail)
	if interacEmail != "" {
		address, err := netmail.ParseAddress(interacEmail)
		if err != nil || address.Address != interacEmail || len(interacEmail) > 254 {
			RespondWithError(w, http.StatusBadRequest, ERRORDUESINTERACEMAIL)
			return
		}
	}
	if err := model.SetDuesSetting(ctx, model.DuesSettingInteracEmail, interacEmail); err != nil {
		common.Warn("Error saving dues settings: %s", err.Error())
		RespondWithError(w, http.StatusInternalServerError, ERRORSETDUESSETTINGS)
		return
	}
	RespondWithJSON(w, http.StatusOK, duesSettings{InteracEmail: interacEmail})
}

// GetMemberAttendance returns the number of events at which the presence of a
// member was confirmed, and the date of the first one. A member can only read
// their own, an admin can read anyone's.
func GetMemberAttendance(w http.ResponseWriter, r *http.Request) {
	ctx, span := tracer.Start(r.Context(), "GetMemberAttendance")
	defer span.End()

	memberUUID := mux.Vars(r)["member_uuid"]

	tokenAuth, err := ExtractToken(ctx, r)
	if err != nil {
		RespondWithTokenError(w, err)
		return
	}
	if !common.StringInSlice(model.MEMBERSTYPEADMIN, tokenAuth.Permissions) && memberUUID != tokenAuth.UserId {
		RespondWithError(w, http.StatusForbidden, ERRORUNAUTHORIZED)
		return
	}

	attendance, err := model.GetAttendance(ctx, memberUUID)
	if err != nil {
		common.Warn("Error getting attendance: %s", err.Error())
		RespondWithError(w, http.StatusInternalServerError, ERRORGETDUES)
		return
	}
	result := model.Attendance{MemberUUID: memberUUID}
	if len(attendance) > 0 {
		result = attendance[0]
	}
	RespondWithJSON(w, http.StatusOK, result)
}

// GetDuesAttendance returns the attendance of every member who came at least
// once (admin only).
func GetDuesAttendance(w http.ResponseWriter, r *http.Request) {
	ctx, span := tracer.Start(r.Context(), "GetDuesAttendance")
	defer span.End()

	attendance, err := model.GetAttendance(ctx, "")
	if err != nil {
		common.Warn("Error getting attendance: %s", err.Error())
		RespondWithError(w, http.StatusInternalServerError, ERRORGETDUES)
		return
	}
	RespondWithJSON(w, http.StatusOK, attendance)
}
