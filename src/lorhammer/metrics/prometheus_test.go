package metrics

import (
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"
)

type fakePrometheusHistogram struct {
	t     *testing.T
	start time.Time
}

func (fakePrometheusHistogram) Desc() *prometheus.Desc           { return nil }
func (fakePrometheusHistogram) Write(*dto.Metric) error          { return nil }
func (fakePrometheusHistogram) Describe(chan<- *prometheus.Desc) {}
func (fakePrometheusHistogram) Collect(chan<- prometheus.Metric) {}
func (f fakePrometheusHistogram) Observe(observedTime float64) {
	if time.Now().Sub(f.start).Seconds()*1000 < observedTime {
		f.t.Fatal("Observed time must be now minus start time multiply by 1000 (to be milliseconds)")
	}
}

func TestPrometheusImpl_StartTimer(t *testing.T) {
	p := NewPrometheus()
	p.(*prometheusImpl).udpPushAckDuration = fakePrometheusHistogram{start: time.Now(), t: t}
	f := p.StartPushAckTimer()
	time.Sleep(100 * time.Millisecond)
	f()
}
