package main

import (
	"github.com/vilisseranen/castellers/model"
	"github.com/vilisseranen/castellers/routes"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	_ "github.com/mattn/go-sqlite3"
)

type App struct {
	Router *mux.Router
}

func (a *App) Initialize(dbname string) {

	model.InitializeDB(dbname)
	a.Router = routes.CreateRouter()
}

func (a *App) Run(addr string) {
	log.Fatal(http.ListenAndServe(addr, a.Router))
}
