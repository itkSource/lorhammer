package command

import (
	"github.com/Sirupsen/logrus"
	"lorhammer/src/model"
	"lorhammer/src/tools"
)

func LaunchScenario(mqttClient tools.Mqtt, init model.Init) error {
	if err := mqttClient.PublishSubCmd(tools.MQTT_INIT_TOPIC, model.INIT, init); err != nil {
		return err
	}
	LOG.WithField("init", init).WithField("toTopic", tools.MQTT_INIT_TOPIC).Info("Send init message")
	return nil
}

func StopScenario(mqttClient tools.Mqtt) {
	if err := mqttClient.PublishCmd(tools.MQTT_INIT_TOPIC, model.STOP); err != nil {
		logrus.WithFields(logrus.Fields{
			"ref": "orchestrator/orchestrator:stopScenario()",
			"err": err,
		}).Error("Couldn't publish stop command")
	} else {
		logrus.WithFields(logrus.Fields{
			"ref":     "orchestrator/orchestrator:stopScenario()",
			"toTopic": tools.MQTT_INIT_TOPIC,
		}).Info("Send stop message")
	}
}

func ShutdownLorhammers(mqttClient tools.Mqtt) {
	if err := mqttClient.PublishCmd(tools.MQTT_INIT_TOPIC, model.SHUTDOWN); err != nil {
		logrus.WithFields(logrus.Fields{
			"ref": "orchestrator/orchestrator:shutdownLorhammers()",
			"err": err,
		}).Error("Couldn't publish shutdown command")
	} else {
		logrus.WithFields(logrus.Fields{
			"ref":     "orchestrator/orchestrator:shutdownLorhammers()",
			"toTopic": tools.MQTT_INIT_TOPIC,
		}).Info("Send shutdown message")
	}
}
