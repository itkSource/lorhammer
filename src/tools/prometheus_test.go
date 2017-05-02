package tools

import (
	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"
	"testing"
	"time"
)

type fakePrometheusHistogram struct {
	t     *testing.T
	start time.Time
}

func (_ fakePrometheusHistogram) Desc() *prometheus.Desc           { return nil }
func (_ fakePrometheusHistogram) Write(*dto.Metric) error          { return nil }
func (_ fakePrometheusHistogram) Describe(chan<- *prometheus.Desc) {}
func (_ fakePrometheusHistogram) Collect(chan<- prometheus.Metric) {}
func (f fakePrometheusHistogram) Observe(observedTime float64) {
	if time.Now().Sub(f.start).Seconds()*1000 < observedTime {
		f.t.Fatal("Observed time must be now minus start time multiply by 1000 (to be milliseconds)")
	}
}

func TestPrometheusImpl_StartTimer(t *testing.T) {
	p := NewPrometheus()
	p.(*prometheusImpl).udpDuration = fakePrometheusHistogram{start: time.Now(), t: t}
	f := p.StartTimer()
	time.Sleep(100 * time.Millisecond)
	f()
}
