package routes

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/vilisseranen/castellers/controller"
)

func CreateRouter(staticDir string) *mux.Router {

	r := mux.NewRouter()

	// No auth
	r.HandleFunc("/api/initialize", controller.Initialize).Methods("POST")
	r.HandleFunc("/api/initialize", controller.IsInitialized).Methods("GET")
	r.HandleFunc("/api/events", controller.GetEvents).Methods("GET")
	r.HandleFunc("/api/events/{uuid:[0-9a-f]+}", controller.GetEvent).Methods("GET")
	r.HandleFunc("/api/roles", controller.GetRoles).Methods("GET")

	// Requires member uuid
	r.HandleFunc("/api/events/{event_uuid:[0-9a-f]+}/members/{member_uuid:[0-9a-f]+}", checkMember(controller.ParticipateEvent)).Methods("POST")
	r.HandleFunc("/api/members/{member_uuid:[0-9a-f]+}", checkMember(controller.GetMember)).Methods("GET")

	// Requires admin uuid
	r.HandleFunc("/api/admins/{admin_uuid:[0-9a-f]+}/events", checkAdmin(controller.CreateEvent)).Methods("POST")
	r.HandleFunc("/api/admins/{admin_uuid:[0-9a-f]+}/members", checkAdmin(controller.CreateMember)).Methods("POST")
	r.HandleFunc("/api/admins/{admin_uuid:[0-9a-f]+}/events/{uuid:[0-9a-f]+}", checkAdmin(controller.UpdateEvent)).Methods("PUT")
	r.HandleFunc("/api/admins/{admin_uuid:[0-9a-f]+}/events/{uuid:[0-9a-f]+}", checkAdmin(controller.DeleteEvent)).Methods("DELETE")
	r.HandleFunc("/api/admins/{admin_uuid:[0-9a-f]+}/events/{event_uuid:[0-9a-f]+}/members/{member_uuid:[0-9a-f]+}", checkAdmin(controller.PresenceEvent)).Methods("POST")

	// Static site
	r.PathPrefix("/").Handler(http.FileServer(http.Dir(staticDir)))

	return r
}
