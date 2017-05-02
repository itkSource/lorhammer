package tools

import (
	"github.com/Sirupsen/logrus"
	"github.com/prometheus/client_golang/prometheus"
	"time"
)

var _LOG_PROMETHEUS = logrus.WithField("logger", "tools/prometheus")

type Prometheus interface {
	StartTimer() func()
	AddGateway(nb int)
	SubGateway(nb int)
	AddNodes(nb int)
	SubNodes(nb int)
	AddLongRequest(nb int)
}

type prometheusImpl struct {
	udpDuration   prometheus.Histogram
	nbGateways    prometheus.Gauge
	nbNodes       prometheus.Gauge
	nbLongRequest prometheus.Counter
}

func NewPrometheus() Prometheus {
	udpDuration := prometheus.NewHistogram(prometheus.HistogramOpts{
		Name:    "lorhammer_durations",
		Help:    "Lora latency distributions.",
		Buckets: prometheus.LinearBuckets(0, 100, 10), // 10 buckets, each 100msc wide.
	})
	prometheus.MustRegister(udpDuration)
	nbGateways := prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "lorhammer_gateway",
		Help: "Lora simulated gateways.",
	})
	prometheus.MustRegister(nbGateways)
	nbNodes := prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "lorhammer_node",
		Help: "Lora simulated nodes.",
	})
	prometheus.MustRegister(nbNodes)
	nbLongRequest := prometheus.NewCounter(prometheus.CounterOpts{
		Name: "lorhammer_long_request",
		Help: "Lora nb lora request witch take more than 2sc.",
	})
	prometheus.MustRegister(nbLongRequest)
	return &prometheusImpl{
		udpDuration:   udpDuration,
		nbGateways:    nbGateways,
		nbNodes:       nbNodes,
		nbLongRequest: nbLongRequest,
	}
}

func (prom *prometheusImpl) StartTimer() func() {
	start := time.Now()
	return func() {
		t := time.Now().Sub(start).Seconds() * 1000
		_LOG_PROMETHEUS.WithField("time", t).Debug("Time")
		prom.udpDuration.Observe(t)
	}
}

func (prom *prometheusImpl) AddGateway(nb int) {
	prom.nbGateways.Add(float64(nb))
}

func (prom *prometheusImpl) SubGateway(nb int) {
	prom.nbGateways.Sub(float64(nb))
}

func (prom *prometheusImpl) AddNodes(nb int) {
	prom.nbNodes.Add(float64(nb))
}

func (prom *prometheusImpl) SubNodes(nb int) {
	prom.nbNodes.Sub(float64(nb))
}

func (prom *prometheusImpl) AddLongRequest(nb int) {
	prom.nbLongRequest.Add(float64(nb))
}
