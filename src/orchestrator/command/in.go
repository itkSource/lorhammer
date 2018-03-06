package command

import (
	"encoding/json"
	"fmt"
	"lorhammer/src/model"
	"lorhammer/src/tools"

	"github.com/sirupsen/logrus"
)

var loggerIn = logrus.WithField("logger", "orchestrator/command/in")

//ApplyCmd launch a model.CMD received from a lorhammer
func ApplyCmd(command model.CMD, mqtt tools.Mqtt, provision func(model.Register) error, newLorhammer func(model.NewLorhammer) error) error {
	switch command.CmdName {
	case model.NEWLORHAMMER:
		{
			var instance model.NewLorhammer
			if err := json.Unmarshal(command.Payload, &instance); err != nil {
				return err
			}
			loggerIn.WithField("topic", instance.CallbackTopic).Info("New Lorhammer")
			newLorhammer(instance)
			return mqtt.PublishCmd(instance.CallbackTopic, model.LORHAMMERADDED)
		}
	case model.REGISTER:
		var sensorsToRegister model.Register
		if err := json.Unmarshal(command.Payload, &sensorsToRegister); err != nil {
			return err
		}
		loggerIn.WithField("nbGateways", len(sensorsToRegister.Gateways)).Info("Received registration command")

		if err := provision(sensorsToRegister); err != nil {
			return err
		}
		loggerIn.WithField("nbGateways", len(sensorsToRegister.Gateways)).Info("Provisioning done")

		startMessage := model.Start{
			ScenarioUUID: sensorsToRegister.ScenarioUUID,
		}

		if err := mqtt.PublishSubCmd(sensorsToRegister.CallBackTopic, model.START, startMessage); err != nil {
			return err
		}
		loggerIn.Info("Start message sent")

	default:
		return fmt.Errorf("Unknown command %s", command.CmdName)
	}
	return nil
}
