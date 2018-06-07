package main

import (
	"database/sql"
	"encoding/json"
	"github.com/vilisseranen/castellers/model"
	"log"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	_ "github.com/mattn/go-sqlite3"
)

type App struct {
	Router *mux.Router
	DB     *sql.DB
}

func (a *App) Initialize(dbname string) {

	var err error
	a.DB, err = sql.Open("sqlite3", dbname)
	if err != nil {
		log.Fatal(err)
	}

	a.Router = mux.NewRouter()
	a.initializeRoutes()
}

func (a *App) initializeRoutes() {
	a.Router.HandleFunc("/events", a.getEvents).Methods("GET")
	a.Router.HandleFunc("/events/{uuid:[0-9a-f]+}", a.getEvent).Methods("GET")
	a.Router.HandleFunc("/admins/{uuid:[0-9a-f]+}/events", a.createEvent).Methods("POST")
	a.Router.HandleFunc("/admins/{uuid:[0-9a-f]+}/members", a.createMember).Methods("POST")
	a.Router.HandleFunc("/events/{event_uuid:[0-9a-f]+}/members/{member_uuid:[0-9a-f]+}", a.participateEvent).Methods("POST")
	a.Router.HandleFunc("/events/{uuid:[0-9a-f]+}", a.updateEvent).Methods("PUT")
	a.Router.HandleFunc("/events/{uuid:[0-9a-f]+}", a.deleteEvent).Methods("DELETE")
	a.Router.HandleFunc("/members/{uuid:[0-9a-f]+}", a.getMember).Methods("GET")
}

func (a *App) Run(addr string) {
	log.Fatal(http.ListenAndServe(addr, a.Router))
}

func (a *App) getEvent(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	uuid := vars["uuid"]
	e := model.Event{UUID: uuid}
	if err := e.Get(a.DB); err != nil {
		switch err {
		case sql.ErrNoRows:
			respondWithError(w, http.StatusNotFound, "Event not found")
		default:
			respondWithError(w, http.StatusInternalServerError, err.Error())
		}
		return
	}
	respondWithJSON(w, http.StatusOK, e)
}

func (a *App) getMember(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	uuid := vars["uuid"]
	m := model.Member{UUID: uuid}
	if err := m.Get(a.DB); err != nil {
		switch err {
		case sql.ErrNoRows:
			respondWithError(w, http.StatusNotFound, "Member not found")
		default:
			respondWithError(w, http.StatusInternalServerError, err.Error())
		}
		return
	}
	respondWithJSON(w, http.StatusOK, m)
}

func (a *App) getEvents(w http.ResponseWriter, r *http.Request) {
	count, _ := strconv.Atoi(r.FormValue("count"))
	start, _ := strconv.Atoi(r.FormValue("start"))
	if count < 1 {
		count = 10
	}
	e := model.Event{}
	events, err := e.GetEvents(a.DB, start, count)
	if err != nil {
		switch err {
		default:
			respondWithError(w, http.StatusInternalServerError, err.Error())
		}
		return
	}
	respondWithJSON(w, http.StatusOK, events)
}

func (a *App) createEvent(w http.ResponseWriter, r *http.Request) {
	// Check if admin exists
	vars := mux.Vars(r)
	uuid := vars["uuid"]
	admin := model.Admin{UUID: uuid}
	if err := admin.Get(a.DB); err != nil {
		switch err {
		case sql.ErrNoRows:
			respondWithError(w, http.StatusUnauthorized, "This admin is not authorized to create events.")
		default:
			respondWithError(w, http.StatusInternalServerError, err.Error())
		}
		return
	}
	// Create the event
	var e model.Event
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&e); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}
	defer r.Body.Close()
	if err := e.CreateEvent(a.DB); err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondWithJSON(w, http.StatusCreated, e)
}

func (a *App) createMember(w http.ResponseWriter, r *http.Request) {
	// Check if admin exists
	vars := mux.Vars(r)
	uuid := vars["uuid"]
	admin := model.Admin{UUID: uuid}
	if err := admin.Get(a.DB); err != nil {
		switch err {
		case sql.ErrNoRows:
			respondWithError(w, http.StatusUnauthorized, "This admin is not authorized to create members.")
		default:
			respondWithError(w, http.StatusInternalServerError, err.Error())
		}
		return
	}
	// Create the event
	var m model.Member
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&m); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}
	defer r.Body.Close()
	if err := m.CreateMember(a.DB); err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondWithJSON(w, http.StatusCreated, m)
}

func (a *App) participateEvent(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	event_uuid := vars["event_uuid"]
	member_uuid := vars["member_uuid"]
	event := model.Event{UUID: event_uuid}
	member := model.Member{UUID: member_uuid}
	if err := member.Get(a.DB); err != nil {
		switch err {
		case sql.ErrNoRows:
			respondWithError(w, http.StatusUnauthorized, "You are not authorized to register for this event.")
		default:
			respondWithError(w, http.StatusInternalServerError, err.Error())
		}
		return
	}
	if err := event.Get(a.DB); err != nil {
		switch err {
		case sql.ErrNoRows:
			respondWithError(w, http.StatusUnauthorized, "You are not authorized to register for this event.")
		default:
			respondWithError(w, http.StatusInternalServerError, err.Error())
		}
		return
	}

	var p model.Participation
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&p); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}
	defer r.Body.Close()

	p.EventUUID = event_uuid
	p.MemberUUID = member_uuid

	if err := p.Participate(a.DB); err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondWithJSON(w, http.StatusCreated, p)
}

func (a *App) updateEvent(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	uuid := vars["uuid"]
	var e model.Event
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&e); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}
	defer r.Body.Close()
	e.UUID = uuid
	if err := e.UpdateEvent(a.DB); err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondWithJSON(w, http.StatusOK, e)
}

func (a *App) deleteEvent(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	uuid := vars["uuid"]
	e := model.Event{UUID: uuid}
	if err := e.DeleteEvent(a.DB); err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondWithJSON(w, http.StatusOK, nil)
}

func respondWithError(w http.ResponseWriter, code int, message string) {
	respondWithJSON(w, code, map[string]string{"error": message})
}

func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	response, _ := json.Marshal(payload)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(response)
}
