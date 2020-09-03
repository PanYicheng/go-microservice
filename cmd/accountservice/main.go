package main

import (
	"github.com/sirupsen/logrus"
	"github.com/alexflint/go-arg"
	"os"
	"os/signal"
	"syscall"
	"github.com/PanYicheng/go-microservice/internal/app/accountservice/dbclient"
	"github.com/PanYicheng/go-microservice/internal/app/accountservice/service"
	"github.com/PanYicheng/go-microservice/internal/pkg/config"
	"github.com/PanYicheng/go-microservice/internal/pkg/messaging"
	"github.com/PanYicheng/go-microservice/internal/pkg/tracing"
	cb "github.com/PanYicheng/go-microservice/internal/pkg/circuitbreaker"
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
	initializeBoltClient()
	mc := initializeMessaging(cfg)
	service.MessagingClient = mc

	cb.ConfigureHystrix([]string{"imageservice", "quotes-service"}, mc)
	handleSigterm(func() {
        cb.Deregister(mc)
		if mc != nil {
			mc.Close()
		}
    })
	service.StartWebServer(cfg.ServicePort)
}

// Creates instance and calls the OpenBoltDb and Seed funcs
func initializeBoltClient() {
	service.DBClient = &dbclient.BoltClient{}
	service.DBClient.OpenBoltDb()
	service.DBClient.Seed()
}

func initializeMessaging(cfg *config.Config) *messaging.MessagingClient {
    if cfg.AmqpUrl == "" {
        panic("No 'amqpurl' in command line or 'AMQP_URL' in env, cannot start")
    }

    mc := &messaging.MessagingClient{}
    mc.ConnectToBroker(cfg.AmqpUrl)
    return mc
}

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
