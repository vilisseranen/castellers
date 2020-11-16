package routes

import (
	"net/http"

	"github.com/gorilla/mux"

	"github.com/vilisseranen/castellers/common"
	"github.com/vilisseranen/castellers/controller"
	"github.com/vilisseranen/castellers/model"
)

type handler func(w http.ResponseWriter, r *http.Request)

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
