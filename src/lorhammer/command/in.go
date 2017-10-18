package command

import (
	"encoding/json"
	"github.com/Sirupsen/logrus"
	"lorhammer/src/lorhammer/scenario"
	"lorhammer/src/model"
	"lorhammer/src/tools"
	"os"
	"sync"
)

var LOG = logrus.WithField("logger", "lorhammer/command/in")
var scenarios = sync.Map{}

func ApplyCmd(command model.CMD, mqtt tools.Mqtt, hostname string, prometheus tools.Prometheus) {
	var err error
	switch command.CmdName {
	case model.INIT:
		err = applyInitCmd(command, mqtt, hostname)
	case model.START:
		applyStartCmd(command, prometheus)
	case model.STOP:
		applyStopCmd(prometheus)
	case model.SHUTDOWN:
		applyStopCmd(prometheus)
		applyShutdownCmd()
	default:
		LOG.WithField("cmd", command.CmdName).Error("Unknown command")
	}
	if err != nil {
		LOG.WithError(err).WithField("cmd", command).Error("Apply cmd fail")
	}
}

func applyInitCmd(command model.CMD, mqtt tools.Mqtt, hostname string) error {
	var initMessage model.Init
	if err := json.Unmarshal(command.Payload, &initMessage); err != nil {
		return err
	}

	LOG.WithFields(logrus.Fields{
		"nbGateway": initMessage.NbGateway,
		"nbMinNode": initMessage.NbNode[0],
		"nbMaxNode": initMessage.NbNode[1],
		"nsAddress": initMessage.NsAddress,
	}).Info("Received an init system command")

	sc, err := scenario.NewScenario(initMessage)
	if err != nil {
		return err
	}

	gateways := make([]model.Gateway, len(sc.Gateways))
	for i, gateway := range sc.Gateways {
		gateways[i] = *gateway
	}

	registerCmd := model.Register{
		CallBackTopic: tools.MQTT_START_TOPIC + "/" + hostname,
		Gateways:      gateways,
		ScenarioUUID:  sc.Uuid,
	}

	err = mqtt.PublishSubCmd(tools.MQTT_ORCHESTRATOR_TOPIC, model.REGISTER, registerCmd)
	if err != nil {
		return err
	} else {
		scenarios.Store(sc.Uuid, sc)
		LOG.WithField("toTopic", tools.MQTT_ORCHESTRATOR_TOPIC).Info("Sent registration command to orchestrator")
	}

	return nil
}

func applyStartCmd(command model.CMD, prometheus tools.Prometheus) {
	var startMessage model.Start
	if err := json.Unmarshal(command.Payload, &startMessage); err != nil {
		LOG.WithError(err).Error("Can't unmarshal init command")
	} else {
		if sc, isPresent := scenarios.Load(startMessage.ScenarioUUID); isPresent {
			LOG.Warn("Start scenario")
			sc.(*scenario.Scenario).Join(prometheus)
			ctx := sc.(*scenario.Scenario).Cron(prometheus)
			go func() {
				LOG.Debug("Blocking routine waiting for cancel function")
				<-ctx.Done()
				LOG.Debug("Releasing blocking routine after cancel function call")
				stopScenario(sc.(*scenario.Scenario), prometheus)
			}()
		} else {
			LOG.WithField("uuid", startMessage.ScenarioUUID).Error("Can't find scenario")
		}
	}
}

func stopScenario(scenario *scenario.Scenario, prometheus tools.Prometheus) {
	LOG.WithField("scenario", scenario.Uuid).Warn("Stopping scenario")
	scenario.Stop(prometheus)
	scenarios.Delete(scenario.Uuid)
}

func applyStopCmd(prometheus tools.Prometheus) {
	LOG.Warn("Stop scenarios")
	scenarios.Range(func(key interface{}, value interface{}) bool {
		stopScenario(value.(*scenario.Scenario), prometheus)
		return true
	})
}

func applyShutdownCmd() {
	LOG.Warn("Shutdown")
	p, err := os.FindProcess(os.Getpid())
	if err != nil {
		LOG.WithError(err).Error("Can't get current process")
	}
	p.Signal(os.Interrupt) // will DeRegister consul
}
