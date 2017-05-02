package provisioning

import (
	"encoding/json"
	"lorhammer/src/model"
)

const NoneType = Type("none")

type none struct{}

func NewNone(_ json.RawMessage) (provisioner, error) { return none{}, nil }

func (_ none) Provision(sensorsToRegister model.Register) error { return nil }

func (_ none) DeProvision() error { return nil }
