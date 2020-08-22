package service

import (
	// "net/http"
	. "github.com/PanYicheng/go-microservice/common/router"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// Initialize our routes
var routes = Routes{
	Route{
		"GetAccount",            // Name
		"GET",                   // HTTP method
		"/accounts/{accountId}", // Route pattern
		GetAccount,
		true,
	},
	Route{
		"HealthCheck",
		"GET",
		"/health",
		HealthCheck,
		false,
	},
	Route{
		"Testability",
		"GET",
		"/testability/healthy/{state}",
		SetHealthyState,
		false,
	},
	Route{
        "Prometheus",
        "GET",
        "/metrics",
        promhttp.Handler().ServeHTTP,
        false,
    },
}
