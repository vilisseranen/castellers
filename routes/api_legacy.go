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
	r.HandleFunc("/api/members/{member_uuid:[0-9a-f]+}", checkTokenType(controller.GetMember, model.MemberTypeMember)).Methods("GET")
	r.HandleFunc("/api/events/{event_uuid:[0-9a-f]+}/members/{member_uuid:[0-9a-f]+}", checkTokenType(controller.ParticipateEvent, model.MemberTypeMember, controller.ParticipateEventPermission)).Methods("POST")
	r.HandleFunc("/api/members/{member_uuid:[0-9a-f]+}", checkTokenType(controller.EditMember, model.MemberTypeMember)).Methods("PUT")
	r.HandleFunc("/api/members/{member_uuid:[0-9a-f]+}/events", checkTokenType(controller.GetEvents, model.MemberTypeMember)).Methods("GET")
	r.HandleFunc("/api/members/{member_uuid:[0-9a-f]+}/change_password", checkTokenType(controller.ResetCredentials, model.MemberTypeMember)).Methods("POST")

	// Requires a token with admin permission
	r.HandleFunc("/api/admins/{admin_uuid:[0-9a-f]+}/events", checkTokenType(controller.CreateEvent, model.MemberTypeAdmin)).Methods("POST")
	r.HandleFunc("/api/admins/{admin_uuid:[0-9a-f]+}/events", checkTokenType(controller.GetEvents, model.MemberTypeAdmin)).Methods("GET")
	r.HandleFunc("/api/admins/{admin_uuid:[0-9a-f]+}/events/{uuid:[0-9a-f]+}", checkTokenType(controller.UpdateEvent, model.MemberTypeAdmin)).Methods("PUT")
	r.HandleFunc("/api/admins/{admin_uuid:[0-9a-f]+}/events/{uuid:[0-9a-f]+}", checkTokenType(controller.DeleteEvent, model.MemberTypeAdmin)).Methods("DELETE")
	r.HandleFunc("/api/admins/{admin_uuid:[0-9a-f]+}/events/{event_uuid:[0-9a-f]+}/members", checkTokenType(controller.GetEventParticipation, model.MemberTypeAdmin)).Methods("GET")
	r.HandleFunc("/api/admins/{admin_uuid:[0-9a-f]+}/events/{event_uuid:[0-9a-f]+}/members/{member_uuid:[0-9a-f]+}", checkTokenType(controller.PresenceEvent, model.MemberTypeAdmin)).Methods("POST")
	r.HandleFunc("/api/admins/{admin_uuid:[0-9a-f]+}/members", checkTokenType(controller.CreateMember, model.MemberTypeAdmin)).Methods("POST")
	r.HandleFunc("/api/admins/{admin_uuid:[0-9a-f]+}/members", checkTokenType(controller.GetMembers, model.MemberTypeAdmin)).Methods("GET")
	r.HandleFunc("/api/admins/{admin_uuid:[0-9a-f]+}/members/{member_uuid:[0-9a-f]+}", checkTokenType(controller.GetMember, model.MemberTypeAdmin)).Methods("GET")
	r.HandleFunc("/api/admins/{admin_uuid:[0-9a-f]+}/members/{member_uuid:[0-9a-f]+}", checkTokenType(controller.EditMember, model.MemberTypeAdmin)).Methods("PUT")
	r.HandleFunc("/api/admins/{admin_uuid:[0-9a-f]+}/members/{member_uuid:[0-9a-f]+}", checkTokenType(controller.DeleteMember, model.MemberTypeAdmin)).Methods("DELETE")
	r.HandleFunc("/api/admins/{admin_uuid:[0-9a-f]+}/members/{member_uuid:[0-9a-f]+}/registration", checkTokenType(controller.SendRegistrationEmail, model.MemberTypeAdmin)).Methods("GET")

	// castells API

	// public
	r.HandleFunc("/api/castells/types", controller.GetCastellTypeList).Methods("GET")
	r.HandleFunc("/api/castells/type/{type:[0-9]+d[0-9]+}", controller.GetCastellType).Methods("GET")

	// member
	r.HandleFunc("/api/castells/model/{uuid:[0-9a-f]+}", checkTokenType(controller.GetCastellModel, model.MemberTypeMember)).Methods("GET")
	r.HandleFunc("/api/castells/models", checkTokenType(controller.GetCastellModels, model.MemberTypeMember)).Methods("GET")

	// admin
	r.HandleFunc("/api/castells/model", checkTokenType(controller.CreateCastellModel, model.MemberTypeAdmin)).Methods("POST")
	r.HandleFunc("/api/castells/model/{uuid:[0-9a-f]+}", checkTokenType(controller.DeleteCastellModel, model.MemberTypeAdmin)).Methods("DELETE")
	r.HandleFunc("/api/castells/model/{uuid:[0-9a-f]+}", checkTokenType(controller.EditCastellModel, model.MemberTypeAdmin)).Methods("PUT")
}
