package log

import (
	"fmt"
	"io"
	"time"

	"github.com/caio/go-tdigest"
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
	counts          map[string]int64
	tdigests        map[string]*tdigest.TDigest
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
	l.tdigests = make(map[string]*tdigest.TDigest)
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
			metricType := entry.Metric.Type

			if metricType == metrics.Counter {
				l.trackCount(metricName, entry.Value)
			} else if metricType == metrics.Trend {
				l.trackTdigest(metricName, entry.Value)
			}
		}
	}
}

func (l *Logger) trackCount(name string, value float64) {
	_, exists := l.counts[name]
	if !exists {
		l.counts[name] = int64(value)
	} else {
		l.counts[name] += int64(value)
	}
}

func (l *Logger) trackTdigest(name string, value float64) {
	_, exists := l.tdigests[name]
	if !exists {
		td, _ := tdigest.New()
		td.Add(value)
		l.tdigests[name] = td
	} else {
		l.tdigests[name].Add(value)
	}
}

func (l *Logger) flushMetrics() {
	// i'm not sure if this approach is going to work fully, b/c I'm not sure if
	//   AddMetricSamples will catch metrics as they're emitted and put them into counts
	//   as flushMetrics is operating <=> not sure if this approach will always cleanly catch
	//   metrics for rounded/pure 10 second intervals
	for metricName, count := range l.counts {
		// would want to send stuff elsewhere but for now just write to stdout aggregated counts
		//fmt.Fprintf(l.out, key, value)
		fmt.Println(metricName, count)
	}
	for metricName, tdigest := range l.tdigests {
		fmt.Println("centroids for metric: ", metricName)
		tdigest.ForEachCentroid(logCentroidDetails)
	}
	l.counts = make(map[string]int64) // reassign to empty to reset
	l.tdigests = make(map[string]*tdigest.TDigest)
	// <=> this reassignment is potentially problematic - not sure if it would wipe metrics that
	// get added to l.counts() after the for loop to the counts map runs
}

func logCentroidDetails(mean float64, count uint64) bool {
	fmt.Println("mean:", mean, "count:", count)
	return true
}
