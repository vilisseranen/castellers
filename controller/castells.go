package controller

import (
	"database/sql"
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/vilisseranen/castellers/common"
	"github.com/vilisseranen/castellers/model"
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
	vars := mux.Vars(r)
	castell_name := vars["type"]
	castell_type := model.CastellType{Name: castell_name}
	if err := castell_type.Get(); err != nil {
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
	castell_type := model.CastellType{}
	castell_type_list, err := castell_type.GetTypeList()
	if err != nil {
		common.Warn("Error getting castell list: %s", err.Error())
		RespondWithError(w, http.StatusInternalServerError, ERRORGETCASTELLLIST)
		return
	}
	RespondWithJSON(w, http.StatusOK, castell_type_list)
}

func CreateCastellModel(w http.ResponseWriter, r *http.Request) {
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
	if err := c.Create(); err != nil {
		common.Warn("Cannot create castell model: %s", err.Error())
		RespondWithError(w, http.StatusInternalServerError, ERRORCREATECASTELLMODEL)
		return
	}
	RespondWithJSON(w, http.StatusCreated, c)
}

func EditCastellModel(w http.ResponseWriter, r *http.Request) {
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
	if err := c.Edit(); err != nil {
		common.Warn("Cannot update castell model: %s", err.Error())
		RespondWithError(w, http.StatusInternalServerError, ERRORUPDATECASTELLMODEL)
		return
	}
	RespondWithJSON(w, http.StatusAccepted, c)
}

func GetCastellModels(w http.ResponseWriter, r *http.Request) {
	m := model.CastellModel{}
	models, err := m.GetAll()
	if err != nil {
		common.Warn("Cannot get castell models: %s", err.Error())
		RespondWithError(w, http.StatusInternalServerError, ERRORGETCASTELLMODEL)
		return
	}
	RespondWithJSON(w, http.StatusOK, models)
}

func GetCastellModel(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	castell_uuid := vars["uuid"]
	m := model.CastellModel{UUID: castell_uuid}
	if err := m.Get(); err != nil {
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
	vars := mux.Vars(r)
	castell_uuid := vars["uuid"]
	m := model.CastellModel{UUID: castell_uuid}
	if err := m.Get(); err != nil {
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
	if err := m.Delete(); err != nil {
		common.Warn("Castell deleting castell: %s", err.Error())
		RespondWithError(w, http.StatusInternalServerError, ERRORDELETECASTELLMODEL)
		return
	}
	RespondWithJSON(w, http.StatusOK, nil)
}

func AttachCastellModelToEvent(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	model_uuid := vars["model_uuid"]
	m := model.CastellModel{UUID: model_uuid}
	if err := m.Get(); err != nil {
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
	if err := e.Get(); err != nil {
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
	if err := m.AttachToEvent(&e); err != nil {
		common.Warn("Error adding castell to event: %s", err.Error())
		RespondWithError(w, http.StatusInternalServerError, ERRORADDCASTELLTOEVENT)
		return
	}
	RespondWithJSON(w, http.StatusOK, nil)
}

func DettachCastellModelFromEvent(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	model_uuid := vars["model_uuid"]
	m := model.CastellModel{UUID: model_uuid}
	if err := m.Get(); err != nil {
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
	if err := e.Get(); err != nil {
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
	if err := m.DettachFromEvent(&e); err != nil {
		common.Warn("Error dettaching castell from event: %s", err.Error())
		RespondWithError(w, http.StatusInternalServerError, ERRORREMOVECASTELLTOEVENT)
		return
	}
	RespondWithJSON(w, http.StatusOK, nil)
}
