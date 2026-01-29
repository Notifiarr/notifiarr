package mnd

import (
	"expvar"
	"strings"

	"github.com/prometheus/client_golang/prometheus"
)

// Additional Prometheus counters for specific operations.
// These might be the same as things in expvar already.
//
//nolint:gochecknoglobals
var (
	SnapshotRuns = prometheus.NewCounter(prometheus.CounterOpts{
		Namespace: "notifiarr_triggers",
		Name:      "snapshot_runs_total",
		Help:      "Total number of snapshot collection runs.",
	})
	StuckItems = prometheus.NewCounter(prometheus.CounterOpts{
		Namespace: "notifiarr_triggers",
		Name:      "stuck_items_processed_total",
		Help:      "Total number of times stuck items processing triggered.",
	})
)

// Custom collector that exports all expvar metrics to Prometheus.
type expvarCollector struct {
	logFiles      *prometheus.Desc
	apiHits       *prometheus.Desc
	httpRequests  *prometheus.Desc
	timerEvents   *prometheus.Desc
	timerCounts   *prometheus.Desc
	website       *prometheus.Desc
	serviceChecks *prometheus.Desc
	apps          *prometheus.Desc
	fileWatcher   *prometheus.Desc
}

func newExpvarCollector() *expvarCollector {
	return &expvarCollector{
		logFiles: prometheus.NewDesc(
			"notifiarr_log_files",
			"Log file information",
			[]string{"name"}, nil,
		),
		apiHits: prometheus.NewDesc(
			"notifiarr_api_hits",
			"Incoming API requests",
			[]string{"name"}, nil,
		),
		httpRequests: prometheus.NewDesc(
			"notifiarr_http_requests",
			"Incoming HTTP requests",
			[]string{"name"}, nil,
		),
		timerEvents: prometheus.NewDesc(
			"notifiarr_timer_events",
			"Triggers and timers executed",
			[]string{"trigger", "name"}, nil,
		),
		timerCounts: prometheus.NewDesc(
			"notifiarr_timer_counts",
			"Triggers and timers counters",
			[]string{"name"}, nil,
		),
		website: prometheus.NewDesc(
			"notifiarr_website",
			"Outbound requests to website",
			[]string{"name"}, nil,
		),
		serviceChecks: prometheus.NewDesc(
			"notifiarr_service_checks",
			"Service check responses",
			[]string{"service", "name"}, nil,
		),
		apps: prometheus.NewDesc(
			"notifiarr_apps",
			"Starr app requests",
			[]string{"app", "name"}, nil,
		),
		fileWatcher: prometheus.NewDesc(
			"notifiarr_file_watcher",
			"File watcher metrics",
			[]string{"name"}, nil,
		),
	}
}

func (c *expvarCollector) Describe(metrics chan<- *prometheus.Desc) {
	metrics <- c.logFiles
	metrics <- c.apiHits
	metrics <- c.httpRequests
	metrics <- c.timerEvents
	metrics <- c.timerCounts
	metrics <- c.website
	metrics <- c.serviceChecks
	metrics <- c.apps
	metrics <- c.fileWatcher
}

func (c *expvarCollector) Collect(metrics chan<- prometheus.Metric) {
	collectMap(metrics, LogFiles, c.logFiles)
	collectMap(metrics, APIHits, c.apiHits)
	collectMap(metrics, HTTPRequests, c.httpRequests)
	collectSplitMap(metrics, TimerEvents, c.timerEvents)
	collectMap(metrics, TimerCounts, c.timerCounts)
	collectMap(metrics, Website, c.website)
	collectSplitMap(metrics, ServiceChecks, c.serviceChecks)
	collectSplitMap(metrics, Apps, c.apps)
	collectMap(metrics, FileWatcher, c.fileWatcher)
}

func collectMap(metrics chan<- prometheus.Metric, m *expvar.Map, desc *prometheus.Desc) {
	m.Do(func(keyval expvar.KeyValue) {
		val := getExpvarValue(keyval.Value)
		if val >= 0 {
			metrics <- prometheus.MustNewConstMetric(desc, prometheus.GaugeValue, val, keyval.Key)
		}
	})
}

//nolint:mnd
func collectSplitMap(metrics chan<- prometheus.Metric, m *expvar.Map, desc *prometheus.Desc) {
	m.Do(func(keyval expvar.KeyValue) {
		keys := strings.SplitN(keyval.Key, "&&", 2)
		if len(keys) != 2 {
			return
		}

		val := getExpvarValue(keyval.Value)
		if val >= 0 {
			metrics <- prometheus.MustNewConstMetric(desc, prometheus.GaugeValue, val, keys[0], keys[1])
		}
	})
}

func getExpvarValue(v expvar.Var) float64 {
	switch val := v.(type) {
	case *expvar.Int:
		return float64(val.Value())
	case *expvar.Float:
		return val.Value()
	case expvar.Func:
		if i, ok := val.Value().(int64); ok {
			return float64(i)
		}
	}

	return -1
}

//nolint:gochecknoinits
func init() {
	// Go and Process collectors are registered by default.
	prometheus.MustRegister(
		newExpvarCollector(),
		SnapshotRuns,
		StuckItems,
	)
}
