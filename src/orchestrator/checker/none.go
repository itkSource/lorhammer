package checker

import (
	"encoding/json"
	"lorhammer/src/tools"
)

const NoneType = Type("none")

type none struct{}

func newNone(_ tools.Consul, _ json.RawMessage) (Checker, error) {
	return none{}, nil
}

func (_ none) Start() error {
	return nil
}

func (_ none) Check() ([]CheckerSuccess, []CheckerError) {
	return make([]CheckerSuccess, 0), make([]CheckerError, 0)
}
