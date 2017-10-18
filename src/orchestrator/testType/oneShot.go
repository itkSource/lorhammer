package testType

import (
	"github.com/sirupsen/logrus"
	"lorhammer/src/model"
	"lorhammer/src/orchestrator/command"
	"lorhammer/src/tools"
)

const TypeOneShot Type = "oneShot"

var LOG_ONE_SHOT = logrus.WithField("logger", "orchestrator/testType/oneShot")

func startOneShot(_ Test, init model.Init, mqtt tools.Mqtt) {
	if err := command.LaunchScenario(mqtt, init); err != nil {
		LOG_ONE_SHOT.WithError(err).Error("Can't launch scenario")
	}
}
