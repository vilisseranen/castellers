package controller

import (
	"database/sql"
	"net/http"

	"github.com/gorilla/mux"
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

func GetCastellModel(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	castell_name := vars["name"]
	castell_model := model.CastellModel{Name: castell_name}
	if err := castell_model.Get(); err != nil {
		switch err {
		case sql.ErrNoRows:
			RespondWithError(w, http.StatusNotFound, "Castell not found")
		default:
			RespondWithError(w, http.StatusInternalServerError, err.Error())
		}
		return
	}
	RespondWithJSON(w, http.StatusOK, castell_model)
}
