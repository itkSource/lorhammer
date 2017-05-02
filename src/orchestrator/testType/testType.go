package testType

import (
	"encoding/json"
	"fmt"
	"lorhammer/src/model"
	"lorhammer/src/tools"
	"time"
)

type Type string

type Test struct {
	testType   Type
	repeatTime time.Duration
	rampTime   time.Duration
}

type testJson struct {
	TestType   Type   `json:"type"`
	RepeatTime string `json:"repeatTime"`
	RampTime   string `json:"rampTime"`
}

func (test *Test) UnmarshalJSON(b []byte) error {
	var serialized testJson
	if err := json.Unmarshal(b, &serialized); err != nil {
		return err
	}

	test.testType = serialized.TestType

	var err error
	if test.rampTime, err = time.ParseDuration(serialized.RampTime); err != nil {
		return err
	}
	if test.repeatTime, err = time.ParseDuration(serialized.RepeatTime); err != nil {
		return err
	}

	return nil
}

var testers = make(map[Type]func(test Test, init model.Init, mqttClient tools.Mqtt))

func init() {
	testers[TypeNone] = startNone
	testers[TypeOneShot] = startOneShot
	testers[TypeRepeat] = startRepeat
	testers[TypeRamp] = startRamp
}

func Start(test Test, init model.Init, mqttClient tools.Mqtt) error {
	if tester := testers[test.testType]; tester != nil {
		go tester(test, init, mqttClient)
		return nil
	}
	return fmt.Errorf("Unknown test type %s", test.testType)
}
