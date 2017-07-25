package testSuite

import (
	"encoding/json"
	"github.com/google/uuid"
	"lorhammer/src/model"
	"lorhammer/src/orchestrator/checker"
	"lorhammer/src/orchestrator/deploy"
	"lorhammer/src/orchestrator/provisioning"
	"lorhammer/src/orchestrator/testType"
	"time"
)

// Describe a test to execute scenarios
type TestSuite struct {
	Uuid                     string             `json:"uuid"`
	Test                     testType.Test      `json:"test"`
	StopAllLorhammerTime     time.Duration      `json:"stopAllLorhammerTime"`
	SleepBeforeCheckTime     time.Duration      `json:"sleepBeforeCheckTime"`
	ShutdownAllLorhammerTime time.Duration      `json:"shutdownAllLorhammerTime"`
	SleepAtEndTime           time.Duration      `json:"sleepAtEndTime"`
	Init                     model.Init         `json:"init"`
	Check                    checker.Model      `json:"check"`
	Provisioning             provisioning.Model `json:"provisioning"`
	Deploy                   deploy.Model       `json:"deploy"`
}

type jsonTestSuite struct {
	Test                     testType.Test      `json:"test"`
	StopAllLorhammerTime     string             `json:"stopAllLorhammerTime"`
	SleepBeforeCheckTime     string             `json:"sleepBeforeCheckTime"`
	ShutdownAllLorhammerTime string             `json:"shutdownAllLorhammerTime"`
	SleepAtEndTime           string             `json:"sleepAtEndTime"`
	Init                     model.Init         `json:"init"`
	Check                    checker.Model      `json:"check"`
	Provisioning             provisioning.Model `json:"provisioning"`
	Deploy                   deploy.Model       `json:"deploy"`
}

func FromFile(configFile []byte) ([]TestSuite, error) {
	var tests = make([]jsonTestSuite, 0)
	if err := json.Unmarshal(configFile, &tests); err != nil {
		return nil, err
	}
	var res = make([]TestSuite, len(tests))
	for i, test := range tests {
		stopAllLorhammerTime, err := time.ParseDuration(test.StopAllLorhammerTime)
		if err != nil {
			return nil, err
		}
		sleepBeforeCheckTime, err := time.ParseDuration(test.SleepBeforeCheckTime)
		if err != nil {
			return nil, err
		}
		shutdownAllLorhammerTime, err := time.ParseDuration(test.ShutdownAllLorhammerTime)
		if err != nil {
			return nil, err
		}
		sleepAtEndTime, err := time.ParseDuration(test.SleepAtEndTime)
		if err != nil {
			return nil, err
		}
		res[i] = TestSuite{
			Uuid:                     uuid.New().String(),
			Test:                     test.Test,
			StopAllLorhammerTime:     stopAllLorhammerTime,
			SleepBeforeCheckTime:     sleepBeforeCheckTime,
			ShutdownAllLorhammerTime: shutdownAllLorhammerTime,
			SleepAtEndTime:           sleepAtEndTime,
			Init:                     test.Init,
			Check:                    test.Check,
			Provisioning:             test.Provisioning,
			Deploy:                   test.Deploy,
		}
	}
	return res, nil
}
