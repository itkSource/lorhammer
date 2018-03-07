package command

import (
	"lorhammer/src/model"
	"lorhammer/src/tools"
	"sync"

	"github.com/sirupsen/logrus"
)

var loggerOut = logrus.WithField("logger", "orchestrator/command/out")
var muLorhammers = sync.Mutex{}
var lorhammers = make([]model.NewLorhammer, 0)

//NewLorhammer add a new lorhammer topic to be able to init scenario
func NewLorhammer(newLorhammer model.NewLorhammer) error {
	muLorhammers.Lock()
	defer muLorhammers.Unlock()
	lorhammers = append(lorhammers, newLorhammer)
	return nil
}

//NbLorhammer return the current number of lorhammer listening for init scenario
func NbLorhammer() int {
	muLorhammers.Lock()
	defer muLorhammers.Unlock()
	return len(lorhammers)
}

//LaunchScenario emit a model.INIT command for lorhammers over mqtt
func LaunchScenario(mqttClient tools.Mqtt, inits []model.Init) error {
	muLorhammers.Lock()
	defer muLorhammers.Unlock()
	currentLorhammer := 0
	for _, init := range inits {
		lorhammer := lorhammers[currentLorhammer]
		if err := mqttClient.PublishSubCmd(lorhammer.CallbackTopic, model.INIT, init); err != nil {
			return err
		}
		loggerOut.WithField("init", init.Description).WithField("toTopic", lorhammer.CallbackTopic).Info("Send init message")
		currentLorhammer++
		if currentLorhammer > len(lorhammers)-1 {
			currentLorhammer = 0
		}
	}
	return nil
}

//StopScenario emit a model.STOP command for lorhammers over mqtt
func StopScenario(mqttClient tools.Mqtt) {
	muLorhammers.Lock()
	defer muLorhammers.Unlock()
	if err := mqttClient.PublishCmd(tools.MqttLorhammerTopic, model.STOP); err != nil {
		loggerOut.WithError(err).Error("Couldn't publish stop command")
	} else {
		loggerOut.WithField("toTopic", tools.MqttLorhammerTopic).Info("Send stop message")
	}
}

//ShutdownLorhammers emit a model.SHUTDOWN command for lorhammers over mqtt
func ShutdownLorhammers(mqttClient tools.Mqtt) {
	muLorhammers.Lock()
	defer muLorhammers.Unlock()
	if err := mqttClient.PublishCmd(tools.MqttLorhammerTopic, model.SHUTDOWN); err != nil {
		loggerOut.WithError(err).Error("Couldn't publish shutdown command")
	} else {
		loggerOut.WithField("toTopic", tools.MqttLorhammerTopic).Info("Send shutdown message")
	}
}
