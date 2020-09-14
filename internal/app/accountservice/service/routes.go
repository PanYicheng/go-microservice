package service

import (
	// "net/http"
	. "github.com/PanYicheng/go-microservice/internal/pkg/route"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// Initialize our routes
var routes = Routes{
	Route{
		"GetAccount",            // Name
		"GET",                   // HTTP method
		"/accounts/{accountId}", // Route pattern
		GetAccount,              // HandlerFunc
		true,                    // Monitor
		true,                    // Trace
	},
	Route{
		"HealthCheck",
		"GET",
		"/health",
		HealthCheck,
		false,
		false,
	},
	Route{
		"Testability",
		"GET",
		"/testability/healthy/{state}",
		SetHealthyState,
		false,
		false,
	},
	Route{
        "Prometheus",
        "GET",
        "/metrics",
        promhttp.Handler().ServeHTTP,
        false,
		false,
    },
}
