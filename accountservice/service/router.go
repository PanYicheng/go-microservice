package service

import (
	"github.com/gorilla/mux"
	"net/http"
	"github.com/PanYicheng/go-microservice/common/tracing"
)

// NewRouter is the function that returns a pointer to a mux.Router we can use as a handler.
func NewRouter() *mux.Router {
	// Create an instance of the Gorilla router
	router := mux.NewRouter().StrictSlash(true)
	//router.Use(tracing.ServerMiddleware)
	// Iterate over the routes we declared in routes.go and attach them to the router instance
	for _, route := range routes {
		// Attach each route, uses a Builder-like pattern to set each route up.
		router.Methods(route.Method).
			Path(route.Pattern).
			Name(route.Name).
			//HandlerFunc(route.HandlerFunc)
			Handler(loadTracing(route.HandlerFunc))
	}
	return router
}

func loadTracing(next http.Handler) http.Handler {
    return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
        span := tracing.StartHTTPTrace(req, "GetAccount")
        defer span.Finish()

        ctx := tracing.UpdateContext(req.Context(), span)
        next.ServeHTTP(rw, req.WithContext(ctx))
    })
}
