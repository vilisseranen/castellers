package controller

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/gorilla/mux"

	"github.com/vilisseranen/castellers/common"
	"github.com/vilisseranen/castellers/model"
)

const (
	ERRORGETEVENTTEMPLATES     = "error getting event templates"
	ERRORSAVEEVENTTEMPLATE     = "error saving event template"
	ERRORDELETEEVENTTEMPLATE   = "error deleting event template"
	ERROREVENTTEMPLATENOTFOUND = "event template not found"
	ERROREVENTTEMPLATETITLE    = "a template already has this title"
)

// Longest title of a template.
const eventTemplateTitleMaxLength = 100

// decodeEventTemplate reads a template from the request and checks it. It
// answers the request itself and returns false when the template is not valid.
func decodeEventTemplate(ctx context.Context, w http.ResponseWriter, r *http.Request, uuid string) (model.EventTemplate, bool) {
	var t model.EventTemplate
	if err := json.NewDecoder(r.Body).Decode(&t); err != nil {
		RespondWithError(w, http.StatusBadRequest, ERRORINVALIDPAYLOAD)
		return t, false
	}
	t.UUID = uuid
	t.Title = strings.TrimSpace(t.Title)
	if t.Title == "" || len([]rune(t.Title)) > eventTemplateTitleMaxLength ||
		!common.StringInSlice(t.Type, model.ValidEventTypes) ||
		t.StartDate > t.EndDate ||
		(t.UniformRequired != 0 && t.UniformRequired != 1) {
		RespondWithError(w, http.StatusBadRequest, ERRORINVALIDPAYLOAD)
		return t, false
	}
	taken, err := model.EventTemplateTitleTaken(ctx, t.Title, uuid)
	if err != nil {
		common.Warn("Error checking the title of an event template: %s", err.Error())
		RespondWithError(w, http.StatusInternalServerError, ERRORSAVEEVENTTEMPLATE)
		return t, false
	}
	if taken {
		RespondWithError(w, http.StatusConflict, ERROREVENTTEMPLATETITLE)
		return t, false
	}
	return t, true
}

// GetEventTemplates returns the templates proposed when creating an event
// (admin only).
func GetEventTemplates(w http.ResponseWriter, r *http.Request) {
	ctx, span := tracer.Start(r.Context(), "GetEventTemplates")
	defer span.End()

	templates, err := model.GetEventTemplates(ctx)
	if err != nil {
		common.Warn("Error getting event templates: %s", err.Error())
		RespondWithError(w, http.StatusInternalServerError, ERRORGETEVENTTEMPLATES)
		return
	}
	RespondWithJSON(w, http.StatusOK, templates)
}

// CreateEventTemplate saves a new template (admin only). Two templates cannot
// have the same title.
func CreateEventTemplate(w http.ResponseWriter, r *http.Request) {
	ctx, span := tracer.Start(r.Context(), "CreateEventTemplate")
	defer span.End()

	t, ok := decodeEventTemplate(ctx, w, r, "")
	if !ok {
		return
	}
	if err := t.Create(ctx); err != nil {
		common.Warn("Error creating event template: %s", err.Error())
		RespondWithError(w, http.StatusInternalServerError, ERRORSAVEEVENTTEMPLATE)
		return
	}
	RespondWithJSON(w, http.StatusCreated, t)
}

// UpdateEventTemplate replaces a template (admin only). Events created from
// it are not changed.
func UpdateEventTemplate(w http.ResponseWriter, r *http.Request) {
	ctx, span := tracer.Start(r.Context(), "UpdateEventTemplate")
	defer span.End()

	t, ok := decodeEventTemplate(ctx, w, r, mux.Vars(r)["uuid"])
	if !ok {
		return
	}
	if err := t.Update(ctx); err != nil {
		if err == model.ErrEventTemplateNotFound {
			RespondWithError(w, http.StatusNotFound, ERROREVENTTEMPLATENOTFOUND)
			return
		}
		common.Warn("Error updating event template: %s", err.Error())
		RespondWithError(w, http.StatusInternalServerError, ERRORSAVEEVENTTEMPLATE)
		return
	}
	RespondWithJSON(w, http.StatusOK, t)
}

// DeleteEventTemplate removes a template (admin only).
func DeleteEventTemplate(w http.ResponseWriter, r *http.Request) {
	ctx, span := tracer.Start(r.Context(), "DeleteEventTemplate")
	defer span.End()

	if err := model.DeleteEventTemplate(ctx, mux.Vars(r)["uuid"]); err != nil {
		if err == model.ErrEventTemplateNotFound {
			RespondWithError(w, http.StatusNotFound, ERROREVENTTEMPLATENOTFOUND)
			return
		}
		common.Warn("Error deleting event template: %s", err.Error())
		RespondWithError(w, http.StatusInternalServerError, ERRORDELETEEVENTTEMPLATE)
		return
	}
	RespondWithJSON(w, http.StatusOK, map[string]string{"result": "success"})
}
