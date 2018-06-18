package scenario

import (
	"lorhammer/src/model"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
)

var models = []model.Init{
	{
		NsAddress:          "127.0.0.1:0",
		NbGateway:          1,
		NbNode:             [2]int{1, 1},
		ScenarioSleepTime:  [2]string{"0", "0"},
		GatewaySleepTime:   [2]string{"0", "0"},
		ReceiveTimeoutTime: "1s",
	}, {
		NsAddress:          "127.0.0.1:0",
		NbGateway:          1,
		NbNode:             [2]int{0, 0},
		ScenarioSleepTime:  [2]string{"0", "0"},
		GatewaySleepTime:   [2]string{"0", "0"},
		ReceiveTimeoutTime: "1s",
	}, {
		NsAddress:          "127.0.0.1:0",
		NbGateway:          1,
		NbNode:             [2]int{0, 100},
		ScenarioSleepTime:  [2]string{"0", "0"},
		GatewaySleepTime:   [2]string{"0", "0"},
		ReceiveTimeoutTime: "1s",
	},
}

var wrongReceiveTimeModels = []model.Init{
	{
		NsAddress:          "127.0.0.1:0",
		NbGateway:          1,
		NbNode:             [2]int{1, 1},
		ScenarioSleepTime:  [2]string{"0", "0"},
		GatewaySleepTime:   [2]string{"0", "0"},
		ReceiveTimeoutTime: "",
	},
	{
		NsAddress:          "127.0.0.1:0",
		NbGateway:          1,
		NbNode:             [2]int{1, 1},
		ScenarioSleepTime:  [2]string{"0", "0"},
		GatewaySleepTime:   [2]string{"0", "0"},
		ReceiveTimeoutTime: "toto",
	},
}

type fakePrometheus struct {
	nbGateway chan int
	nbNodes   chan int
}

func (prom *fakePrometheus) StartPushAckTimer() func()     { return nil }
func (prom *fakePrometheus) StartPullRespTimer() func()    { return nil }
func (prom *fakePrometheus) AddGateway(nb int)             { go func() { prom.nbGateway <- nb }() }
func (prom *fakePrometheus) SubGateway(nb int)             { go func() { prom.nbGateway <- nb }() }
func (prom *fakePrometheus) AddNodes(nb int)               { go func() { prom.nbNodes <- nb }() }
func (prom *fakePrometheus) SubNodes(nb int)               { go func() { prom.nbNodes <- nb }() }
func (prom *fakePrometheus) AddPushAckLongRequest(nb int)  {}
func (prom *fakePrometheus) AddPullRespLongRequest(nb int) {}

type fakeWriter struct{}

func (f fakeWriter) Write(p []byte) (n int, err error) { return len(p), nil }

func TestCreation(t *testing.T) {
	logrus.SetOutput(fakeWriter{}) // shut up logrus 🙊

	nbGatewaysChan := make(chan int)
	nbNodesChan := make(chan int)

	defer close(nbGatewaysChan)
	defer close(nbNodesChan)

	var fakeProm = &fakePrometheus{
		nbGateway: nbGatewaysChan,
		nbNodes:   nbNodesChan,
	}

	for _, init := range models {

		sc, err := NewScenario(init)

		if err != nil {
			t.Fatal("Valid init should not return err")
		}

		if sc == nil {
			t.Fatal("A valid scenario must return a no nil scenario")
		}

		if len(sc.Gateways) != init.NbGateway {
			t.Fatal("A scenario must have same nb gateways as init")
		}

		if sc.nbGateways() != init.NbGateway {
			t.Fatal("Facitlity method nbGateways must return same nb gateways as init")
		}

		var nbNode int
		for _, gateway := range sc.Gateways {
			nbNode = nbNode + len(gateway.Nodes)
			if len(gateway.Nodes) < init.NbNode[0] || len(gateway.Nodes) > init.NbNode[1] {
				t.Fatal("Each gateway must have good nbNodes comparing to init")
			}
		}

		if nbNode != sc.nbNodes() {
			t.Fatal("Facitlity method nbNodes must be exactly")
		}

		sc.Cron(fakeProm)

		// sleep because of go routine
		time.Sleep(1 * time.Millisecond)

		promNbGateways := <-nbGatewaysChan
		promNbNodes := <-nbNodesChan

		if promNbGateways != init.NbGateway {
			t.Fatalf("Scenario send %d nb gateway to prometheus instead of %d", promNbGateways, init.NbGateway)
		}

		if promNbNodes < init.NbNode[0] || promNbNodes > init.NbNode[1] {
			t.Fatal("Scenario must call prometheus with good nb nodes")
		}

		sc.Stop(fakeProm)

		promNbGateways -= <-nbGatewaysChan
		promNbNodes -= <-nbNodesChan

		// sleep because of go routine
		time.Sleep(1 * time.Millisecond)

		if promNbGateways != 0 {
			t.Fatalf("Scenario have %d nb gateway instead of 0", promNbGateways)
		}

		if promNbNodes != 0 {
			t.Fatalf("Scenario have %d nb nodes instead of 0", promNbNodes)
		}

		// sleep because of go routine
		time.Sleep(100 * time.Millisecond)
	}
}

func TestWrongReceiveTimeCreation(t *testing.T) {
	logrus.SetOutput(fakeWriter{}) // shut up logrus 🙊

	for _, init := range wrongReceiveTimeModels {

		sc, err := NewScenario(init)

		if err == nil {
			t.Fatal("Error expected on wrong receive timeout time")
		}

		if sc != nil {
			t.Fatal("Nil scenario expected on wrong receive timeout time")
		}
	}
}
