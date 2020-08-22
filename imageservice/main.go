/**
The MIT License (MIT)

Copyright (c) 2016 Callista Enterprise

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.
*/
package main

import (
	"flag"
	"os"
	"os/signal"
	"syscall"
	"github.com/sirupsen/logrus"
	"github.com/PanYicheng/go-microservice/common/config"
	"github.com/PanYicheng/go-microservice/common/messaging"
	"github.com/PanYicheng/go-microservice/common/tracing"
	"github.com/PanYicheng/go-microservice/imageservice/service"
	"github.com/spf13/viper"
)

var appName = "imageservice"

func init() {
	profile := flag.String("profile", "dev", "Environment profile, something similar to spring profiles")
	configServerUrl := flag.String("configServerUrl", "http://127.0.0.1:8888", "Address to config server")
	configBranch := flag.String("configBranch", "P11", "git branch to fetch configuration from")

	flag.Parse()

	viper.Set("profile", *profile)
	viper.Set("configServerUrl", *configServerUrl)
	viper.Set("configBranch", *configBranch)

	if *profile == "dev" {
		logrus.SetFormatter(&logrus.TextFormatter{
			TimestampFormat: "2006-01-02T15:04:05.000",
			FullTimestamp:   true,
		})
	} else {
		logrus.SetFormatter(&logrus.JSONFormatter{})
	}
}

func main() {
	logrus.Infof("Starting %v", appName)

	config.LoadConfigurationFromBranch(viper.GetString("configServerUrl"), appName, viper.GetString("profile"), viper.GetString("configBranch"))
	initializeMessaging()
	initializeTracing()

	// Makes sure connection is closed when service exits.
	handleSigterm(func() {
		if service.MessagingClient != nil {
			service.MessagingClient.Close()
		}
	})

	service.StartWebServer(viper.GetString("server_port")) // Starts HTTP service  (async)
}

func initializeMessaging() {
	if !viper.IsSet("amqp_server_url") {
		panic("No 'amqp_server_url' set in configuration, cannot start")
	}

	service.MessagingClient = &messaging.MessagingClient{}
	service.MessagingClient.ConnectToBroker(viper.GetString("amqp_server_url"))
	service.MessagingClient.Subscribe(viper.GetString("config_event_bus"), "topic", appName, config.HandleRefreshEvent)
}

func initializeTracing() {
	tracing.InitTracing(viper.GetString("zipkin_server_url"), appName)
}

func handleSigterm(handleExit func()) {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	signal.Notify(c, syscall.SIGTERM)
	go func() {
		<-c
		handleExit()
		os.Exit(1)
	}()
}