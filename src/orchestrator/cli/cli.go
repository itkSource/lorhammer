package cli

import (
	"bufio"
	"github.com/Sirupsen/logrus"
	"lorhammer/src/orchestrator/command"
	"lorhammer/src/orchestrator/prometheus"
	"lorhammer/src/tools"
	"os"
	"strconv"
)

var LOG = logrus.WithField("logger", "orchestrator/cli/cli")

func Start(mqttClient tools.Mqtt, consulClient tools.Consul, prometheusApiClient prometheus.ApiClient) {
	scanner := bufio.NewScanner(os.Stdin)
	LOG.Warn("What do you wan to do ?")
	LOG.Warn("1 - stop all scenarios")
	LOG.Warn("2 - shutdown all lorhammers")
	LOG.Warn("3 - nb lorhammers")
	LOG.Warn("4 - clean consul services")
	switch scanChoose(scanner, []string{"1", "2", "3", "4"}) {
	case "1":
		command.StopScenario(mqttClient)
	case "2":
		command.ShutdownLorhammers(mqttClient)
	case "3":
		fetchAndDisplayNbLorhammer(prometheusApiClient)
	case "4":
		cleanConsulServices(scanner, consulClient)
	}
	Start(mqttClient, consulClient, prometheusApiClient) // infinite recursion
}

func scanChoose(scanner *bufio.Scanner, chooses []string) string {
	mapChooses := make(map[string]struct{})
	for _, s := range chooses {
		mapChooses[s] = struct{}{}
	}
	scanner.Scan()
	choose := scanner.Text()
	if _, ok := mapChooses[choose]; ok {
		return choose
	} else {
		LOG.WithFields(logrus.Fields{
			"response": choose,
		}).Error("Enter a valid choose please")
		return scanChoose(scanner, chooses)
	}
}

func fetchAndDisplayNbLorhammer(prometheusApiClient prometheus.ApiClient) {
	if nbLorhammer, err := prometheusApiClient.ExecQuery("count(lorhammer_durations_count)"); err != nil {
		LOG.WithError(err).Error("Error while retreiving nb lorhammer")
	} else {
		LOG.Warnf("They are %d lorhammers", int(nbLorhammer))
	}
}

func cleanConsulServices(scanner *bufio.Scanner, consulClient tools.Consul) {
	for { // block until user enter 0
		if services, err := consulClient.AllServices(); err != nil {
			LOG.WithError(err).Error("Can't retreive all services")
		} else {
			chooses := make([]string, len(services)+1)
			LOG.Warn("0 - return to main menu")
			chooses[0] = "0"
			for i, service := range services {
				LOG.Warnf("%d - unregister %s : %s", i+1, service.ServiceID, service.ServiceName)
				chooses[i+1] = strconv.Itoa(i + 1)
			}
			choose := scanChoose(scanner, chooses)
			if index, err := strconv.Atoi(choose); err != nil {
				LOG.WithError(err).Error("Choose not a number")
			} else {
				if index == 0 {
					return
				}
				if err := consulClient.DeRegister(services[index-1].ServiceID); err != nil {
					LOG.WithField("service", services[index-1]).WithError(err).Error("Can't unregister")
				}
			}
		}
	}
}
