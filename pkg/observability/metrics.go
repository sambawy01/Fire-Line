package observability

import (
	"encoding/json"
	"net/http"
	"sort"
	"sync"
	"time"
)

// Metrics is a simple in-process metrics collector with no external dependencies.
type Metrics struct {
	mu             sync.Mutex
	requestCount   map[string]int64
	requestLatency map[string][]float64
	errorCount     map[string]int64
	startTime      time.Time
}

// NewMetrics creates a new Metrics instance.
func NewMetrics() *Metrics {
	return &Metrics{
		requestCount:   make(map[string]int64),
		requestLatency: make(map[string][]float64),
		errorCount:     make(map[string]int64),
		startTime:      time.Now(),
	}
}

// RecordRequest records a single HTTP request's method, path, status, and duration.
func (m *Metrics) RecordRequest(method, path string, status int, duration time.Duration) {
	key := method + " " + path
	m.mu.Lock()
	defer m.mu.Unlock()
	m.requestCount[key]++
	if status >= 500 {
		m.errorCount[key]++
	}
	m.requestLatency[key] = append(m.requestLatency[key], duration.Seconds())
	// Keep only last 1000 samples per endpoint
	if len(m.requestLatency[key]) > 1000 {
		m.requestLatency[key] = m.requestLatency[key][len(m.requestLatency[key])-1000:]
	}
}

// Handler returns an http.HandlerFunc that serves collected metrics as JSON.
func (m *Metrics) Handler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		m.mu.Lock()
		defer m.mu.Unlock()

		type endpointMetric struct {
			Endpoint     string  `json:"endpoint"`
			Requests     int64   `json:"requests"`
			Errors       int64   `json:"errors"`
			AvgLatencyMs float64 `json:"avg_latency_ms"`
			P95LatencyMs float64 `json:"p95_latency_ms"`
		}

		var metrics []endpointMetric
		for key, count := range m.requestCount {
			em := endpointMetric{
				Endpoint: key,
				Requests: count,
				Errors:   m.errorCount[key],
			}
			if latencies := m.requestLatency[key]; len(latencies) > 0 {
				var sum float64
				for _, l := range latencies {
					sum += l
				}
				em.AvgLatencyMs = (sum / float64(len(latencies))) * 1000

				// P95 calculation
				sorted := make([]float64, len(latencies))
				copy(sorted, latencies)
				sort.Float64s(sorted)
				idx := int(float64(len(sorted)) * 0.95)
				if idx >= len(sorted) {
					idx = len(sorted) - 1
				}
				em.P95LatencyMs = sorted[idx] * 1000
			}
			metrics = append(metrics, em)
		}

		resp := map[string]interface{}{
			"uptime_seconds": time.Since(m.startTime).Seconds(),
			"endpoints":      metrics,
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}
}

// MetricsMiddleware returns HTTP middleware that records request metrics.
func MetricsMiddleware(m *Metrics) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			sw := &statusWriter{ResponseWriter: w, status: 200}
			next.ServeHTTP(sw, r)
			m.RecordRequest(r.Method, r.URL.Path, sw.status, time.Since(start))
		})
	}
}

type statusWriter struct {
	http.ResponseWriter
	status int
}

func (sw *statusWriter) WriteHeader(code int) {
	sw.status = code
	sw.ResponseWriter.WriteHeader(code)
}
