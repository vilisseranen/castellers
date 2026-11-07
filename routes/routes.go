package routes

import (
	"net/http"

	"github.com/gorilla/mux"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gorilla/mux/otelmux"
	"go.opentelemetry.io/otel"
)

const (
	BASE_PATH = "/api/"
)

var tracer = otel.Tracer("castellers")

func CreateRouter(staticDir string) *mux.Router {
	r := mux.NewRouter()
	// Keep the span names used before otelmux 0.72 (route template only, without the HTTP method).
	r.Use(otelmux.Middleware("castellers", otelmux.WithSpanNameFormatter(func(route string, _ *http.Request) string {
		return route
	})))
	return r
}
