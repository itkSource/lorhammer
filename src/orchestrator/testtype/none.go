package testtype

import (
	"lorhammer/src/model"
	"lorhammer/src/tools"

	"github.com/sirupsen/logrus"
)

const typeNone Type = "none"

var logNone = logrus.WithField("logger", "orchestrator/testType/none")

func startNone(_ Test, _ []model.Init, _ tools.Mqtt) {
	logNone.WithField("type", "none").Warn("Nothing to test")
}
