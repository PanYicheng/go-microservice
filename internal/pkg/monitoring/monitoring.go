package monitoring

import (
	"github.com/sirupsen/logrus"
    "github.com/PanYicheng/go-microservice/internal/pkg/route"
    "github.com/prometheus/client_golang/prometheus"
    "net/http"
    "strconv"
    "time"
)

type OtherVec struct {
	totalRequest    *prometheus.CounterVec
	inflightRequest *prometheus.GaugeVec
	responseTime    *prometheus.HistogramVec
	valid           bool
}

var otherVec OtherVec
var serviceName string

func init() {
	otherVec.valid = false
}

func BuildOtherVec(name string) {
	logrus.Debugf("BuildOtherVec is called for %s.\n", name)
	serviceName = name
	otherVec.totalRequest = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: name + "_processed_request",
			Help: "Total number of request processed by the service",
		},
		[]string{},
	)

	otherVec.inflightRequest = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: name + "_inflight_request",
			Help: "Total number of inflight request",
		},
		[]string{},
	)

	otherVec.responseTime = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    name + "_response_time",
			Help:    "Reponse time of request",
			Buckets: []float64{0.2, 0.5, 1},
		},
		[]string{},
	)
	prometheus.MustRegister(otherVec.totalRequest)
	prometheus.MustRegister(otherVec.inflightRequest)
	prometheus.MustRegister(otherVec.responseTime)
	otherVec.valid = true
}

func BuildSummaryVec(metricName string, metricHelp string) *prometheus.SummaryVec {
    summaryVec := prometheus.NewSummaryVec(
        prometheus.SummaryOpts{
            Name:       serviceName + "_" + metricName + "_summary",
            Help:       metricHelp,
			Objectives: map[float64]float64{0.5: 0.05, 0.9:0.01, 0.95: 0.005, 0.99:0.001},
			MaxAge:     5 * time.Second, // Percentile metrics will live for 5s
        },
        []string{"service_summary"},
    )
    prometheus.Register(summaryVec)
    return summaryVec
}

func BuildHistoVec(metricName string, metricHelp string) *prometheus.HistogramVec {
	histoVec := prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:      serviceName + "_" + metricName + "_histogram",
			Help:      metricHelp,
			Buckets:   []float64{.001, .0025, .005, .0075, .01, .025, .05, .075, .1, .25, .5, .75, 1, 1.5, 2, 2.5, 5},
		},
		[]string{"service_histogram"},
		)
	prometheus.Register(histoVec)
	return histoVec
}

// monitorOtherVec sets package level prometheus variables.
func monitorOtherVec() {
	logrus.Debug("monitorOtherVec")
	if otherVec.valid == true {
		otherVec.totalRequest.WithLabelValues().Inc()
	}
}

// WithMonitoringSummary optionally adds a middleware that stores request duration and response size
// into the supplied summaryVec
func WithMonitoringSummary(next http.HandlerFunc, route route.Route, summary *prometheus.SummaryVec) http.HandlerFunc {

	// Just return the next handler if route shouldn't be monitored
	if !route.Monitor {
		return next
	}

	return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		//set other prometheus vector
		monitorOtherVec()

		start := time.Now()
		// next.ServeHTTP(rw, req)
		next(rw, req)
		duration := time.Since(start)

		// Store duration of request
		summary.WithLabelValues("duration").Observe(duration.Seconds())

		// Store size of response, if possible.
		size, err := strconv.Atoi(rw.Header().Get("Content-Length"))
		if err == nil {
			summary.WithLabelValues("size").Observe(float64(size))
		}
	})
}

// WithMonitoringHisto optionally adds a middleware that stores request duration and response size
// into the supplied histoVec
func WithMonitoringHisto(next http.HandlerFunc, route route.Route, histo *prometheus.HistogramVec) http.HandlerFunc {

	// Just return the next handler if route shouldn't be monitored
	if !route.Monitor {
		return next
	}

	return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		//set other prometheus vector
		monitorOtherVec()

		start := time.Now()
		// next.ServeHTTP(rw, req)
		next(rw, req)
		duration := time.Since(start)

		// Store duration of request
		histo.WithLabelValues("duration").Observe(duration.Seconds())

		// Store size of response, if possible.
		size, err := strconv.Atoi(rw.Header().Get("Content-Length"))
		if err == nil {
			histo.WithLabelValues("size").Observe(float64(size))
		}
	})
}

func UnregisterGoCollector() {
	prometheus.Unregister(prometheus.NewGoCollector());
}