package deploy

import (
	"encoding/json"
	"github.com/Sirupsen/logrus"
	"lorhammer/src/tools"
	"os/exec"
)

const TypeLocal = "local"

var _LOG_LOCAL = logrus.WithField("logger", "orchestrator/deploy/local")

type localImpl struct {
	PathFile               string `json:"pathFile"`
	NbInstanceToLaunch     int    `json:"nbInstanceToLaunch"`
	CleanPreviousInstances bool   `json:"cleanPreviousInstances"`

	consulAddress string
	cmdFabric     func(name string, arg ...string) *exec.Cmd
}

func NewLocalFromJson(serialized json.RawMessage, consulClient tools.Consul) (Deployer, error) {
	d := &localImpl{}
	if err := json.Unmarshal(serialized, d); err != nil {
		return nil, err
	}
	d.consulAddress = consulClient.GetAddress()
	d.cmdFabric = exec.Command
	return d, nil
}

func (local *localImpl) RunBefore() error {
	if local.CleanPreviousInstances {
		cmd := local.cmdFabric("bash", "-c", "if pgrep lorhammer; then pkill lorhammer; fi")
		cmd.Run()
	}
	return nil
}

func (local *localImpl) Deploy() error {
	errs := LocalRunError{Errors: make([]error, 0)}
	chanErr := make(chan error)
	defer close(chanErr)

	for i := 0; i < local.NbInstanceToLaunch; i++ {
		go func() {
			_LOG_LOCAL.WithField("cmd", local.PathFile).WithField("args", "-consul "+local.consulAddress).WithField("nb", local.NbInstanceToLaunch).Info("Will exec cmd")
			var cmd = local.cmdFabric(local.PathFile, "-consul", local.consulAddress)
			if err := cmd.Start(); err != nil {
				_LOG_LOCAL.WithError(err).Error("Local output error when launching")
				chanErr <- err
			} else {
				chanErr <- nil
				if errWait := cmd.Wait(); errWait != nil {
					_LOG_LOCAL.WithError(errWait).Error("Local output error when wait")
				}
			}
		}()
	}

	for i := 0; i < local.NbInstanceToLaunch; i++ {
		if err := <-chanErr; err != nil {
			errs.Errors = append(errs.Errors, err)
		}
	}

	if len(errs.Errors) > 0 {
		return errs
	} else {
		return nil
	}
}

func (local *localImpl) RunAfter() error {
	return nil
}

type LocalRunError struct {
	Errors []error
}

func (localErr LocalRunError) Error() string {
	s := "LocalRunError: \n"
	for _, err := range localErr.Errors {
		s = s + " \n " + err.Error()
	}
	return s
}
