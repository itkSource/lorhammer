package deploy

import (
	"encoding/json"
	"fmt"
	"github.com/sirupsen/logrus"
	"lorhammer/src/tools"
	"os/exec"
)

const TypeDistant = "distant"

var _LOG_DISTANT = logrus.WithField("logger", "orchestrator/deploy/distant")

type distantImpl struct {
	SshKeyPath        string `json:"sshKeyPath"`
	User              string `json:"user"`
	IpServer          string `json:"ipServer"`
	PathFile          string `json:"pathFile"`
	PathWhereScp      string `json:"pathWhereScp"`
	BeforeCmd         string `json:"beforeCmd"`
	AfterCmd          string `json:"afterCmd"`
	NbDistantToLaunch int    `json:"nbDistantToLaunch"`

	cmdFabric func(name string, arg ...string) *exec.Cmd
}

func NewDistantFromJson(serialized json.RawMessage, _ tools.Consul) (Deployer, error) {
	if d, err := newDistantImpl(serialized); err != nil {
		return nil, err
	} else {
		return d, nil
	}
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
	c := fmt.Sprintf("%s@%s:%s", distant.User, distant.IpServer, distant.PathWhereScp)
	_LOG_DISTANT.WithField("cmd", "scp -q -i "+distant.SshKeyPath+" -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null "+distant.PathFile+" "+c).Info("Will exec cmd")
	cmd := distant.cmdFabric("scp", "-q", "-i", distant.SshKeyPath, "-o", "StrictHostKeyChecking=no", "-o", "UserKnownHostsFile=/dev/null", distant.PathFile, c)
	if stdoutStderr, err := cmd.CombinedOutput(); err != nil {
		_LOG_DISTANT.WithField("output", fmt.Sprintf("%s", stdoutStderr)).Info("Scp output")
		return err
	}
	return nil
}

func (distant *distantImpl) RunAfter() error {
	return distant.runCmd(distant.AfterCmd)
}

type DistantRunError struct {
	Errors []error
}

func (distErr DistantRunError) Error() string {
	s := "DistantRunError: \n"
	for _, err := range distErr.Errors {
		s = s + " \n " + err.Error()
	}
	return s
}

func (distant *distantImpl) runCmd(cmd string) error {
	errs := DistantRunError{Errors: make([]error, 0)}
	ip := fmt.Sprintf("%s@%s", distant.User, distant.IpServer)
	_LOG_DISTANT.WithField("cmd", "ssh -q -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null "+ip+" "+cmd).Info("Will exec cmd")

	chanErr := make(chan error)
	defer close(chanErr)

	for i := 0; i < distant.NbDistantToLaunch; i++ {
		go func() {
			cmd := distant.cmdFabric("ssh", "-q", "-o", "StrictHostKeyChecking=no", "-o", "UserKnownHostsFile=/dev/null", ip, cmd)
			if stdoutStderr, err := cmd.CombinedOutput(); err != nil {
				_LOG_DISTANT.WithField("output", fmt.Sprintf("%s", stdoutStderr)).Info("Ssh output")
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
	} else {
		return nil
	}
}
