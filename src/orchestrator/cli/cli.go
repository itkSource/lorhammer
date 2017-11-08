package cli

import (
	"bufio"
	"lorhammer/src/orchestrator/command"
	"lorhammer/src/orchestrator/prometheus"
	"lorhammer/src/tools"
	"os"
	"strconv"

	"github.com/sirupsen/logrus"
)

var logger = logrus.WithField("logger", "orchestrator/cli/cli")

//Start launch the cli mode and ask to user to select actions
func Start(mqttClient tools.Mqtt, consulClient tools.Consul) {
	scanner := bufio.NewScanner(os.Stdin)
	logger.Warn("What do you wan to do ?")
	logger.Warn("1 - stop all scenarios")
	logger.Warn("2 - shutdown all lorhammers")
	logger.Warn("3 - nb lorhammers")
	logger.Warn("4 - clean consul services")
	switch scanChoose(scanner, []string{"1", "2", "3", "4"}) {
	case "1":
		command.StopScenario(mqttClient)
	case "2":
		command.ShutdownLorhammers(mqttClient)
	case "3":
		fetchAndDisplayNbLorhammer(consulClient)
	case "4":
		cleanConsulServices(scanner, consulClient)
	}
	Start(mqttClient, consulClient) // infinite recursion
}

func scanChoose(scanner *bufio.Scanner, chooses []string) string {
	mapChooses := make(map[string]struct{})
	for _, s := range chooses {
		mapChooses[s] = struct{}{}
	}
	scanner.Scan()
	choose := scanner.Text()
	_, ok := mapChooses[choose]
	if ok {
		return choose
	}
	logger.WithFields(logrus.Fields{
		"response": choose,
	}).Error("Enter a valid choose please")
	return scanChoose(scanner, chooses)
}

func fetchAndDisplayNbLorhammer(consulClient tools.Consul) {
	prometheusAPIClient, err := prometheus.NewAPIClient(consulClient)
	if err != nil {
		logger.WithError(err).Error("Error while constructing new prometheus api client")
	}
	if nbLorhammer, err := prometheusAPIClient.ExecQuery("count(lorhammer_durations_count)"); err != nil {
		logger.WithError(err).Error("Error while retreiving nb lorhammer")
	} else {
		logger.Warnf("They are %d lorhammers", int(nbLorhammer))
	}
}

func cleanConsulServices(scanner *bufio.Scanner, consulClient tools.Consul) {
	for { // block until user enter 0
		if services, err := consulClient.AllServices(); err != nil {
			logger.WithError(err).Error("Can't retrieve all services")
		} else {
			chooses := make([]string, len(services)+1)
			logger.Warn("0 - return to main menu")
			chooses[0] = "0"
			for i, service := range services {
				logger.Warnf("%d - unregister %s : %s", i+1, service.ServiceID, service.ServiceName)
				chooses[i+1] = strconv.Itoa(i + 1)
			}
			choose := scanChoose(scanner, chooses)
			if index, err := strconv.Atoi(choose); err != nil {
				logger.WithError(err).Error("Choose not a number")
			} else {
				if index == 0 {
					return
				}
				if err := consulClient.DeRegister(services[index-1].ServiceID); err != nil {
					logger.WithField("service", services[index-1]).WithError(err).Error("Can't unregister")
				}
			}
		}
	}
}
