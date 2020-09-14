package service

import (
	. "github.com/PanYicheng/go-microservice/internal/pkg/route"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// Initialize our routes
var routes = Routes{
	Route{
		"Index",       // Name
		"GET",         // HTTP Method
		"/",           // Route pattern
		Index,         // HandlerFunc
		true,          // Monitor 
		true,          // Trace
	},
	Route{
		"SetResponseTime",       // Name
		"POST",         // HTTP Method
		"/responsetime",           // Route pattern
		SetResponseTime,         // HandlerFunc
		true,          // Monitor 
		false,          // Trace
	},
	Route{
		"GetServiceInfo",       // Name
		"GET",         // HTTP Method
		"/info",           // Route pattern
		GetServiceInfo,         // HandlerFunc
		true,          // Monitor 
		false,          // Trace
	},
	Route{
		"GetCircuitInfo",       // Name
		"GET",         // HTTP Method
		"/circuitinfo",           // Route pattern
		GetCircuitInfo,         // HandlerFunc
		true,          // Monitor 
		false,          // Trace
	},
	Route{
		"SetCircuitInfo",       // Name
		"POST",         // HTTP Method
		"/circuitinfo/{name}",           // Route pattern
		SetCircuitInfo,         // HandlerFunc
		true,          // Monitor 
		false,          // Trace
	},
	Route{
		"SetConcurrency",       // Name
		"POST",         // HTTP Method
		"/concurrency",           // Route pattern
		SetConcurrency,         // HandlerFunc
		true,          // Monitor 
		false,          // Trace
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
