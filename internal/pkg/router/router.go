package router

import (
	"github.com/gorilla/mux"
	"github.com/PanYicheng/go-microservice/internal/pkg/tracing"
	"github.com/PanYicheng/go-microservice/internal/pkg/monitoring"
	. "github.com/PanYicheng/go-microservice/internal/pkg/route"
	"github.com/sirupsen/logrus"
)

// NewRouter is the function that returns a pointer to a mux.Router we can use as a handler.
func NewRouter(routes Routes) *mux.Router {
	// Create an instance of the Gorilla router
	router := mux.NewRouter().StrictSlash(true)
	// Iterate over the routes we declared in routes.go and attach them to the router instance
	for _, route := range routes {
		// Init Summary
		summaryVec := monitoring.BuildSummaryVec(route.Name, route.Method + " " + route.Pattern + "summary")
		// Init Histogram
		histoVec := monitoring.BuildHistoVec(route.Name, route.Method + " " + route.Pattern + "histogram")

		// Attach each route, uses a Builder-like pattern to set each route up.
		router.Methods(route.Method).
			Path(route.Pattern).
			Name(route.Name).
			HandlerFunc(
				tracing.WithTracing(
					monitoring.WithMonitoringHisto(
						monitoring.WithMonitoringSummary(
							route.HandlerFunc,
							route,
							summaryVec),
						route,
						histoVec),
					route,
				),
			)
	}
	logrus.Infoln("Successfully initialized routes with Prometheus and Zipking tracing.")
	return router
}
