package service

import (
        "net/http"
        "log"
        "github.com/sirupsen/logrus"
)

func StartWebServer(port string) {

        r := NewRouter()
        http.Handle("/", r)
        logrus.Infof("Starting HTTP service at %v" , port)
        err := http.ListenAndServe(":" + port, nil)
        if err != nil {
                log.Println("An error occured starting HTTP listener at port " + port)
                log.Println("Error: " + err.Error())
        }
}
