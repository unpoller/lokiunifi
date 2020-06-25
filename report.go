package lokiunifi

import (
	"strings"
	"time"

	"github.com/unifi-poller/poller"
	"github.com/unifi-poller/unifi"
)

// LogStream contains a stream of logs (like a log file).
// This app uses one stream per log entry because each log may have different labels.
type LogStream struct {
	Labels  map[string]string `json:"stream"` // "the file name"
	Entries [][]string        `json:"values"` // "the log lines"
}

// Logs is the main logs-holding structure.
type Logs struct {
	Streams []LogStream `json:"streams"` // "multiple files"
}

// Report is the temporary data generated and sent to Loki at every interval.
type Report struct {
	Logs
	Counts map[string]int
	Oldest time.Time
	poller.Logger
}

// NewReport makes a new report.
func (l *Loki) NewReport() *Report {
	return &Report{
		Logger: l.Collect,
		Oldest: l.last,
	}
}

func (r *Report) LogOutput(start time.Time) {
	r.Logf("Events sent to Loki. Event: %d, IDS: %d, Alarm: %d, Anomaly: %d, Dur: %v",
		r.Counts[typeEvent], r.Counts[typeIDS], r.Counts[typeAlarm], r.Counts[typeAnomaly],
		time.Since(start).Round(time.Millisecond))
}

// ProcessEventLogs loops the event Logs, matches the interface
// type, calls the appropriate method for the data, and compiles the report.
// This runs once per interval, if there was no collection error.
func (r *Report) ProcessEventLogs(events *poller.Events) {
	for _, e := range events.Logs {
		switch event := e.(type) {
		case *unifi.IDS:
			r.IDS(event)
		case *unifi.Event:
			r.Event(event)
		case *unifi.Alarm:
			r.Alarm(event)
		case *unifi.Anomaly:
			r.Anomaly(event)
		default: // unlikely.
			r.LogErrorf("unknown event type: %T", e)
		}
	}
}

// CleanLabels removes any tag that is empty.
func CleanLabels(labels map[string]string) map[string]string {
	for i := range labels {
		if strings.TrimSpace(labels[i]) == "" {
			delete(labels, i)
		}
	}

	return labels
}
