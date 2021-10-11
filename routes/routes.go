package routes

import (
	"github.com/gorilla/mux"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gorilla/mux/otelmux"
	"go.opentelemetry.io/otel"
)

var tracer = otel.Tracer("castellers")

func CreateRouter(staticDir string) *mux.Router {
	r := mux.NewRouter()
	r.Use(otelmux.Middleware("castellers"))
	return r
}
