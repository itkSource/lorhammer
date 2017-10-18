package deploy

import (
	"encoding/json"
	"fmt"
	"github.com/sirupsen/logrus"
	"lorhammer/src/tools"
	"time"
)

var _LOG_DEPLOY = logrus.WithField("logger", "orchestrator/deploy/deploy")

type Type string

type Model struct {
	Type                 Type
	SleepAfterDeployTime time.Duration
	Config               json.RawMessage
}

type modelJson struct {
	Type                 Type            `json:"type"`
	SleepAfterDeployTime string          `json:"sleepAfterDeployTime"`
	Config               json.RawMessage `json:"config"`
}

func (m *Model) UnmarshalJSON(b []byte) error {
	mjson := &modelJson{}
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

type Deployer interface {
	RunBefore() error
	Deploy() error
	RunAfter() error
}

var deployers = make(map[Type]func(config json.RawMessage, consulClient tools.Consul) (Deployer, error))

func init() {
	deployers[TypeNone] = NewNone
	deployers[TypeDistant] = NewDistantFromJson
	deployers[TypeAmazon] = NewAmazonFromJson
	deployers[TypeLocal] = NewLocalFromJson
}

func Start(model Model, consulClient tools.Consul) error {
	var d Deployer
	var err error
	if dep := deployers[model.Type]; dep == nil {
		return fmt.Errorf("Unknown type %s for deployer", model.Type)
	} else {
		d, err = dep(model.Config, consulClient)
	}
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
	_LOG_DEPLOY.WithField("duration", model.SleepAfterDeployTime).Info("Sleep to lets prometheus discover new lorhammer")
	time.Sleep(model.SleepAfterDeployTime)
	return nil
}
