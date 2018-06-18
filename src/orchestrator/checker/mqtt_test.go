package checker

import (
	"encoding/json"
	"errors"
	"lorhammer/src/model"
	"lorhammer/src/tools"
	"testing"
)

type fakePrometheus struct {
	mqttOk   bool
	mqttFail bool
}

func (prom *fakePrometheus) AddMQTTMessageOK() {
	prom.mqttOk = true
}

func (prom *fakePrometheus) AddMQTTMessageFailed() {
	prom.mqttFail = true
}

func TestNewMqtt(t *testing.T) {
	k, err := newMqtt(json.RawMessage([]byte(`{"address": "127.0.0.1:1883"}`)), &fakePrometheus{})
	if err != nil {
		t.Fatalf("Good config should not return err : %s", err.Error())
	}
	if k == nil {
		t.Fatal("Good config should return kafka checker")
	}
}

func TestNewMqttError(t *testing.T) {
	k, err := newMqtt(json.RawMessage([]byte(`{`)), &fakePrometheus{})
	if err == nil {
		t.Fatal("Bad config should return err")
	}
	if k != nil {
		t.Fatal("Bad config should not return kafka checker")
	}
}

func TestStartMqttError(t *testing.T) {
	k, _ := newMqtt(json.RawMessage([]byte(`{"address": "127.0.0.1:1883"}`)), &fakePrometheus{})
	k.(*mqttChecker).clientFactory = func(url string, clientID string) (tools.Mqtt, error) {
		return nil, errors.New("error mqtt")
	}
	err := k.Start()
	if err == nil {
		t.Fatal("Error from mqtt client must be reported")
	}
}

type fakeMqtt struct {
	t           *testing.T
	errorHandle error
	handlers    []func(message []byte)
}

func (m *fakeMqtt) GetAddress() string { return "" }
func (m *fakeMqtt) Connect() error     { return nil }
func (m *fakeMqtt) Disconnect()        {}
func (m *fakeMqtt) Handle(topics []string, handle func(message []byte)) error {
	if m.errorHandle != nil {
		return m.errorHandle
	}
	m.handlers = append(m.handlers, handle)
	return nil
}
func (m *fakeMqtt) HandleCmd(topics []string, handle func(cmd model.CMD)) error { return nil }
func (m *fakeMqtt) PublishCmd(topic string, cmdName model.CommandName) error    { return nil }
func (m *fakeMqtt) PublishSubCmd(topic string, cmdName model.CommandName, subCmd interface{}) error {
	return nil
}

func TestStartMqttHandleError(t *testing.T) {
	k, _ := newMqtt(json.RawMessage([]byte(`{"address": "127.0.0.1:1883"}`)), &fakePrometheus{})
	k.(*mqttChecker).clientFactory = func(url string, clientID string) (tools.Mqtt, error) {
		return &fakeMqtt{t: t, errorHandle: errors.New("fake error")}, nil
	}
	err := k.Start()
	if err == nil {
		t.Fatal("Error from mqtt client when handle must be reported")
	}
}

func TestCheck(t *testing.T) {
	prom := &fakePrometheus{}
	k, err := newMqtt(json.RawMessage([]byte(`{"address": "127.0.0.1:1883", "checks": [{"description": "test", "text": "hi", "remove": [" from test"]}]}`)), prom)
	if err != nil {
		t.Fatal("Good conf should not return err", err)
	}
	fakeMqttInstance := &fakeMqtt{t: t}
	k.(*mqttChecker).clientFactory = func(url string, clientID string) (tools.Mqtt, error) {
		return fakeMqttInstance, nil
	}
	k.Start()
	fakeMqttInstance.handlers[0]([]byte("hi from test"))
	fakeMqttInstance.handlers[0]([]byte("fail"))

	success, errs := k.Check()
	if len(success) != 1 || len(success[0].Details()) != 1 {
		t.Fatal("1 good result should return 1 success")
	}
	if len(errs) != 1 || len(errs[0].Details()) != 2 {
		t.Fatal("1 error result should return 1 error")
	}
	if !prom.mqttFail {
		t.Fatal("MQTT not failed")
	}
	if !prom.mqttOk {
		t.Fatal("MQTT not OK")
	}

}
