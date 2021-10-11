package routes

import (
	"github.com/gorilla/mux"
	"go.elastic.co/apm/module/apmgorilla"
)

const (
	APM_SPAN_TYPE_REQUEST = "request"
)

func CreateRouter(staticDir string) *mux.Router {
	r := mux.NewRouter()
	// tracer, _ := apm.NewTracerOptions(apm.TracerOptions{ServiceName: controller.APM_SPAN_TYPE})
	r.Use(apmgorilla.Middleware())
	return r
}
