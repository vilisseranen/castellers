package routes

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/vilisseranen/castellers/controller"
)

func CreateRouter(staticDir string) *mux.Router {

	r := mux.NewRouter()

	// No auth
	r.HandleFunc("/initialize", controller.Initialize).Methods("POST")
	r.HandleFunc("/events", controller.GetEvents).Methods("GET")
	r.HandleFunc("/events/{uuid:[0-9a-f]+}", controller.GetEvent).Methods("GET")

	// Requires member uuid
	r.HandleFunc("/events/{event_uuid:[0-9a-f]+}/members/{member_uuid:[0-9a-f]+}", checkMember(controller.ParticipateEvent)).Methods("POST")
	r.HandleFunc("/members/{member_uuid:[0-9a-f]+}", checkMember(controller.GetMember)).Methods("GET")

	// Requires admin uuid
	r.HandleFunc("/admins/{admin_uuid:[0-9a-f]+}/events", checkAdmin(controller.CreateEvent)).Methods("POST")
	r.HandleFunc("/admins/{admin_uuid:[0-9a-f]+}/members", checkAdmin(controller.CreateMember)).Methods("POST")
	r.HandleFunc("/admins/{admin_uuid:[0-9a-f]+}/events/{uuid:[0-9a-f]+}", checkAdmin(controller.UpdateEvent)).Methods("PUT")
	r.HandleFunc("/admins/{admin_uuid:[0-9a-f]+}/events/{uuid:[0-9a-f]+}", checkAdmin(controller.DeleteEvent)).Methods("DELETE")
	r.HandleFunc("/admins/{admin_uuid:[0-9a-f]+}/events/{event_uuid:[0-9a-f]+}/members/{member_uuid:[0-9a-f]+}", checkAdmin(controller.PresenceEvent)).Methods("POST")

	// Static site
	r.PathPrefix("/").Handler(http.FileServer(http.Dir(staticDir)))

	return r
}
