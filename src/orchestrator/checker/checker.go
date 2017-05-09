package checker

import (
	"encoding/json"
	"fmt"
	"lorhammer/src/tools"
)

type Type string

type Model struct {
	Type   Type            `json:"type"`
	Config json.RawMessage `json:"config"`
}

type CheckerSuccess interface {
	Details() map[string]interface{}
}

type CheckerError interface {
	Details() map[string]interface{}
}

type Checker interface {
	Start() error
	Check() ([]CheckerSuccess, []CheckerError)
}

var checkers = make(map[Type]func(consulClient tools.Consul, config json.RawMessage) (Checker, error))

func init() {
	checkers[NoneType] = newNone
	checkers[PrometheusType] = newPrometheus
	checkers[KafkaType] = newKafka
}

func Get(consulClient tools.Consul, checker Model) (Checker, error) {
	if checkers[checker.Type] == nil {
		return nil, fmt.Errorf("Unknown checker type %s", checker.Type)
	}
	c, err := checkers[checker.Type](consulClient, checker.Config)
	if err != nil {
		return nil, err
	}
	if err := c.Start(); err != nil {
		return nil, err
	}
	return c, nil
}
