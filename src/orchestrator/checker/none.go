package checker

import (
	"encoding/json"
	"lorhammer/src/tools"
)

const noneType = Type("none")

type none struct{}

func newNone(_ tools.Consul, _ json.RawMessage) (Checker, error) {
	return none{}, nil
}

func (none) Start() error {
	return nil
}

func (none) Check() ([]Success, []Error) {
	return make([]Success, 0), make([]Error, 0)
}
