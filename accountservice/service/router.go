package service

import (
	"github.com/gorilla/mux"
	//"net/http"
	"github.com/PanYicheng/go-microservice/common/tracing"
	"github.com/PanYicheng/go-microservice/common/monitoring"
	"github.com/sirupsen/logrus"
)

// NewRouter is the function that returns a pointer to a mux.Router we can use as a handler.
func NewRouter() *mux.Router {
	// Create an instance of the Gorilla router
	router := mux.NewRouter().StrictSlash(true)
	// Iterate over the routes we declared in routes.go and attach them to the router instance
	for _, route := range routes {
		summaryVec := monitoring.BuildSummaryVec(route.Name, route.Method+" "+route.Pattern)

		// Attach each route, uses a Builder-like pattern to set each route up.
		router.Methods(route.Method).
			Path(route.Pattern).
			Name(route.Name).
			HandlerFunc(monitoring.WithMonitoring(route.HandlerFunc, route, summaryVec))
			//Handler(loadTracing(route.HandlerFunc))
	}
	router.Use(tracing.ServerMiddleware)
	logrus.Infoln("Successfully initialized routes with Prometheus.")
	return router
}

//func loadTracing(next http.Handler) http.Handler {
//    return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
//        span := tracing.StartHTTPTrace(req, "GetAccount")
//        defer span.Finish()
//
//        ctx := tracing.UpdateContext(req.Context(), span)
//        next.ServeHTTP(rw, req.WithContext(ctx))
//    })
//}
