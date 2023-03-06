package routes

import (
	"net/http"

	"github.com/vilisseranen/castellers/common"
	"github.com/vilisseranen/castellers/controller"
)

type handler func(w http.ResponseWriter, r *http.Request)

func checkTokenType(h handler, requestedType ...string) handler {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, span := tracer.Start(r.Context(), "checkTokenType")
		defer span.End()
		common.Debug("Validating token in Authorization Header: %s", r.Header.Get("Authorization"))
		tokenAuth, err := controller.ExtractToken(ctx, r)
		if err != nil {
			// TODO: find a better way to determine if the token has expired.
			if err.Error() == "Token is expired" {
				controller.RespondWithError(w, http.StatusForbidden, controller.ERRORUNAUTHORIZED)
			} else {
				controller.RespondWithError(w, http.StatusUnauthorized, controller.ERRORUNAUTHORIZED)
			}
			common.Debug("Token invalid: %s", err.Error())
			return
		}
		if !common.StringInBothSlices(requestedType, tokenAuth.Permissions) {
			common.Warn("Token not allowed to access this resource")
			controller.RespondWithError(w, http.StatusUnauthorized, controller.ERRORUNAUTHORIZED)
			return
		}
		// Move this validation in the controller for resources accessing to members only converning themselves
		// if common.StringInSlice(model.MEMBERSTYPEREGULAR, requestedType) {
		// 	vars := mux.Vars(r)
		// 	uuid := vars["member_uuid"]
		// 	if uuid != "" && uuid != tokenAuth.UserId {
		// 		common.Error("Token not allowed to access this resource 2")
		// 		controller.RespondWithError(w, http.StatusUnauthorized, controller.ERRORUNAUTHORIZED)
		// 		return
		// 	}
		// }
		common.Debug("Token is valid and match required type")
		span.End()
		h(w, r)
	}
}
