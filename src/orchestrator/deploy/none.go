package deploy

import (
	"encoding/json"
	"lorhammer/src/tools"
)

const typeNone = Type("none")

type none struct{}

func (none) RunBefore() error { return nil }
func (none) Deploy() error    { return nil }
func (none) RunAfter() error  { return nil }

func newNone(json.RawMessage, tools.Mqtt) (deployer, error) {
	return none{}, nil
}
