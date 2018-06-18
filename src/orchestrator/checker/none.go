package checker

import (
	"encoding/json"
	"lorhammer/src/orchestrator/metrics"
)

const noneType = Type("none")

type none struct{}

func newNone(_ json.RawMessage, _ metrics.Prometheus) (Checker, error) {
	return none{}, nil
}

func (none) Start() error {
	return nil
}

func (none) Check() ([]Success, []Error) {
	return make([]Success, 0), make([]Error, 0)
}
