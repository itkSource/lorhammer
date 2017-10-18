package provisioning

import (
	"encoding/json"
	"errors"
	"lorhammer/src/model"
	"testing"
)

func TestFake(t *testing.T) {
	m := Model{
		Type: Type("fake"),
	}

	if err := Provision("1", m, model.Register{}); err == nil {
		t.Fatal("Type fake should return an error")
	}

	if err := DeProvision("1"); err == nil {
		t.Fatal("Type fake should return an error")
	}
}

func TestNone(t *testing.T) {
	m := Model{
		Type: noneType,
	}

	if err := Provision("2", m, model.Register{}); err != nil {
		t.Log(err)
		t.Fatal("Type none should not return an error")
	}

	if err := DeProvision("2"); err != nil {
		t.Log(err)
		t.Fatal("Type none should not return an error")
	}
}

func TestOther(t *testing.T) {
	m := Model{
		Type: "other",
	}

	provisioners["other"] = func(conf json.RawMessage) (provisioner, error) {
		return nil, errors.New("Bad unmarshal config")
	}

	if err := Provision("3", m, model.Register{}); err == nil {
		t.Fatal("Type other should return an error")
	}

	if err := DeProvision("3"); err == nil {
		t.Fatal("Type other should return an error")
	}
}

func TestGoRoutineSafe(t *testing.T) {
	m := Model{
		Type: noneType,
	}

	if err := Provision("4", m, model.Register{}); err != nil {
		t.Log(err)
		t.Fatal("Type none should not return an error")
	}
	firstInstance, _ := instances.Load("4")
	//reuse same instance to provision again
	if err := Provision("4", m, model.Register{}); err != nil {
		t.Log(err)
		t.Fatal("Type none should not return an error")
	}
	if secondInstance, _ := instances.Load("4"); firstInstance != secondInstance {
		t.Fatal("Provision should reuse same instance before deProvisioning")
	}

	finish := make(chan error, 1)
	defer close(finish)

	go func() {
		if err := DeProvision("4"); err != nil {
			finish <- err
		} else {
			finish <- nil
		}
	}()

	if err := <-finish; err != nil {
		t.Log(err)
		t.Fatal("Type none should not return an error")
	}
}

func TestCleanAfterDeProvision(t *testing.T) {
	m := Model{
		Type: noneType,
	}

	if err := Provision("5", m, model.Register{}); err != nil {
		t.Log(err)
		t.Fatal("Type none should not return an error")
	}

	if err := DeProvision("5"); err != nil {
		t.Log(err)
		t.Fatal("Type none should not return an error")
	}

	if _, ok := instances.Load("5"); ok {
		t.Fatal("After DeProvision instances should be cleaned")
	}
}
