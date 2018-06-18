package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
)

//Prometheus export prometheus metrics
type Prometheus interface {
	AddMQTTMessageOK()
	AddMQTTMessageFailed()
}

type prometheusImpl struct {
	mqttMessagesOK     prometheus.Counter
	mqttMessagesFailed prometheus.Counter
}

//NewPrometheus return a Prometheus instance
func NewPrometheus() Prometheus {
	mqttMessagesOK := prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "orchestrator_mqtt_ok",
		Help: "Count MQTT messages OK.",
	})
	prometheus.MustRegister(mqttMessagesOK)
	mqttMessagesFailed := prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "orchestrator_mqtt_failed",
		Help: "Count MQTT messages failed.",
	})
	prometheus.MustRegister(mqttMessagesFailed)
	return &prometheusImpl{
		mqttMessagesOK:     mqttMessagesOK,
		mqttMessagesFailed: mqttMessagesFailed,
	}
}

func (prom *prometheusImpl) AddMQTTMessageOK() {
	prom.mqttMessagesOK.Add(1)
}

func (prom *prometheusImpl) AddMQTTMessageFailed() {
	prom.mqttMessagesFailed.Add(1)
}
