package deploy

import (
	"encoding/json"
	"fmt"
	"lorhammer/src/tools"
	"os/exec"

	"github.com/sirupsen/logrus"
)

const typeDistant = "distant"

var logDistant = logrus.WithField("logger", "orchestrator/deploy/distant")

type arrayDistantImpl struct {
	Instances []distantImpl `json:"instances"`
	cmdFabric func(name string, arg ...string) *exec.Cmd
}

type distantImpl struct {
	SSHKeyPath        string `json:"sshKeyPath"`
	User              string `json:"user"`
	IPServer          string `json:"ipServer"`
	PathFile          string `json:"pathFile"`
	PathWhereScp      string `json:"pathWhereScp"`
	BeforeCmd         string `json:"beforeCmd"`
	AfterCmd          string `json:"afterCmd"`
	NbDistantToLaunch int    `json:"nbDistantToLaunch"`
}

func newDistantFromJSON(serialized json.RawMessage, _ tools.Mqtt) (deployer, error) {
	d, err := newDistantImpl(serialized)
	if err != nil {
		return nil, err
	}
	return d, nil
}

func newDistantImpl(serialized json.RawMessage) (*arrayDistantImpl, error) {
	d := &arrayDistantImpl{}
	if err := json.Unmarshal(serialized, d); err != nil {
		return nil, err
	}
	d.cmdFabric = exec.Command
	return d, nil
}

func (distant *arrayDistantImpl) RunBefore() error {
	for _, instance := range distant.Instances {
		err := instance.runCmd(instance.BeforeCmd, distant.cmdFabric)
		if err != nil {
			return err
		}
	}
	return nil
}

func (distant *arrayDistantImpl) Deploy() error {

	for _, instance := range distant.Instances {
		c := fmt.Sprintf("%s@%s:%s", instance.User, instance.IPServer, instance.PathWhereScp)
		logDistant.WithField("cmd", "scp -q -i "+instance.SSHKeyPath+" -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null "+instance.PathFile+" "+c).Info("Will exec cmd")
		cmd := distant.cmdFabric("scp", "-q", "-i", instance.SSHKeyPath, "-o", "StrictHostKeyChecking=no", "-o", "UserKnownHostsFile=/dev/null", instance.PathFile, c)
		if stdoutStderr, err := cmd.CombinedOutput(); err != nil {
			logDistant.WithField("output", fmt.Sprintf("%s", stdoutStderr)).Info("Scp output")
			return err
		}
	}
	return nil
}

func (distant *arrayDistantImpl) RunAfter() error {
	for _, instance := range distant.Instances {
		err := instance.runCmd(instance.AfterCmd, distant.cmdFabric)
		if err != nil {
			return err
		}
	}
	return nil
}

type distantRunError struct {
	Errors []error
}

func (distErr distantRunError) Error() string {
	s := "DistantRunError: \n"
	for _, err := range distErr.Errors {
		s = s + " \n " + err.Error()
	}
	return s
}

func (distant *distantImpl) runCmd(cmd string, execFunc func(name string, arg ...string) *exec.Cmd) error {
	errs := distantRunError{Errors: make([]error, 0)}
	ip := fmt.Sprintf("%s@%s", distant.User, distant.IPServer)
	logDistant.WithField("cmd", "ssh -q -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null "+ip+" "+cmd).Info("Will exec cmd")

	chanErr := make(chan error)
	defer close(chanErr)

	for i := 0; i < distant.NbDistantToLaunch; i++ {
		go func() {
			cmd := execFunc("ssh", "-q", "-o", "StrictHostKeyChecking=no", "-o", "UserKnownHostsFile=/dev/null", ip, cmd)
			if stdoutStderr, err := cmd.CombinedOutput(); err != nil {
				logDistant.WithField("output", fmt.Sprintf("%s", stdoutStderr)).Info("Ssh output")
				chanErr <- err
			} else {
				chanErr <- nil
			}
		}()
	}

	for i := 0; i < distant.NbDistantToLaunch; i++ {
		if err := <-chanErr; err != nil {
			errs.Errors = append(errs.Errors, err)
		}
	}

	if len(errs.Errors) > 0 {
		return errs
	}
	return nil
}
