// done
package main

import (
	"encoding/json"
	"math/rand"

	"github.com/sirupsen/logrus"

	// "fmt"
	"io/ioutil"
	"time"

	"github.com/PanYicheng/go-microservice/internal/app/customservice/model"
	"github.com/PanYicheng/go-microservice/internal/app/customservice/service"
	"github.com/PanYicheng/go-microservice/internal/pkg/monitoring"
	"github.com/PanYicheng/go-microservice/internal/pkg/tracing"
	"github.com/afex/hystrix-go/hystrix"
)

func main() {
	rand.Seed(time.Now().UTC().UnixNano())

	logrus.SetLevel(logrus.DebugLevel)
	jsonData, err := ioutil.ReadFile("/data/conf.json")
	if err != nil {
		logrus.Fatal(err)
	}
	err = json.Unmarshal(jsonData, &service.ServiceConfig)
	if err != nil {
		logrus.Fatal(err)
	}
	logrus.Info("Reading conf.json: ", string(jsonData))
	setCircuit()
	tracing.InitTracing(service.ServiceConfig.ZipkinServer, service.ServiceConfig.Name)
	// Get rid of go gc metrics in Prometheus data payload
	monitoring.UnregisterGoCollector()
	monitoring.BuildOtherVec(service.ServiceConfig.Name)

	// go register()
	go getinfo()
	go getcircuitinfo()

	logrus.Printf("%s is servicing at port:%s ...\n", service.ServiceConfig.Name, service.ServiceConfig.Port)
	service.StartWebServer(service.ServiceConfig.Port)
}

// setCircuit sets the hystrix circuitbreaker config for child service callings.
func setCircuit() {
	config := hystrix.CommandConfig{
		Timeout:                1000,
		MaxConcurrentRequests:  1000,
		SleepWindow:            5000,
		ErrorPercentThreshold:  5,
		RequestVolumeThreshold: 5,
	}
	for _, call := range service.ServiceConfig.CallList {
		name := service.ServiceConfig.Name + "To" + call
		hystrix.ConfigureCommand(name, config)
	}
}

// getinfo reloads configuration per second.
func getinfo() {
	for {
		time.Sleep(time.Second)
		logrus.Debug("getinfo runs.")
		var newservice model.Service
		jsonData, err := ioutil.ReadFile("/data/conf.json")
		if err != nil {
			continue
		}
		err = json.Unmarshal(jsonData, &newservice)
		if err != nil {
			continue
		}
		service.ServiceConfig = newservice
	}
}

// getcircuitinfo reloads circuit breaker config per second.
func getcircuitinfo() {
	for {
		time.Sleep(time.Second)
		logrus.Debug("getcircuitinfo runs.\n")
		var circuitinfo model.CircuitInfo
		jsonData, err := ioutil.ReadFile("/data/circuitinfo.json")
		if err != nil {
			continue
		}
		err = json.Unmarshal(jsonData, &circuitinfo)
		if err != nil {
			continue
		}

		settings := hystrix.GetCircuitSettings()
		oldcircuitinfo := settings[circuitinfo.Name]

		if oldcircuitinfo == nil || circuitinfo.MaxConcurrentRequests != oldcircuitinfo.MaxConcurrentRequests ||
			circuitinfo.ErrorPercentThreshold != oldcircuitinfo.ErrorPercentThreshold ||
			circuitinfo.RequestVolumeThreshold != int(oldcircuitinfo.RequestVolumeThreshold) ||
			circuitinfo.Timeout != int(oldcircuitinfo.Timeout/1000000) ||
			circuitinfo.SleepWindow != int(oldcircuitinfo.SleepWindow/1000000) {

			config := hystrix.CommandConfig{
				Timeout:                circuitinfo.Timeout,
				MaxConcurrentRequests:  circuitinfo.MaxConcurrentRequests,
				SleepWindow:            circuitinfo.SleepWindow,
				ErrorPercentThreshold:  circuitinfo.ErrorPercentThreshold,
				RequestVolumeThreshold: circuitinfo.RequestVolumeThreshold,
			}

			hystrix.ConfigureCommand(circuitinfo.Name, config)

		}

	}
}
