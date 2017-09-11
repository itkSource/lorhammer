package provisioning

import (
	"encoding/json"
	"errors"
	"lorhammer/src/model"
	"sync"
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
var instances = sync.Map{}

func init() {
	provisioners[NoneType] = NewNone
	provisioners[loraserverType] = newLoraserver
	provisioners[SemtechV4Type] = NewSemtechV4
	provisioners[HttpType] = NewHttpProvisioner
}

func Provision(uuid string, provisioning Model, sensorsToRegister model.Register) error {
	if pro := provisioners[provisioning.Type]; pro != nil {
		if instance, ok := instances.Load(uuid); ok {
			if err := instance.(provisioner).Provision(sensorsToRegister); err != nil {
				return err
			} else {
				return nil
			}
		} else if instance, err := pro(provisioning.Config); err != nil {
			return err
		} else {
			instances.Store(uuid, instance)
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
	if instance, ok := instances.Load(uuid); ok {
		instances.Delete(uuid)
		return instance.(provisioner).DeProvision()
	}
	return errors.New("You must Provision before DeProvision")
}
