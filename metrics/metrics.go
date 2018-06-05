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
	// Measures the number of tar files that was processed by embargo app engine.
	// Provides metrics:
	//   embargo_tar_total
	// Example usage:
	//   metrics.Metrics_embargoTarTotal.WithLabelValues("sidestream", "status").Inc()
	Metrics_embargoTarTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "embargo_tar_total",
			Help: "Number of tar files that were processed by embargo app engine.",
		},
		// "sidestream", "success/error"
		[]string{"dataset", "status"})

	// Measures the number of tar files that was output by embargo app engine.
	// Provides metrics:
	//   embargo_tar_output_total
	// Example usage:
	//   metrics.Metrics_embargoTarTotal.WithLabelValues("sidestream", "status").Inc()
	Metrics_embargoTarOutputTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "embargo_tar_output_total",
			Help: "Number of tar files that were processed by embargo app engine.",
		},
		// "sidestream", "public/private"
		[]string{"dataset", "status"})

	// Measures the number of tests that were processed through embargo app engine.
	// Provides metrics:
	//   embargo_test_total
	// Example usage:
	//   metrics.Metrics_embargoTestTotal.WithLabelValues("sidestream", "status").Inc()
	Metrics_embargoTestTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "embargo_test_total",
			Help: "Number of sidestream tests that were processed by embargo app engine.",
		},
		// "sidestream", "public/private"
		[]string{"dataset", "status"})

	// IPv6ErrorsTotal counts the kinds of errors encountered when normalizing IPv6 addresses.
	// Provides metrics:
	//   embargo_ipv6_errors_total
	// Example usage:
	//   metrics.IPv6ErrorsTotal.WithLabelValues(err.Error()).Inc()
	IPv6ErrorsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "embargo_ipv6_errors_total",
			Help: "Number of failures normalizing IPv6 addresses.",
		},
		[]string{"error"})
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
	prometheus.MustRegister(Metrics_embargoTarTotal)
	prometheus.MustRegister(Metrics_embargoTarOutputTotal)
	prometheus.MustRegister(Metrics_embargoTestTotal)

	go http.ListenAndServe(":9090", mux)
}
