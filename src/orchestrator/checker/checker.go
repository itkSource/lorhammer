package checker

import (
	"encoding/json"
	"fmt"
	"lorhammer/src/orchestrator/metrics"
)

//Type is a type to define a checker
type Type string

//Model is the structure loaded from json file
type Model struct {
	Type   Type            `json:"type"`
	Config json.RawMessage `json:"config"`
}

//Success is the interface fo details success depending on implementation
type Success interface {
	Details() map[string]interface{}
}

//Error is the interface fo details error depending on implementation
type Error interface {
	Details() map[string]interface{}
}

//Checker check if data is correct depending on implementation
type Checker interface {
	Start() error
	Check() ([]Success, []Error)
}

var checkers = make(map[Type]func(config json.RawMessage, prometheus metrics.Prometheus) (Checker, error))

func init() {
	checkers[noneType] = newNone
	checkers[prometheusType] = newPrometheus
	checkers[kafkaType] = newKafka
	checkers[mqttType] = newMqtt
}

//Get return a checker if the Model is an implementation of Checker
func Get(checker Model, prometheus metrics.Prometheus) (Checker, error) {
	if checkers[checker.Type] == nil {
		return nil, fmt.Errorf("Unknown checker type %s", checker.Type)
	}
	c, err := checkers[checker.Type](checker.Config, prometheus)
	if err != nil {
		return nil, err
	}
	if err := c.Start(); err != nil {
		return nil, err
	}
	return c, nil
}
