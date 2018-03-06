package testtype

import (
	"lorhammer/src/model"
	"lorhammer/src/orchestrator/command"
	"lorhammer/src/tools"

	"github.com/sirupsen/logrus"
)

const typeOneShot Type = "oneShot"

var logOneShot = logrus.WithField("logger", "orchestrator/testtype/oneShot")

func startOneShot(_ Test, init []model.Init, mqtt tools.Mqtt) {
	if err := command.LaunchScenario(mqtt, init); err != nil {
		logOneShot.WithError(err).Error("Can't launch scenario")
	}
}
