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

	requiredConfigs := []string{"encryption.key", "encryption.key_salt", "encryption.password_pepper", "jwt.access_secret", "jwt.refresh_secret"}
	for _, config := range requiredConfigs {
		if len(common.GetConfigString(config)) == 0 {
			log.Fatalf("The configuration value for %s is required", config)
		}
	}

	err := common.InitializeLogger()
	if err != nil {
		log.Fatalf("Error configuring the logger: %v", err)
	}

	model.InitializeDB(common.GetConfigString("db_name"))
	a.Router = routes.CreateRouter("static")
	routes.AttachV1API(a.Router)

	f, err := os.OpenFile(common.GetConfigString("log_file"), os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		common.Fatal("Error opening file: %v", err)
	}

	controller.InitializeRedis()
	common.InitializeTranslations()

	// Define logger
	a.handler = handlers.ProxyHeaders(a.Router)
	a.handler = handlers.CombinedLoggingHandler(f, a.handler)

	// Define CORS handlers
	headersOk := handlers.AllowedHeaders([]string{"X-Requested-With", "Content-Type", "Authorization"})
	originsOk := handlers.AllowedOrigins([]string{common.GetConfigString("domain")})
	methodsOk := handlers.AllowedMethods([]string{"DELETE", "GET", "HEAD", "POST", "PUT", "OPTIONS"})
	allowCredentials := handlers.AllowCredentials()
	a.handler = handlers.CORS(originsOk, headersOk, methodsOk, allowCredentials)(a.handler)
}

func (a *App) Run(addr string) {
	a.scheduler.Start()
	log.Fatal(http.ListenAndServe(addr, a.handler))
}
