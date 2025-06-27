package metrics

import (
	"fmt"
	"sync"
	"sync/atomic"
)

// MetricType defines the type of metric.
type MetricType int

const (
	Counter MetricType = iota
	Gauge
)

// Metric represents a single metric entry.
type Metric struct {
	Name  string
	Type  MetricType
	atomicValue atomic.Int64 // Using atomic for thread-safe operations
}

// MetricsRegistry holds all registered metrics.
type MetricsRegistry struct {
	metrics map[string]*Metric
	mu      sync.RWMutex
}

// NewMetricsRegistry creates a new MetricsRegistry.
func NewMetricsRegistry() *MetricsRegistry {
	return &MetricsRegistry{
		metrics: make(map[string]*Metric),
	}
}

// RegisterCounter registers a new counter metric.
func (mr *MetricsRegistry) RegisterCounter(name string) (*Metric, error) {
	mr.mu.Lock()
	defer mr.mu.Unlock()

	if _, exists := mr.metrics[name]; exists {
		return nil, fmt.Errorf("metric '%s' already registered", name)
	}

	metric := &Metric{Name: name, Type: Counter}
	mr.metrics[name] = metric
	return metric, nil
}

// RegisterGauge registers a new gauge metric.
func (mr *MetricsRegistry) RegisterGauge(name string) (*Metric, error) {
	mr.mu.Lock()
	defer mr.mu.Unlock()

	if _, exists := mr.metrics[name]; exists {
		return nil, fmt.Errorf("metric '%s' already registered", name)
	}

	metric := &Metric{Name: name, Type: Gauge}
	mr.metrics[name] = metric
	return metric, nil
}

// GetMetric retrieves a metric by name.
func (mr *MetricsRegistry) GetMetric(name string) *Metric {
	mr.mu.RLock()
	defer mr.mu.RUnlock()
	return mr.metrics[name]
}

// Inc increments a counter metric by 1.
func (m *Metric) Inc() {
	if m.Type == Counter {
		m.atomicValue.Add(1)
	}
}

// Add adds a value to a counter metric.
func (m *Metric) Add(delta int64) {
	if m.Type == Counter {
		m.atomicValue.Add(delta)
	}
}

// Set sets the value of a gauge metric.
func (m *Metric) Set(value int64) {
	if m.Type == Gauge {
		m.atomicValue.Store(value)
	}
}

// Value returns the current value of the metric.
func (m *Metric) Value() int64 {
	return m.atomicValue.Load()
}

// Global metrics registry instance
var defaultRegistry = NewMetricsRegistry()

// RegisterCounter registers a counter metric with the default registry.
func RegisterCounter(name string) (*Metric, error) {
	return defaultRegistry.RegisterCounter(name)
}

// RegisterGauge registers a gauge metric with the default registry.
func RegisterGauge(name string) (*Metric, error) {
	return defaultRegistry.RegisterGauge(name)
}

// GetMetric retrieves a metric from the default registry.
func GetMetric(name string) *Metric {
	return defaultRegistry.GetMetric(name)
}

// Inc increments a counter metric by 1 using the default registry.
func Inc(name string) {
	if m := GetMetric(name); m != nil {
		m.Inc()
	}
}

// Add adds a value to a counter metric using the default registry.
func Add(name string, delta int64) {
	if m := GetMetric(name); m != nil {
		m.Add(delta)
	}
}

// Set sets the value of a gauge metric using the default registry.
func Set(name string, value int64) {
	if m := GetMetric(name); m != nil {
		m.Set(value)
	}
}

// Value returns the current value of a metric from the default registry.
func Value(name string) int64 {
	if m := GetMetric(name); m != nil {
		return m.Value()
	}
	return 0
}

// Collect returns a map of all metric names and their current values.
func Collect() map[string]int64 {
	defaultRegistry.mu.RLock()
	defer defaultRegistry.mu.RUnlock()

	data := make(map[string]int64, len(defaultRegistry.metrics))
	for name, metric := range defaultRegistry.metrics {
		data[name] = metric.Value()
	}
	return data
}