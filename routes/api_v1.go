package routes

import (
	"fmt"

	"github.com/gorilla/mux"

	"github.com/vilisseranen/castellers/controller"
	"github.com/vilisseranen/castellers/model"
)

func AttachV1API(r *mux.Router) {

	const (
		API       = "v1"
		BASE_PATH = "/api/" + API
	)
	// castells API

	r.HandleFunc(fmt.Sprintf("%s/castells/types", BASE_PATH), controller.GetCastellTypeList).Methods("GET")
	r.HandleFunc(fmt.Sprintf("%s/castells/types/{type:[0-9]+d[0-9]+}", BASE_PATH), controller.GetCastellType).Methods("GET")
	r.HandleFunc(fmt.Sprintf("%s/castells/models", BASE_PATH), checkTokenType(controller.GetCastellModels, model.MEMBERSTYPEREGULAR)).Methods("GET")
	r.HandleFunc(fmt.Sprintf("%s/castells/models", BASE_PATH), checkTokenType(controller.GetCastellModels, model.MEMBERSTYPEREGULAR)).Queries("event", "{event_uuid:[0-9a-f]+}").Methods("GET")
	r.HandleFunc(fmt.Sprintf("%s/castells/models", BASE_PATH), checkTokenType(controller.CreateCastellModel, model.MEMBERSTYPEADMIN)).Methods("POST")
	r.HandleFunc(fmt.Sprintf("%s/castells/models/{uuid:[0-9a-f]+}", BASE_PATH), checkTokenType(controller.GetCastellModel, model.MEMBERSTYPEREGULAR)).Methods("GET")
	r.HandleFunc(fmt.Sprintf("%s/castells/models/{uuid:[0-9a-f]+}", BASE_PATH), checkTokenType(controller.DeleteCastellModel, model.MEMBERSTYPEADMIN)).Methods("DELETE")
	r.HandleFunc(fmt.Sprintf("%s/castells/models/{uuid:[0-9a-f]+}", BASE_PATH), checkTokenType(controller.EditCastellModel, model.MEMBERSTYPEADMIN)).Methods("PUT")
	r.HandleFunc(fmt.Sprintf("%s/castells/models/{model_uuid:[0-9a-f]+}/events/{event_uuid:[0-9a-f]+}", BASE_PATH), checkTokenType(controller.AttachCastellModelToEvent, model.MEMBERSTYPEADMIN)).Methods("POST")
	r.HandleFunc(fmt.Sprintf("%s/castells/models/{model_uuid:[0-9a-f]+}/events/{event_uuid:[0-9a-f]+}", BASE_PATH), checkTokenType(controller.DettachCastellModelFromEvent, model.MEMBERSTYPEADMIN)).Methods("DELETE")

	// Initialize, login, tokens, version
	r.HandleFunc(fmt.Sprintf("%s/initialize", BASE_PATH), controller.Initialize).Methods("POST")
	r.HandleFunc(fmt.Sprintf("%s/initialize", BASE_PATH), controller.IsInitialized).Methods("GET")
	r.HandleFunc(fmt.Sprintf("%s/login", BASE_PATH), controller.Login).Methods("POST")
	r.HandleFunc(fmt.Sprintf("%s/logout", BASE_PATH), controller.Logout).Methods("POST")
	r.HandleFunc(fmt.Sprintf("%s/refresh", BASE_PATH), controller.RefreshToken).Methods("POST")
	r.HandleFunc(fmt.Sprintf("%s/forgot_password", BASE_PATH), controller.ForgotPassword).Methods("POST")
	r.HandleFunc(fmt.Sprintf("%s/version", BASE_PATH), controller.Version).Methods("GET")
	r.HandleFunc(fmt.Sprintf("%s/reset_credentials", BASE_PATH), checkTokenType(controller.ResetCredentials, controller.ResetCredentialsPermission)).Methods("POST")
	r.HandleFunc(fmt.Sprintf("%s/change_password", BASE_PATH), checkTokenType(controller.ResetCredentials, model.MEMBERSTYPEREGULAR)).Methods("POST")

	// Events
	r.HandleFunc(fmt.Sprintf("%s/events", BASE_PATH), controller.GetEvents).Methods("GET")
	r.HandleFunc(fmt.Sprintf("%s/events/{uuid:[0-9a-f]+}", BASE_PATH), controller.GetEvent).Methods("GET")
	r.HandleFunc(fmt.Sprintf("%s/events", BASE_PATH), checkTokenType(controller.CreateEvent, model.MEMBERSTYPEADMIN)).Methods("POST")
	r.HandleFunc(fmt.Sprintf("%s/events/{uuid:[0-9a-f]+}", BASE_PATH), checkTokenType(controller.UpdateEvent, model.MEMBERSTYPEADMIN)).Methods("PUT")
	r.HandleFunc(fmt.Sprintf("%s/events/{uuid:[0-9a-f]+}", BASE_PATH), checkTokenType(controller.DeleteEvent, model.MEMBERSTYPEADMIN)).Methods("DELETE")
	r.HandleFunc(fmt.Sprintf("%s/events/{event_uuid:[0-9a-f]+}/members", BASE_PATH), checkTokenType(controller.GetEventParticipation, model.MEMBERSTYPEADMIN)).Methods("GET")
	r.HandleFunc(fmt.Sprintf("%s/events/{event_uuid:[0-9a-f]+}/members/{member_uuid:[0-9a-f]+}", BASE_PATH), checkTokenType(controller.PresenceEvent, model.MEMBERSTYPEADMIN)).Methods("POST")

	// Members
	r.HandleFunc(fmt.Sprintf("%s/members", BASE_PATH), checkTokenType(controller.GetMembers, model.MEMBERSTYPEADMIN)).Methods("GET")
	r.HandleFunc(fmt.Sprintf("%s/members", BASE_PATH), checkTokenType(controller.CreateMember, model.MEMBERSTYPEADMIN)).Methods("POST")
	r.HandleFunc(fmt.Sprintf("%s/members/roles", BASE_PATH), controller.GetRoles).Methods("GET")
	r.HandleFunc(fmt.Sprintf("%s/members/{member_uuid:[0-9a-f]+}", BASE_PATH), checkTokenType(controller.GetMember, model.MEMBERSTYPEREGULAR)).Methods("GET")
	r.HandleFunc(fmt.Sprintf("%s/members/{member_uuid:[0-9a-f]+}", BASE_PATH), checkTokenType(controller.EditMember, model.MEMBERSTYPEREGULAR)).Methods("PUT")
	r.HandleFunc(fmt.Sprintf("%s/members/{member_uuid:[0-9a-f]+}", BASE_PATH), checkTokenType(controller.DeleteMember, model.MEMBERSTYPEADMIN)).Methods("DELETE")
	r.HandleFunc(fmt.Sprintf("%s/members/{member_uuid:[0-9a-f]+}/registration", BASE_PATH), checkTokenType(controller.SendRegistrationEmail, model.MEMBERSTYPEADMIN)).Methods("GET")
	r.HandleFunc(fmt.Sprintf("%s/members/events/{event_uuid:[0-9a-f]+}", BASE_PATH), checkTokenType(controller.ParticipateEvent, model.MEMBERSTYPEREGULAR, controller.ParticipateEventPermission)).Methods("POST")

}
