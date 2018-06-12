package routes

import (
	"github.com/gorilla/mux"
	"github.com/vilisseranen/castellers/controller"
)

func CreateRouter() *mux.Router {

	r := mux.NewRouter()
	// No auth
	r.HandleFunc("/events", controller.GetEvents).Methods("GET")
	r.HandleFunc("/events/{uuid:[0-9a-f]+}", controller.GetEvent).Methods("GET")

	// Requires member uuid
	r.HandleFunc("/events/{event_uuid:[0-9a-f]+}/members/{member_uuid:[0-9a-f]+}", controller.ParticipateEvent).Methods("POST")
	r.HandleFunc("/members/{uuid:[0-9a-f]+}", controller.GetMember).Methods("GET")

	// Requires admin uuid
	r.HandleFunc("/admins/{admin_uuid:[0-9a-f]+}/events", controller.CreateEvent).Methods("POST")
	r.HandleFunc("/admins/{admin_uuid:[0-9a-f]+}/members", controller.CreateMember).Methods("POST")
	r.HandleFunc("/admins/{admin_uuid:[0-9a-f]+}/events/{uuid:[0-9a-f]+}", controller.UpdateEvent).Methods("PUT")
	r.HandleFunc("/admins/{admin_uuid:[0-9a-f]+}/events/{uuid:[0-9a-f]+}", controller.DeleteEvent).Methods("DELETE")

	return r
}
