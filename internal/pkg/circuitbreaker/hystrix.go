package circuitbreaker

import (
	//"context"
	"encoding/json"
	"fmt"
	"github.com/PanYicheng/go-microservice/internal/pkg/messaging"
	"github.com/PanYicheng/go-microservice/internal/pkg/netutil"
	"github.com/PanYicheng/go-microservice/internal/pkg/tracing"
	"github.com/afex/hystrix-go/hystrix"
	"github.com/eapache/go-resiliency/retrier"
	"github.com/sirupsen/logrus"
	"io/ioutil"
	"net"
	"net/http"
	"time"
)

var Client http.Client

var RETRIES = 3

func CallUsingCircuitBreaker(breakerName string, req *http.Request) ([]byte, error) {
	output := make(chan []byte, 1)
	errors := hystrix.Go(
		breakerName,
		func() error {
			//tracing.AddTracingToReqFromContext(ctx, req)
			err := callWithRetries(req, output)
			// For hystrix, forward the err from the retrier. It's nil if OK.
			return err
		},
		func(err error) error {
			logrus.Errorf("In fallback function for breaker %v, error: %v", breakerName, err.Error())
			circuit, _, _ := hystrix.GetCircuit(breakerName)
			logrus.Errorf("Circuit state is: %v", circuit.IsOpen())
			return err
		},
	)

	select {
	case out := <-output:
		logrus.Debugf("Call in breaker %v successful", breakerName)
		return out, nil

	case err := <-errors:
		logrus.Debugf("Got error on channel in breaker %v. Msg: %v", breakerName, err.Error())
		return nil, err
	}
}

func callWithRetries(req *http.Request, output chan []byte) error {
	r := retrier.New(retrier.ConstantBackoff(RETRIES, 100*time.Millisecond), nil)
	attempt := 0
	err := r.Run(func() error {
		attempt++
		//resp, err := Client.Do(req)
		resp, err := tracing.TracingClient.DoWithAppSpan(req, req.Host)
		if err == nil && resp.StatusCode < 299 {
			responseBody, err := ioutil.ReadAll(resp.Body)
			if err == nil {
				output <- responseBody
				return nil
			}
			return err
		} else if err == nil {
			err = fmt.Errorf("Status was %v", resp.StatusCode)
		}

		logrus.Errorf("Retrier failed, attempt %v", attempt)

		return err
	})
	return err
}

func ConfigureHystrix(commands []string, mc messaging.IMessagingClient) {

	for _, command := range commands {
		hystrix.ConfigureCommand(command, hystrix.CommandConfig{
			Timeout:                resolveProperty(command, "Timeout"),
			MaxConcurrentRequests:  resolveProperty(command, "MaxConcurrentRequests"),
			ErrorPercentThreshold:  resolveProperty(command, "ErrorPercentThreshold"),
			RequestVolumeThreshold: resolveProperty(command, "RequestVolumeThreshold"),
			SleepWindow:            resolveProperty(command, "SleepWindow"),
		})
		logrus.Printf("Circuit %v settings: %v", command, hystrix.GetCircuitSettings()[command])
	}

	hystrixStreamHandler := hystrix.NewStreamHandler()
	hystrixStreamHandler.Start()
	go http.ListenAndServe(net.JoinHostPort("", "8181"), hystrixStreamHandler)
	logrus.Infoln("Launched hystrixStreamHandler at 8181")

	// Publish presence on RabbitMQ
	publishDiscoveryToken(mc)
}

func Deregister(amqpClient messaging.IMessagingClient) {
	ip := netutil.GetOutboundIP()
	token := DiscoveryToken{
		State:   "DOWN",
		Address: ip,
	}
	bytes, _ := json.Marshal(token)
	amqpClient.PublishOnQueue(bytes, "discovery")
}

func publishDiscoveryToken(mc messaging.IMessagingClient) {
	ip := netutil.GetOutboundIP()
	token := DiscoveryToken{
		State:   "UP",
		Address: ip,
	}
	bytes, _ := json.Marshal(token)
	go func() {
		for {
			mc.PublishOnQueue(bytes, "discovery")
			time.Sleep(time.Second * 30)
		}
	}()
}

func resolveProperty(command string, prop string) int {
	return getDefaultHystrixConfigPropertyValue(prop)
}

func getDefaultHystrixConfigPropertyValue(prop string) int {
	switch prop {
	case "Timeout":
		return hystrix.DefaultTimeout
	case "MaxConcurrentRequests":
		return hystrix.DefaultMaxConcurrent
	case "RequestVolumeThreshold":
		return hystrix.DefaultVolumeThreshold
	case "SleepWindow":
		return hystrix.DefaultSleepWindow
	case "ErrorPercentThreshold":
		return hystrix.DefaultErrorPercentThreshold
	}
	panic("Got unknown hystrix property: " + prop + ". Panicing!")
}

type DiscoveryToken struct {
	State   string `json:"state"` // UP, RUNNING, DOWN ??
	Address string `json:"address"`
}
