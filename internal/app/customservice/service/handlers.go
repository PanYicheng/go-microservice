package service

import (
	"io"
	"io/ioutil"
	"sync"
	"os"
	"context"
	"encoding/json"
	"github.com/sirupsen/logrus"
	"net/http"
	"strconv"
	"time"
	. "github.com/PanYicheng/go-microservice/internal/app/customservice/model"
	"github.com/PanYicheng/go-microservice/internal/pkg/netutil"
	// "github.com/PanYicheng/go-microservice/internal/pkg/tracing"
	cb "github.com/PanYicheng/go-microservice/internal/pkg/circuitbreaker"
	"github.com/afex/hystrix-go/hystrix"
	"github.com/gorilla/mux"
)

var client = &http.Client{}
var ServiceConfig Service

func init() {
	var transport http.RoundTripper = &http.Transport{
		DisableKeepAlives: true,
	}
	client.Transport = transport
}

// Index is the main HTTP handler function. It calls all children services, sequentailly or concurrently.
func Index(w http.ResponseWriter, r *http.Request) {
	// Create zipkin spans.
	// span := getSpan(r)
	// defer span.Finish()

	// 模拟服务响应时间
	time.Sleep(time.Millisecond * time.Duration(ServiceConfig.ResponseTime))

	var response Response
	if ServiceConfig.Concurrency == true {
		response = subIndexConcurrency(r.Context())
	} else {
		response = subIndexSequential(r.Context())
	}
	response.ServiceName = ServiceConfig.Name
	response.Ip = netutil.GetOutboundIP()
	bytes, _ := json.MarshalIndent(response, "", "    ")
	writeJSONResponse(w, 200, bytes)
}

// getOneChild gets responses from one child service specified in srvAddr.
func getOneChild(ctx context.Context, srvAddr string) Response {
	logrus.Debugf("Calling child service: %s", srvAddr)
	req, err := http.NewRequestWithContext(ctx, "GET", "http://"+srvAddr, nil)
	body, err := cb.CallUsingCircuitBreaker(ServiceConfig.Name+"To"+srvAddr, req)
	var childResp = Response{
		ServiceName: srvAddr,
		Ip: "fallback IP",
		Data: nil,
		ErrorInfo: "",
		Children: nil,
	}
	if err == nil {
		err := json.Unmarshal(body, &childResp)
		if err != nil {
			logrus.Error(err.Error())
			childResp.ErrorInfo = err.Error()
		}
	} else {
		logrus.Error(err.Error())
		childResp.ErrorInfo = err.Error()
	}
	return childResp
}

// subIndexSequential implements a sequential child service calls.
func subIndexSequential(ctx context.Context) Response {
	response := Response{}
	for _, call := range ServiceConfig.CallList {
		childResp := getOneChild(ctx, call)
		response.Children = append(response.Children, childResp)
	}
	return response
}

// indexConcurrency implements a concurrent child service calls.
func subIndexConcurrency(ctx context.Context) Response {
	response := Response{}
	var wg sync.WaitGroup
	respChan := make(chan Response, len(ServiceConfig.CallList))
	for _, call := range ServiceConfig.CallList {
		wg.Add(1)
		go func(call string) {
			respChan <- getOneChild(ctx, call)
		}(call)
	}
	go func(){
		for resp := range respChan {
			wg.Done()
			response.Children = append(response.Children, resp)
		}
	}()
	wg.Wait()
	return response
}

// SetResponseTime sets the internal response time of this service.
func SetResponseTime(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(io.LimitReader(r.Body, 1024))
	if err != nil {
		writeJSONResponse(w, 500, []byte(err.Error()))
		return
	}

	var srv Service
	err = json.Unmarshal(body, &srv)
	if err != nil {
		writeJSONResponse(w, 500, []byte(err.Error()))
		return
	}

	err = r.Body.Close()
	if err != nil {
		writeJSONResponse(w, 500, []byte(err.Error()))
		return
	}

	ServiceConfig.ResponseTime = srv.ResponseTime

	confFd, err := os.OpenFile("data/conf.json", os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0644)
	if err != nil {
		writeJSONResponse(w, 500, []byte(err.Error()))
		return
	}
	defer confFd.Close()

	// 写 conf.json
	bytes, err := json.MarshalIndent(ServiceConfig, "", "    ")
	if err != nil {
		writeJSONResponse(w, 500, []byte(err.Error()))
		return
	}
	_, err = confFd.Write(bytes)
	if err != nil {
		writeJSONResponse(w, 500, []byte(err.Error()))
		return
	}

	writeJSONResponse(w, 200, []byte("success\n"))
}

// GetServiceInfo returns service info.
func GetServiceInfo(w http.ResponseWriter, r *http.Request) {
	bytes, err := json.MarshalIndent(ServiceConfig, "", "    ")
	if err != nil {
		writeJSONResponse(w, 500, []byte(err.Error()))
		return
	}
	writeJSONResponse(w, 200, bytes)
}

