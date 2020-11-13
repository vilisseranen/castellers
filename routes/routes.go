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

	// Special tokens
	r.HandleFunc("/api/create_credentials", checkTokenType(controller.CreateCredentials, controller.CreateCredentialsPermission)).Methods("POST")
	r.HandleFunc("/api/reset_credentials", checkTokenType(controller.ResetCredentials, controller.ResetCredentialsPermission)).Methods("POST")

	// Requires a member token
	r.HandleFunc("/api/test", checkTokenType(controller.Test, model.MemberTypeMember)).Methods("GET")

	// Requires member uuid
	r.HandleFunc("/api/events/{event_uuid:[0-9a-f]+}/members/{member_uuid:[0-9a-f]+}", checkMember(controller.ParticipateEvent)).Methods("POST")
	r.HandleFunc("/api/members/{member_uuid:[0-9a-f]+}", checkMember(controller.GetMember)).Methods("GET")
	r.HandleFunc("/api/members/{member_uuid:[0-9a-f]+}", checkMember(controller.EditMember)).Methods("PUT")
	r.HandleFunc("/api/members/{member_uuid:[0-9a-f]+}/events", checkMember(controller.GetEvents)).Methods("GET")

	// Requires admin uuid
	r.HandleFunc("/api/admins/{admin_uuid:[0-9a-f]+}/events", checkAdmin(controller.CreateEvent)).Methods("POST")
	r.HandleFunc("/api/admins/{admin_uuid:[0-9a-f]+}/events", checkAdmin(controller.GetEvents)).Methods("GET")
	r.HandleFunc("/api/admins/{admin_uuid:[0-9a-f]+}/events/{uuid:[0-9a-f]+}", checkAdmin(controller.UpdateEvent)).Methods("PUT")
	r.HandleFunc("/api/admins/{admin_uuid:[0-9a-f]+}/events/{uuid:[0-9a-f]+}", checkAdmin(controller.DeleteEvent)).Methods("DELETE")
	r.HandleFunc("/api/admins/{admin_uuid:[0-9a-f]+}/events/{event_uuid:[0-9a-f]+}/members", checkAdmin(controller.GetEventParticipation)).Methods("GET")
	r.HandleFunc("/api/admins/{admin_uuid:[0-9a-f]+}/events/{event_uuid:[0-9a-f]+}/members/{member_uuid:[0-9a-f]+}", checkAdmin(controller.PresenceEvent)).Methods("POST")
	r.HandleFunc("/api/admins/{admin_uuid:[0-9a-f]+}/members", checkAdmin(controller.CreateMember)).Methods("POST")
	r.HandleFunc("/api/admins/{admin_uuid:[0-9a-f]+}/members", checkAdmin(controller.GetMembers)).Methods("GET")
	r.HandleFunc("/api/admins/{admin_uuid:[0-9a-f]+}/members/{member_uuid:[0-9a-f]+}", checkAdmin(controller.GetMember)).Methods("GET")
	r.HandleFunc("/api/admins/{admin_uuid:[0-9a-f]+}/members/{member_uuid:[0-9a-f]+}", checkAdmin(controller.EditMember)).Methods("PUT")
	r.HandleFunc("/api/admins/{admin_uuid:[0-9a-f]+}/members/{member_uuid:[0-9a-f]+}", checkAdmin(controller.DeleteMember)).Methods("DELETE")
	r.HandleFunc("/api/admins/{admin_uuid:[0-9a-f]+}/members/{member_uuid:[0-9a-f]+}/registration", checkAdmin(controller.SendRegistrationEmail)).Methods("GET")

	return r
}
