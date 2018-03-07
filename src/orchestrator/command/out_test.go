package command

import (
	"encoding/json"
	"lorhammer/src/model"
	"testing"
	"time"
)

func TestNbLorhammer(t *testing.T) {
	nbLorhammer := NbLorhammer()
	if nbLorhammer != 0 {
		t.Fatal("At start no lorhammer")
	}
	NewLorhammer(model.NewLorhammer{CallbackTopic: "topic1"})
	NewLorhammer(model.NewLorhammer{CallbackTopic: "topic1"})
	NewLorhammer(model.NewLorhammer{CallbackTopic: "topic1"})
	nbLorhammer = NbLorhammer()
	if nbLorhammer != 3 {
		t.Fatal("Should have 3 lorhammers")
	}

	lorhammers = make([]model.NewLorhammer, 0) // reinitialize lorhammers
	nbToLaunch := 100
	go func() {
		for i := 0; i < nbToLaunch; i++ {
			NewLorhammer(model.NewLorhammer{CallbackTopic: "topic1"})
			time.Sleep(1 * time.Millisecond)
		}
	}()
	for {
		nbLorhammer = NbLorhammer()
		if nbLorhammer > nbToLaunch {
			t.Fatal("Too much lorhammer")
		} else if nbLorhammer == nbToLaunch {
			break
		}
		time.Sleep(1 * time.Microsecond)
	}
}

func TestLaunchLaunchScenario(t *testing.T) {
	init := model.Init{}
	serialized, err := json.Marshal(init)
	if err != nil {
		t.Fatal("init model should be marshalled")
	}
	mqtt := &fakeMqtt{
		t: t,
		test: mqttTest{
			publishCmdName: model.INIT,
			publishPayload: string(serialized),
		},
	}
	lorhammers = make([]model.NewLorhammer, 0) // reinitialize lorhammers
	NewLorhammer(model.NewLorhammer{CallbackTopic: "topic1"})
	err = LaunchScenario(mqtt, []model.Init{init, init, init})
	if err != nil {
		t.Fatal("A valid model.init should not return err")
	}
}

func TestLaunchLaunchScenarioError(t *testing.T) {
	init := model.Init{}
	mqtt := &fakeMqtt{
		t: t,
		test: mqttTest{
			publishError: true,
		},
	}
	NewLorhammer(model.NewLorhammer{CallbackTopic: "topic1"})
	err := LaunchScenario(mqtt, []model.Init{init})
	if err == nil {
		t.Fatal("If mqtt return err out should return err")
	}
}

func TestStopScenario(t *testing.T) {
	mqtt := &fakeMqtt{
		t: t,
		test: mqttTest{
			publishCmdName: model.STOP,
		},
	}
	StopScenario(mqtt)
}

func TestStopScenarioError(t *testing.T) {
	mqtt := &fakeMqtt{
		t: t,
		test: mqttTest{
			publishError: true,
		},
	}
	StopScenario(mqtt)
}

func TestShutdownLorhammers(t *testing.T) {
	mqtt := &fakeMqtt{
		t: t,
		test: mqttTest{
			publishCmdName: model.SHUTDOWN,
		},
	}
	ShutdownLorhammers(mqtt)
}

func TestShutdownLorhammersError(t *testing.T) {
	mqtt := &fakeMqtt{
		t: t,
		test: mqttTest{
			publishError: true,
		},
	}
	ShutdownLorhammers(mqtt)
}
