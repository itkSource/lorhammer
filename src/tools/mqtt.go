package tools

import (
	"encoding/json"
	"github.com/Sirupsen/logrus"
	mqttLib "github.com/eclipse/paho.mqtt.golang"
	"lorhammer/src/model"
)

const (
	MQTT_START_TOPIC        = "/lorhammer"
	MQTT_INIT_TOPIC         = "/lorhammer/all"
	MQTT_ORCHESTRATOR_TOPIC = "/lorhammer/orchestrator"
)

var _LOG_MQTT = logrus.WithField("logger", "tools/mqtt")

type Mqtt interface {
	Connect() error
	HandleCmd(topics []string, handle func(cmd model.CMD)) error
	PublishCmd(topic string, cmdName model.CommandName) error
	PublishSubCmd(topic string, cmdName model.CommandName, subCmd interface{}) error
}

type mqttImpl struct {
	client mqttLib.Client
}

func NewMqtt(hostname string, consulClient Consul) (Mqtt, error) {
	url, err := consulClient.ServiceFirst("mqtt", "tcp://")

	if err != nil {
		return nil, err
	}

	clientId := MQTT_START_TOPIC + "/" + hostname

	// uncomment next line to see all mqtt logs (very verbose)
	// mqttLib.DEBUG = log.New(os.Stderr, "", log.LstdFlags)

	connOpts := mqttLib.NewClientOptions().AddBroker(url).SetClientID(clientId).SetOnConnectHandler(func(client mqttLib.Client) {
		_LOG_MQTT.WithFields(logrus.Fields{
			"mqtt":     url,
			"ClientID": clientId,
		}).Info("Connected to Mqtt broker")
	}).SetConnectionLostHandler(func(client mqttLib.Client, reason error) {
		_LOG_MQTT.WithFields(logrus.Fields{
			"err": reason.Error(),
		}).Warn("Connection mqtt lost")
	})

	client := mqttLib.NewClient(connOpts)

	return &mqttImpl{
		client: client,
	}, nil
}

func (mqtt *mqttImpl) Connect() error {
	if token := mqtt.client.Connect(); token.Wait() && token.Error() != nil {
		return token.Error()
	}
	return nil
}

func (mqtt *mqttImpl) HandleCmd(topics []string, handle func(cmd model.CMD)) error {
	filters := make(map[string]byte)
	for _, topic := range topics {
		filters[topic] = byte(0)
	}
	if token := mqtt.client.SubscribeMultiple(filters, func(client mqttLib.Client, message mqttLib.Message) {
		var command model.CMD
		if err := json.Unmarshal(message.Payload(), &command); err != nil {
			_LOG_MQTT.WithFields(logrus.Fields{
				"err": err,
				"msg": string(message.Payload()),
			}).Warn("Skeep message because can't unMarshalling incoming message")
		} else {
			handle(command)
		}
	}); token.Wait() && token.Error() != nil {
		return token.Error()
	}
	return nil
}

func (mqtt *mqttImpl) publish(topic string, message []byte) error {
	mqtt.client.Publish(topic, 0, false, message)
	return nil
}

func (mqtt *mqttImpl) publishFullCmd(topic string, cmd model.CMD) error {
	message, err := json.Marshal(cmd)
	if err != nil {
		return err
	}
	_LOG_MQTT.WithFields(logrus.Fields{
		"topic": topic,
		"cmd":   cmd.CmdName,
	}).Info("Send mqtt cmd")
	return mqtt.publish(topic, message)
}

func (mqtt *mqttImpl) PublishCmd(topic string, cmdName model.CommandName) error {
	cmd := model.CMD{
		CmdName: cmdName,
	}
	return mqtt.publishFullCmd(topic, cmd)
}

func (mqtt *mqttImpl) PublishSubCmd(topic string, cmdName model.CommandName, subCmd interface{}) error {
	message, err := json.Marshal(subCmd)
	if err != nil {
		return err

	}
	cmd := model.CMD{
		CmdName: cmdName,
		Payload: message,
	}
	return mqtt.publishFullCmd(topic, cmd)
}
