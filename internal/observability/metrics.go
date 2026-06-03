package observability

import (
	"net/http"
	"strings"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var registry = prometheus.NewRegistry()

var (
	executionsTotal = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "code_execution_requests_total",
			Help: "Total number of recorded code executions.",
		},
	)
	executionStatusTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "code_execution_status_total",
			Help: "Number of executions grouped by final status.",
		},
		[]string{"status"},
	)
	executionLatency = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "code_execution_latency_seconds",
			Help:    "Execution latency in seconds grouped by language.",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"language"},
	)
)

func init() {
	registry.MustRegister(executionsTotal)
	registry.MustRegister(executionStatusTotal)
	registry.MustRegister(executionLatency)
}

func Handler() http.Handler {
	return promhttp.HandlerFor(registry, promhttp.HandlerOpts{})
}

func RecordExecution(language string, status string, duration time.Duration) {
	executionsTotal.Inc()
	executionStatusTotal.WithLabelValues(normalizeStatus(status)).Inc()
	executionLatency.WithLabelValues(normalizeLanguage(language)).Observe(duration.Seconds())
}

func normalizeLanguage(language string) string {
	language = strings.TrimSpace(strings.ToLower(language))
	if language == "" {
		return "unknown"
	}
	return language
}

func normalizeStatus(status string) string {
	status = strings.TrimSpace(strings.ToLower(status))
	switch status {
	case "successfully executed":
		return "success"
	case "compilation error":
		return "compile_error"
	case "runtime error":
		return "runtime_error"
	case "time limit exceeded":
		return "tle"
	case "memory limit":
		return "mle"
	case "system error":
		return "system_error"
	case "queued":
		return "queued"
	case "processing":
		return "processing"
	case "stored":
		return "stored"
	default:
		if status == "" {
			return "unknown"
		}
		return strings.ReplaceAll(status, " ", "_")
	}
}
