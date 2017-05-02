package testType

import (
	"github.com/Sirupsen/logrus"
	"lorhammer/src/model"
	"lorhammer/src/tools"
)

const TypeNone Type = "none"

var LOG_NONE = logrus.WithField("logger", "orchestrator/testType/none")

func startNone(_ Test, _ model.Init, _ tools.Mqtt) {
	LOG_NONE.WithField("type", "none").Warn("Nothing to test")
}
