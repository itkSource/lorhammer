package testSuite

import (
	"fmt"
	"lorhammer/src/model"
	"lorhammer/src/orchestrator/provisioning"
	"testing"
)

type testLaunch struct {
	description      string
	testValid        bool
	test             string
	rampTime         string
	repeatTime       string
	stopAll          string
	shutdownAll      string
	init             string
	provisioning     string
	needProvisioning bool
	prometheusCheck  string
	deploy           string
}

var testsLaunch = []testLaunch{
	{
		testValid:        true,
		description:      "Must run",
		test:             `{"type": "oneShot", "rampTime": "0", "repeatTime": "0"}`,
		stopAll:          "0",
		shutdownAll:      "0",
		init:             `{"nsAddress": "127.0.0.1:1700","nbGateway": 1,"nbNodePerGateway": [1, 1],"sleepTime": [100, 500]}`,
		provisioning:     `{"type": "none"}`,
		needProvisioning: true,
		prometheusCheck:  `[]`,
		deploy:           `{"type": "none"}`,
	},
	{
		testValid:        false,
		description:      "Fake deploy",
		test:             `{"type": "oneShot", "rampTime": "0", "repeatTime": "0"}`,
		stopAll:          "0",
		shutdownAll:      "0",
		init:             `{"nsAddress": "127.0.0.1:1700","nbGateway": 1,"nbNodePerGateway": [1, 1],"sleepTime": [100, 500]}`,
		provisioning:     `{"type": "none"}`,
		needProvisioning: true,
		prometheusCheck:  `[]`,
		deploy:           `{"type": "fake"}`,
	},
	{
		testValid:        false,
		description:      "Fake testType",
		test:             `{"type": "fake", "rampTime": "0", "repeatTime": "0"}`,
		stopAll:          "0",
		shutdownAll:      "0",
		init:             `{"nsAddress": "127.0.0.1:1700","nbGateway": 1,"nbNodePerGateway": [1, 1],"sleepTime": [100, 500]}`,
		provisioning:     `{"type": "none"}`,
		needProvisioning: true,
		prometheusCheck:  `[]`,
		deploy:           `{"type": "none"}`,
	},
	{
		testValid:        true,
		description:      "Must run stopTime > 0",
		test:             `{"type": "oneShot", "rampTime": "0", "repeatTime": "0"}`,
		stopAll:          "1ms",
		shutdownAll:      "0",
		init:             `{"nsAddress": "127.0.0.1:1700","nbGateway": 1,"nbNodePerGateway": [1, 1],"sleepTime": [100, 500]}`,
		provisioning:     `{"type": "none"}`,
		needProvisioning: true,
		prometheusCheck:  `[]`,
		deploy:           `{"type": "none"}`,
	},
	{
		testValid:        false,
		description:      "DeProvision without provision and run stopTime > 0",
		test:             `{"type": "oneShot", "rampTime": "0", "repeatTime": "0"}`,
		stopAll:          "1ms",
		shutdownAll:      "0",
		init:             `{"nsAddress": "127.0.0.1:1700","nbGateway": 1,"nbNodePerGateway": [1, 1],"sleepTime": [100, 500]}`,
		provisioning:     `{"type": "none"}`,
		needProvisioning: false,
		prometheusCheck:  `[]`,
		deploy:           `{"type": "none"}`,
	},
	{
		testValid:        true,
		description:      "Prometheus check",
		test:             `{"type": "oneShot", "rampTime": "0", "repeatTime": "0"}`,
		stopAll:          "0",
		shutdownAll:      "0",
		init:             `{"nsAddress": "127.0.0.1:1700","nbGateway": 1,"nbNodePerGateway": [1, 1],"sleepTime": [100, 500]}`,
		provisioning:     `{"type": "none"}`,
		needProvisioning: false,
		prometheusCheck:  `[{"query": "sum(lorhammer_long_request) + sum(lorhammer_durations_count)", "resultMin": 1, "resultMax": 1, "description": "nb messages"}]`,
		deploy:           `{"type": "none"}`,
	},
	{
		testValid:        true,
		description:      "Prometheus check not in domain but not throw error because its written in file",
		test:             `{"type": "oneShot", "rampTime": "0", "repeatTime": "0"}`,
		stopAll:          "0",
		shutdownAll:      "0",
		init:             `{"nsAddress": "127.0.0.1:1700","nbGateway": 1,"nbNodePerGateway": [1, 1],"sleepTime": [100, 500]}`,
		provisioning:     `{"type": "none"}`,
		needProvisioning: false,
		prometheusCheck:  `[{"query": "sum(lorhammer_long_request) + sum(lorhammer_durations_count)", "resultMin": 0, "resultMax": 0, "description": "nb messages"}]`,
		deploy:           `{"type": "none"}`,
	},
}

var templateLaunch = `[{"test": %s,"rampTime": "%s","repeatTime": "%s","stopAllLorhammerTime": "%s","shutdownAllLorhammerTime": "%s","init": %s,"provisioning": %s,"prometheusCheck": %s, "deploy": %s}]`

func TestLaunchTest(t *testing.T) {
	for _, test := range testsLaunch {
		var ct = test
		data := []byte(fmt.Sprintf(templateLaunch, ct.test, ct.rampTime, ct.repeatTime, ct.stopAll, ct.shutdownAll, ct.init, ct.provisioning, ct.prometheusCheck, ct.deploy))
		tests, err := FromFile(data)
		if err != nil {
			t.Fatalf(`valid scenario should not return err %s for : "%s"`, err, ct.description)
		}
		if len(tests) != 1 {
			t.Fatalf(`1 valid scenario should return 1 valid testSuite for : "%s"`, ct.description)
		}
		if test.needProvisioning {
			provisioning.Provision(tests[0].Uuid, tests[0].Provisioning, model.Register{})
		}
		report, err := LaunchTest(nil, &fakeMqtt{}, &tests[0], fakePrometheus{}, nil)
		if ct.testValid && err != nil {
			t.Fatal("valid test should not throw err")
		} else if ct.testValid && report == nil {
			t.Fatal("valid test should return report")
		} else if !ct.testValid && err == nil {
			t.Fatal("not valid test should throw err")
		} else if !ct.testValid && report != nil {
			t.Fatal("not valid test should not return report")
		}

	}
}

type fakeMqtt struct {
}

func (m *fakeMqtt) Connect() error                                              { return nil }
func (m *fakeMqtt) HandleCmd(topics []string, handle func(cmd model.CMD)) error { return nil }
func (m *fakeMqtt) PublishCmd(topic string, cmdName model.CommandName) error    { return nil }
func (m *fakeMqtt) PublishSubCmd(topic string, cmdName model.CommandName, subCmd interface{}) error {
	return nil
}

type fakePrometheus struct {
}

func (_ fakePrometheus) ExecQuery(query string) (float64, error) { return 1, nil }
