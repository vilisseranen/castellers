package controller

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/vilisseranen/castellers/common"
	"github.com/vilisseranen/castellers/model"
)

func GetMember(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	uuid := vars["member_uuid"]
	m := model.Member{UUID: uuid}
	if err := m.Get(); err != nil {
		RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}
	RespondWithJSON(w, http.StatusOK, m)
}

func CreateMember(w http.ResponseWriter, r *http.Request) {
	var m model.Member
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&m); err != nil {
		RespondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}
	defer r.Body.Close()
	m.UUID = common.GenerateUUID()
	m.Code = common.GenerateCode()
	if err := m.CreateMember(); err != nil {
		RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}
	RespondWithJSON(w, http.StatusCreated, m)
}
