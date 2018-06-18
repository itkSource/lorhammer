package scenario

import (
	"lorhammer/src/lorhammer/lora"
	"lorhammer/src/model"
	"lorhammer/src/tools"
	"time"

	"context"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"lorhammer/src/lorhammer/metrics"
)

var logger = logrus.WithField("logger", "lorhammer/scenario/scenario")

//Scenario struc define scenari with metadata
type Scenario struct {
	UUID                 string
	Gateways             []*lora.LorhammerGateway
	poison               chan bool
	ScenarioSleepTime    [2]time.Duration
	GatewaySleepTime     [2]time.Duration
	NbScenarioReplayLaps int
	RxpkDate             uint64
	WithJoin             bool
	MessageFcnt          uint32
	AppsKey              string
	Nwskey               string
	Payloads             []model.Payload
}

//NewScenario provide new Scenario with param defined in model.Init
func NewScenario(init model.Init) (*Scenario, error) {
	gateways := make([]*lora.LorhammerGateway, init.NbGateway)
	for i := 0; i < len(gateways); i++ {
		if _, err := time.ParseDuration(init.ReceiveTimeoutTime); err != nil {
			return nil, err
		}
		gateways[i] = lora.NewGateway(int(tools.Random64(int64(init.NbNode[0]), int64(init.NbNode[1]))), init)
	}
	scenarioSleepTimeMin, err := time.ParseDuration(init.ScenarioSleepTime[0])
	if err != nil {
		return nil, err
	}
	scenarioSleepTimeMax, err := time.ParseDuration(init.ScenarioSleepTime[1])
	if err != nil {
		return nil, err
	}
	gatewaySleepTimeMin, err := time.ParseDuration(init.GatewaySleepTime[0])
	if err != nil {
		return nil, err
	}
	gatewaySleepTimeMax, err := time.ParseDuration(init.GatewaySleepTime[1])
	if err != nil {
		return nil, err
	}
	return &Scenario{
		UUID:                 uuid.New().String(),
		Gateways:             gateways,
		poison:               make(chan bool),
		ScenarioSleepTime:    [2]time.Duration{scenarioSleepTimeMin, scenarioSleepTimeMax},
		GatewaySleepTime:     [2]time.Duration{gatewaySleepTimeMin, gatewaySleepTimeMax},
		NbScenarioReplayLaps: init.NbScenarioReplayLaps,
		WithJoin:             init.WithJoin,
		MessageFcnt:          0,
		Nwskey:               init.Nwskey,
		AppsKey:              init.AppsKey,
		Payloads:             init.Payloads,
	}, nil
}

//Cron start scenario in go routine and start gateway every `scenario.ScenarioSleepTime`
func (p *Scenario) Cron(prometheus metrics.Prometheus) context.Context {
	prometheus.AddGateway(p.nbGateways())
	prometheus.AddNodes(p.nbNodes())
	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		p.start(prometheus, cancel)
		quit := false
		for {
			select {
			case <-time.After(tools.RandomDuration(p.ScenarioSleepTime[0], p.ScenarioSleepTime[1])):
				p.start(prometheus, cancel)
				p.MessageFcnt++
			case <-p.poison:
				quit = true
			}
			if quit {
				break
			}
		}
	}()
	return ctx
}

//Stop stop scenario launched in Start
func (p *Scenario) Stop(prometheus metrics.Prometheus) {
	p.poison <- true
	defer close(p.poison)
	prometheus.SubGateway(p.nbGateways())
	prometheus.SubNodes(p.nbNodes())
}

//Join launch all gateways join method
func (p *Scenario) Join(prometheus metrics.Prometheus) {
	logger.WithField("nbGateways", len(p.Gateways)).Info("All gateways are joining the application server")

	for _, gateway := range p.Gateways {
		gateway.Join(prometheus, p.WithJoin)
	}
}

func (p *Scenario) start(prometheus metrics.Prometheus, cancelFunction context.CancelFunc) {
	// all gateways have ended, the scenario is stopped properly by calling the stop method
	if doAllGatewaysHaveEnded(p) {
		cancelFunction()
		logger.Info("AllGatewaysHaveEnded")
		return
	}

	logger.WithField("nbGateways", len(p.Gateways)).Info("Gateways started")

	for _, gateway := range p.Gateways {
		time.Sleep(tools.RandomDuration(p.GatewaySleepTime[0], p.GatewaySleepTime[1]))
		go gateway.Start(prometheus, p.MessageFcnt)
		p.MessageFcnt++
	}
}

func doAllGatewaysHaveEnded(p *Scenario) bool {
	//infinite case when PayloadsReplayMaxRound is set to 0 or inferior
	if p.NbScenarioReplayLaps <= 0 {
		return false
	}
	for _, gateway := range p.Gateways {
		// if only one node has not reached the expected number of rounds, break the loop and return false
		if !gateway.AllLapsCompleted {
			return false
		}
	}
	return true
}

func (p *Scenario) nbGateways() int {
	return len(p.Gateways)
}

func (p *Scenario) nbNodes() int {
	nbNode := 0
	for _, gateway := range p.Gateways {
		nbNode += len(gateway.Nodes)
	}
	return nbNode
}
