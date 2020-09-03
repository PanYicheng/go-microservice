package main

import (
	"os"
	"os/signal"
	"syscall"
	"github.com/sirupsen/logrus"
	"github.com/alexflint/go-arg"
	"github.com/PanYicheng/go-microservice/internal/pkg/config"
	// "github.com/PanYicheng/go-microservice/internal/pkg/messaging"
	"github.com/PanYicheng/go-microservice/internal/pkg/tracing"
	"github.com/PanYicheng/go-microservice/internal/app/imageservice/service"
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
	// mc := initializeMessaging(cfg)
	// service.MessagingClient = mc

	// Makes sure AMQP connection is closed when service exits.
	handleSigterm(func() {
		// if mc != nil {
		// 	mc.Close()
		// }
	})

	service.StartWebServer(cfg.ServicePort)
}

// func initializeMessaging(cfg *config.Config) *messaging.MessagingClient {
//     if cfg.AmqpUrl == "" {
//         panic("No 'amqpserverurl' in command line or 'AMQP_SERVER_URL' in env, cannot start")
//     }
// 
//     mc := &messaging.MessagingClient{}
//     mc.ConnectToBroker(cfg.AmqpUrl)
//     return mc
// }

func initializeTracing(cfg *config.Config) {
	tracing.InitTracing(cfg.ZipkinUrl, cfg.ServiceName)
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
