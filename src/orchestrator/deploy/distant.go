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

type distantImpl struct {
	SSHKeyPath        string `json:"sshKeyPath"`
	User              string `json:"user"`
	IPServer          string `json:"ipServer"`
	PathFile          string `json:"pathFile"`
	PathWhereScp      string `json:"pathWhereScp"`
	BeforeCmd         string `json:"beforeCmd"`
	AfterCmd          string `json:"afterCmd"`
	NbDistantToLaunch int    `json:"nbDistantToLaunch"`

	cmdFabric func(name string, arg ...string) *exec.Cmd
}

func newDistantFromJSON(serialized json.RawMessage, _ tools.Consul) (deployer, error) {
	d, err := newDistantImpl(serialized)
	if err != nil {
		return nil, err
	}
	return d, nil
}

func newDistantImpl(serialized json.RawMessage) (*distantImpl, error) {
	d := &distantImpl{}
	if err := json.Unmarshal(serialized, d); err != nil {
		return nil, err
	}
	d.cmdFabric = exec.Command
	return d, nil
}

func (distant *distantImpl) RunBefore() error {
	return distant.runCmd(distant.BeforeCmd)
}

func (distant *distantImpl) Deploy() error {
	c := fmt.Sprintf("%s@%s:%s", distant.User, distant.IPServer, distant.PathWhereScp)
	logDistant.WithField("cmd", "scp -q -i "+distant.SSHKeyPath+" -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null "+distant.PathFile+" "+c).Info("Will exec cmd")
	cmd := distant.cmdFabric("scp", "-q", "-i", distant.SSHKeyPath, "-o", "StrictHostKeyChecking=no", "-o", "UserKnownHostsFile=/dev/null", distant.PathFile, c)
	if stdoutStderr, err := cmd.CombinedOutput(); err != nil {
		logDistant.WithField("output", fmt.Sprintf("%s", stdoutStderr)).Info("Scp output")
		return err
	}
	return nil
}

func (distant *distantImpl) RunAfter() error {
	return distant.runCmd(distant.AfterCmd)
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

func (distant *distantImpl) runCmd(cmd string) error {
	errs := distantRunError{Errors: make([]error, 0)}
	ip := fmt.Sprintf("%s@%s", distant.User, distant.IPServer)
	logDistant.WithField("cmd", "ssh -q -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null "+ip+" "+cmd).Info("Will exec cmd")

	chanErr := make(chan error)
	defer close(chanErr)

	for i := 0; i < distant.NbDistantToLaunch; i++ {
		go func() {
			cmd := distant.cmdFabric("ssh", "-q", "-o", "StrictHostKeyChecking=no", "-o", "UserKnownHostsFile=/dev/null", ip, cmd)
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
