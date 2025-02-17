package cache

import (
	"net/http"
	"strconv"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var _ MetricsHandler = &noopMetrics{}

// MetricsHandler allows [LRUCache] to report metrics for monitoring.
// Handler should be go routine safe because while cache has its own internal lock
// the metrics are read by other go routines outside of cache's control.
// For test we use [noopMetrics]
// For http server, we use Prometheus/OpenTelemetry
type MetricsHandler interface {
	// Get

	AddNotFound()
	AddHit()

	// Set

	AddSet()
	AddSetExists()

	// Delete

	AddDelete()
	AddEvict()
	AddExpire(lazy bool)

	// Size

	SetSize(size int)
}

// MetricsExporter allows http server to export metrics.
// It should use same underlying implementation as [MetricsHandler].
type MetricsExporter interface {
	HTTPHandler() http.Handler
}

// do nothing

type noopMetrics struct{}

func (n *noopMetrics) AddNotFound()        {}
func (n *noopMetrics) AddHit()             {}
func (n *noopMetrics) AddSet()             {}
func (n *noopMetrics) AddSetExists()       {}
func (n *noopMetrics) AddDelete()          {}
func (n *noopMetrics) AddEvict()           {}
func (n *noopMetrics) AddExpire(lazy bool) {}
func (n *noopMetrics) SetSize(size int)    {}

type prometheusMetrics struct {
	notFound  *prometheus.CounterVec
	hit       *prometheus.CounterVec
	set       *prometheus.CounterVec
	setExists *prometheus.CounterVec
	delete    *prometheus.CounterVec
	evict     *prometheus.CounterVec
	expire    *prometheus.CounterVec
	size      *prometheus.GaugeVec
}

// NewPrometheusMetrics creates a new prometheus metrics handler
// that supports both [MetricsHandler] and [MetricsExporter]
func NewPrometheusMetrics() *prometheusMetrics {
	p := &prometheusMetrics{
		notFound: prometheus.NewCounterVec(prometheus.CounterOpts{
			Namespace: "cache",
			Subsystem: "lru",
			Name:      "not_found",
			Help:      "Number of not found keys",
		}, nil),
		hit: prometheus.NewCounterVec(prometheus.CounterOpts{
			Namespace: "cache",
			Subsystem: "lru",
			Name:      "hit",
			Help:      "Number of hit keys",
		}, nil),
		set: prometheus.NewCounterVec(prometheus.CounterOpts{
			Namespace: "cache",
			Subsystem: "lru",
			Name:      "set",
			Help:      "Number of set keys",
		}, nil),
		setExists: prometheus.NewCounterVec(prometheus.CounterOpts{
			Namespace: "cache",
			Subsystem: "lru",
			Name:      "set_exists",
			Help:      "Number of set keys that already exist",
		}, nil),
		delete: prometheus.NewCounterVec(prometheus.CounterOpts{
			Namespace: "cache",
			Subsystem: "lru",
			Name:      "delete",
			Help:      "Number of delete keys",
		}, nil),
		evict: prometheus.NewCounterVec(prometheus.CounterOpts{
			Namespace: "cache",
			Subsystem: "lru",
			Name:      "evict",
			Help:      "Number of evicted keys",
		}, nil),
		expire: prometheus.NewCounterVec(prometheus.CounterOpts{
			Namespace: "cache",
			Subsystem: "lru",
			Name:      "expire",
			Help:      "Number of expired keys",
		}, []string{"lazy"}),
		size: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: "cache",
			Subsystem: "lru",
			Name:      "size",
			Help:      "Number of keys in the cache",
		}, nil),
	}

	prometheus.MustRegister(p.notFound, p.hit, p.set, p.setExists, p.delete, p.evict, p.expire, p.size)
	return p
}

func (m *prometheusMetrics) HTTPHandler() http.Handler {
	return promhttp.Handler()
}

func (m *prometheusMetrics) AddNotFound() {
	m.notFound.WithLabelValues().Inc()
}

func (m *prometheusMetrics) AddHit() {
	m.hit.WithLabelValues().Inc()
}

func (m *prometheusMetrics) AddSet() {
	m.set.WithLabelValues().Inc()
}

func (m *prometheusMetrics) AddSetExists() {
	m.setExists.WithLabelValues().Inc()
}

func (m *prometheusMetrics) AddDelete() {
	m.delete.WithLabelValues().Inc()
}

func (m *prometheusMetrics) AddEvict() {
	m.evict.WithLabelValues().Inc()
}

func (m *prometheusMetrics) AddExpire(lazy bool) {
	m.expire.WithLabelValues(strconv.FormatBool(lazy)).Inc()
}

func (m *prometheusMetrics) SetSize(size int) {
	m.size.WithLabelValues().Set(float64(size))
}
