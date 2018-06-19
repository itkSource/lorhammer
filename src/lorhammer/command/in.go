package command

import (
	"encoding/json"
	"lorhammer/src/lorhammer/scenario"
	"lorhammer/src/model"
	"lorhammer/src/tools"
	"os"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
	"lorhammer/src/lorhammer/metrics"
)

var logger = logrus.WithField("logger", "lorhammer/command/in")
var scenarios = sync.Map{}

//Start send model.NEWLORHAMMER command every second until orchestrator has respond model.LORHAMMERADDED command
func Start(mqtt tools.Mqtt, hostname string, maxWaitOrchestratorTime time.Duration) chan bool {
	lorhammerAddedChan := make(chan bool)
	go func() {
		defer close(lorhammerAddedChan)
		start := time.Now()
		for {
			select {
			case <-time.After(1 * time.Second):
				mqtt.PublishSubCmd(tools.MqttOrchestratorTopic, model.NEWLORHAMMER, model.NewLorhammer{CallbackTopic: tools.MqttLorhammerTopic + "/" + hostname})
				if time.Now().Sub(start) > maxWaitOrchestratorTime {
					logger.Error("Max time to wait ended, I will not registered on an orchestrator...")
					return
				}
			case <-lorhammerAddedChan:
				return
			}
		}
	}()
	return lorhammerAddedChan
}

//ApplyCmd take a model.CMD and execute it
func ApplyCmd(command model.CMD, mqtt tools.Mqtt, hostname string, lorhammerAddedChan chan bool, prometheus metrics.Prometheus) {
	var err error
	switch command.CmdName {
	case model.LORHAMMERADDED:
		lorhammerAddedChan <- true
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
		logger.WithField("cmd", command.CmdName).Error("Unknown command")
	}
	if err != nil {
		logger.WithError(err).WithField("cmd", command.CmdName).WithField("body", string(command.Payload)).Error("Apply cmd fail")
	}
}

func applyInitCmd(command model.CMD, mqtt tools.Mqtt, hostname string) error {
	var initMessage model.Init
	if err := json.Unmarshal(command.Payload, &initMessage); err != nil {
		return err
	}

	logger.WithFields(logrus.Fields{
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
		gateways[i] = gateway.ConvertToGateway()
	}

	registerCmd := model.Register{
		CallBackTopic: tools.MqttLorhammerTopic + "/" + hostname,
		Gateways:      gateways,
		ScenarioUUID:  sc.UUID,
	}

	err = mqtt.PublishSubCmd(tools.MqttOrchestratorTopic, model.REGISTER, registerCmd)
	if err != nil {
		return err
	}

	scenarios.Store(sc.UUID, sc)
	logger.WithField("toTopic", tools.MqttOrchestratorTopic).Info("Sent registration command to orchestrator")
	return nil
}

func applyStartCmd(command model.CMD, prometheus metrics.Prometheus) {
	var startMessage model.Start
	if err := json.Unmarshal(command.Payload, &startMessage); err != nil {
		logger.WithError(err).Error("Can't unmarshal init command")
	} else {
		if sc, isPresent := scenarios.Load(startMessage.ScenarioUUID); isPresent {
			logger.Warn("Start scenario")
			sc.(*scenario.Scenario).Join(prometheus)
			ctx := sc.(*scenario.Scenario).Cron(prometheus)
			go func() {
				logger.Debug("Blocking routine waiting for cancel function")
				<-ctx.Done()
				logger.Debug("Releasing blocking routine after cancel function call")
				stopScenario(sc.(*scenario.Scenario), prometheus)
			}()
		} else {
			logger.WithField("uuid", startMessage.ScenarioUUID).Error("Can't find scenario")
		}
	}
}

func stopScenario(scenario *scenario.Scenario, prometheus metrics.Prometheus) {
	logger.WithField("scenario", scenario.UUID).Warn("Stopping scenario")
	scenario.Stop(prometheus)
	scenarios.Delete(scenario.UUID)
}

func applyStopCmd(prometheus metrics.Prometheus) {
	logger.Warn("Stop scenarios")
	scenarios.Range(func(key interface{}, value interface{}) bool {
		stopScenario(value.(*scenario.Scenario), prometheus)
		return true
	})
}

func applyShutdownCmd() {
	logger.Warn("Shutdown")
	p, err := os.FindProcess(os.Getpid())
	if err != nil {
		logger.WithError(err).Error("Can't get current process")
	}
	p.Signal(os.Interrupt) // will DeRegister consul
}
