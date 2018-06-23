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
	Router *mux.Router
}

func (a *App) Initialize(dbname, staticDir string) {

	model.InitializeDB(dbname)
	a.Router = routes.CreateRouter(staticDir)
}

func (a *App) Run(addr, logFile string) {
	f, err := os.OpenFile(logFile, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("Error opening file: %v", err)
	}
	defer f.Close()

	log.Fatal(http.ListenAndServe(addr, handlers.CombinedLoggingHandler(f, a.Router)))
}
