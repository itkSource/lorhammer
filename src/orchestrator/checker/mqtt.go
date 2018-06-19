package checker

import (
	"encoding/json"
	"lorhammer/src/tools"
	"regexp"

	"github.com/sirupsen/logrus"
	"lorhammer/src/orchestrator/metrics"
)

const mqttType = Type("mqtt")

var logMqtt = logrus.WithField("logger", "orchestrator/checker/mqtt")

type mqttChecker struct {
	clientFactory func(url string, clientID string) (tools.Mqtt, error)
	client        tools.Mqtt
	config        mqttConfig
	prometheus    metrics.Prometheus
	success       []Success
	fails         []Error
}

type mqttConfig struct {
	Address string      `json:"address"`
	Channel string      `json:"channel"`
	Checks  []mqttCheck `json:"checks"`
}

type mqttCheck struct {
	Description string   `json:"description"`
	Remove      []string `json:"remove"`
	Text        string   `json:"text"`
}

type mqttSuccess struct {
	check mqttCheck
}

func (m mqttSuccess) Details() map[string]interface{} {
	details := make(map[string]interface{})
	details["success"] = m.check.Description
	return details
}

type mqttError struct {
	reason string
	value  string
}

func (m mqttError) Details() map[string]interface{} {
	details := make(map[string]interface{})
	details["reason"] = m.reason
	details["value"] = m.value
	return details
}

func newMqtt(rawConfig json.RawMessage, prometheus metrics.Prometheus) (Checker, error) {
	conf := mqttConfig{}
	if err := json.Unmarshal(rawConfig, &conf); err != nil {
		return nil, err
	}
	mqtt := &mqttChecker{
		clientFactory: tools.NewMqttBasic,
		config:        conf,
		prometheus:    prometheus,
		success:       make([]Success, 0),
		fails:         make([]Error, 0),
	}
	return mqtt, nil
}

func (mqtt *mqttChecker) Start() error {
	client, err := mqtt.clientFactory(mqtt.config.Address, string(tools.RandomBytes(12)))
	if err != nil {
		return err
	}
	mqtt.client = client
	err = mqtt.client.Connect()
	if err != nil {
		return err
	}
	return client.Handle([]string{mqtt.config.Channel}, mqtt.handle)
}

func (mqtt *mqttChecker) handle(message []byte) {
	atLeastMatch := false
	for _, check := range mqtt.config.Checks {
		/**Here we strip the value to check from all the dynamically produced values (applicationID, devEUI...)
		These values are specified in the remove field through the json scenario in the mqtt check section **/
		var s = string(message)
		for _, dynamicValueToRemove := range check.Remove {
			var re = regexp.MustCompile(dynamicValueToRemove)
			s = re.ReplaceAllLiteralString(s, ``)
		}
		logMqtt.Warn(s)
		if s == check.Text {
			atLeastMatch = true
			mqtt.success = append(mqtt.success, mqttSuccess{check: check})
			logMqtt.WithField("description", check.Description).Info("Success")
			mqtt.prometheus.AddMQTTMessageOK()
			break
		}
	}
	if !atLeastMatch {
		logMqtt.Error("Result mismatch")
		mqtt.fails = append(mqtt.fails, mqttError{reason: "Result mismatch", value: string(message)})
		mqtt.prometheus.AddMQTTMessageFailed()
	}
}

func (mqtt *mqttChecker) Check() ([]Success, []Error) {
	mqtt.client.Disconnect()
	return mqtt.success, mqtt.fails
}
