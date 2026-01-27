//nolint:gochecknoglobals
package mnd

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
)

// Prometheus metrics for monitoring notifiarr performance.
var (
	SnapshotRuns = prometheus.NewCounter(prometheus.CounterOpts{
		Namespace: "notifiarr",
		Name:      "snapshot_runs_total",
		Help:      "Total number of snapshot collection runs.",
	})
	QueuePolls = prometheus.NewCounter(prometheus.CounterOpts{
		Namespace: "notifiarr",
		Name:      "queue_polls_total",
		Help:      "Total number of starr queue poll operations.",
	})
	HTTPRequestsTotal = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: "notifiarr",
		Name:      "http_requests_total",
		Help:      "Total number of HTTP requests by status code.",
	}, []string{"code"})
)

func init() {
	prometheus.MustRegister(
		collectors.NewGoCollector(),
		collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}),
		SnapshotRuns,
		QueuePolls,
		HTTPRequestsTotal,
	)
}
