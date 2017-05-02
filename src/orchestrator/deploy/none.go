package deploy

import (
	"encoding/json"
	"lorhammer/src/tools"
)

const TypeNone = Type("none")

type none struct{}

func (_ none) RunBefore() error { return nil }
func (_ none) Deploy() error    { return nil }
func (_ none) RunAfter() error  { return nil }

func NewNone(_ json.RawMessage, _ tools.Consul) (Deployer, error) {
	return none{}, nil
}
