package provisioning

import (
	"encoding/json"
	"errors"
	"lorhammer/src/model"
	"sync"
)

//Type represent provioner type from json
type Type string

//Model is the representation of provisioner in json config file
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
	provisioners[noneType] = newNone
	provisioners[loraserverType] = newLoraserver
	provisioners[httpType] = newHTTPProvisioner
}

//Provision start a provisioner
func Provision(uuid string, provisioning Model, sensorsToRegister model.Register) error {
	err := errors.New("Unknown Provisioning type")
	if provisionerFabrik, ok := provisioners[provisioning.Type]; ok {
		instance, instanceExist := instances.Load(uuid)
		if !instanceExist {
			instance, err = provisionerFabrik(provisioning.Config)
			if err != nil {
				return err
			}
			instances.Store(uuid, instance)
		}
		return instance.(provisioner).Provision(sensorsToRegister)
	}
	return err
}

//DeProvision delete all references of a previous Provision()
func DeProvision(uuid string) error {
	if instance, ok := instances.Load(uuid); ok {
		instances.Delete(uuid)
		return instance.(provisioner).DeProvision()
	}
	return errors.New("You must Provision before DeProvision")
}
