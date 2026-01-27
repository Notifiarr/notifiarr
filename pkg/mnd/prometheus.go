//nolint:gochecknoglobals
package mnd

import (
	"expvar"
	"strings"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
)

// Additional Prometheus counters for specific operations.
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

func (c *expvarCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.logFiles
	ch <- c.apiHits
	ch <- c.httpRequests
	ch <- c.timerEvents
	ch <- c.timerCounts
	ch <- c.website
	ch <- c.serviceChecks
	ch <- c.apps
	ch <- c.fileWatcher
}

func (c *expvarCollector) Collect(ch chan<- prometheus.Metric) {
	collectMap(ch, LogFiles, c.logFiles)
	collectMap(ch, APIHits, c.apiHits)
	collectMap(ch, HTTPRequests, c.httpRequests)
	collectSplitMap(ch, TimerEvents, c.timerEvents)
	collectMap(ch, TimerCounts, c.timerCounts)
	collectMap(ch, Website, c.website)
	collectSplitMap(ch, ServiceChecks, c.serviceChecks)
	collectSplitMap(ch, Apps, c.apps)
	collectMap(ch, FileWatcher, c.fileWatcher)
}

func collectMap(ch chan<- prometheus.Metric, m *expvar.Map, desc *prometheus.Desc) {
	m.Do(func(kv expvar.KeyValue) {
		val := getExpvarValue(kv.Value)
		if val >= 0 {
			ch <- prometheus.MustNewConstMetric(desc, prometheus.GaugeValue, val, kv.Key)
		}
	})
}

//nolint:mnd
func collectSplitMap(ch chan<- prometheus.Metric, m *expvar.Map, desc *prometheus.Desc) {
	m.Do(func(kv expvar.KeyValue) {
		keys := strings.SplitN(kv.Key, "&&", 2)
		if len(keys) != 2 {
			return
		}

		val := getExpvarValue(kv.Value)
		if val >= 0 {
			ch <- prometheus.MustNewConstMetric(desc, prometheus.GaugeValue, val, keys[0], keys[1])
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

func init() {
	prometheus.MustRegister(
		collectors.NewGoCollector(),
		collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}),
		newExpvarCollector(),
		SnapshotRuns,
		QueuePolls,
	)
}
