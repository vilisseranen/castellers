package main

import (
	"github.com/vilisseranen/castellers/model"
	"github.com/vilisseranen/castellers/routes"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	_ "github.com/mattn/go-sqlite3"
)

type App struct {
	Router  *mux.Router
	handler http.Handler
}

func (a *App) Initialize(dbname, logFile string) {

	model.InitializeDB(dbname)
	a.Router = routes.CreateRouter("static")

	f, err := os.OpenFile("/var/log/"+logFile, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("Error opening file: %v", err)
	}
	a.handler = handlers.CombinedLoggingHandler(f, a.Router)
}

func (a *App) Run(addr string) {

	log.Fatal(http.ListenAndServe(addr, a.handler))
}
