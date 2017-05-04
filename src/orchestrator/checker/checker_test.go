package checker

import (
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
