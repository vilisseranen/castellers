package routes

import (
	"github.com/gorilla/mux"

	"github.com/vilisseranen/castellers/controller"
	"github.com/vilisseranen/castellers/model"
)

func CreateRouter(staticDir string) *mux.Router {

	r := mux.NewRouter()

	// No auth
	r.HandleFunc("/api/initialize", controller.Initialize).Methods("POST")
	r.HandleFunc("/api/initialize", controller.IsInitialized).Methods("GET")
	r.HandleFunc("/api/events", controller.GetEvents).Methods("GET")
	r.HandleFunc("/api/events/{uuid:[0-9a-f]+}", controller.GetEvent).Methods("GET")
	r.HandleFunc("/api/roles", controller.GetRoles).Methods("GET")
	r.HandleFunc("/api/login", controller.Login).Methods("POST")
	r.HandleFunc("/api/logout", controller.Logout).Methods("POST")
	r.HandleFunc("/api/refresh", controller.RefreshToken).Methods("POST")

	r.HandleFunc("/api/test", controller.Test).Methods("GET")

	// Special tokens
	r.HandleFunc("/api/create_credentials", checkTokenType(controller.CreateCredentials, controller.CreateCredentialsPermission)).Methods("POST")
	r.HandleFunc("/api/reset_credentials", checkTokenType(controller.ResetCredentials, controller.ResetCredentialsPermission)).Methods("POST")

	// Requires a token with member permission
	r.HandleFunc("/api/members/{member_uuid:[0-9a-f]+}", checkTokenType(controller.GetMember, model.MemberTypeMember)).Methods("GET")
	r.HandleFunc("/api/events/{event_uuid:[0-9a-f]+}/members/{member_uuid:[0-9a-f]+}", checkTokenType(controller.ParticipateEvent, model.MemberTypeMember)).Methods("POST")
	r.HandleFunc("/api/members/{member_uuid:[0-9a-f]+}", checkTokenType(controller.EditMember, model.MemberTypeMember)).Methods("PUT")
	r.HandleFunc("/api/members/{member_uuid:[0-9a-f]+}/events", checkTokenType(controller.GetEvents, model.MemberTypeMember)).Methods("GET")

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

	return r
}
