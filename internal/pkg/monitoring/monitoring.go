package monitoring

import (
    "github.com/PanYicheng/go-microservice/internal/pkg/route"
    "github.com/prometheus/client_golang/prometheus"
    "github.com/spf13/viper"
    "net/http"
    "strconv"
    "time"
)

func BuildSummaryVec(metricName string, metricHelp string) *prometheus.SummaryVec {
    summaryVec := prometheus.NewSummaryVec(
        prometheus.SummaryOpts{
            Namespace: viper.GetString("service_name"),
            Name:      metricName,
            Help:      metricHelp,
			Objectives: map[float64]float64{0.5: 0.05, 0.9:0.01, 0.99:0.001, },
        },
        []string{"service"},
    )
    prometheus.Register(summaryVec)
    return summaryVec
}

// WithMonitoring optionally adds a middleware that stores request duration and response size into the supplied
// summaryVec
func WithMonitoring(next http.HandlerFunc, route route.Route, summary *prometheus.SummaryVec) http.HandlerFunc {

    // Just return the next handler if route shouldn't be monitored
    if !route.Monitor {
        return next
    }

    return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
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
