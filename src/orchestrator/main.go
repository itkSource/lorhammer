package main

import (
	"flag"
	"io/ioutil"
	"lorhammer/src/model"
	"lorhammer/src/orchestrator/checker"
	"lorhammer/src/orchestrator/cli"
	"lorhammer/src/orchestrator/command"
	"lorhammer/src/orchestrator/provisioning"
	"lorhammer/src/orchestrator/testSuite"
	"lorhammer/src/tools"
	"os"
	"runtime"
	"time"

	"github.com/sirupsen/logrus"
)

var version string // set at build time
var commit string  // set at build time
var date string    // set at build time

var logger = logrus.WithField("logger", "orchestrator/main")

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
		logger.Error("You need to specify at least -consul with ip:port")
		return
	}
	consulClient, err := tools.NewConsul(*consulAddr)
	if err != nil {
		logger.WithError(err).Panic("Can't build consul")
	}

	logrus.Warn("Welcome to the Lorhammer's Orchestrator")

	var currentTestSuite testSuite.TestSuite

	// MQTT PART
	mqttClient, err := tools.NewMqtt(host, consulClient)
	if err != nil {
		logger.WithError(err).Error("Can't build mqtt client")
	} else {
		if errMqtt := mqttClient.Connect(); errMqtt != nil {
			logger.WithError(errMqtt).Error("Error while connecting to mqtt")
		}
		if errHandleCmd := mqttClient.HandleCmd([]string{tools.MqttOrchestratorTopic}, func(cmd model.CMD) {
			if errApplyCmd := command.ApplyCmd(cmd, mqttClient, func(register model.Register) error {
				return provisioning.Provision(currentTestSuite.UUID, currentTestSuite.Provisioning, register)
			}); errApplyCmd != nil {
				logger.WithField("cmd", string(cmd.Payload)).WithError(errApplyCmd).Error("ApplyCmd error")
			}
		}); errHandleCmd != nil {
			logger.WithError(errHandleCmd).Error("Error while subscribing to topic")
		} else {
			logger.WithField("topic", tools.MqttOrchestratorTopic).Info("Listen mqtt")
		}
	}

	// GRAFANA
	grafanaClient, err := tools.NewGrafana(consulClient)
	if err != nil {
		logger.WithError(err).Error("Error while constructing new grafana api client")
	}

	// SCENARIO
	if *scenarioFromFile != "" {
		configFile, err := ioutil.ReadFile(*scenarioFromFile)
		if err != nil {
			logger.WithError(err).Panic("Error while reading test suite file")
		}
		tests, err := testSuite.FromFile(configFile)
		if err != nil {
			logger.WithError(err).WithField("file", *scenarioFromFile).Panic("Error while parsing test suite file")
		}
		checkErrors := make([]checker.Error, 0)
		for _, test := range tests {
			currentTestSuite = test
			testReport, err := test.LaunchTest(consulClient, mqttClient, grafanaClient)
			if err != nil {
				logger.WithError(err).Error("Error during test")
			} else if err := testReport.WriteFile(*reportFile); err != nil {
				logger.WithError(err).Error("Can't report test")
			} else {
				checkErrors = append(checkErrors, testReport.ChecksError...)
			}
			time.Sleep(test.SleepAtEndTime)
		}
		os.Exit(len(checkErrors))
	}
	if *startCli {
		cli.Start(mqttClient, consulClient)
	}
}
