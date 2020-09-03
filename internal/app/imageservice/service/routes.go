package service

import (
	"net/http"
	. "github.com/PanYicheng/go-microservice/internal/pkg/route"
)

var routes = Routes{
	Route{
		"ProcessImage",
		"POST",
		"/image",
		ProcessImage,
		true,
		false,
	},
	Route{
		"ProcessImage",
		"GET",
		"/file/{filename}",
		ProcessImageFromFile,
		true,
		false,
	},
	Route{
		"GetAccountImage",
		"GET",
		"/accounts/{accountId}",
		GetAccountImage,
		true,
		true,
	},
	Route{
		"HealthCheck",
		"GET",
		"/health",
		func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "text/plain; charset=UTF-8")
			w.Write([]byte("OK"))
		},
		false,
		false,
	},
}
