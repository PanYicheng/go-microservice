package service

import (
	"encoding/json"
	"github.com/sirupsen/logrus"
	"net/http"
	"strconv"
	"time"
	"github.com/PanYicheng/go-microservice/accountservice/dbclient"
	"github.com/PanYicheng/go-microservice/accountservice/model"
	"github.com/PanYicheng/go-microservice/common/messaging"
	"github.com/PanYicheng/go-microservice/common/util"
	cb "github.com/PanYicheng/go-microservice/common/circuitbreaker"
	"github.com/gorilla/mux"
)

// DBClient acts as database client
var DBClient dbclient.IBoltClient

// MessagingClient acts as messaging queue client
var MessagingClient messaging.IMessagingClient
var client = &http.Client{}

var fallbackQuote = model.Quote{
	Language: "en",
	ServedBy: "circuit-breaker",
	Text:     "May the source be with you, always."}

func init() {
	var transport http.RoundTripper = &http.Transport{
		DisableKeepAlives: true,
	}
	client.Transport = transport
}

// GetAccount handlers http request of /accounts/xxx
func GetAccount(w http.ResponseWriter, r *http.Request) {

	// Read the 'accountID' path parameter from the mux map
	var accountID = mux.Vars(r)["accountId"]

	// Read the account struct BoltDB
	account, err := DBClient.QueryAccount(accountID)
	account.ServedBy = util.GetIP()
	// If err, return a 404
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	account.Quote = getQuote()
	account.ImageUrl = getImageUrl(accountID)
	notifyVIP(account)
	// If found, marshal into JSON, write headers and content
	data, _ := json.Marshal(account)
	writeJSONResponse(w, http.StatusOK, data)
}

// If our hard-coded "VIP" account, spawn a goroutine to send a message.
func notifyVIP(account model.Account) {
	if account.Id == "10000" {
		go func(account model.Account) {
			vipNotification := model.VipNotification{AccountId: account.Id, ReadAt: time.Now().UTC().String()}
			data, _ := json.Marshal(vipNotification)
			err := MessagingClient.PublishOnQueue(data, "vipQueue")
			if err != nil {
				logrus.Println(err.Error())
			}
		}(account)
	}
}

func getQuote() (model.Quote) {
	body, err := cb.CallUsingCircuitBreaker("quotes-service", "http://quotes-service:8080/api/quote?strength=13", "GET")
	if err == nil {
		quote := model.Quote{}
		json.Unmarshal(body, &quote)
		return quote
	} else {
		return fallbackQuote
	}
}

func getImageUrl(accountId string) (string) {
        body, err := cb.CallUsingCircuitBreaker("imageservice", "http://imageservice:7777/accounts/" + accountId, "GET")
        if err == nil {
        return string(body)
    } else {
        return "http://path.to.placeholder"
    }
}

// HealthCheck is the http handlers for http request /health
func HealthCheck(w http.ResponseWriter, r *http.Request) {
	// Since we're here, we already know that HTTP service is up. Let's just check the state of the boltdb connection
	dbUp := DBClient.Check()
	if dbUp && isHealthy {
		data, _ := json.Marshal(healthCheckResponse{Status: "UP"})
		writeJSONResponse(w, http.StatusOK, data)
	} else {
		data, _ := json.Marshal(healthCheckResponse{Status: "Database unaccessible"})
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

var isHealthy = true // NEW

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
