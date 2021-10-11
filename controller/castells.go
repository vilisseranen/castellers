package controller

import (
	"database/sql"
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/vilisseranen/castellers/common"
	"github.com/vilisseranen/castellers/model"
	"go.elastic.co/apm"
)

const (
	ERRORGETCASTELLTYPE       = "Error getting castell type"
	ERRORGETCASTELLLIST       = "Error getting castell list"
	ERRORGETCASTELLMODEL      = "Error getting castell model"
	ERRORCASTELLMODELNOTFOUND = "Castell model not found"
	ERRORCASTELLTYPENOTFOUND  = "Castell type not found"
	ERRORDELETECASTELLMODEL   = "Error deleting castell model"
	ERRORCREATECASTELLMODEL   = "Error creating castell model"
	ERRORUPDATECASTELLMODEL   = "Error editing castell model"
	ERRORADDCASTELLTOEVENT    = "Error adding castell model to event"
	ERRORREMOVECASTELLTOEVENT = "Error removing castell model from event"
)

func GetCastellType(w http.ResponseWriter, r *http.Request) {
	span, ctx := apm.StartSpan(r.Context(), "GetCastellType", APM_SPAN_TYPE_REQUEST)
	defer span.End()
	vars := mux.Vars(r)
	castell_name := vars["type"]
	castell_type := model.CastellType{Name: castell_name}
	if err := castell_type.Get(ctx); err != nil {
		switch err {
		case sql.ErrNoRows:
			common.Debug("Castell type not found: %s", err.Error())
			RespondWithError(w, http.StatusNotFound, ERRORCASTELLTYPENOTFOUND)
		default:
			common.Warn("Castell type get error: %s", err.Error())
			RespondWithError(w, http.StatusInternalServerError, ERRORGETCASTELLTYPE)
		}
		return
	}
	RespondWithJSON(w, http.StatusOK, castell_type)
}

func GetCastellTypeList(w http.ResponseWriter, r *http.Request) {
	span, ctx := apm.StartSpan(r.Context(), "GetCastellTypeList", APM_SPAN_TYPE_REQUEST)
	defer span.End()
	castell_type := model.CastellType{}
	castell_type_list, err := castell_type.GetTypeList(ctx)
	if err != nil {
		common.Warn("Error getting castell list: %s", err.Error())
		RespondWithError(w, http.StatusInternalServerError, ERRORGETCASTELLLIST)
		return
	}
	RespondWithJSON(w, http.StatusOK, castell_type_list)
}

func CreateCastellModel(w http.ResponseWriter, r *http.Request) {
	span, ctx := apm.StartSpan(r.Context(), "CreateCastellModel", APM_SPAN_TYPE_REQUEST)
	defer span.End()
	var c model.CastellModel
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&c); err != nil {
		common.Debug("Error decoding castell: %s", err.Error())
		RespondWithError(w, http.StatusBadRequest, ERRORINVALIDPAYLOAD)
		return
	}
	defer r.Body.Close()
	c.UUID = common.GenerateUUID()
	// Validate input
	if c.Name == "" || c.Type == "" || len(c.PositionMembers) == 0 {
		common.Debug("Castell has empty name or type: %s", c)
		RespondWithError(w, http.StatusBadRequest, ERRORINVALIDPAYLOAD)
		return
	}
	// Create the model now
	if err := c.Create(ctx); err != nil {
		common.Warn("Cannot create castell model: %s", err.Error())
		RespondWithError(w, http.StatusInternalServerError, ERRORCREATECASTELLMODEL)
		return
	}
	RespondWithJSON(w, http.StatusCreated, c)
}

func EditCastellModel(w http.ResponseWriter, r *http.Request) {
	span, ctx := apm.StartSpan(r.Context(), "EditCastellModel", APM_SPAN_TYPE_REQUEST)
	defer span.End()
	var c model.CastellModel
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&c); err != nil {
		common.Debug("Error decoding castell: %s", err.Error())
		RespondWithError(w, http.StatusBadRequest, ERRORINVALIDPAYLOAD)
		return
	}
	defer r.Body.Close()
	// Validate input
	if c.Name == "" || c.Type == "" || len(c.PositionMembers) == 0 {
		common.Debug("Castell has empty name or type: %s", c)
		RespondWithError(w, http.StatusBadRequest, ERRORINVALIDPAYLOAD)
		return
	}
	// Update the model now
	if err := c.Edit(ctx); err != nil {
		common.Warn("Cannot update castell model: %s", err.Error())
		RespondWithError(w, http.StatusInternalServerError, ERRORUPDATECASTELLMODEL)
		return
	}
	RespondWithJSON(w, http.StatusAccepted, c)
}

func GetCastellModels(w http.ResponseWriter, r *http.Request) {
	span, ctx := apm.StartSpan(r.Context(), "GetCastellModels", APM_SPAN_TYPE_REQUEST)
	defer span.End()
	event := r.FormValue("event")
	m := model.CastellModel{}
	models := []model.CastellModel{}
	if event != "" {
		common.Debug("Getting models for event %s", event)
		e := model.Event{UUID: event}
		err := e.Get(ctx)
		if err != nil {
			common.Info("Error retrieving event: %s", err.Error())
			RespondWithError(w, http.StatusNotFound, ERROREVENTNOTFOUND)
			return
		}
		models, err = m.GetAllFromEvent(ctx, e)
		if err != nil {
			common.Warn("Cannot get castell models: %s", err.Error())
			RespondWithError(w, http.StatusInternalServerError, ERRORGETCASTELLMODEL)
			return
		}
		RespondWithJSON(w, http.StatusOK, models)
	} else {
		models, err := m.GetAll(ctx)
		if err != nil {
			common.Warn("Cannot get castell models: %s", err.Error())
			RespondWithError(w, http.StatusInternalServerError, ERRORGETCASTELLMODEL)
			return
		}
		RespondWithJSON(w, http.StatusOK, models)
	}
}

