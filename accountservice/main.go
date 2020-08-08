package main

import (
	"github.com/sirupsen/logrus"

	"flag"
	"os"
	"os/signal"
	"syscall"

	"github.com/PanYicheng/go-microservice/accountservice/dbclient"
	"github.com/PanYicheng/go-microservice/accountservice/service"
	"github.com/PanYicheng/go-microservice/common/config"
	"github.com/PanYicheng/go-microservice/common/messaging"
	"github.com/PanYicheng/go-microservice/common/tracing"
	cb "github.com/PanYicheng/go-microservice/common/circuitbreaker"
	"github.com/spf13/viper"
)

var appName = "accountservice"

// Init function, runs before main()
func init() {
	// Read command line flags
	profile := flag.String("profile", "test", "Environment profile, something similar to spring profiles")
	configServerUrl := flag.String("configServerUrl", "", "Address to config server")
	configBranch := flag.String("configBranch", "master", "git branch to fetch configuration from")
	flag.Parse()

	// Pass the flag values into viper.
	viper.Set("profile", *profile)
	viper.Set("configServerUrl", *configServerUrl)
	viper.Set("configBranch", *configBranch)

	if *profile == "dev" {
		logrus.SetFormatter(&logrus.TextFormatter{
			TimestampFormat: "2006-01-02T15:04:05.000",
			FullTimestamp:   true,
		})
		logrus.SetLevel(logrus.DebugLevel)
	} else {
		logrus.SetFormatter(&logrus.JSONFormatter{})
	}
}

func main() {
	logrus.Printf("Starting %v\n", appName)
	// load the config if configServerUrl set
	if len(viper.GetString("configServerUrl")) > 0 {
		config.LoadConfigurationFromBranch(
		viper.GetString("configServerUrl"),
		appName,
		viper.GetString("profile"),
		viper.GetString("configBranch"))
	} else {
		// set custome configs if configServerUrl not set
		viper.Set("amqp_server_url", "http://172.17.7.221:5772")
		viper.Set("config_event_bus", "springCloudBus")
		viper.Set("server_port", 6767)
		viper.Set("zipkin_server_url", "http://localhost:9411")

	}
	initializeBoltClient()
	initializeMessaging()
	initializeTracing()
	cb.ConfigureHystrix([]string{"imageservice", "quotes-service"}, service.MessagingClient)
	go config.StartListener(appName, viper.GetString("amqp_server_url"), viper.GetString("config_event_bus"))
	handleSigterm(func() {
        cb.Deregister(service.MessagingClient)
        service.MessagingClient.Close()
    })
	service.StartWebServer(viper.GetString("server_port"))
}

// Creates instance and calls the OpenBoltDb and Seed funcs
func initializeBoltClient() {
	service.DBClient = &dbclient.BoltClient{}
	service.DBClient.OpenBoltDb()
	service.DBClient.Seed()
}

// Call this from the main method.
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

// Handles Ctrl+C or most other means of "controlled" shutdown gracefully. Invokes the supplied func before exiting.
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
