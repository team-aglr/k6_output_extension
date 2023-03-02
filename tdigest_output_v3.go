package log

import (
	"fmt"
	"io"
	"time"

	"go.k6.io/k6/metrics"
	"go.k6.io/k6/output"
)

// init is called by the Go runtime at application startup
func init() {
	output.RegisterExtension("logger", New)
}

// Logger = Output struct - for now just try to write aggregated k6 metric samples to stdout
type Logger struct {
	out             io.Writer
	periodicFlusher *output.PeriodicFlusher
	counts          map[string]int64 // this is illegal, but the xk6 build won't error out/just get panic error when try to run a test with this output extension
	// client *statsd.Client
}

// New returns a new instance of Logger
func New(params output.Params) (output.Output, error) {
	return NewLogger(params), nil
}

func NewLogger(params output.Params) *Logger {
	var l Logger
	l.out = params.StdOut
	l.counts = make(map[string]int64)
	return &l
}

// Description returns a short human-readable description of the output
func (*Logger) Description() string {
	return "logger"
}

// Start tries to open a connection to specified statsd service and starts the goroutine for
// metric flushing.
func (l *Logger) Start() error {
	//l.out = output.Params.StdOut
	pf, err := output.NewPeriodicFlusher(time.Duration(1*1e10), l.flushMetrics) // time.Duration is in nanoseconds; 10 billion ns = 10s
	if err != nil {
		return err
	}
	l.periodicFlusher = pf // according to docs at link it seems like the periodic flusher should be started when you call NewPeriodicFlusher
	// https://pkg.go.dev/go.k6.io/k6/output#PeriodicFlusher <=> this is alos how the k6 statsd output initializes the periodic flusher

	return nil
}

// Stop flushes any remaining metrics and stops the goroutine.
func (l *Logger) Stop() error {
	l.periodicFlusher.Stop()
	return nil
}

// AddMetricSamples receives metric samples from the k6 Engine as they're emitted
// <=> need to have this defined on the output struct
//
//	<=> if try to build w/o AddMetricSamples get error that prevents build from happening
func (l *Logger) AddMetricSamples(samples []metrics.SampleContainer) {
	for _, sample := range samples {
		all := sample.GetSamples()

		for _, entry := range all {
			metricName := entry.Metric.Name
			if entry.Metric.Type == metrics.Counter {
				_, exists := l.counts[metricName]
				if !exists {
					l.counts[metricName] = int64(entry.Value)
				} else {
					l.counts[metricName] += int64(entry.Value)
				}
			}
		}
	}
}

func (l *Logger) flushMetrics() {
	// i don't think this is going to work fully, b/c of how AddMetricSamples is working
	//   AddMetricSamples will catch metrics as they're emitted and put them into counts
	//    so I don't think this approach is going to cleanly catch metrics for 10 second intervals
	for key, value := range l.counts {
		// would want to send stuff elsewhere but for now just write to stdout aggregated counts
		//fmt.Fprintf(l.out, key, value)
		fmt.Println(key, value)
	}
	l.counts = make(map[string]int64) // reassign to empty to reset
	// <=> this is also potentially problematic bc it could potentially wipe metrics that
	// get added after the for loop to the counts map
}
