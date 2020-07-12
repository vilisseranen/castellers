package app

import (
	"log"
	"net/http"
	"os"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	_ "github.com/mattn/go-sqlite3"

	"github.com/vilisseranen/castellers/common"
	"github.com/vilisseranen/castellers/controller"
	"github.com/vilisseranen/castellers/model"
	"github.com/vilisseranen/castellers/routes"
)

type App struct {
	Router    *mux.Router
	handler   http.Handler
	scheduler controller.Scheduler
}

func (a *App) Initialize() {

	common.ReadConfig()

	if len(common.GetConfigString("encryption.key")) == 0 || len(common.GetConfigString("encryption.key_salt")) == 0 || len(common.GetConfigString("encryption.password_pepper")) == 0 {
		log.Fatalf("Please provide `encryption.key`, `encryption.key_salt` and `encryption.password_pepper` for encrypting database.")
	}

	err := common.InitializeLogger()
	if err != nil {
		log.Fatalf("Error configuring the logger: %v", err)
	}
	common.Info("Message from app %s et %s", "toto", "tutu")

	model.InitializeDB(common.GetConfigString("db_name"))
	a.Router = routes.CreateRouter("static")

	f, err := os.OpenFile(common.GetConfigString("log_file"), os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("Error opening file: %v", err)
	}

	// Define logger
	a.handler = handlers.CombinedLoggingHandler(f, a.Router)

	// Define CORS handlers
	headersOk := handlers.AllowedHeaders([]string{"Content-Type"})
	originsOk := handlers.AllowedOrigins([]string{common.GetConfigString("domain")})
	methodsOk := handlers.AllowedMethods([]string{"DELETE", "GET", "HEAD", "POST", "PUT", "OPTIONS"})
	a.handler = handlers.CORS(originsOk, headersOk, methodsOk)(a.handler)
}

func (a *App) Run(addr string) {
	a.scheduler.Start()
	log.Fatal(http.ListenAndServe(addr, a.handler))
}
