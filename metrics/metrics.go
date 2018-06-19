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
	// Measures the number of tar files that was processed by embargo service.
	// Provides metrics:
	//   embargo_tar_input_total
	// Example usage:
	//   metrics.Metrics_embargoTarTotal.WithLabelValues("sidestream", "success").Inc()
	Metrics_embargoTarInputTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "embargo_tar_input_total",
			Help: "Number of tar files that were processed by embargo app engine.",
		},
		// "sidestream", "success/error"
		[]string{"dataset", "status"})

	// Measures the number of output tar files by embargo service.
	// Provides metrics:
	//   embargo_tar_output_total
	// Example usage:
	//   metrics.Metrics_embargoTarOutputTotal.WithLabelValues("sidestream", "public").Inc()
	Metrics_embargoTarOutputTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "embargo_tar_output_total",
			Help: "Number of tar output files by embargo app engine.",
		},
		// "sidestream", "public/private"
		[]string{"dataset", "status"})

	// Measures the number of web100 files that were processed by embargo service.
	// Provides metrics:
	//   embargo_file_total
	// Example usage:
	//   metrics.Metrics_embargoTestTotal.WithLabelValues("sidestream", "private").Inc()
	Metrics_embargoFileTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "embargo_file_total",
			Help: "Number of web100 sidestream files that were processed by embargo app engine.",
		},
		// "sidestream", "public/private"
		[]string{"dataset", "status"})

	// Measures the number of tar files that was unembargoed by daily unembargo cron job.
	// Provides metrics:
	//   unembargo_tar_total
	// Example usage:
	//   metrics.Metrics_unembargoTarTotal.WithLabelValues("sidestream").Inc()
	Metrics_unembargoTarTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "unembargo_tar_total",
			Help: "Number of sidestream tar files that were unembargoed.",
		},
		// "sidestream"
		[]string{"dataset"})

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
	prometheus.MustRegister(Metrics_embargoTarInputTotal)
	prometheus.MustRegister(Metrics_embargoTarOutputTotal)
	prometheus.MustRegister(Metrics_embargoFileTotal)
	prometheus.MustRegister(Metrics_unembargoTarTotal)

	go http.ListenAndServe(":9090", mux)
}
