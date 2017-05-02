package provisioning

import (
	"encoding/json"
	"errors"
	"github.com/orcaman/concurrent-map"
	"lorhammer/src/model"
)

type Type string

type Model struct {
	Type   Type            `json:"type"`
	Config json.RawMessage `json:"config"`
}

type provisioner interface {
	Provision(sensorsToRegister model.Register) error
	DeProvision() error
}

var provisioners = make(map[Type]func(config json.RawMessage) (provisioner, error))
var instances = cmap.New()

func init() {
	provisioners[NoneType] = NewNone
	provisioners[LoraserverType] = NewLoraserver
	provisioners[SemtechV4Type] = NewSemtechV4
}

func Provision(uuid string, provisioning Model, sensorsToRegister model.Register) error {
	if pro := provisioners[provisioning.Type]; pro != nil {
		if instance, err := pro(provisioning.Config); err != nil {
			return err
		} else {
			instances.Set(uuid, instance)
			if err := instance.Provision(sensorsToRegister); err != nil {
				return err
			} else {
				return nil
			}
		}

	}
	return errors.New("Unknown Provisioning type")
}

func DeProvision(uuid string) error {
	//TODO this function is called twice (at stopTime and shutdownTime), the second time the instance
	//is not there anymore so the error is logged. need to be fixed
	if instance, ok := instances.Pop(uuid); ok {
		return instance.(provisioner).DeProvision()
	}
	return errors.New("You must Provision before DeProvision")
}