func GetCastellModel(w http.ResponseWriter, r *http.Request) {
	span, ctx := apm.StartSpan(r.Context(), "GetCastellModel", APM_SPAN_TYPE_REQUEST)
	defer span.End()
	vars := mux.Vars(r)
	castell_uuid := vars["uuid"]
	m := model.CastellModel{UUID: castell_uuid}
	if err := m.Get(ctx); err != nil {
		switch err {
		case sql.ErrNoRows:
			common.Debug("No castell model found: %s", err.Error())
			RespondWithError(w, http.StatusNotFound, ERRORCASTELLMODELNOTFOUND)
		default:
			common.Warn("Cannot get castell model: %s", err.Error())
			RespondWithError(w, http.StatusInternalServerError, ERRORGETCASTELLMODEL)
		}
		return
	}
	RespondWithJSON(w, http.StatusOK, m)
}

func DeleteCastellModel(w http.ResponseWriter, r *http.Request) {
	span, ctx := apm.StartSpan(r.Context(), "DeleteCastellModel", APM_SPAN_TYPE_REQUEST)
	defer span.End()
	vars := mux.Vars(r)
	castell_uuid := vars["uuid"]
	m := model.CastellModel{UUID: castell_uuid}
	if err := m.Get(ctx); err != nil {
		switch err {
		case sql.ErrNoRows:
			common.Debug("No castell model found: %s", err.Error())
			RespondWithError(w, http.StatusNotFound, ERRORCASTELLMODELNOTFOUND)
		default:
			common.Warn("Error getting castell model: %s", err.Error())
			RespondWithError(w, http.StatusInternalServerError, ERRORGETCASTELLMODEL)
		}
		return
	}
	if err := m.Delete(ctx); err != nil {
		common.Warn("Castell deleting castell: %s", err.Error())
		RespondWithError(w, http.StatusInternalServerError, ERRORDELETECASTELLMODEL)
		return
	}
	RespondWithJSON(w, http.StatusOK, nil)
}

func AttachCastellModelToEvent(w http.ResponseWriter, r *http.Request) {
	span, ctx := apm.StartSpan(r.Context(), "AttachCastellModelToEvent", APM_SPAN_TYPE_REQUEST)
	defer span.End()
	vars := mux.Vars(r)
	model_uuid := vars["model_uuid"]
	m := model.CastellModel{UUID: model_uuid}
	if err := m.Get(ctx); err != nil {
		switch err {
		case sql.ErrNoRows:
			common.Debug("No castell model found: %s", err.Error())
			RespondWithError(w, http.StatusNotFound, ERRORCASTELLMODELNOTFOUND)
		default:
			common.Warn("Error getting castell model: %s", err.Error())
			RespondWithError(w, http.StatusInternalServerError, ERRORGETCASTELLMODEL)
		}
		return
	}
	event_uuid := vars["event_uuid"]
	e := model.Event{UUID: event_uuid}
	if err := e.Get(ctx); err != nil {
		switch err {
		case sql.ErrNoRows:
			common.Debug("No event found: %s", err.Error())
			RespondWithError(w, http.StatusNotFound, ERROREVENTNOTFOUND)
		default:
			common.Warn("Error getting event: %s", err.Error())
			RespondWithError(w, http.StatusInternalServerError, ERRORGETEVENT)
		}
		return
	}
	if err := m.AttachToEvent(ctx, &e); err != nil {
		common.Warn("Error adding castell to event: %s", err.Error())
		RespondWithError(w, http.StatusInternalServerError, ERRORADDCASTELLTOEVENT)
		return
	}
	RespondWithJSON(w, http.StatusOK, nil)
}

func DettachCastellModelFromEvent(w http.ResponseWriter, r *http.Request) {
	span, ctx := apm.StartSpan(r.Context(), "DettachCastellModelFromEvent", APM_SPAN_TYPE_REQUEST)
	defer span.End()
	vars := mux.Vars(r)
	model_uuid := vars["model_uuid"]
	m := model.CastellModel{UUID: model_uuid}
	if err := m.Get(ctx); err != nil {
		switch err {
		case sql.ErrNoRows:
			common.Debug("No castell model found: %s", err.Error())
			RespondWithError(w, http.StatusNotFound, ERRORCASTELLMODELNOTFOUND)
		default:
			common.Warn("Error getting castell model: %s", err.Error())
			RespondWithError(w, http.StatusInternalServerError, ERRORGETCASTELLMODEL)
		}
		return
	}
	event_uuid := vars["event_uuid"]
	e := model.Event{UUID: event_uuid}
	if err := e.Get(ctx); err != nil {
		switch err {
		case sql.ErrNoRows:
			common.Debug("No event found: %s", err.Error())
			RespondWithError(w, http.StatusNotFound, ERROREVENTNOTFOUND)
		default:
			common.Warn("Error getting event: %s", err.Error())
			RespondWithError(w, http.StatusInternalServerError, ERRORGETEVENT)
		}
		return
	}
	if err := m.DettachFromEvent(ctx, &e); err != nil {
		common.Warn("Error dettaching castell from event: %s", err.Error())
		RespondWithError(w, http.StatusInternalServerError, ERRORREMOVECASTELLTOEVENT)
		return
	}
	RespondWithJSON(w, http.StatusOK, nil)
}
