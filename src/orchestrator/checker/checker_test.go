package checker

import (
	"encoding/json"
	"errors"
	"lorhammer/src/orchestrator/metrics"
	"testing"
)

func TestGetFake(t *testing.T) {
	check, err := Get(Model{Type: "Fake"}, nil)
	if err == nil {
		t.Fatal("Fake type should return error")
	}
	if check != nil {
		t.Fatal("Fake type should not return checker")
	}
}

func TestGetNone(t *testing.T) {
	check, err := Get(Model{Type: noneType}, nil)
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

func (o other) Start() error              { return o.startError }
func (other) Check() ([]Success, []Error) { return nil, nil }

func TestOtherError(t *testing.T) {
	checkers[Type("other")] = func(config json.RawMessage, prometheus metrics.Prometheus) (Checker, error) {
		return nil, errors.New("error")
	}
	check, err := Get(Model{Type: "other"}, nil)
	if err == nil {
		t.Fatal("other type should return error")
	}
	if check != nil {
		t.Fatal("other type should not return checker")
	}
}

func TestOtherStartError(t *testing.T) {
	checkers[Type("other")] = func(config json.RawMessage, prometheus metrics.Prometheus) (Checker, error) {
		return other{startError: errors.New("error")}, nil
	}
	check, err := Get(Model{Type: "other"}, nil)
	if err == nil {
		t.Fatal("other type should return error on start")
	}
	if check != nil {
		t.Fatal("other type should not return checker")
	}
}
