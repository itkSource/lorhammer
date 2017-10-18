package provisioning

import (
	"encoding/json"
	"lorhammer/src/model"
)

const noneType = Type("none")

type none struct{}

func newNone(json.RawMessage) (provisioner, error) { return none{}, nil }

func (none) Provision(sensorsToRegister model.Register) error { return nil }

func (none) DeProvision() error { return nil }
