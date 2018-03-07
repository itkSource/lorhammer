package tools

import (
	"encoding/json"
	"lorhammer/src/model"

	mqttLib "github.com/eclipse/paho.mqtt.golang"
	"github.com/sirupsen/logrus"
)

//Channels mqtt to use
const (
	MqttLorhammerTopic    = "/lorhammer"
	MqttOrchestratorTopic = "/lorhammer/orchestrator"
)

var logMqtt = logrus.WithField("logger", "tools/mqtt")

//Mqtt is responsible of communication with the mqtt server
type Mqtt interface {
	Connect() error
	Disconnect()
	GetAddress() string
	Handle(topics []string, handle func(message []byte)) error
	HandleCmd(topics []string, handle func(cmd model.CMD)) error
	PublishCmd(topic string, cmdName model.CommandName) error
	PublishSubCmd(topic string, cmdName model.CommandName, subCmd interface{}) error
}

type mqttImpl struct {
	url    string
	client mqttLib.Client
}

//NewMqtt return a Mqtt based on mqttAddr (protocol://ip:port) and set clientID with hostname
func NewMqtt(hostname string, mqttAddr string) (Mqtt, error) {
	clientID := hostname + "_" + string(RandomBytes(8))
	return NewMqttBasic(mqttAddr, clientID)
}

//NewMqttBasic return a Mqtt client
func NewMqttBasic(url string, clientID string) (Mqtt, error) {
	// uncomment next line to see all mqtt logs (very verbose)
	// mqttLib.DEBUG = log.New(os.Stderr, "", log.LstdFlags)

	connOpts := mqttLib.NewClientOptions().AddBroker(url).SetClientID(clientID).SetOnConnectHandler(func(client mqttLib.Client) {
		logMqtt.WithField("mqtt", url).WithField("ClientID", clientID).Info("Connected to Mqtt broker")
	}).SetConnectionLostHandler(func(client mqttLib.Client, reason error) {
		logMqtt.WithError(reason).Warn("Connection mqtt lost")
	})

	client := mqttLib.NewClient(connOpts)

	return &mqttImpl{
		url:    url,
		client: client,
	}, nil
}

func (mqtt *mqttImpl) GetAddress() string {
	return mqtt.url
}

func (mqtt *mqttImpl) Disconnect() {
	mqtt.client.Disconnect(0)
}

func (mqtt *mqttImpl) Connect() error {
	if token := mqtt.client.Connect(); token.Wait() && token.Error() != nil {
		return token.Error()
	}
	return nil
}

func (mqtt *mqttImpl) Handle(topics []string, handle func(message []byte)) error {
	filters := make(map[string]byte)
	for _, topic := range topics {
		filters[topic] = byte(0)
	}
	if token := mqtt.client.SubscribeMultiple(filters, func(client mqttLib.Client, message mqttLib.Message) {
		handle(message.Payload())
	}); token.Wait() && token.Error() != nil {
		return token.Error()
	}
	return nil
}

func (mqtt *mqttImpl) HandleCmd(topics []string, handle func(cmd model.CMD)) error {
	return mqtt.Handle(topics, func(message []byte) {
		var command model.CMD
		if err := json.Unmarshal(message, &command); err != nil {
			logMqtt.WithField("msg", string(message)).WithError(err).Warn("Skeep message because can't unMarshalling incoming message")
		} else {
			handle(command)
		}
	})
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
	logMqtt.WithField("topic", topic).WithField("cmd", cmd.CmdName).Info("Send mqtt cmd")
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
