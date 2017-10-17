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

	"github.com/Sirupsen/logrus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var version string // set at build time
var commit string  // set at build time
var date string    // set at build time

var logger = logrus.WithField("logger", "lorhammer/main")

func main() {
	showVersion := flag.Bool("version", false, "Show current version and build time")
	localIP := flag.String("local-ip", "", "The address used by others tools to access lorhammer instance")
	consulAddr := flag.String("consul", "", "The ip:port of consul")
	nbGateway := flag.Int("nb-gateway", 0, "The number of gateway to launch")
	minNbNode := flag.Int("min-nb-node", 1, "The minimal number of node by gateway")
	maxNbNode := flag.Int("max-nb-node", 1, "The maximal number of node by gateway")
	nsAddress := flag.String("ns-address", "127.0.0.1:1700", "NetworkServer ip:port address")
	logInfo := flag.Bool("vv", false, "log infos")
	logDebug := flag.Bool("vvv", false, "log debugs")
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
	httpPort, err := tools.FreeTcpPort()
	if err != nil {
		logger.WithError(err).Error("Free tcp port error")
	} else {
		logger.WithField("port", httpPort).Info("Tcp port reserved")
	}

	// IP
	ip, err := tools.DetectIp(*localIP)
	if err != nil {
		logger.WithError(err).Error("Ip error")
	} else {
		logger.WithField("ip", ip).Info("Ip discovered")
	}

	// HOSTNAME
	hostname, err := tools.Hostname(ip, httpPort)
	if err != nil {
		logger.WithError(err).Error("Hostname error")
	} else {
		logger.WithField("hostname", hostname).Info("Unique hostname generated")
	}

	// PROMETHEUS
	prometheus := tools.NewPrometheus()

	// CONSUL/MQTT
	if *consulAddr == "" && *nbGateway <= 0 {
		logger.Error("You need to specify at least -consul with ip:port")
		return
	}
	consulClient, err := tools.NewConsul(*consulAddr)
	if err != nil {
		logger.WithError(err).Warn("Consul not found, lorhammer is in standalone mode")
	} else {
		if err := consulClient.Register(ip, hostname, httpPort); err != nil {
			logger.WithError(err).Warn("Consul register error, lorhammer is in standalone mode")
		} else {
			mqttClient, err := tools.NewMqtt(hostname, consulClient)
			if err != nil {
				logger.WithError(err).Warn("Mqtt not found, lorhammer is in standalone mode")
			} else {
				if err := mqttClient.Connect(); err != nil {
					logger.WithError(err).Warn("Can't connect to mqtt, lorhammer is in standalone mode")
				}
				listenMqtt(mqttClient, []string{tools.MQTT_INIT_TOPIC, tools.MQTT_START_TOPIC + "/" + hostname}, hostname, prometheus)
			}
		}
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
			LOG.Info("Blocking routine waiting for cancel function")
			<-ctx.Done()
			LOG.Info("Releasing blocking routine after cancel function call")
			cmd := model.CMD{
				CmdName: model.STOP,
			}
			LOG.WithField("cmd ", cmd).Info("Apply Cmd Called")
			// mqtt client is unneeded in case of shutdown command
			command.ApplyCmd(cmd, nil, hostname, prometheus)
		}()
	} else {
		logger.Warn("No gateway, orchestrator will start scenarii")
	}

	// HTTP PART
	http.Handle("/metrics", promhttp.Handler())
	logger.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", httpPort), nil))
}

func listenMqtt(mqttClient tools.Mqtt, topics []string, hostname string, prometheus tools.Prometheus) {
	if err := mqttClient.HandleCmd(topics, func(cmd model.CMD) {
		command.ApplyCmd(cmd, mqttClient, hostname, prometheus)
	}); err != nil {
		logger.WithError(err).WithField("topics", topics).Error("Error while subscribing")
	} else {
		logrus.WithField("topics", topics).Info("Listen mqtt")
	}
}
