package log

import (
	"fmt"
	"io"
	"time"
	"strconv"

	"go.k6.io/k6/metrics"
	"go.k6.io/k6/output"
)

// init is called by the Go runtime at application startup.
func init() {
	output.RegisterExtension("logger", New)
}

// Logger writes k6 metric samples to stdout.
type Logger struct {
	out    io.Writer
	buffer []metrics.SampleContainer
}

// New returns a new instance of Logger.
func New(params output.Params) (output.Output, error) {
	emptyBuffer := []metrics.SampleContainer // this is illegal <=> xk6 won't build because of this
	return &Logger{params.StdOut, emptyBuffer}, nil
}

// Description returns a short human-readable description of the output.
func (*Logger) Description() string {
	return "logger"
}

// AddMetricSamples receives metric samples from the k6 Engine as they're emitted.
func (l *Logger) AddMetricSamples(samples []metrics.SampleContainer) {
	l.buffer = append(l.buffer, samples...)

	for range time.Tick(time.Second * 10) {
		l.flushMetrics(l.buffer)
		l.buffer = []metrics.SampleContainer // will this erase any metrics that were emitted into the buffer after l.flushMetrics was called?
	}
}

// Start initializes any state needed for the output, establishes network
// connections, etc.
func (l *Logger) Start() error {
	return nil
}

func (l *Logger) flushMetrics(samples []metrics.SampleContainer) {
	counts := make(map[string]int64)


	for _, sample := range samples {
		sc := sample.GetSamples()
		for _, entry := range sc {
			metricName := entry.Metric.Name
			if entry.Metric.Type == metrics.Counter {
				value, exists := counts[metricName]
				if !exists {
					counts[metricName] = value
				} else {
					counts[metricName] += value
				}
			}
		}
	}
	fmt.Println(counts["http_reqs"])
	fmt.Fprintf(l.out, strconv.Itoa(int(counts["http_reqs"])))
}

// Stop finalizes any tasks in progress, closes network connections, etc.
func (*Logger) Stop() error {
	return nil
}
