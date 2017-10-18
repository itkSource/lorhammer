package command

import (
	"lorhammer/src/model"
	"lorhammer/src/tools"

	"github.com/sirupsen/logrus"
)

var loggerOut = logrus.WithField("logger", "orchestrator/command/out")

//LaunchScenario emit a model.INIT command for lorhammers over mqtt
func LaunchScenario(mqttClient tools.Mqtt, init model.Init) error {
	if err := mqttClient.PublishSubCmd(tools.MqttInitTopic, model.INIT, init); err != nil {
		return err
	}
	loggerIn.WithField("init", init).WithField("toTopic", tools.MqttInitTopic).Info("Send init message")
	return nil
}

//StopScenario emit a model.STOP command for lorhammers over mqtt
func StopScenario(mqttClient tools.Mqtt) {
	if err := mqttClient.PublishCmd(tools.MqttInitTopic, model.STOP); err != nil {
		loggerOut.WithError(err).Error("Couldn't publish stop command")
	} else {
		loggerOut.WithField("toTopic", tools.MqttInitTopic).Info("Send stop message")
	}
}

//ShutdownLorhammers emit a model.SHUTDOWN command for lorhammers over mqtt
func ShutdownLorhammers(mqttClient tools.Mqtt) {
	if err := mqttClient.PublishCmd(tools.MqttInitTopic, model.SHUTDOWN); err != nil {
		loggerOut.WithError(err).Error("Couldn't publish shutdown command")
	} else {
		loggerOut.WithField("toTopic", tools.MqttInitTopic).Info("Send shutdown message")
	}
}
