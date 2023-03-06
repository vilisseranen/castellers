package routes

import (
	"github.com/gorilla/mux"

	"github.com/vilisseranen/castellers/controller"
	"github.com/vilisseranen/castellers/model"
)

func AttachV1API(r *mux.Router) {

	API_VERSION := "v1"

	// castells API

	s := r.PathPrefix(BASE_PATH + API_VERSION).Subrouter()

	s.HandleFunc("/castells/types", controller.GetCastellTypeList).Methods("GET")
	s.HandleFunc("/castells/types/{type:[0-9]+d[0-9]+}", controller.GetCastellType).Methods("GET")
	s.HandleFunc("/castells/models", checkTokenType(controller.GetCastellModels, model.MEMBERSTYPEREGULAR)).Methods("GET")
	s.HandleFunc("/castells/models", checkTokenType(controller.CreateCastellModel, model.MEMBERSTYPEADMIN)).Methods("POST")
	s.HandleFunc("/castells/models/{uuid:[0-9a-f]+}", checkTokenType(controller.GetCastellModel, model.MEMBERSTYPEREGULAR)).Methods("GET")
	s.HandleFunc("/castells/models/{uuid:[0-9a-f]+}", checkTokenType(controller.DeleteCastellModel, model.MEMBERSTYPEADMIN)).Methods("DELETE")
	s.HandleFunc("/castells/models/{uuid:[0-9a-f]+}", checkTokenType(controller.EditCastellModel, model.MEMBERSTYPEADMIN)).Methods("PUT")
	s.HandleFunc("/castells/models/{model_uuid:[0-9a-f]+}/events/{event_uuid:[0-9a-f]+}", checkTokenType(controller.AttachCastellModelToEvent, model.MEMBERSTYPEADMIN)).Methods("POST")
	s.HandleFunc("/castells/models/{model_uuid:[0-9a-f]+}/events/{event_uuid:[0-9a-f]+}", checkTokenType(controller.DettachCastellModelFromEvent, model.MEMBERSTYPEADMIN)).Methods("DELETE")

	// Initialize, login, tokens, version
	s.HandleFunc("/initialize", controller.Initialize).Methods("POST")
	s.HandleFunc("/initialize", controller.IsInitialized).Methods("GET")
	s.HandleFunc("/login", controller.Login).Methods("POST")
	s.HandleFunc("/logout", controller.Logout).Methods("POST")
	s.HandleFunc("/refresh", controller.RefreshToken).Methods("POST")
	s.HandleFunc("/forgot_password", controller.ForgotPassword).Methods("POST")
	s.HandleFunc("/version", controller.Version).Methods("GET")
	s.HandleFunc("/reset_credentials", checkTokenType(controller.ResetCredentials, controller.ResetCredentialsPermission)).Methods("POST")
	s.HandleFunc("/change_password", checkTokenType(controller.ResetCredentials, model.MEMBERSTYPEREGULAR)).Methods("POST")

	// Events
	s.HandleFunc("/events", controller.GetEvents).Methods("GET")
	s.HandleFunc("/events/{uuid:[0-9a-f]+}", controller.GetEvent).Methods("GET")
	s.HandleFunc("/events", checkTokenType(controller.CreateEvent, model.MEMBERSTYPEADMIN)).Methods("POST")
	s.HandleFunc("/events/{uuid:[0-9a-f]+}", checkTokenType(controller.UpdateEvent, model.MEMBERSTYPEADMIN)).Methods("PUT")
	s.HandleFunc("/events/{uuid:[0-9a-f]+}", checkTokenType(controller.DeleteEvent, model.MEMBERSTYPEADMIN)).Methods("DELETE")
	s.HandleFunc("/events/{event_uuid:[0-9a-f]+}/members", checkTokenType(controller.GetEventParticipation, model.MEMBERSTYPEADMIN)).Methods("GET")
	s.HandleFunc("/events/{event_uuid:[0-9a-f]+}/members/{member_uuid:[0-9a-f]+}", checkTokenType(controller.PresenceEvent, model.MEMBERSTYPEADMIN)).Methods("POST")

	// Members
	s.HandleFunc("/members", checkTokenType(controller.GetMembers, model.MEMBERSTYPEREGULAR)).Methods("GET")
	s.HandleFunc("/members", checkTokenType(controller.CreateMember, model.MEMBERSTYPEADMIN)).Methods("POST")
	s.HandleFunc("/members/roles", controller.GetRoles).Methods("GET")
	s.HandleFunc("/members/{member_uuid:[0-9a-f]+}", checkTokenType(controller.GetMember, model.MEMBERSTYPEREGULAR)).Methods("GET")
	s.HandleFunc("/members/{member_uuid:[0-9a-f]+}", checkTokenType(controller.EditMember, model.MEMBERSTYPEREGULAR)).Methods("PUT")
	s.HandleFunc("/members/{member_uuid:[0-9a-f]+}", checkTokenType(controller.DeleteMember, model.MEMBERSTYPEADMIN)).Methods("DELETE")
	s.HandleFunc("/members/{member_uuid:[0-9a-f]+}/registration", checkTokenType(controller.SendRegistrationEmail, model.MEMBERSTYPEADMIN)).Methods("GET")
	s.HandleFunc("/members/{member_uuid:[0-9a-f]+}/events/{event_uuid:[0-9a-f]+}", checkTokenType(controller.ParticipateEvent, model.MEMBERSTYPEREGULAR, controller.ParticipateEventPermission)).Methods("POST")
	s.HandleFunc("/members/{responsible_uuid:[0-9a-f]+}/dependents/{dependent_uuid:[0-9a-f]+}", checkTokenType(controller.AddRemoveDependent, model.MEMBERSTYPEADMIN)).Methods("POST")
	s.HandleFunc("/members/{responsible_uuid:[0-9a-f]+}/dependents/{dependent_uuid:[0-9a-f]+}", checkTokenType(controller.AddRemoveDependent, model.MEMBERSTYPEADMIN)).Methods("DELETE")

	// Deprecated
	s.HandleFunc("/members/events/{event_uuid:[0-9a-f]+}", checkTokenType(controller.ParticipateEvent, model.MEMBERSTYPEREGULAR, controller.ParticipateEventPermission)).Methods("POST")
}
