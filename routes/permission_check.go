package routes

import (
	"database/sql"
	"net/http"

	"github.com/gorilla/mux"

	"github.com/vilisseranen/castellers/common"
	"github.com/vilisseranen/castellers/controller"
	"github.com/vilisseranen/castellers/model"
)

type handler func(w http.ResponseWriter, r *http.Request)

func checkAdmin(h handler) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		uuid := vars["admin_uuid"]
		member := model.Member{UUID: uuid}
		if err := member.Get(); err != nil {
			switch err {
			case sql.ErrNoRows:
				controller.RespondWithError(w, http.StatusUnauthorized, controller.UnauthorizedMessage)
				return
			default:
				controller.RespondWithError(w, http.StatusInternalServerError, err.Error())
				return
			}
		}
		code := r.Header.Get("X-Member-Code")
		if code != member.Code {
			controller.RespondWithError(w, http.StatusUnauthorized, controller.UnauthorizedMessage)
			return
		}
		if member.Type != model.MemberTypeAdmin {
			controller.RespondWithError(w, http.StatusUnauthorized, controller.UnauthorizedMessage)
			return
		}
		r.Header.Add("Permission", member.Type)
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
				controller.RespondWithError(w, http.StatusUnauthorized, controller.UnauthorizedMessage)
				return
			default:
				controller.RespondWithError(w, http.StatusInternalServerError, err.Error())
				return
			}
		}
		code := r.Header.Get("X-Member-Code")
		if code != member.Code {
			controller.RespondWithError(w, http.StatusUnauthorized, controller.UnauthorizedMessage)
			return
		}
		if member.Activated == 0 {
			if err := member.Activate(); err != nil {
				controller.RespondWithError(w, http.StatusInternalServerError, err.Error())
				return
			}
		}
		r.Header.Add("Permission", member.Type)
		h(w, r)
	}
}

func checkTokenType(h handler, requestedType string) handler {
	return func(w http.ResponseWriter, r *http.Request) {
		tokenAuth, err := controller.ExtractToken(r)
		if err != nil {
			common.Debug("Token invalid: %s", err.Error())
			controller.RespondWithError(w, http.StatusUnauthorized, controller.UnauthorizedMessage)
			return
		}
		if !common.StringInSlice(requestedType, tokenAuth.Permissions) {
			controller.RespondWithError(w, http.StatusUnauthorized, controller.UnauthorizedMessage)
			return
		}
		if requestedType == model.MemberTypeMember {
			vars := mux.Vars(r)
			uuid := vars["member_uuid"]
			if uuid != tokenAuth.UserId {
				controller.RespondWithError(w, http.StatusUnauthorized, controller.UnauthorizedMessage)
				return
			}
		}
		h(w, r)
	}
}
