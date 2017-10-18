package deploy

import (
	"encoding/json"
	"errors"
	"lorhammer/src/tools"
	"testing"
)

func TestModel_UnmarshalJSON(t *testing.T) {
	json := []byte(`{"type": "none"}`)
	m := Model{}
	err := m.UnmarshalJSON(json)
	if err != nil {
		t.Fatal("unmarshal none should not throw error")
	}
	if m.Type != typeNone {
		t.Fatal("json none deploy type should return none type")
	}
}

func TestModel_UnmarshalJSONWithDuration(t *testing.T) {
	json := []byte(`{"type": "none", "sleepAfterDeployTime": "10s"}`)
	m := Model{}
	err := m.UnmarshalJSON(json)
	if err != nil {
		t.Fatal("unmarshal none should not throw error")
	}
	if m.Type != typeNone {
		t.Fatal("json none deploy type should return none type")
	}
	if m.SleepAfterDeployTime.Seconds() != 10 {
		t.Fatal("duration 20seconds in json should return 20 second in go")
	}
}

func TestModel_UnmarshalJSONError(t *testing.T) {
	json := []byte(`{`)
	m := Model{}
	err := m.UnmarshalJSON(json)
	if err == nil {
		t.Fatal("unmarshal bad json should throw error")
	}
}

func TestModel_UnmarshalJSONErrorDuration(t *testing.T) {
	json := []byte(`{"type": "none", "sleepAfterDeployTime": "sc"}`)
	m := Model{}
	err := m.UnmarshalJSON(json)
	if err == nil {
		t.Fatal("unmarshal bad json duration should throw error")
	}
}

func TestFake(t *testing.T) {
	m := Model{
		Type:                 Type("fake"),
		SleepAfterDeployTime: 0,
	}

	err := Start(m, nil)

	if err == nil {
		t.Fatal("Fake deployer should throw error")
	}
}

func TestNone(t *testing.T) {
	m := Model{
		Type:                 Type("none"),
		SleepAfterDeployTime: 0,
	}

	err := Start(m, nil)

	if err != nil {
		t.Fatal("None deployer should not throw error")
	}
}

type otherDeploy struct {
	errBefore error
	errDeploy error
	errAfter  error
}

func (o otherDeploy) RunBefore() error { return o.errBefore }
func (o otherDeploy) Deploy() error    { return o.errDeploy }
func (o otherDeploy) RunAfter() error  { return o.errAfter }

func TestOther(t *testing.T) {
	m := Model{
		Type:                 Type("other"),
		SleepAfterDeployTime: 0,
	}

	other := func(json.RawMessage, tools.Consul) (deployer, error) {
		return otherDeploy{}, nil
	}

	deployers["other"] = other

	err := Start(m, nil)

	if err != nil {
		t.Fatal("Other deployer should not throw error")
	}
}

func TestOtherErr(t *testing.T) {
	m := Model{
		Type:                 Type("other"),
		SleepAfterDeployTime: 0,
	}

	other := func(json.RawMessage, tools.Consul) (deployer, error) {
		return nil, errors.New("error creating func")
	}

	deployers["other"] = other

	err := Start(m, nil)

	if err == nil {
		t.Fatal("Other deployer should throw before error")
	}
}

func TestOtherErrBefore(t *testing.T) {
	m := Model{
		Type:                 Type("other"),
		SleepAfterDeployTime: 0,
	}

	other := func(json.RawMessage, tools.Consul) (deployer, error) {
		return otherDeploy{errBefore: errors.New("before")}, nil
	}

	deployers["other"] = other

	err := Start(m, nil)

	if err == nil {
		t.Fatal("Other deployer should throw before error")
	}
}

func TestOtherErrDeploy(t *testing.T) {
	m := Model{
		Type:                 Type("other"),
		SleepAfterDeployTime: 0,
	}

	other := func(json.RawMessage, tools.Consul) (deployer, error) {
		return otherDeploy{errDeploy: errors.New("deploy")}, nil
	}

	deployers["other"] = other

	err := Start(m, nil)

	if err == nil {
		t.Fatal("Other deployer should throw deploy error")
	}
}

func TestOtherErrAfter(t *testing.T) {
	m := Model{
		Type:                 Type("other"),
		SleepAfterDeployTime: 0,
	}

	other := func(json.RawMessage, tools.Consul) (deployer, error) {
		return otherDeploy{errAfter: errors.New("after")}, nil
	}

	deployers["other"] = other

	err := Start(m, nil)

	if err == nil {
		t.Fatal("Other deployer should throw after error")
	}
}
