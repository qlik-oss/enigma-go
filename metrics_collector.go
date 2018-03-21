package enigma

import (
	"context"
	"fmt"
	"sync"
	"time"
)

type (
	metricsCollectorID struct{}

	// MetricsCollector is used to extract performance metrics for the last invocation in the context
	MetricsCollector struct {
		sync.Mutex
		metrics *InvocationMetrics
	}

	// InvocationMetrics contains performance information about the last invocation in the context
	InvocationMetrics struct {
		InvocationRequestTimestamp  time.Time
		SocketWriteTimestamp        time.Time
		SocketReadTimestamp         time.Time
		InvocationResponseTimestamp time.Time
		RequestMessageSize          int
		ResponseMessageSize         int
	}
)

// ToString returns a human-friendly string representation of elapsed times
func (m *InvocationMetrics) ToString() string {
	return fmt.Sprintf("On air time: %v, Total time: %v", m.SocketReadTimestamp.Sub(m.SocketWriteTimestamp), m.InvocationResponseTimestamp.Sub(m.InvocationRequestTimestamp))
}

// Metrics extracts performance information
func (c *MetricsCollector) Metrics() *InvocationMetrics {
	c.Lock()
	defer c.Unlock()
	return c.metrics
}

func getMetricsCollector(ctx context.Context) *MetricsCollector {
	value := ctx.Value(metricsCollectorID{})
	if value != nil {
		return value.(*MetricsCollector)
	}
	return nil
}

// WithMetricsCollector provides a new context with the a MetricsCollector that records performance metrics for invocations
func WithMetricsCollector(ctx context.Context) (context.Context, *MetricsCollector) {
	metricsCollector := &MetricsCollector{metrics: &InvocationMetrics{}}
	ctxWithMetrics := context.WithValue(ctx, metricsCollectorID{}, metricsCollector)
	return ctxWithMetrics, metricsCollector
}
