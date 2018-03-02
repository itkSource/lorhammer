package deploy

import (
	"encoding/json"
	"lorhammer/src/tools"
	"os/exec"
	"strconv"

	"github.com/sirupsen/logrus"
)

const typeLocal = "local"

var logLocal = logrus.WithField("logger", "orchestrator/deploy/local")

type localImpl struct {
	PathFile               string `json:"pathFile"`
	NbInstanceToLaunch     int    `json:"nbInstanceToLaunch"`
	CleanPreviousInstances bool   `json:"cleanPreviousInstances"`
	LocalIP                string `json:"localIp"`
	Port                   int    `json:"port"`

	consulAddress string
	cmdFabric     func(name string, arg ...string) *exec.Cmd
}

func newLocalFromJSON(serialized json.RawMessage, consulClient tools.Consul) (deployer, error) {
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
	errs := localRunError{Errors: make([]error, 0)}
	chanErr := make(chan error)
	defer close(chanErr)

	for i := 0; i < local.NbInstanceToLaunch; i++ {
		go func() {
			args := []string{"-consul", local.consulAddress}
			if local.LocalIP != "" {
				args = append(args, "-local-ip", local.LocalIP)
			}
			if local.Port != 0 {
				args = append(args, "-port", strconv.Itoa(local.Port))
			}
			logLocal.WithField("cmd", local.PathFile).WithField("args", args).WithField("nb", local.NbInstanceToLaunch).Debug("Will exec cmd")
			var cmd = local.cmdFabric(local.PathFile, args...)
			if err := cmd.Start(); err != nil {
				logLocal.WithError(err).Error("Local output error when launching")
				chanErr <- err
			} else {
				chanErr <- nil
				if errWait := cmd.Wait(); errWait != nil {
					logLocal.WithError(errWait).Error("Local output error when wait")
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
	}
	return nil
}

func (local *localImpl) RunAfter() error {
	return nil
}

type localRunError struct {
	Errors []error
}

func (localErr localRunError) Error() string {
	s := "LocalRunError: \n"
	for _, err := range localErr.Errors {
		s = s + " \n " + err.Error()
	}
	return s
}
