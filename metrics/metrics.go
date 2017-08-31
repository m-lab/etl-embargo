// The metrics package defines prometheus metric types and provides
// convenience methods to add accounting to embargo service..
package metrics

import (
	"math"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

func init() {
	// Register the metrics defined with Prometheus's default registry.
	prometheus.MustRegister(EmbargoSuccess)
	prometheus.MustRegister(EmbargoError)
}

var (

	// Measures the number of files that was processed by embargo app engine successfully.
	// Provides metrics:
	//   embargo_success_total
	// Example usage:
	//   metrics.EmbargoSuccess.Inc() / .Dec()
	EmbargoSuccess = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "embargo_success_total",
		Help: "Number of files that was processed by embargo app engine successfully.",
	},
		// "sidestream", "Monday"
		[]string{"experiment", "day_of_week"})

	// Measures the number of files that was not processed by embargo app engine successfully.
	// Provides metrics:
	//   embargo_error_total
	// Example usage:
	//   metrics.EmbargoError.Inc() / .Dec()
	EmbargoError = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "embargo_error_total",
		Help: "Number of files that was not processed by embargo app engine successfully.",
	},
		// "sidestream", "Monday"
		[]string{"experiment", "day_of_week"})
)
