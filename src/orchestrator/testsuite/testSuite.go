package testsuite

import (
	"encoding/json"
	"lorhammer/src/model"
	"lorhammer/src/orchestrator/checker"
	"lorhammer/src/orchestrator/deploy"
	"lorhammer/src/orchestrator/provisioning"
	"lorhammer/src/orchestrator/testtype"
	"time"

	"github.com/google/uuid"
)

//TestSuite describe a test to execute scenarios
type TestSuite struct {
	UUID                     string             `json:"uuid"`
	Test                     testtype.Test      `json:"test"`
	StopAllLorhammerTime     time.Duration      `json:"stopAllLorhammerTime"`
	SleepBeforeCheckTime     time.Duration      `json:"sleepBeforeCheckTime"`
	ShutdownAllLorhammerTime time.Duration      `json:"shutdownAllLorhammerTime"`
	SleepAtEndTime           time.Duration      `json:"sleepAtEndTime"`
	RequieredLorhammer       int                `json:"requieredLorhammer"`
	MaxWaitLorhammerTime     time.Duration      `json:"maxWaitLorhammerTime"`
	Init                     []model.Init       `json:"init"`
	Check                    checker.Model      `json:"check"`
	Provisioning             provisioning.Model `json:"provisioning"`
	Deploy                   deploy.Model       `json:"deploy"`
}

type jsonTestSuite struct {
	Test                     testtype.Test      `json:"test"`
	StopAllLorhammerTime     string             `json:"stopAllLorhammerTime"`
	SleepBeforeCheckTime     string             `json:"sleepBeforeCheckTime"`
	ShutdownAllLorhammerTime string             `json:"shutdownAllLorhammerTime"`
	SleepAtEndTime           string             `json:"sleepAtEndTime"`
	RequieredLorhammer       int                `json:"requieredLorhammer"`
	MaxWaitLorhammerTime     string             `json:"maxWaitLorhammerTime"`
	Init                     []model.Init       `json:"init"`
	Check                    checker.Model      `json:"check"`
	Provisioning             provisioning.Model `json:"provisioning"`
	Deploy                   deploy.Model       `json:"deploy"`
}

//FromFile build []testSuite from a json file
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
		maxWaitLorhammerTime, err := time.ParseDuration(test.MaxWaitLorhammerTime)
		if err != nil {
			return nil, err
		}
		res[i] = TestSuite{
			UUID:                     uuid.New().String(),
			Test:                     test.Test,
			StopAllLorhammerTime:     stopAllLorhammerTime,
			SleepBeforeCheckTime:     sleepBeforeCheckTime,
			ShutdownAllLorhammerTime: shutdownAllLorhammerTime,
			SleepAtEndTime:           sleepAtEndTime,
			RequieredLorhammer:       test.RequieredLorhammer,
			MaxWaitLorhammerTime:     maxWaitLorhammerTime,
			Init:                     test.Init,
			Check:                    test.Check,
			Provisioning:             test.Provisioning,
			Deploy:                   test.Deploy,
		}
	}
	return res, nil
}
