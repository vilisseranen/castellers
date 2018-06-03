package main

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	_ "github.com/mattn/go-sqlite3"
)

const UUID_SIZE = 50

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

func (a *App) Run(addr string) {
	log.Fatal(http.ListenAndServe(addr, a.Router))
}

func (a *App) initializeRoutes() {
	a.Router.HandleFunc("/events", a.getEvents).Methods("GET")
	a.Router.HandleFunc("/event/{uuid:[0-9a-f]+}", a.getEvent).Methods("GET")
	a.Router.HandleFunc("/event", a.createEvent).Methods("POST")
	a.Router.HandleFunc("/event/{uuid:[0-9a-f]+}", a.updateEvent).Methods("PUT")
	a.Router.HandleFunc("/event/{uuid:[0-9a-f]+}", a.deleteEvent).Methods("DELETE")
}

func (a *App) getEvent(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	uuid := vars["uuid"]
	e := event{UUID: uuid}
	if err := e.getEvent(a.DB); err != nil {
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

func (a *App) getEvents(w http.ResponseWriter, r *http.Request) {
	count, _ := strconv.Atoi(r.FormValue("count"))
	start, _ := strconv.Atoi(r.FormValue("start"))
	e := event{}
	events, err := e.getEvents(a.DB, start, count)
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
	var e event
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&e); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}
	defer r.Body.Close()
	if err := e.createEvent(a.DB); err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondWithJSON(w, http.StatusCreated, e)
}

func (a *App) updateEvent(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	uuid := vars["uuid"]
	var e event
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&e); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}
	defer r.Body.Close()
	e.UUID = uuid
	if err := e.updateEvent(a.DB); err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondWithJSON(w, http.StatusOK, e)
}

func (a *App) deleteEvent(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	uuid := vars["uuid"]
	e := event{UUID: uuid}
	if err := e.deleteEvent(a.DB); err != nil {
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
