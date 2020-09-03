package service

import (
	"github.com/sirupsen/logrus"
	. "github.com/PanYicheng/go-microservice/internal/pkg/router"

	"net/http"
)

// StartWebServer is the function that start http server on port
func StartWebServer(port string) {
	r := NewRouter(routes)
	http.Handle("/", r)
	logrus.Println("Starting HTTP service at " + port)
	err := http.ListenAndServe(":"+port, nil) // Goroutine will block here
	if err != nil {
		logrus.Println("An error occured starting HTTP listener at port " + port)
		logrus.Println("Error: " + err.Error())
	}
}
