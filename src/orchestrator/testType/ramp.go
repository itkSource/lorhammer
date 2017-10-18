package testType

import (
	"github.com/sirupsen/logrus"
	"lorhammer/src/model"
	"lorhammer/src/orchestrator/command"
	"lorhammer/src/tools"
	"math"
	"time"
)

const TypeRamp Type = "ramp"

var LOG_RAMP = logrus.WithField("logger", "orchestrator/testType/ramp")

type ramp struct {
	nbGateway        int
	timeMinute       float64
	currentRest      float64
	gatewaysLaunched int
	intervalRampTime time.Duration
}

func startRamp(test Test, init model.Init, mqttClient tools.Mqtt) {
	ramp := &ramp{
		nbGateway:        init.NbGateway,
		timeMinute:       test.rampTime.Minutes(),
		currentRest:      float64(0),
		gatewaysLaunched: 0,
		intervalRampTime: time.Duration(1 * time.Minute),
	}
	ramp.start(init, mqttClient)
}

func (r *ramp) start(init model.Init, mqttClient tools.Mqtt) {
	needRelaunch := r.launch(mqttClient, init)
	if needRelaunch {
		for range time.Tick(r.intervalRampTime) {
			if needRelaunch := r.launch(mqttClient, init); needRelaunch == false {
				break
			}
		}
	}
}

func (r *ramp) launch(mqttClient tools.Mqtt, init model.Init) bool {
	nbGatewayToLaunch := r.nextTick()
	LOG_RAMP.WithField("nbGateway", nbGatewayToLaunch).Info("Launch ramp")
	if nbGatewayToLaunch > 0 {
		init := model.Init{
			NbGateway:         nbGatewayToLaunch,
			NbNode:            init.NbNode,
			NsAddress:         init.NsAddress,
			ScenarioSleepTime: init.ScenarioSleepTime,
			GatewaySleepTime:  init.GatewaySleepTime,
		}
		if err := command.LaunchScenario(mqttClient, init); err != nil {
			LOG_RAMP.WithError(err).Error("Can't launch scenario")
		}
	}
	return r.willLaunch()
}

func (r *ramp) nextTick() int {
	if r.timeMinute == 0 {
		r.gatewaysLaunched += r.nbGateway
		return r.nbGateway
	}
	res := float64(r.nbGateway) / r.timeMinute
	nbGatewayToLaunch := math.Trunc(res)
	rest := res - nbGatewayToLaunch
	r.currentRest = r.currentRest + rest
	if r.currentRest >= float64(1) {
		nbFromRest := math.Trunc(r.currentRest)
		nbGatewayToLaunch = nbGatewayToLaunch + nbFromRest
		r.currentRest = r.currentRest - nbFromRest
	}
	r.gatewaysLaunched += int(nbGatewayToLaunch)
	return int(nbGatewayToLaunch)
}

func (r *ramp) willLaunch() bool {
	return r.gatewaysLaunched < r.nbGateway
}
