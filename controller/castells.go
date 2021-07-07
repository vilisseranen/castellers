package controller

import (
	"database/sql"
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/vilisseranen/castellers/common"
	"github.com/vilisseranen/castellers/model"
)

func GetCastellType(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	castell_name := vars["type"]
	castell_type := model.CastellType{Name: castell_name}
	if err := castell_type.Get(); err != nil {
		switch err {
		case sql.ErrNoRows:
			RespondWithError(w, http.StatusNotFound, "Castell not found")
		default:
			RespondWithError(w, http.StatusInternalServerError, err.Error())
		}
		return
	}
	RespondWithJSON(w, http.StatusOK, castell_type)
}

func GetCastellTypeList(w http.ResponseWriter, r *http.Request) {
	castell_type := model.CastellType{}
	castell_type_list, err := castell_type.GetTypeList()
	if err != nil {
		switch err {
		default:
			RespondWithError(w, http.StatusInternalServerError, err.Error())
		}
		return
	}
	RespondWithJSON(w, http.StatusOK, castell_type_list)
}

func CreateCastellModel(w http.ResponseWriter, r *http.Request) {
	var c model.CastellModel
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&c); err != nil {
		common.Debug("Error decoding castell: %s", err.Error())
		RespondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}
	defer r.Body.Close()
	c.UUID = common.GenerateUUID()
	// Create the model now
	if err := c.Create(); err != nil {
		RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}
	RespondWithJSON(w, http.StatusCreated, c)
}

func EditCastellModel(w http.ResponseWriter, r *http.Request) {
	var c model.CastellModel
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&c); err != nil {
		common.Debug("Error decoding castell: %s", err.Error())
		RespondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}
	defer r.Body.Close()
	// Update the model now
	if err := c.Edit(); err != nil {
		RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}
	RespondWithJSON(w, http.StatusAccepted, c)
}

func GetCastellModels(w http.ResponseWriter, r *http.Request) {
	m := model.CastellModel{}
	models, err := m.GetAll()
	if err != nil {
		switch err {
		default:
			RespondWithError(w, http.StatusInternalServerError, err.Error())
		}
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
		default:
			RespondWithError(w, http.StatusInternalServerError, err.Error())
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
			RespondWithError(w, http.StatusNotFound, "Castell not found")
		default:
			RespondWithError(w, http.StatusInternalServerError, err.Error())
		}
		return
	}
	if err := m.Delete(); err != nil {
		RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}
	RespondWithJSON(w, http.StatusOK, nil)
}
