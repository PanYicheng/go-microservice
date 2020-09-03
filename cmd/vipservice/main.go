package main

import (
	"github.com/sirupsen/logrus"
	"github.com/alexflint/go-arg"
	"os"
	"os/signal"
	"syscall"
	"github.com/PanYicheng/go-microservice/internal/pkg/config"
	"github.com/PanYicheng/go-microservice/internal/pkg/messaging"
	"github.com/PanYicheng/go-microservice/internal/pkg/tracing"
	"github.com/PanYicheng/go-microservice/internal/app/vipservice/service"
	"github.com/streadway/amqp"
)

func main() {
	// Initialize config struct and populate it from env vars and flags.
    cfg := &config.Config{}
    arg.MustParse(cfg)

	if cfg.Environment == "dev" {
		logrus.SetFormatter(&logrus.TextFormatter{
			TimestampFormat: "2006-01-02T15:04:05.000",
			FullTimestamp:   true,
		})
		logrus.SetLevel(logrus.DebugLevel)
	} else {
		logrus.SetFormatter(&logrus.JSONFormatter{})
		logrus.SetLevel(logrus.InfoLevel)
	}
	logrus.Printf("Starting %v in %v\n", cfg.ServiceName, cfg.Environment)

	initializeTracing(cfg)
	initializeMessaging(cfg)
	// Makes sure connection is closed when service exits.
	handleSigterm(func() {
		if service.MessagingClient != nil {
			service.MessagingClient.Close()
		}
	})
	service.StartWebServer(cfg.ServicePort)
}

// The callback function that's invoked whenever we get a message on the "vipQueue"
func onMessage(delivery amqp.Delivery) {
	logrus.Printf("Got a message: %v\n", string(delivery.Body))
}

// Call this from the main method.
func initializeMessaging(cfg *config.Config) {
	if cfg.AmqpUrl == "" {
        panic("No 'amqpurl' in command line or 'AMQP_URL' in env, cannot start")
	}

	service.MessagingClient = &messaging.MessagingClient{}
	service.MessagingClient.ConnectToBroker(cfg.AmqpUrl)

	// Call the subscribe method with queue name and callback function
	err := service.MessagingClient.SubscribeToQueue("vipQueue", cfg.ServiceName, onMessage)
	if err != nil {
		logrus.Println("Could not start subscribe to vip_queue")
	}
}

func initializeTracing(cfg *config.Config) {
	tracing.InitTracing(cfg.ZipkinUrl, cfg.ServiceName)
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
