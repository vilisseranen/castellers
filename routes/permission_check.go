package routes

import (
	"database/sql"
	"github.com/gorilla/mux"
	"github.com/vilisseranen/castellers/controller"
	"github.com/vilisseranen/castellers/model"
	"net/http"
)

type Handler func(w http.ResponseWriter, r *http.Request)

const UNAUTHORIZED_MESSAGE = "You are not authorized to perform this action."

func checkAdmin(h Handler) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		uuid := vars["admin_uuid"]
		member := model.Member{UUID: uuid}
		if err := member.Get(); err != nil {
			switch err {
			case sql.ErrNoRows:
				controller.RespondWithError(w, http.StatusUnauthorized, UNAUTHORIZED_MESSAGE)
				return
			default:
				controller.RespondWithError(w, http.StatusInternalServerError, err.Error())
				return
			}
		}
		code := r.Header.Get("X-Member-Code")
		if code != member.Code {
			controller.RespondWithError(w, http.StatusUnauthorized, UNAUTHORIZED_MESSAGE)
			return
		}
		if member.Type != model.MEMBER_TYPE_ADMIN {
			controller.RespondWithError(w, http.StatusUnauthorized, UNAUTHORIZED_MESSAGE)
			return
		}
		h(w, r)
	}
}

func checkMember(h Handler) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		uuid := vars["member_uuid"]
		member := model.Member{UUID: uuid}
		if err := member.Get(); err != nil {
			switch err {
			case sql.ErrNoRows:
				controller.RespondWithError(w, http.StatusUnauthorized, UNAUTHORIZED_MESSAGE)
				return
			default:
				controller.RespondWithError(w, http.StatusInternalServerError, err.Error())
				return
			}
		}
		code := r.Header.Get("X-Member-Code")
		if code != member.Code {
			controller.RespondWithError(w, http.StatusUnauthorized, UNAUTHORIZED_MESSAGE)
			return
		}
		h(w, r)
	}
}
