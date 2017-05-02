package command

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/Sirupsen/logrus"
	"lorhammer/src/model"
	"lorhammer/src/tools"
)

var LOG = logrus.WithField("logger", "orchestrator/command/in")

func ApplyCmd(command model.CMD, mqtt tools.Mqtt, provision func(model.Register) error) error {
	switch command.CmdName {

	case model.REGISTER:
		var sensorsToRegister model.Register
		if err := json.Unmarshal(command.Payload, &sensorsToRegister); err != nil {
			return err
		}
		LOG.WithField("nbGateways", len(sensorsToRegister.Gateways)).Info("Received registration command")

		if err := provision(sensorsToRegister); err != nil {
			return err
		}

		startMessage := model.Start{
			ScenarioUUID: sensorsToRegister.ScenarioUUID,
		}

		if err := mqtt.PublishSubCmd(sensorsToRegister.CallBackTopic, model.START, startMessage); err != nil {
			return err
		}
		LOG.Info("Start message sent")

	default:
		return errors.New(fmt.Sprintf("Unknown command %s", command.CmdName))
	}
	return nil
}
