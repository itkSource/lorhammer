package deploy

import (
	"encoding/json"
	"os/exec"
	"strings"
	"testing"
)

func newDistantFromJSONTest(j string) (deployer, error) {
	raw := json.RawMessage([]byte(j))
	return newDistantFromJSON(raw, &fakeMqtt{})
}

func TestNewDistantFromJson(t *testing.T) {
	d, err := newDistantFromJSONTest(`{}`)
	if err != nil {
		t.Fatal("good distant deployer json should not throw error")
	}
	if d == nil {
		t.Fatal("good distant deployer json should be created")
	}
}

func TestNewDistantFromJsonError(t *testing.T) {
	d, err := newDistantFromJSONTest(`{`)
	if err == nil {
		t.Fatal("bad distant deployer json should throw error")
	}
	if d != nil {
		t.Fatal("bad distant deployer json should not be created")
	}
}

func TestDistantImpl_RunBefore(t *testing.T) {
	d, err := newDistantFromJSONTest(`{ "instances": [ { "beforeCmd": "/", "nbDistantToLaunch": 1 } ] }`)
	if err != nil {
		t.Fatalf("good local deployer json should not throw error : %s", err)
	}
	hasBeenCalledWith := ""
	d.(*arrayDistantImpl).cmdFabric = func(name string, arg ...string) *exec.Cmd {
		hasBeenCalledWith = name
		return exec.Command("ls")
	}
	if err := d.RunBefore(); err != nil {
		t.Fatal("DistantDeploy run before should not return error")
	}
	if hasBeenCalledWith != "ssh" {
		t.Fatalf("Before should call ssh cmd instead of %s", hasBeenCalledWith)
	}
}

func TestDistantImpl_RunBeforeError(t *testing.T) {
	d, err := newDistantFromJSONTest(`{ "instances": [ {"beforeCmd": "/", "nbDistantToLaunch": 2} ] } `)
	if err != nil {
		t.Fatal("good local deployer json should not throw error")
	}
	d.(*arrayDistantImpl).cmdFabric = func(name string, arg ...string) *exec.Cmd {
		return exec.Command("")
	}
	if err := d.RunBefore(); err == nil {
		t.Fatal("DistantDeploy run before should return error")
	} else {
		if len(err.(distantRunError).Errors) != 2 {
			t.Fatal("DistantDeploy run before should return 2 errors")
		}
		if strings.Count(err.Error(), "\n") < 2 {
			t.Fatal("Error reporting should contains 2 \n at least")
		}
	}
}

func TestDistantImpl_Deploy(t *testing.T) {
	d, err := newDistantFromJSONTest(`{ "instances": [ {"beforeCmd": "/", "nbDistantToLaunch": 2} ] }`)
	if err != nil {
		t.Fatal("good local deployer json should not throw error")
	}
	d.(*arrayDistantImpl).cmdFabric = func(name string, arg ...string) *exec.Cmd {
		return exec.Command("ls")
	}
	if err := d.Deploy(); err != nil {
		t.Fatal("DistantDeploy deploy() should not return error")
	}
}

func TestDistantImpl_DeployError(t *testing.T) {
	d, err := newDistantFromJSONTest(`{ "instances": [ {"beforeCmd": "/", "nbDistantToLaunch": 2} ] }`)
	if err != nil {
		t.Fatalf("good local deployer json should not throw error : %s", err)
	}
	d.(*arrayDistantImpl).cmdFabric = func(name string, arg ...string) *exec.Cmd {
		return exec.Command("")
	}
	if err := d.Deploy(); err == nil {
		t.Fatal("DistantDeploy deploy() should return error")
	}
}

func TestDistantImpl_RunAfter(t *testing.T) {
	d, err := newDistantFromJSONTest(`{ "instances": [ {"afterCmd": "/", "nbDistantToLaunch": 1} ] }`)
	if err != nil {
		t.Fatalf("good local deployer json should not throw error %s", err)
	}
	hasBeenCalledWith := ""
	d.(*arrayDistantImpl).cmdFabric = func(name string, arg ...string) *exec.Cmd {
		hasBeenCalledWith = name
		return exec.Command("ls")
	}
	if err := d.RunAfter(); err != nil {
		t.Fatal("DistantDeploy run after should not return error")
	}
	if hasBeenCalledWith != "ssh" {
		t.Fatalf("After should call ssh cmd instead of %s", hasBeenCalledWith)
	}
}

func TestDistantImpl_RunAfterError(t *testing.T) {
	d, err := newDistantFromJSONTest(`{ "instances": [ {"afterCmd": "/", "nbDistantToLaunch": 2} ] }`)
	if err != nil {
		t.Fatal("good local deployer json should not throw error")
	}
	d.(*arrayDistantImpl).cmdFabric = func(name string, arg ...string) *exec.Cmd {
		return exec.Command("")
	}
	if err := d.RunAfter(); err == nil {
		t.Fatal("DistantDeploy run after should return error")
	} else {
		if len(err.(distantRunError).Errors) != 2 {
			t.Fatal("DistantDeploy run after should return 2 errors")
		}
		if strings.Count(err.Error(), "\n") < 2 {
			t.Fatal("Error reporting should contains 2 \n at least")
		}
	}
}
