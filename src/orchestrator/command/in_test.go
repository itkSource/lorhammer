package command

import (
	"encoding/json"
	"errors"
	"lorhammer/src/model"
	"testing"

	"github.com/sirupsen/logrus"
)

type mqttTest struct {
	valid          bool
	cmdName        model.CommandName
	payload        string
	provision      bool
	publishError   bool
	publishCmdName model.CommandName
	publishPayload string
	provisionError error
}

var tests = []mqttTest{
	{
		valid:          true,
		cmdName:        model.REGISTER,
		payload:        `{"scenarioid":"1","gateways":[],"callBackTopic":"cbt"}`,
		provision:      true,
		publishError:   false,
		publishCmdName: model.START,
		publishPayload: `{"scenarioid":"1"}`,
		provisionError: nil,
	}, {
		valid:          false,
		cmdName:        model.REGISTER,
		payload:        ``,
		provisionError: nil,
	}, {
		valid:          false,
		cmdName:        model.REGISTER,
		payload:        `{}`,
		publishError:   true,
		provisionError: nil,
	}, {
		valid:          false,
		cmdName:        model.CommandName(""),
		provisionError: nil,
	}, {
		valid:          false,
		cmdName:        model.NEWLORHAMMER,
		payload:        `{`,
		provisionError: nil,
	}, {
		valid:          false,
		cmdName:        model.REGISTER,
		payload:        `{"scenarioid":"1","gateways":[],"callBackTopic":"cbt"}`,
		provisionError: errors.New("fake error provisioning"),
	},
}

type fakeMqtt struct {
	t    *testing.T
	test mqttTest
}

func (m *fakeMqtt) GetAddress() string                                          { return "" }
func (m *fakeMqtt) Connect() error                                              { return nil }
func (m *fakeMqtt) Disconnect()                                                 {}
func (m *fakeMqtt) Handle(topics []string, handle func(messgae []byte)) error   { return nil }
func (m *fakeMqtt) HandleCmd(topics []string, handle func(cmd model.CMD)) error { return nil }
func (m *fakeMqtt) PublishCmd(topic string, cmdName model.CommandName) error {
	if m.test.publishError {
		return errors.New("Error")
	}
	if m.test.publishCmdName != cmdName {
		m.t.Fatalf("%s command should send %s command instead of %s command", m.test.cmdName, m.test.publishCmdName, cmdName)
	}
	if m.test.payload != "" {
		m.t.Fatalf("%s command must have payload", cmdName)
	}
	return nil
}

func (m *fakeMqtt) PublishSubCmd(topic string, cmdName model.CommandName, subCmd interface{}) error {
	if m.test.publishError {
		return errors.New("Error")
	}
	if m.test.publishCmdName != cmdName {
		m.t.Fatalf("%s command should send %s command instead of %s command", m.test.cmdName, m.test.publishCmdName, cmdName)
	}
	message, err := json.Marshal(subCmd)
	if err != nil {
		return err

	}
	if string(message) != m.test.publishPayload {
		m.t.Fatalf("bad payload %s instead of %s", string(message), m.test.publishPayload)
	}
	return nil
}

type fakeWriter struct{}

func (f fakeWriter) Write(p []byte) (n int, err error) { return len(p), nil }

func TestRegister(t *testing.T) {
	logrus.SetOutput(fakeWriter{}) // shut up logrus 🙊

	for i, test := range tests {
		cmd := model.CMD{
			CmdName: test.cmdName,
			Payload: json.RawMessage([]byte(test.payload)),
		}
		mqtt := &fakeMqtt{
			t:    t,
			test: test,
		}
		hasCallProvision := false
		hasCallNewLorhammer := false
		err := ApplyCmd(cmd, mqtt, func(register model.Register) error {
			hasCallProvision = true
			return test.provisionError
		}, func(instance model.NewLorhammer) error {
			hasCallNewLorhammer = true
			return nil
		})

		if test.valid && err != nil {
			t.Fatalf("a valid test should not return err for test %d", i)
		} else if !test.valid && err == nil {
			t.Fatalf("an invalid test should return err for test %d", i)
		}

		if test.valid && test.provision && !hasCallProvision {
			t.Fatalf("a valid test should call provision() method for test %d", i)
		}

		if hasCallNewLorhammer {
			t.Fatal("a valid test don't call new lorhammer")
		}
	}
}

func TestNewLorhammer(t *testing.T) {
	logrus.SetOutput(fakeWriter{}) // shut up logrus 🙊

	cmd := model.CMD{
		CmdName: model.NEWLORHAMMER,
		Payload: json.RawMessage([]byte(`{"CallbackTopic":"topic1"}`)),
	}
	mqtt := &fakeMqtt{
		t: t,
		test: mqttTest{
			valid:          true,
			publishCmdName: model.LORHAMMERADDED,
		},
	}
	hasCallProvision := false
	hasCallNewLorhammer := false
	err := ApplyCmd(cmd, mqtt, func(register model.Register) error {
		hasCallProvision = true
		return nil
	}, func(instance model.NewLorhammer) error {
		hasCallNewLorhammer = true
		return nil
	})

	if err != nil {
		t.Fatal("a valid test should not return err", err)
	}

	if hasCallProvision {
		t.Fatal("new lorhammer must not call provision")
	}

	if !hasCallNewLorhammer {
		t.Fatal("a valid test should call new lorhammer")
	}
}
