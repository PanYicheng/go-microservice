package service

import (
        "net/http"
        "log"
        "github.com/sirupsen/logrus"
		"github.com/PanYicheng/go-microservice/internal/pkg/router"
)

func StartWebServer(port string) {
        r := router.NewRouter(routes)
        http.Handle("/", r)
        logrus.Infof("Starting HTTP service at %v" , port)
        err := http.ListenAndServe(":" + port, nil)
        if err != nil {
                log.Println("An error occured starting HTTP listener at port " + port)
                log.Println("Error: " + err.Error())
        }
}
