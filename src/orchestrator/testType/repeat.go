package testType

import (
	"github.com/Sirupsen/logrus"
	"lorhammer/src/model"
	"lorhammer/src/orchestrator/command"
	"lorhammer/src/tools"
	"time"
)

const TypeRepeat Type = "repeat"

var LOG_REPEAT = logrus.WithField("logger", "orchestrator/testType/repeat")

func startRepeat(test Test, init model.Init, mqtt tools.Mqtt) {
	for range time.Tick(test.repeatTime) {
		if err := command.LaunchScenario(mqtt, init); err != nil {
			LOG_REPEAT.WithError(err).Error("Can't launch scenario")
		}
	}
}
