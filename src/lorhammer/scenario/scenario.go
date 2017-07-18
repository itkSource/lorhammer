package scenario

import (
	"github.com/Sirupsen/logrus"
	"github.com/google/uuid"
	"lorhammer/src/lorhammer/lora"
	"lorhammer/src/model"
	"lorhammer/src/tools"
	"time"
)

type Scenario struct {
	Uuid              string
	Gateways          []*model.Gateway
	poison            chan bool
	ScenarioSleepTime [2]time.Duration
	GatewaySleepTime  [2]time.Duration
	RxpkDate          uint64
	WithJoin          bool
	MessageFcnt       uint32
	AppsKey           string
	Nwskey            string
	Payloads          []string
}

func NewScenario(init model.Init) (*Scenario, error) {
	gateways := make([]*model.Gateway, init.NbGateway)
	for i := 0; i < len(gateways); i++ {
		if parsedTime, err := time.ParseDuration(init.ReceiveTimeoutTime); err != nil {
			return nil, err
		} else {
			gateways[i] = lora.NewGateway(tools.Random(init.NbNode[0], init.NbNode[1]), init.NsAddress, init.AppsKey, init.Nwskey, init.Payloads, init.RxpkDate, parsedTime)
		}
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
		Uuid:              uuid.New().String(),
		Gateways:          gateways,
		poison:            make(chan bool),
		ScenarioSleepTime: [2]time.Duration{scenarioSleepTimeMin, scenarioSleepTimeMax},
		GatewaySleepTime:  [2]time.Duration{gatewaySleepTimeMin, gatewaySleepTimeMax},
		WithJoin:          init.WithJoin,
		MessageFcnt:       0,
		Nwskey:            init.Nwskey,
		AppsKey:           init.AppsKey,
		Payloads:          init.Payloads,
	}, nil
}

func (p *Scenario) Cron(prometheus tools.Prometheus) {
	prometheus.AddGateway(p.nbGateways())
	prometheus.AddNodes(p.nbNodes())
	go func() {
		p.start(prometheus)
		quit := false
		for {
			select {
			case <-time.After(tools.RandomDuration(p.ScenarioSleepTime[0], p.ScenarioSleepTime[1])):
				p.start(prometheus)
				p.MessageFcnt++
			case <-p.poison:
				quit = true
			}
			if quit {

				break
			}
		}
	}()
}

func (p *Scenario) Stop(prometheus tools.Prometheus) {
	p.poison <- true
	defer close(p.poison)
	prometheus.SubGateway(p.nbGateways())
	prometheus.SubNodes(p.nbNodes())
}

func (p *Scenario) Join(prometheus tools.Prometheus) {
	logrus.WithFields(logrus.Fields{
		"ref":        "scenario/scenario:scenario()",
		"nbGateways": len(p.Gateways),
	}).Info("All gateways are joining the application server")

	for i := 0; i < len(p.Gateways); i++ {
		lora.Join(p.Gateways[i], prometheus, p.WithJoin)
	}
}

func (p *Scenario) start(prometheus tools.Prometheus) {
	logrus.WithFields(logrus.Fields{
		"ref":        "scenario/scenario:scenario()",
		"nbGateways": len(p.Gateways),
	}).Info("Gateways started")
	for i := 0; i < len(p.Gateways); i++ {
		time.Sleep(tools.RandomDuration(p.GatewaySleepTime[0], p.GatewaySleepTime[1]))
		go lora.Start(p.Gateways[i], prometheus, p.MessageFcnt)
		p.MessageFcnt++
	}
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
