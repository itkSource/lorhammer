package deploy

import (
	"encoding/json"
	"lorhammer/src/model"
	"os/exec"
	"strings"
	"testing"
	"time"
)

type sharedCounter struct {
	count int
}

type fakeMqtt struct {
}

func (m *fakeMqtt) GetAddress() string                                          { return "" }
func (m *fakeMqtt) Connect() error                                              { return nil }
func (m *fakeMqtt) Disconnect()                                                 {}
func (m *fakeMqtt) Handle(topics []string, handle func(messgae []byte)) error   { return nil }
func (m *fakeMqtt) HandleCmd(topics []string, handle func(cmd model.CMD)) error { return nil }
func (m *fakeMqtt) PublishCmd(topic string, cmdName model.CommandName) error    { return nil }
func (m *fakeMqtt) PublishSubCmd(topic string, cmdName model.CommandName, subCmd interface{}) error {
	return nil
}

func newLocalFromJSONTest(j string) (deployer, error) {
	raw := json.RawMessage([]byte(j))
	return newLocalFromJSON(raw, &fakeMqtt{})
}

func testCall(countChan chan bool, nbRequired int, t *testing.T) {
	countCall := 0
	if nbRequired > 0 {
		for i := 0; i < nbRequired; i++ {
			select {
			case <-countChan:
				countCall++
			case <-time.After(100 * time.Millisecond):
			}
		}
	} else {
		select {
		case <-countChan:
			countCall++
		case <-time.After(100 * time.Millisecond):
			//check if no call was made
		}
	}
	if countCall != nbRequired {
		t.Fatalf("Cmd should be called %d times instead of %d", nbRequired, countCall)
	}
}

func TestNewLocalFromJson(t *testing.T) {
	d, err := newLocalFromJSONTest(`{}`)
	if err != nil {
		t.Fatal("good local deployer json should not throw error")
	}
	if d == nil {
		t.Fatal("good local deployer json should be created")
	}
}

func TestNewLocalFromJsonError(t *testing.T) {
	d, err := newLocalFromJSONTest(`{`)
	if err == nil {
		t.Fatal("bad local deployer json should throw error")
	}
	if d != nil {
		t.Fatal("bad local deployer json should not be created")
	}
}

func TestLocalImpl_RunBefore(t *testing.T) {
	d, err := newLocalFromJSONTest(`{"pathFile": "/", "nbInstanceToLaunch": 3, "cleanPreviousInstances": true}`)
	if err != nil {
		t.Fatal("good local deployer json should not throw error")
	}
	countChan := make(chan bool)
	defer close(countChan)
	d.(*localImpl).cmdFabric = func(name string, arg ...string) *exec.Cmd {
		go func() {
			countChan <- true
		}()
		return exec.Command("ls")
	}
	if err := d.RunBefore(); err != nil {
		t.Fatal("LocalDeploy run before should not return error")
	}
	testCall(countChan, 1, t)
}

func TestLocalImpl_RunBeforeNoClean(t *testing.T) {
	d, err := newLocalFromJSONTest(`{"pathFile": "/", "nbInstanceToLaunch": 3, "cleanPreviousInstances": false}`)
	if err != nil {
		t.Fatal("good local deployer json should not throw error")
	}
	countChan := make(chan bool)
	defer close(countChan)
	d.(*localImpl).cmdFabric = func(name string, arg ...string) *exec.Cmd {
		go func() {
			countChan <- true
		}()
		return nil
	}
	if err := d.RunBefore(); err != nil {
		t.Fatal("LocalDeploy run before should not return error")
	}
	testCall(countChan, 0, t)
}

func TestLocalImpl_Deploy(t *testing.T) {
	d, err := newLocalFromJSONTest(`{"pathFile": "/", "nbInstanceToLaunch": 3, "cleanPreviousInstances": false}`)
	if err != nil {
		t.Fatal("good local deployer json should not throw error")
	}
	countChan := make(chan bool)
	defer close(countChan)
	d.(*localImpl).cmdFabric = func(name string, arg ...string) *exec.Cmd {
		go func() {
			countChan <- true
		}()
		return exec.Command("ls")
	}
	if err := d.Deploy(); err != nil {
		t.Fatal("LocalDeploy Deploy() should not return error")
	}
	testCall(countChan, 3, t)
}

func TestLocalImpl_DeployErr(t *testing.T) {
	d, err := newLocalFromJSONTest(`{"pathFile": "/", "nbInstanceToLaunch": 3, "cleanPreviousInstances": false}`)
	if err != nil {
		t.Fatal("good local deployer json should not throw error")
	}
	countChan := make(chan bool)
	defer close(countChan)
	d.(*localImpl).cmdFabric = func(name string, arg ...string) *exec.Cmd {
		go func() {
			countChan <- true
		}()
		return exec.Command("")
	}
	if err := d.Deploy(); err == nil {
		t.Fatal("LocalDeploy Deploy() with bad command should return error")
	} else {
		if len(err.(localRunError).Errors) != 3 {
			t.Fatal("3 instances errors should return 3 errors")
		}
		if strings.Count(err.Error(), "\n") < 3 {
			t.Fatal("Error reporting should contains 3 \n at least")
		}
	}
	testCall(countChan, 3, t)
}

func TestLocalImpl_RunAfter(t *testing.T) {
	d, err := newLocalFromJSONTest(`{"pathFile": "/", "nbInstanceToLaunch": 3, "cleanPreviousInstances": false}`)
	if err != nil {
		t.Fatal("good local deployer json should not throw error")
	}
	if err := d.RunAfter(); err != nil {
		t.Fatal("run after should not throw error")
	}
}
