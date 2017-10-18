package testSuite

import (
	"errors"
	"fmt"
	"lorhammer/src/model"
	"lorhammer/src/orchestrator/provisioning"
	"lorhammer/src/tools"
	"testing"
	"time"
)

type testLaunch struct {
	description      string
	testValid        bool
	test             string
	rampTime         string
	repeatTime       string
	stopAll          string
	beforeCheck      string
	shutdownAll      string
	sleep            string
	init             string
	provisioning     string
	needProvisioning bool
	check            string
	deploy           string
	grafana          tools.GrafanaClient
}

var testsLaunch = []testLaunch{
	{
		testValid:        true,
		description:      "Must run",
		test:             `{"type": "oneShot", "rampTime": "0", "repeatTime": "0"}`,
		stopAll:          "0",
		beforeCheck:      "0",
		shutdownAll:      "0",
		sleep:            "0",
		init:             `{"nsAddress": "127.0.0.1:1700","nbGateway": 1,"nbNodePerGateway": [1, 1],"sleepTime": [100, 500]}`,
		provisioning:     `{"type": "none"}`,
		needProvisioning: true,
		check:            `{"type": "none"}`,
		deploy:           `{"type": "none"}`,
		grafana:          nil,
	},
	{
		testValid:        false,
		description:      "Fake deploy",
		test:             `{"type": "oneShot", "rampTime": "0", "repeatTime": "0"}`,
		stopAll:          "0",
		beforeCheck:      "0",
		shutdownAll:      "0",
		sleep:            "0",
		init:             `{"nsAddress": "127.0.0.1:1700","nbGateway": 1,"nbNodePerGateway": [1, 1],"sleepTime": [100, 500]}`,
		provisioning:     `{"type": "none"}`,
		needProvisioning: true,
		check:            `{"type": "none"}`,
		deploy:           `{"type": "fake"}`,
		grafana:          nil,
	},
	{
		testValid:        false,
		description:      "Fake testType",
		test:             `{"type": "fake", "rampTime": "0", "repeatTime": "0"}`,
		stopAll:          "0",
		beforeCheck:      "0",
		shutdownAll:      "0",
		sleep:            "0",
		init:             `{"nsAddress": "127.0.0.1:1700","nbGateway": 1,"nbNodePerGateway": [1, 1],"sleepTime": [100, 500]}`,
		provisioning:     `{"type": "none"}`,
		needProvisioning: true,
		check:            `{"type": "none"}`,
		deploy:           `{"type": "none"}`,
		grafana:          nil,
	},
	{
		testValid:        true,
		description:      "Must run stopTime > 0",
		test:             `{"type": "oneShot", "rampTime": "0", "repeatTime": "0"}`,
		stopAll:          "1ms",
		beforeCheck:      "0",
		shutdownAll:      "0",
		sleep:            "0",
		init:             `{"nsAddress": "127.0.0.1:1700","nbGateway": 1,"nbNodePerGateway": [1, 1],"sleepTime": [100, 500]}`,
		provisioning:     `{"type": "none"}`,
		needProvisioning: true,
		check:            `{"type": "none"}`,
		deploy:           `{"type": "none"}`,
		grafana:          nil,
	},
	{
		testValid:        false,
		description:      "DeProvision without provision and run stopTime > 0",
		test:             `{"type": "oneShot", "rampTime": "0", "repeatTime": "0"}`,
		stopAll:          "1ms",
		beforeCheck:      "0",
		shutdownAll:      "0",
		sleep:            "0",
		init:             `{"nsAddress": "127.0.0.1:1700","nbGateway": 1,"nbNodePerGateway": [1, 1],"sleepTime": [100, 500]}`,
		provisioning:     `{"type": "none"}`,
		needProvisioning: false,
		check:            `{"type": "none"}`,
		deploy:           `{"type": "none"}`,
		grafana:          nil,
	},
	{
		testValid:        true,
		description:      "Prometheus check",
		test:             `{"type": "oneShot", "rampTime": "0", "repeatTime": "0"}`,
		stopAll:          "0",
		beforeCheck:      "0",
		shutdownAll:      "0",
		sleep:            "0",
		init:             `{"nsAddress": "127.0.0.1:1700","nbGateway": 1,"nbNodePerGateway": [1, 1],"sleepTime": [100, 500]}`,
		provisioning:     `{"type": "none"}`,
		needProvisioning: false,
		check:            `{"type": "prometheus", "config": [{"query": "sum(lorhammer_long_request) + sum(lorhammer_durations_count)", "resultMin": 1, "resultMax": 1, "description": "nb messages"}]}`,
		deploy:           `{"type": "none"}`,
		grafana:          nil,
	},
	{
		testValid:        true,
		description:      "Prometheus check not in domain but not throw error because its written in file",
		test:             `{"type": "oneShot", "rampTime": "0", "repeatTime": "0"}`,
		stopAll:          "0",
		beforeCheck:      "0",
		shutdownAll:      "0",
		sleep:            "0",
		init:             `{"nsAddress": "127.0.0.1:1700","nbGateway": 1,"nbNodePerGateway": [1, 1],"sleepTime": [100, 500]}`,
		provisioning:     `{"type": "none"}`,
		needProvisioning: false,
		check:            `{"type": "prometheus", "config": [{"query": "sum(lorhammer_long_request) + sum(lorhammer_durations_count)", "resultMin": 0, "resultMax": 0, "description": "nb messages"}]}`,
		deploy:           `{"type": "none"}`,
		grafana:          nil,
	},
	{
		testValid:        false,
		description:      "Fake checker should return error",
		test:             `{"type": "oneShot", "rampTime": "0", "repeatTime": "0"}`,
		stopAll:          "0",
		beforeCheck:      "0",
		shutdownAll:      "0",
		sleep:            "0",
		init:             `{"nsAddress": "127.0.0.1:1700","nbGateway": 1,"nbNodePerGateway": [1, 1],"sleepTime": [100, 500]}`,
		provisioning:     `{"type": "none"}`,
		needProvisioning: false,
		check:            `{"type": "fake"}`,
		deploy:           `{"type": "none"}`,
		grafana:          nil,
	},
	{
		testValid:        true,
		description:      "Grafana error should not be reported",
		test:             `{"type": "oneShot", "rampTime": "0", "repeatTime": "0"}`,
		stopAll:          "0",
		beforeCheck:      "0",
		shutdownAll:      "1ms",
		sleep:            "0",
		init:             `{"nsAddress": "127.0.0.1:1700","nbGateway": 1,"nbNodePerGateway": [1, 1],"sleepTime": [100, 500]}`,
		provisioning:     `{"type": "none"}`,
		needProvisioning: true,
		check:            `{"type": "none"}`,
		deploy:           `{"type": "none"}`,
		grafana:          fakeGrafana{err: errors.New("error grafana")},
	},
}

