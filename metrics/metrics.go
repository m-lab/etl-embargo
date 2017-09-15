// The metrics package defines prometheus metric types and provides
// convenience methods to add accounting to embargo service..
package metrics

import (
	"net/http"
	"net/http/pprof"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	// Measures the number of files that was processed by embargo app engine successfully.
	// Provides metrics:
	//   embargo_success_total
	// Example usage:
	//   metrics.EmbargoSuccess.Inc()
	Metrics_embargoSuccessTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "embargo_Success_Total",
			Help: "Number of files that was processed by embargo app engine successfully.",
		},
		// "sidestream", "Monday"
		[]string{"experiment", "day_of_week"})

	// Measures the number of files that was not processed by embargo app engine successfully.
	// Provides metrics:
	//   embargo_error_total
	// Example usage:
	//   metrics.EmbargoError.Inc()
	Metrics_embargoErrorTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "embargo_Error_Total",
			Help: "Number of files that was not processed by embargo app engine successfully.",
		},
		// "sidestream", "Monday"
		[]string{"experiment", "day_of_week"})
)

func SetupPrometheus() {
	// Define a custom serve mux for prometheus to listen on a separate port.
	// We listen on a separate port so we can forward this port on the host VM.
	// We cannot forward port 8080 because it is used by AppEngine.
	mux := http.NewServeMux()
	// Assign the default prometheus handler to the standard exporter path.
	mux.Handle("/metrics", promhttp.Handler())
	// Assign the pprof handling paths to the external port to access individual
	// instances.
	mux.HandleFunc("/debug/pprof/", pprof.Index)
	mux.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
	mux.HandleFunc("/debug/pprof/profile", pprof.Profile)
	mux.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
	mux.HandleFunc("/debug/pprof/trace", pprof.Trace)

	http.Handle("/metrics", promhttp.Handler())
	// Register the metrics defined with Prometheus's default registry.
	prometheus.MustRegister(Metrics_embargoSuccessTotal)
	prometheus.MustRegister(Metrics_embargoErrorTotal)

	go http.ListenAndServe(":9090", mux)
}
