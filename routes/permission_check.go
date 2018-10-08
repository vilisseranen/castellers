package routes

import (
	"database/sql"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/vilisseranen/castellers/controller"
	"github.com/vilisseranen/castellers/model"
)

type handler func(w http.ResponseWriter, r *http.Request)

const unauthorizedMessage = "You are not authorized to perform this action."

func checkAdmin(h handler) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		uuid := vars["admin_uuid"]
		member := model.Member{UUID: uuid}
		if err := member.Get(); err != nil {
			switch err {
			case sql.ErrNoRows:
				controller.RespondWithError(w, http.StatusUnauthorized, unauthorizedMessage)
				return
			default:
				controller.RespondWithError(w, http.StatusInternalServerError, err.Error())
				return
			}
		}
		code := r.Header.Get("X-Member-Code")
		if code != member.Code {
			controller.RespondWithError(w, http.StatusUnauthorized, unauthorizedMessage)
			return
		}
		if member.Type != model.MEMBER_TYPE_ADMIN {
			controller.RespondWithError(w, http.StatusUnauthorized, unauthorizedMessage)
			return
		}
		h(w, r)
	}
}

func checkMember(h handler) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		uuid := vars["member_uuid"]
		member := model.Member{UUID: uuid}
		if err := member.Get(); err != nil {
			switch err {
			case sql.ErrNoRows:
				controller.RespondWithError(w, http.StatusUnauthorized, unauthorizedMessage)
				return
			default:
				controller.RespondWithError(w, http.StatusInternalServerError, err.Error())
				return
			}
		}
		code := r.Header.Get("X-Member-Code")
		if code != member.Code {
			controller.RespondWithError(w, http.StatusUnauthorized, unauthorizedMessage)
			return
		}
		if member.Activated == 0 {
			if err := member.Activate(); err != nil {
				controller.RespondWithError(w, http.StatusInternalServerError, err.Error())
				return
			}
		}
		h(w, r)
	}
}
