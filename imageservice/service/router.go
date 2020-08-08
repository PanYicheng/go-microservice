package service

import (
	"net/http"
	"github.com/gorilla/mux"
	"github.com/PanYicheng/go-microservice/common/tracing"
)

/**
 * From http://thenewstack.io/make-a-restful-json-api-go/
 */
func NewRouter() *mux.Router {

	router := mux.NewRouter().StrictSlash(true)
	for _, route := range routes {
		var handler http.Handler

		handler = route.HandlerFunc

		router.
			Methods(route.Method).
			Path(route.Pattern).
			Name(route.Name).
			Handler(handler)
	}
	router.Use(tracing.ServerMiddleware)
	return router
}
