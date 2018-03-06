package command

import (
	"encoding/json"
	"lorhammer/src/model"
	"testing"
)

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
	NewLorhammer(model.NewLorhammer{CallbackTopic: "topic1"})
	err = LaunchScenario(mqtt, []model.Init{init})
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
