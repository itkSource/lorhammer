package testtype

import (
	"encoding/json"
	"fmt"
	"lorhammer/src/model"
	"lorhammer/src/tools"
	"time"
)

//Type represent a testType
type Type string

//Test is the representation of a test in config file
type Test struct {
	testType   Type
	repeatTime time.Duration
}

type testJSON struct {
	TestType   Type   `json:"type"`
	RepeatTime string `json:"repeatTime"`
}

//UnmarshalJSON permit to load a Test from a json
func (test *Test) UnmarshalJSON(b []byte) error {
	var serialized testJSON
	if err := json.Unmarshal(b, &serialized); err != nil {
		return err
	}

	test.testType = serialized.TestType

	var err error
	if test.repeatTime, err = time.ParseDuration(serialized.RepeatTime); err != nil {
		return err
	}

	return nil
}

var testers = make(map[Type]func(test Test, init []model.Init, mqttClient tools.Mqtt))

func init() {
	testers[typeNone] = startNone
	testers[typeOneShot] = startOneShot
	testers[typeRepeat] = startRepeat
}

//Start launch a test
func Start(test Test, init []model.Init, mqttClient tools.Mqtt) error {
	if tester := testers[test.testType]; tester != nil {
		go tester(test, init, mqttClient)
		return nil
	}
	return fmt.Errorf("Unknown test type %s", test.testType)
}
