package deploy

import (
	"encoding/json"
	"fmt"
	"lorhammer/src/tools"
	"time"

	"github.com/sirupsen/logrus"
)

var logDeploy = logrus.WithField("logger", "orchestrator/deploy/deploy")

//Type is type to define a deployer
type Type string

//Model represent a deployer in config file
type Model struct {
	Type                 Type
	SleepAfterDeployTime time.Duration
	Config               json.RawMessage
}

type modelJSON struct {
	Type                 Type            `json:"type"`
	SleepAfterDeployTime string          `json:"sleepAfterDeployTime"`
	Config               json.RawMessage `json:"config"`
}

//UnmarshalJSON permit to json to object a depoyer
func (m *Model) UnmarshalJSON(b []byte) error {
	mjson := &modelJSON{}
	err := json.Unmarshal(b, mjson)
	if err != nil {
		return err
	}
	m.Type = mjson.Type
	m.Config = mjson.Config
	if mjson.SleepAfterDeployTime == "" {
		m.SleepAfterDeployTime, _ = time.ParseDuration("0")
		return nil
	}
	d, err := time.ParseDuration(mjson.SleepAfterDeployTime)
	if err != nil {
		return err
	}
	m.SleepAfterDeployTime = d
	return nil
}

type deployer interface {
	RunBefore() error
	Deploy() error
	RunAfter() error
}

var deployers = make(map[Type]func(config json.RawMessage, consulClient tools.Consul) (deployer, error))

func init() {
	deployers[typeNone] = newNone
	deployers[typeDistant] = newDistantFromJSON
	deployers[typeAmazon] = newAmazonFromJSON
	deployers[typeLocal] = newLocalFromJSON
}

//Start launch a deployement
func Start(model Model, consulClient tools.Consul) error {
	dep, ok := deployers[model.Type]
	if !ok {
		return fmt.Errorf("Unknown type %s for deployer", model.Type)
	}
	d, err := dep(model.Config, consulClient)
	if err != nil {
		return err
	}
	if err := d.RunBefore(); err != nil {
		return err
	}
	if err := d.Deploy(); err != nil {
		return err
	}
	if err := d.RunAfter(); err != nil {
		return err
	}
	logDeploy.WithField("duration", model.SleepAfterDeployTime).Info("Sleep to lets new lorhammer start")
	time.Sleep(model.SleepAfterDeployTime)
	return nil
}
