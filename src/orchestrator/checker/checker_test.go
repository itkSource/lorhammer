package checker

import (
	"encoding/json"
	"errors"
	"lorhammer/src/tools"
	"testing"
)

func TestGetFake(t *testing.T) {
	check, err := Get(nil, Model{Type: "Fake"})
	if err == nil {
		t.Fatal("Fake type should return error")
	}
	if check != nil {
		t.Fatal("Fake type should not return checker")
	}
}

func TestGetNone(t *testing.T) {
	check, err := Get(nil, Model{Type: NoneType})
	if err != nil {
		t.Fatal("None type should not return error")
	}
	if check == nil {
		t.Fatal("None type should return valid checker")
	}
	if success, err := check.Check(); len(err) > 0 {
		t.Fatal("None checker should not return err")
	} else if len(success) > 0 {
		t.Fatal("None checker should not return success")
	}
}

type other struct {
	startError error
}

func (o other) Start() error                              { return o.startError }
func (_ other) Check() ([]CheckerSuccess, []CheckerError) { return nil, nil }

func TestOtherError(t *testing.T) {
	checkers[Type("other")] = func(consulClient tools.Consul, config json.RawMessage) (Checker, error) {
		return nil, errors.New("error")
	}
	check, err := Get(nil, Model{Type: "other"})
	if err == nil {
		t.Fatal("other type should return error")
	}
	if check != nil {
		t.Fatal("other type should not return checker")
	}
}

func TestOtherStartError(t *testing.T) {
	checkers[Type("other")] = func(consulClient tools.Consul, config json.RawMessage) (Checker, error) {
		return other{startError: errors.New("error")}, nil
	}
	check, err := Get(nil, Model{Type: "other"})
	if err == nil {
		t.Fatal("other type should return error on start")
	}
	if check != nil {
		t.Fatal("other type should not return checker")
	}
}
