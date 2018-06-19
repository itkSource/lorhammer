package main

import (
	"flag"
	"fmt"
	"lorhammer/src/lorhammer/command"
	"lorhammer/src/lorhammer/scenario"
	"lorhammer/src/model"
	"lorhammer/src/tools"
	"net/http"
	"runtime"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sirupsen/logrus"
	"lorhammer/src/lorhammer/metrics"
)

var version string // set at build time
var commit string  // set at build time
var date string    // set at build time

var logger = logrus.WithField("logger", "lorhammer/main")

func main() {
	showVersion := flag.Bool("version", false, "Show current version and build time")
	port := flag.Int("port", 0, "The port to use to expose prometheus metrics, default 0 means random")
	mqttAddr := flag.String("mqtt", "", "The protocol://ip:port of mqtt")
	nbGateway := flag.Int("nb-gateway", 0, "The number of gateway to launch")
	minNbNode := flag.Int("min-nb-node", 1, "The minimal number of node by gateway")
	maxNbNode := flag.Int("max-nb-node", 1, "The maximal number of node by gateway")
	nsAddress := flag.String("ns-address", "127.0.0.1:1700", "NetworkServer ip:port address")
	logInfo := flag.Bool("vv", false, "log infos")
	logDebug := flag.Bool("vvv", false, "log debugs")
	maxWaitOrchestratorTime := flag.Duration("max-wait-orchestrator", 1*time.Minute, "The maximum time to wait for first communication with orchestrator")
	flag.Parse()

	if *showVersion {
		logrus.WithFields(logrus.Fields{
			"version":    version,
			"commit":     commit,
			"build time": date,
			"go version": runtime.Version(),
		}).Warn("Welcome to the Lorhammer")
		return
	}

	// LOGS
	if *logDebug {
		logrus.SetLevel(logrus.DebugLevel)
	} else if *logInfo {
		logrus.SetLevel(logrus.InfoLevel)
	} else {
		logrus.SetLevel(logrus.WarnLevel)
	}

	// PORT
	var httpPort int
	if *port == 0 {
		p, err := tools.FreeTCPPort()
		if err != nil {
			logger.WithError(err).Error("Free tcp port error")
		} else {
			logger.WithField("port", p).Info("Tcp port reserved")
			httpPort = p
		}
	} else {
		httpPort = *port
	}

	// HOSTNAME
	hostname, err := tools.Hostname(httpPort)
	if err != nil {
		logger.WithError(err).Error("Hostname error")
	} else {
		logger.WithField("hostname", hostname).Info("Unique hostname generated")
	}

	// PROMETHEUS
	prometheus := metrics.NewPrometheus()

	// MQTT
	if *mqttAddr == "" && *nbGateway <= 0 {
		logger.Error("You need to specify at least -mqtt with protocol://ip:port")
		return
	}
	mqttClient, err := tools.NewMqtt(hostname, *mqttAddr)
	if err != nil {
		logger.WithError(err).Warn("Mqtt not found, lorhammer is in standalone mode")
	} else {
		if err := mqttClient.Connect(); err != nil {
			logger.WithError(err).Warn("Can't connect to mqtt, lorhammer is in standalone mode")
		}

		// LINK TO ORCHESTRATOR
		lorhammerAddedChan := command.Start(mqttClient, hostname, *maxWaitOrchestratorTime)
		listenMqtt(mqttClient, []string{tools.MqttLorhammerTopic, tools.MqttLorhammerTopic + "/" + hostname}, hostname, lorhammerAddedChan, prometheus)
	}

	// SCENARIO
	if *nbGateway > 0 {
		logger.Warn("Launch manual scenario")
		sc, err := scenario.NewScenario(model.Init{
			NbGateway:          *nbGateway,
			NbNode:             [2]int{*minNbNode, *maxNbNode},
			NsAddress:          *nsAddress,
			ScenarioSleepTime:  [2]string{"10s", "10s"},
			GatewaySleepTime:   [2]string{"100ms", "500ms"},
			ReceiveTimeoutTime: "1s",
		})
		if err != nil {
			logger.WithError(err).Fatal("Can't create scenario with infos passed in flags")
		}
		ctx := sc.Cron(prometheus)
		go func() {
			logger.Info("Blocking routine waiting for cancel function")
			<-ctx.Done()
			logger.Info("Releasing blocking routine after cancel function call")
			cmd := model.CMD{
				CmdName: model.STOP,
			}
			logger.WithField("cmd ", cmd).Info("Apply Cmd Called")
			// mqtt client and lorhammerAddedChan are unneeded in case of shutdown command
			command.ApplyCmd(cmd, nil, hostname, nil, prometheus)
		}()
	} else {
		logger.Warn("No gateway, orchestrator will start scenarii")
	}

	// HTTP PART
	http.Handle("/metrics", promhttp.Handler())
	logger.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", httpPort), nil))
}

func listenMqtt(mqttClient tools.Mqtt, topics []string, hostname string, lorhammerAddedChan chan bool, prometheus metrics.Prometheus) {
	if err := mqttClient.HandleCmd(topics, func(cmd model.CMD) {
		command.ApplyCmd(cmd, mqttClient, hostname, lorhammerAddedChan, prometheus)
	}); err != nil {
		logger.WithError(err).WithField("topics", topics).Error("Error while subscribing")
	} else {
		logrus.WithField("topics", topics).Info("Listen mqtt")
	}
}
