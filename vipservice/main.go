package main

import (
	"github.com/sirupsen/logrus"
	"flag"
	"os"
	"os/signal"
	"syscall"
	"github.com/PanYicheng/go-microservice/common/config"
	"github.com/PanYicheng/go-microservice/common/messaging"
	"github.com/PanYicheng/go-microservice/vipservice/service"
	"github.com/spf13/viper"
	"github.com/streadway/amqp"
)

var appName = "vipservice"

// Init function, runs before main()
func init() {
	// Read command line flags
	profile := flag.String("profile", "test", "Environment profile, something similar to spring profiles")
	configServerUrl := flag.String("configServerUrl", "http://configserver:8888", "Address to config server")
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
	} else {
		logrus.SetFormatter(&logrus.JSONFormatter{})
	}
}

func main() {
	logrus.Printf("Starting %v\n", appName)
	// load the config
	config.LoadConfigurationFromBranch(
		viper.GetString("configServerUrl"),
		appName,
		viper.GetString("profile"),
		viper.GetString("configBranch"))
	initializeMessaging()
	// Makes sure connection is closed when service exits.
	handleSigterm(func() {
		if service.MessagingClient != nil {
			service.MessagingClient.Close()
		}
	})
	service.StartWebServer(viper.GetString("server_port"))
}

// The callback function that's invoked whenever we get a message on the "vipQueue"
func onMessage(delivery amqp.Delivery) {
	logrus.Printf("Got a message: %v\n", string(delivery.Body))
}

// Call this from the main method.
func initializeMessaging() {
	if !viper.IsSet("amqp_server_url") {
		panic("No 'amqp_server_url' set in configuration, cannot start")
	}

	service.MessagingClient = &messaging.MessagingClient{}
	service.MessagingClient.ConnectToBroker(viper.GetString("amqp_server_url"))
	service.MessagingClient.Subscribe(viper.GetString("config_event_bus"), "topic", appName, config.HandleRefreshEvent)

	// Call the subscribe method with queue name and callback function
	err := service.MessagingClient.SubscribeToQueue("vipQueue", appName, onMessage)
	if err != nil {
		logrus.Println("Could not start subscribe to vip_queue")
	}
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
