package service

import (
	"context"
	"encoding/json"
	"github.com/sirupsen/logrus"
	"net/http"
	"strconv"
	"time"
	"github.com/PanYicheng/go-microservice/internal/app/accountservice/dbclient"
	"github.com/PanYicheng/go-microservice/internal/app/accountservice/model"
	"github.com/PanYicheng/go-microservice/internal/pkg/messaging"
	"github.com/PanYicheng/go-microservice/internal/pkg/netutil"
	"github.com/PanYicheng/go-microservice/internal/pkg/tracing"
	cb "github.com/PanYicheng/go-microservice/internal/pkg/circuitbreaker"
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
	logrus.Debugf("GetAccount: %s", accountID)

	// Read the account struct BoltDB
	account, err := DBClient.QueryAccount(r.Context(), accountID)
	account.ServedBy = netutil.GetOutboundIP()
	// If err, return a 404
	if err != nil {
		logrus.Errorf("Some error occured serving %v: %s", accountID, err.Error())
		w.WriteHeader(http.StatusNotFound)
		return
	}
	account.Quote = getQuote(r.Context())
	account.ImageUrl = getImageUrl(r.Context(), accountID)
	notifyVIP(r.Context(), account)
	// If found, marshal into JSON, write headers and content
	data, _ := json.Marshal(account)
	writeJSONResponse(w, http.StatusOK, data)
}

// If our hard-coded "VIP" account, spawn a goroutine to send a message.
func notifyVIP(ctx context.Context, account model.Account) {
	if account.Id == "10000" {
		tracing.LogEventToOngoingSpan(ctx, "Sent VIP message")
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

func getQuote(ctx context.Context) (model.Quote) {
	logrus.Debug("getQuote")
	// Start a new opentracing child span
	child := tracing.StartSpanFromContextWithLogEvent(ctx, "getQuote", "Client send")
	defer tracing.CloseSpan(child, "Client Receive")

	// Create the http request and pass it to the circuit breaker
	req, err := http.NewRequestWithContext(ctx, "GET", "http://quotes-service:8080/api/quote?strength=4", nil)
	body, err := cb.CallUsingCircuitBreaker("quotes-service", req)
	if err == nil {
		quote := model.Quote{}
		json.Unmarshal(body, &quote)
		return quote
	} else {
		return fallbackQuote
	}
}

func getImageUrl(ctx context.Context, accountID string) (string) {
	logrus.Debugf("getImageUrl: %s", accountID)
	child := tracing.StartSpanFromContextWithLogEvent(ctx, "getImageUrl", "Client send")
	defer tracing.CloseSpan(child, "Client Receive")

	req, err := http.NewRequestWithContext(ctx, "GET", "http://imageservice:7777/accounts/" + accountID, nil)
	body, err := cb.CallUsingCircuitBreaker("imageservice", req)
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
