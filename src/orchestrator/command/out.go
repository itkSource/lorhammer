package command

import (
	"lorhammer/src/model"
	"lorhammer/src/tools"

	"github.com/sirupsen/logrus"
)

func LaunchScenario(mqttClient tools.Mqtt, init model.Init) error {
	if err := mqttClient.PublishSubCmd(tools.MqttInitTopic, model.INIT, init); err != nil {
		return err
	}
	LOG.WithField("init", init).WithField("toTopic", tools.MqttInitTopic).Info("Send init message")
	return nil
}

func StopScenario(mqttClient tools.Mqtt) {
	if err := mqttClient.PublishCmd(tools.MqttInitTopic, model.STOP); err != nil {
		logrus.WithFields(logrus.Fields{
			"ref": "orchestrator/orchestrator:stopScenario()",
			"err": err,
		}).Error("Couldn't publish stop command")
	} else {
		logrus.WithFields(logrus.Fields{
			"ref":     "orchestrator/orchestrator:stopScenario()",
			"toTopic": tools.MqttInitTopic,
		}).Info("Send stop message")
	}
}

func ShutdownLorhammers(mqttClient tools.Mqtt) {
	if err := mqttClient.PublishCmd(tools.MqttInitTopic, model.SHUTDOWN); err != nil {
		logrus.WithFields(logrus.Fields{
			"ref": "orchestrator/orchestrator:shutdownLorhammers()",
			"err": err,
		}).Error("Couldn't publish shutdown command")
	} else {
		logrus.WithFields(logrus.Fields{
			"ref":     "orchestrator/orchestrator:shutdownLorhammers()",
			"toTopic": tools.MqttInitTopic,
		}).Info("Send shutdown message")
	}
}
