package service

import . "github.com/PanYicheng/go-microservice/internal/pkg/route"

// Initialize our routes
var routes = Routes{
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
}
