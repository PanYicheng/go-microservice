package service

import (
	"encoding/json"
	"fmt"
	"github.com/PanYicheng/go-microservice/common/messaging"
	"github.com/gorilla/mux"
	"net/http"
	"strconv"
)

// MessagingClient acts as messaging queue client
var MessagingClient messaging.IMessagingClient
var client = &http.Client{}

func init() {
	var transport http.RoundTripper = &http.Transport{
		DisableKeepAlives: true,
	}
	client.Transport = transport
}

var isHealthy = true

// HealthCheck is the http handlers for http request /health
func HealthCheck(w http.ResponseWriter, r *http.Request) {
	if isHealthy {
		data, _ := json.Marshal(healthCheckResponse{Status: "UP"})
		writeJSONResponse(w, http.StatusOK, data)
	} else {
		data, _ := json.Marshal(healthCheckResponse{Status: "Unhealthy"})
		writeJSONResponse(w, http.StatusServiceUnavailable, data)
	}
}

func writeJSONResponse(w http.ResponseWriter, status int, data []byte) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Content-Length", strconv.Itoa(len(data)))
	w.WriteHeader(status)
	w.Write(data)
}

type healthCheckResponse struct {
	Status string `json:"status"`
}

// SetHealthyState sets the isHealthy variable with http requests
func SetHealthyState(w http.ResponseWriter, r *http.Request) {

	// Read the 'state' path parameter from the mux map and convert to a bool
	var state, err = strconv.ParseBool(mux.Vars(r)["state"])

	// If we couldn't parse the state param, return a HTTP 400
	if err != nil {
		fmt.Println("Invalid request to SetHealthyState, allowed values are true or false")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Otherwise, mutate the package scoped "isHealthy" variable.
	isHealthy = state
	w.WriteHeader(http.StatusOK)
}
