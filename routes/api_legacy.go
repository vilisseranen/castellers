package routes

import (
	"github.com/gorilla/mux"
	"github.com/vilisseranen/castellers/controller"
	"github.com/vilisseranen/castellers/model"
)

func AttachLegacyAPI(r *mux.Router) {
	// No auth
	r.HandleFunc("/api/initialize", controller.Initialize).Methods("POST")
	r.HandleFunc("/api/initialize", controller.IsInitialized).Methods("GET")
	r.HandleFunc("/api/events", controller.GetEvents).Methods("GET")
	r.HandleFunc("/api/events/{uuid:[0-9a-f]+}", controller.GetEvent).Methods("GET")
	r.HandleFunc("/api/roles", controller.GetRoles).Methods("GET")
	r.HandleFunc("/api/login", controller.Login).Methods("POST")
	r.HandleFunc("/api/logout", controller.Logout).Methods("POST")
	r.HandleFunc("/api/refresh", controller.RefreshToken).Methods("POST")
	r.HandleFunc("/api/forgot_password", controller.ForgotPassword).Methods("POST")
	r.HandleFunc("/api/version", controller.Version).Methods("GET")

	// Special tokens
	r.HandleFunc("/api/reset_credentials", checkTokenType(controller.ResetCredentials, controller.ResetCredentialsPermission)).Methods("POST")

	// Requires a token with member permission
	r.HandleFunc("/api/members/{member_uuid:[0-9a-f]+}", checkTokenType(controller.GetMember, model.MEMBERSTYPEREGULAR)).Methods("GET")
	r.HandleFunc("/api/events/{event_uuid:[0-9a-f]+}/members/{member_uuid:[0-9a-f]+}", checkTokenType(controller.ParticipateEvent, model.MEMBERSTYPEREGULAR, controller.ParticipateEventPermission)).Methods("POST")
	r.HandleFunc("/api/members/{member_uuid:[0-9a-f]+}", checkTokenType(controller.EditMember, model.MEMBERSTYPEREGULAR)).Methods("PUT")
	r.HandleFunc("/api/members/{member_uuid:[0-9a-f]+}/events", checkTokenType(controller.GetEvents, model.MEMBERSTYPEREGULAR)).Methods("GET")
	r.HandleFunc("/api/members/{member_uuid:[0-9a-f]+}/change_password", checkTokenType(controller.ResetCredentials, model.MEMBERSTYPEREGULAR)).Methods("POST")

	// Requires a token with admin permission
	r.HandleFunc("/api/admins/{admin_uuid:[0-9a-f]+}/events", checkTokenType(controller.CreateEvent, model.MEMBERSTYPEADMIN)).Methods("POST")
	r.HandleFunc("/api/admins/{admin_uuid:[0-9a-f]+}/events", checkTokenType(controller.GetEvents, model.MEMBERSTYPEADMIN)).Methods("GET")
	r.HandleFunc("/api/admins/{admin_uuid:[0-9a-f]+}/events/{uuid:[0-9a-f]+}", checkTokenType(controller.UpdateEvent, model.MEMBERSTYPEADMIN)).Methods("PUT")
	r.HandleFunc("/api/admins/{admin_uuid:[0-9a-f]+}/events/{uuid:[0-9a-f]+}", checkTokenType(controller.DeleteEvent, model.MEMBERSTYPEADMIN)).Methods("DELETE")
	r.HandleFunc("/api/admins/{admin_uuid:[0-9a-f]+}/events/{event_uuid:[0-9a-f]+}/members", checkTokenType(controller.GetEventParticipation, model.MEMBERSTYPEADMIN)).Methods("GET")
	r.HandleFunc("/api/admins/{admin_uuid:[0-9a-f]+}/events/{event_uuid:[0-9a-f]+}/members/{member_uuid:[0-9a-f]+}", checkTokenType(controller.PresenceEvent, model.MEMBERSTYPEADMIN)).Methods("POST")
	r.HandleFunc("/api/admins/{admin_uuid:[0-9a-f]+}/members", checkTokenType(controller.CreateMember, model.MEMBERSTYPEADMIN)).Methods("POST")
	r.HandleFunc("/api/admins/{admin_uuid:[0-9a-f]+}/members", checkTokenType(controller.GetMembers, model.MEMBERSTYPEADMIN)).Methods("GET")
	r.HandleFunc("/api/admins/{admin_uuid:[0-9a-f]+}/members/{member_uuid:[0-9a-f]+}", checkTokenType(controller.GetMember, model.MEMBERSTYPEADMIN)).Methods("GET")
	r.HandleFunc("/api/admins/{admin_uuid:[0-9a-f]+}/members/{member_uuid:[0-9a-f]+}", checkTokenType(controller.EditMember, model.MEMBERSTYPEADMIN)).Methods("PUT")
	r.HandleFunc("/api/admins/{admin_uuid:[0-9a-f]+}/members/{member_uuid:[0-9a-f]+}", checkTokenType(controller.DeleteMember, model.MEMBERSTYPEADMIN)).Methods("DELETE")
	r.HandleFunc("/api/admins/{admin_uuid:[0-9a-f]+}/members/{member_uuid:[0-9a-f]+}/registration", checkTokenType(controller.SendRegistrationEmail, model.MEMBERSTYPEADMIN)).Methods("GET")

	// castells API

	// public
	r.HandleFunc("/api/castells/types", controller.GetCastellTypeList).Methods("GET")
	r.HandleFunc("/api/castells/type/{type:[0-9]+d[0-9]+}", controller.GetCastellType).Methods("GET")

	// member
	r.HandleFunc("/api/castells/model/{uuid:[0-9a-f]+}", checkTokenType(controller.GetCastellModel, model.MEMBERSTYPEREGULAR)).Methods("GET")
	r.HandleFunc("/api/castells/models", checkTokenType(controller.GetCastellModels, model.MEMBERSTYPEREGULAR)).Methods("GET")

	// admin
	r.HandleFunc("/api/castells/model", checkTokenType(controller.CreateCastellModel, model.MEMBERSTYPEADMIN)).Methods("POST")
	r.HandleFunc("/api/castells/model/{uuid:[0-9a-f]+}", checkTokenType(controller.DeleteCastellModel, model.MEMBERSTYPEADMIN)).Methods("DELETE")
	r.HandleFunc("/api/castells/model/{uuid:[0-9a-f]+}", checkTokenType(controller.EditCastellModel, model.MEMBERSTYPEADMIN)).Methods("PUT")
}
