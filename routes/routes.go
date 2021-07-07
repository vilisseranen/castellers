package routes

import (
	"github.com/gorilla/mux"
)

func CreateRouter(staticDir string) *mux.Router {
	return mux.NewRouter()
}
