package cli

import (
	"bufio"
	"lorhammer/src/orchestrator/command"
	"lorhammer/src/tools"
	"os"

	"github.com/sirupsen/logrus"
)

var logger = logrus.WithField("logger", "orchestrator/cli/cli")

//Start launch the cli mode and ask to user to select actions
func Start(mqttClient tools.Mqtt) {
	scanner := bufio.NewScanner(os.Stdin)
	logger.Warn("What do you wan to do ?")
	logger.Warn("1 - stop all scenarios")
	logger.Warn("2 - shutdown all lorhammers")
	switch scanChoose(scanner, []string{"1", "2"}) {
	case "1":
		command.StopScenario(mqttClient)
	case "2":
		command.ShutdownLorhammers(mqttClient)
	}
	Start(mqttClient) // infinite recursion
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
