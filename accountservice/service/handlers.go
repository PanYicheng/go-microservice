package service
import (
	"github.com/PanYicheng/go-microservice/accountservice/dbclient"
	"github.com/PanYicheng/go-microservice/accountservice/model"
	"github.com/gorilla/mux"
	"net/http"
	"net"
	"encoding/json"
	"strconv"
	"fmt"
	"io/ioutil"
)

var DBClient dbclient.IBoltClient

var client = &http.Client{}

func init() {
        var transport http.RoundTripper = &http.Transport{
                DisableKeepAlives: true,
        }
        client.Transport = transport
}

func GetAccount(w http.ResponseWriter, r *http.Request) {

	// Read the 'accountId' path parameter from the mux map
	var accountId = mux.Vars(r)["accountId"]

	// Read the account struct BoltDB
	account, err := DBClient.QueryAccount(accountId)
	account.ServedBy = getIP()
	// If err, return a 404
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	quote, err := getQuote()
	if err == nil {
			account.Quote = quote
	}

	// If found, marshal into JSON, write headers and content
	data, _ := json.Marshal(account)
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Content-Length", strconv.Itoa(len(data)))
	w.WriteHeader(http.StatusOK)
	w.Write(data)
}

func getIP() string {
        addrs, err := net.InterfaceAddrs()
        if err != nil {
                return "error"
        }
        for _, address := range addrs {
                // check the address type and if it is not a loopback the display it
                if ipnet, ok := address.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
                        if ipnet.IP.To4() != nil {
                                return ipnet.IP.String()
                        }
                }
        }
        panic("Unable to determine local IP address (non loopback). Exiting.")
}

func getQuote() (model.Quote, error) {
        req, _ := http.NewRequest("GET", "http://quotes-service:8080/api/quote?strength=4", nil)
        resp, err := client.Do(req)

        if err == nil && resp.StatusCode == 200 {
                quote := model.Quote{}
                bytes, _ := ioutil.ReadAll(resp.Body)
                json.Unmarshal(bytes, &quote)
                return quote, nil
        } else {
                return model.Quote{}, fmt.Errorf("Some error")
        }
}

func HealthCheck(w http.ResponseWriter, r *http.Request) {
        // Since we're here, we already know that HTTP service is up. Let's just check the state of the boltdb connection
        dbUp := DBClient.Check()
        if dbUp && isHealthy {
                data, _ := json.Marshal(healthCheckResponse{Status: "UP"})
                writeJsonResponse(w, http.StatusOK, data)
        } else {
                data, _ := json.Marshal(healthCheckResponse{Status: "Database unaccessible"})
                writeJsonResponse(w, http.StatusServiceUnavailable, data)
        }
}

func writeJsonResponse(w http.ResponseWriter, status int, data []byte) {
        w.Header().Set("Content-Type", "application/json")
        w.Header().Set("Content-Length", strconv.Itoa(len(data)))
        w.WriteHeader(status)
        w.Write(data)
}

type healthCheckResponse struct {
        Status string `json:"status"`
}

var isHealthy = true // NEW

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
