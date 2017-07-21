package main

import (
	"flag"
	"github.com/Sirupsen/logrus"
	"io/ioutil"
	"lorhammer/src/model"
	"lorhammer/src/orchestrator/cli"
	"lorhammer/src/orchestrator/command"
	"lorhammer/src/orchestrator/provisioning"
	"lorhammer/src/orchestrator/testSuite"
	"lorhammer/src/tools"
	"runtime"
	"time"
)

var version string // set at build time
var commit string  // set at build time
var date string    // set at build time

var LOG = logrus.WithField("logger", "orchestrator/orchestrator")

func main() {
	showVersion := flag.Bool("version", false, "Show current version and build time")
	consulAddr := flag.String("consul", "", "The ip:port of consul")
	scenarioFromFile := flag.String("from-file", "", "A file containing a scenario to launch")
	reportFile := flag.String("report-file", "./report.json", "A file to fill reports tests in json")
	startCli := flag.Bool("cli", false, "Enter in cli mode and access to menu (stop/kill all lorhammers...)")
	flag.Parse()

	if *showVersion {
		logrus.WithFields(logrus.Fields{
			"version":    version,
			"commit":     commit,
			"build time": date,
			"go version": runtime.Version(),
		}).Warn("Welcome to the Lorhammer's Orchestrator")
		return
	}

	// HOSTNAME
	host := "orchestrator"

	// CONSUL PART
	if *consulAddr == "" {
		LOG.Error("You need to specify at least -consul with ip:port")
		return
	}
	consulClient, err := tools.NewConsul(*consulAddr)
	if err != nil {
		LOG.WithError(err).Panic("Can't build consul")
	}

	logrus.Warn("Welcome to the Lorhammer's Orchestrator")

	var currentTestSuite testSuite.TestSuite

	// MQTT PART
	mqttClient, err := tools.NewMqtt(host, consulClient)
	if err != nil {
		LOG.WithError(err).Error("Can't build mqtt client")
	} else {
		if errMqtt := mqttClient.Connect(); errMqtt != nil {
			LOG.WithError(errMqtt).Error("Error while connecting to mqtt")
		}
		if errHandleCmd := mqttClient.HandleCmd([]string{tools.MQTT_ORCHESTRATOR_TOPIC}, func(cmd model.CMD) {
			if errApplyCmd := command.ApplyCmd(cmd, mqttClient, func(register model.Register) error {
				return provisioning.Provision(currentTestSuite.Uuid, currentTestSuite.Provisioning, register)
			}); errApplyCmd != nil {
				LOG.WithField("cmd", cmd).WithError(errApplyCmd).Error("ApplyCmd error")
			}
		}); errHandleCmd != nil {
			LOG.WithError(errHandleCmd).Error("Error while subscribing to topic")
		} else {
			LOG.WithField("topic", tools.MQTT_ORCHESTRATOR_TOPIC).Info("Listen mqtt")
		}
	}

	// GRAFANA
	grafanaClient, err := tools.NewGrafana(consulClient)
	if err != nil {
		LOG.WithError(err).Error("Error while constructing new grafana api client")
	}

	if *scenarioFromFile != "" {
		configFile, err := ioutil.ReadFile(*scenarioFromFile)
		if err != nil {
			LOG.WithError(err).Panic("Error while reading test suite file")
		}
		tests, err := testSuite.FromFile(configFile)
		if err != nil {
			LOG.WithError(err).WithField("file", *scenarioFromFile).Panic("Error while parsing test suite file")
		}
		for _, test := range tests {
			currentTestSuite = test
			testReport, err := test.LaunchTest(consulClient, mqttClient, grafanaClient)
			if err != nil {
				LOG.WithError(err).Error("Error during test")
			} else if err := testSuite.WriteFile(testReport, *reportFile); err != nil {
				LOG.WithError(err).Error("Can't report test")
			}
			time.Sleep(test.SleepAtEndTime)
		}
	}
	if *startCli {
		cli.Start(mqttClient, consulClient)
	}
}
