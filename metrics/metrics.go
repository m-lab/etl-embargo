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
	//   embargo_Success_Total
	// Example usage:
	//   metrics.Metrics_embargoSuccessTotal.WithLabelValues("sidestream", dayOfWeek).Inc()
	Metrics_embargoSuccessTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "embargo_Success_Total",
			Help: "Number of tar files that were processed by embargo app engine successfully.",
		},
		// "sidestream", "Monday"
		[]string{"experiment", "day_of_week"})

	// Measures the number of files that was not processed by embargo app engine successfully.
	// Provides metrics:
	//   embargo_Error_Total
	// Example usage:
	//   metrics.Metrics_embargoErrorTotal.WithLabelValues("sidestream", dayOfWeek).Inc()
	Metrics_embargoErrorTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "embargo_Error_Total",
			Help: "Number of tar files that were not processed by embargo app engine successfully.",
		},
		// "sidestream", "Monday"
		[]string{"experiment", "day_of_week"})

	// Measures the number of tests that were published through embargo app engine successfully.
	// Provides metrics:
	//   embargo_Public_Test_Total
	// Example usage:
	//   metrics.Metrics_embargoPublicTestTotal.WithLabelValues("sidestream", dayOfWeek).Inc()
	Metrics_embargoPublicTestTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "embargo_Public_Test_Total",
			Help: "Number of sidestream tests that were published by embargo app engine successfully.",
		},
		// "sidestream", "Monday"
		[]string{"experiment", "day_of_week"})

	// Measures the number of tests that were embargoed through embargo app engine successfully.
	// Provides metrics:
	//   embargo_Private_Test_Total
	// Example usage:
	//   metrics.Metrics_embargoPrivateTestTotal.WithLabelValues("sidestream", dayOfWeek).Inc()
	Metrics_embargoPrivateTestTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "embargo_Private_Test_Total",
			Help: "Number of sidestream tests that were embargoed by embargo app engine successfully.",
		},
		// "sidestream", "Monday"
		[]string{"experiment", "day_of_week"})

	// IPv6ErrorsTotal counts the kinds of errors encountered when normalizing IPv6 addresses.
	//
	// Example usage:
	//     metrics.IPv6ErrorsTotal.WithLabelValues(err.Error()).Inc()
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
	prometheus.MustRegister(Metrics_embargoSuccessTotal)
	prometheus.MustRegister(Metrics_embargoErrorTotal)
	prometheus.MustRegister(Metrics_embargoPublicTestTotal)
	prometheus.MustRegister(Metrics_embargoPrivateTestTotal)

	go http.ListenAndServe(":9090", mux)
}
