package metrics

import (
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/sirupsen/logrus"
)

var logPrometheus = logrus.WithField("logger", "tools/prometheus")

//Prometheus export prometheus metrics
type Prometheus interface {
	StartPushAckTimer() func()
	StartPullRespTimer() func()
	AddGateway(nb int)
	SubGateway(nb int)
	AddNodes(nb int)
	SubNodes(nb int)
	AddPushAckLongRequest(nb int)
	AddPullRespLongRequest(nb int)
}

type prometheusImpl struct {
	udpPushAckDuration    prometheus.Histogram
	udpPullRespDuration   prometheus.Histogram
	nbGateways            prometheus.Gauge
	nbNodes               prometheus.Gauge
	nbPushAckLongRequest  prometheus.Counter
	nbPullRespLongRequest prometheus.Counter
}

//NewPrometheus return a Prometheus instance
func NewPrometheus() Prometheus {
	udpPushAckDuration := prometheus.NewHistogram(prometheus.HistogramOpts{
		Name:    "lorhammer_pushack_durations",
		Help:    "Lora push ack latency distributions.",
		Buckets: prometheus.LinearBuckets(0, 100, 10), // 10 buckets, each 100msc wide.
	})
	prometheus.MustRegister(udpPushAckDuration)
	udpPullRespDuration := prometheus.NewHistogram(prometheus.HistogramOpts{
		Name:    "lorhammer_pullresp_durations",
		Help:    "Lora pull resp latency distributions.",
		Buckets: prometheus.LinearBuckets(0, 100, 10), // 10 buckets, each 100msc wide.
	})
	prometheus.MustRegister(udpPullRespDuration)
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
	nbPushAckLongRequest := prometheus.NewCounter(prometheus.CounterOpts{
		Name: "lorhammer_pushack_long_request",
		Help: "Lora nb lora push ack request witch take more than 2sc.",
	})
	prometheus.MustRegister(nbPushAckLongRequest)
	nbPullRespLongRequest := prometheus.NewCounter(prometheus.CounterOpts{
		Name: "lorhammer_pullresp_long_request",
		Help: "Lora nb lora pull resp request witch take more than 2sc.",
	})
	prometheus.MustRegister(nbPullRespLongRequest)
	return &prometheusImpl{
		udpPullRespDuration:   udpPullRespDuration,
		udpPushAckDuration:    udpPushAckDuration,
		nbGateways:            nbGateways,
		nbNodes:               nbNodes,
		nbPushAckLongRequest:  nbPushAckLongRequest,
		nbPullRespLongRequest: nbPullRespLongRequest,
	}
}

func (prom *prometheusImpl) StartPushAckTimer() func() {
	start := time.Now()
	return func() {
		t := time.Now().Sub(start).Seconds() * 1000
		logPrometheus.WithField("time", t).WithField("msyType", "Push Ack").Debug("Time")
		prom.udpPushAckDuration.Observe(t)
	}
}

func (prom *prometheusImpl) StartPullRespTimer() func() {
	start := time.Now()
	return func() {
		t := time.Now().Sub(start).Seconds() * 1000
		logPrometheus.WithField("time", t).WithField("msyType", "Pull Resp").Debug("Time")
		prom.udpPullRespDuration.Observe(t)
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

func (prom *prometheusImpl) AddPushAckLongRequest(nb int) {
	prom.nbPushAckLongRequest.Add(float64(nb))
}

func (prom *prometheusImpl) AddPullRespLongRequest(nb int) {
	prom.nbPullRespLongRequest.Add(float64(nb))
}
