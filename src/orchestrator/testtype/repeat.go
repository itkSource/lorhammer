package testtype

import (
	"lorhammer/src/model"
	"lorhammer/src/orchestrator/command"
	"lorhammer/src/tools"
	"time"

	"github.com/sirupsen/logrus"
)

const typeRepeat Type = "repeat"

var logRepeat = logrus.WithField("logger", "orchestrator/testType/repeat")

func startRepeat(test Test, init []model.Init, mqtt tools.Mqtt) {
	for range time.Tick(test.repeatTime) {
		if err := command.LaunchScenario(mqtt, init); err != nil {
			logRepeat.WithError(err).Error("Can't launch scenario")
		}
	}
}