// GetCircuitInfo returns circuit breaker info.
func GetCircuitInfo(w http.ResponseWriter, r *http.Request) {
	settings := hystrix.GetCircuitSettings()

	bytes, err := json.MarshalIndent(settings, "", "    ")
	if err != nil {
		writeJSONResponse(w, 500, []byte(err.Error()))
		return
	}
	writeJSONResponse(w, 200, bytes)
}

// SetCircuitInfo sets circuit breaker settings.
func SetCircuitInfo(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(io.LimitReader(r.Body, 1024))
	if err != nil {
		writeJSONResponse(w, 500, []byte(err.Error()))
		return
	}

	var setting hystrix.Settings
	err = json.Unmarshal(body, &setting)
	if err != nil {
		writeJSONResponse(w, 500, []byte(err.Error()))
		return
	}

	err = r.Body.Close()
	if err != nil {
		writeJSONResponse(w, 500, []byte(err.Error()))
		return
	}

	params := mux.Vars(r)
	name := params["name"]

	settings := hystrix.GetCircuitSettings()
	oldsetting := settings[name]

	conf := hystrix.CommandConfig{}
	if setting.Timeout == 0 {
		conf.Timeout = int(oldsetting.Timeout / 1000000)
	} else {
		conf.Timeout = int(setting.Timeout / 1000000)
	}
	if setting.MaxConcurrentRequests == 0 {
		conf.MaxConcurrentRequests = oldsetting.MaxConcurrentRequests
	} else {
		conf.MaxConcurrentRequests = setting.MaxConcurrentRequests
	}
	if setting.RequestVolumeThreshold == 0 {
		conf.RequestVolumeThreshold = int(oldsetting.RequestVolumeThreshold)
	} else {
		conf.RequestVolumeThreshold = int(setting.RequestVolumeThreshold)
	}
	if setting.SleepWindow == 0 {
		conf.SleepWindow = int(oldsetting.SleepWindow / 1000000)
	} else {
		conf.SleepWindow = int(setting.SleepWindow / 1000000)
	}
	if setting.ErrorPercentThreshold == 0 {
		conf.ErrorPercentThreshold = oldsetting.ErrorPercentThreshold
	} else {
		conf.ErrorPercentThreshold = setting.ErrorPercentThreshold
	}
	hystrix.ConfigureCommand(name, conf)
	writeJSONResponse(w, 200, []byte("success\n"))
}

// SetConcurrency sets the concurrency settings.
func SetConcurrency(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(io.LimitReader(r.Body, 1024))
	if err != nil {
		writeJSONResponse(w, 500, []byte(err.Error()))
		return
	}

	var srv Service
	err = json.Unmarshal(body, &srv)
	if err != nil {
		writeJSONResponse(w, 500, []byte(err.Error()))
		return
	}

	err = r.Body.Close()
	if err != nil {
		writeJSONResponse(w, 500, []byte(err.Error()))
		return
	}

	ServiceConfig.Concurrency = srv.Concurrency

	confFd, err := os.OpenFile("data/conf.json", os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0644)
	if err != nil {
		writeJSONResponse(w, 500, []byte(err.Error()))
		return
	}
	defer confFd.Close()

	// 写 conf.json
	bytes, err := json.MarshalIndent(ServiceConfig, "", "    ")
	if err != nil {
		writeJSONResponse(w, 500, []byte(err.Error()))
		return
	}
	_, err = confFd.Write(bytes)
	if err != nil {
		writeJSONResponse(w, 500, []byte(err.Error()))
		return
	}

	writeJSONResponse(w, 200, []byte("success\n"))
}

// writeJSONResponse is a helper function that writes HTTP code and json data as bytes.
func writeJSONResponse(w http.ResponseWriter, status int, data []byte) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Content-Length", strconv.Itoa(len(data)))
	w.WriteHeader(status)
	w.Write(data)
}

type healthCheckResponse struct {
	Status string `json:"status"`
}

var isHealthy = true // NEW

// HealthCheck is the http handlers for http request /health
func HealthCheck(w http.ResponseWriter, r *http.Request) {
	// Since we're here, we already know that HTTP service is up. Let's just check the state of the boltdb connection
	if isHealthy {
		data, _ := json.Marshal(healthCheckResponse{Status: "UP"})
		writeJSONResponse(w, http.StatusOK, data)
	} else {
		data, _ := json.Marshal(healthCheckResponse{Status: "Down"})
		writeJSONResponse(w, http.StatusServiceUnavailable, data)
	}
}
// SetHealthyState sets the isHealthy variable with http requests
func SetHealthyState(w http.ResponseWriter, r *http.Request) {
	// Read the 'state' path parameter from the mux map and convert to a bool
	var state, err = strconv.ParseBool(mux.Vars(r)["state"])

	// If we couldn't parse the state param, return a HTTP 400
	if err != nil {
		logrus.Println("Invalid request to SetHealthyState, allowed values are true or false")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Otherwise, mutate the package scoped "isHealthy" variable.
	isHealthy = state
	w.WriteHeader(http.StatusOK)
}