var templateLaunch = `[{"test": %s,"rampTime": "%s","repeatTime": "%s","stopAllLorhammerTime": "%s","sleepBeforeCheckTime": "%s","shutdownAllLorhammerTime": "%s","sleepAtEndTime": "%s","init": %s,"provisioning": %s,"check": %s, "deploy": %s}]`

func TestLaunchTest(t *testing.T) {
	for _, test := range testsLaunch {
		var ct = test
		data := []byte(fmt.Sprintf(templateLaunch, ct.test, ct.rampTime, ct.repeatTime, ct.stopAll, ct.beforeCheck, ct.shutdownAll, ct.sleep, ct.init, ct.provisioning, ct.check, ct.deploy))
		tests, err := FromFile(data)
		if err != nil {
			t.Fatalf(`valid scenario should not return err %s for : "%s"`, err, ct.description)
		}
		if len(tests) != 1 {
			t.Fatalf(`1 valid scenario should return 1 valid testSuite for : "%s"`, ct.description)
		}
		if test.needProvisioning {
			provisioning.Provision(tests[0].UUID, tests[0].Provisioning, model.Register{})
		}
		report, err := tests[0].LaunchTest(fakeConsul{}, &fakeMqtt{}, ct.grafana)
		if ct.testValid && err != nil {
			t.Fatalf("valid test should not throw err %s", ct.description)
		} else if ct.testValid && report == nil {
			t.Fatalf("valid test should return report %s", ct.description)
		} else if !ct.testValid && err == nil {
			t.Fatalf("not valid test should throw err %s", ct.description)
		} else if !ct.testValid && report != nil {
			t.Fatalf("not valid test should not return report %s", ct.description)
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

type fakeConsul struct {
	serviceFirstError error
}

func (fakeConsul) GetAddress() string                                      { return "" }
func (fakeConsul) Register(ip string, hostname string, httpPort int) error { return nil }
func (f fakeConsul) ServiceFirst(name string, prefix string) (string, error) {
	return "prometheusUrl", f.serviceFirstError
}
func (fakeConsul) DeRegister(string) error                     { return nil }
func (fakeConsul) AllServices() ([]tools.ConsulService, error) { return nil, nil }

type fakeGrafana struct {
	err error
}

func (g fakeGrafana) MakeSnapshot(startTime time.Time, endTime time.Time) (string, error) {
	return "", g.err
}
