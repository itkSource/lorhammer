package testsuite

import (
	"fmt"
	"testing"
)

type testFromFile struct {
	description          string
	testValid            bool
	test                 string
	repeatTime           string
	stopAll              string
	beforeCheck          string
	shutdownAll          string
	sleep                string
	requieredLorhammer   int
	maxWaitLorhammerTime string
	init                 string
	provisioning         string
	check                string
	deploy               string
}

var testsFromFile = []testFromFile{
	{
		testValid:            true,
		description:          "First simple valid test",
		test:                 `{"type": "oneShot", "repeatTime": "0"}`,
		stopAll:              "0",
		beforeCheck:          "0",
		shutdownAll:          "0",
		sleep:                "0",
		requieredLorhammer:   0,
		maxWaitLorhammerTime: "0",
		init:                 `[{"nsAddress": "127.0.0.1:1700","nbGateway": 1,"nbNodePerGateway": [1, 1],"sleepTime": [100, 500]}]`,
		provisioning:         `{"type": "none"}`,
		check:                `{"type": "none"}`,
		deploy:               `{"type": "none"}`,
	},
	{
		testValid:            true,
		description:          "Test with good repeatTime",
		test:                 `{"type": "repeat", "repeatTime": "1m"}`,
		stopAll:              "0",
		beforeCheck:          "0",
		shutdownAll:          "0",
		sleep:                "0",
		requieredLorhammer:   0,
		maxWaitLorhammerTime: "0",
		init:                 `[{"nsAddress": "127.0.0.1:1700","nbGateway": 1,"nbNodePerGateway": [1, 1],"sleepTime": [100, 500]}]`,
		provisioning:         `{"type": "none"}`,
		check:                `{"type": "none"}`,
		deploy:               `{"type": "none"}`,
	},
	{
		testValid:            false,
		description:          "Invalid test because empty testType",
		test:                 "",
		stopAll:              "0",
		beforeCheck:          "0",
		shutdownAll:          "0",
		sleep:                "0",
		requieredLorhammer:   0,
		maxWaitLorhammerTime: "0",
		init:                 `[{"nsAddress": "127.0.0.1:1700","nbGateway": 1,"nbNodePerGateway": [1, 1],"sleepTime": [100, 500]}]`,
		provisioning:         `{"type": "none"}`,
		check:                `{"type": "none"}`,
		deploy:               `{"type": "none"}`,
	},
	{
		testValid:            false,
		description:          "Invalid test because invalid repeatTime",
		test:                 `{"type": "repeat", "repeatTime": "not good"}`,
		stopAll:              "0",
		beforeCheck:          "0",
		shutdownAll:          "0",
		sleep:                "0",
		requieredLorhammer:   0,
		maxWaitLorhammerTime: "0",
		init:                 `[{"nsAddress": "127.0.0.1:1700","nbGateway": 1,"nbNodePerGateway": [1, 1],"sleepTime": [100, 500]}]`,
		provisioning:         `{"type": "none"}`,
		check:                `{"type": "none"}`,
		deploy:               `{"type": "none"}`,
	},
	{
		testValid:            false,
		description:          "Invalid test because invalid stopAllTime",
		test:                 `{"type": "repeat", "repeatTime": "0"}`,
		stopAll:              "a",
		beforeCheck:          "0",
		shutdownAll:          "0",
		sleep:                "0",
		requieredLorhammer:   0,
		maxWaitLorhammerTime: "0",
		init:                 `[{"nsAddress": "127.0.0.1:1700","nbGateway": 1,"nbNodePerGateway": [1, 1],"sleepTime": [100, 500]}]`,
		provisioning:         `{"type": "none"}`,
		check:                `{"type": "none"}`,
		deploy:               `{"type": "none"}`,
	},
	{
		testValid:            false,
		description:          "Invalid test because invalid shutdownAllTime",
		test:                 `{"type": "repeat", "repeatTime": "0"}`,
		stopAll:              "0",
		beforeCheck:          "0",
		shutdownAll:          "a",
		sleep:                "0",
		requieredLorhammer:   0,
		maxWaitLorhammerTime: "0",
		init:                 `[{"nsAddress": "127.0.0.1:1700","nbGateway": 1,"nbNodePerGateway": [1, 1],"sleepTime": [100, 500]}]`,
		provisioning:         `{"type": "none"}`,
		check:                `{"type": "none"}`,
		deploy:               `{"type": "none"}`,
	},
	{
		testValid:            false,
		description:          "Invalid test because invalid sleepAtEndTime",
		test:                 `{"type": "repeat", "repeatTime": "0"}`,
		stopAll:              "0",
		beforeCheck:          "0",
		shutdownAll:          "0",
		sleep:                "a",
		requieredLorhammer:   0,
		maxWaitLorhammerTime: "0",
		init:                 `[{"nsAddress": "127.0.0.1:1700","nbGateway": 1,"nbNodePerGateway": [1, 1],"sleepTime": [100, 500]}]`,
		provisioning:         `{"type": "none"}`,
		check:                `{"type": "none"}`,
		deploy:               `{"type": "none"}`,
	},
	{
		testValid:            false,
		description:          "Invalid test because invalid SleepBeforeCheckTime",
		test:                 `{"type": "repeat", "repeatTime": "0"}`,
		stopAll:              "0",
		beforeCheck:          "a",
		shutdownAll:          "0",
		sleep:                "0",
		requieredLorhammer:   0,
		maxWaitLorhammerTime: "0",
		init:                 `[{"nsAddress": "127.0.0.1:1700","nbGateway": 1,"nbNodePerGateway": [1, 1],"sleepTime": [100, 500]}]`,
		provisioning:         `{"type": "none"}`,
		check:                `{"type": "none"}`,
		deploy:               `{"type": "none"}`,
	},
	{
		testValid:            false,
		description:          "Invalid test because invalid maxWaitLorhammerTime",
		test:                 `{"type": "repeat", "repeatTime": "0"}`,
		stopAll:              "0",
		beforeCheck:          "0",
		shutdownAll:          "0",
		sleep:                "0",
		requieredLorhammer:   0,
		maxWaitLorhammerTime: "a",
		init:                 `[{"nsAddress": "127.0.0.1:1700","nbGateway": 1,"nbNodePerGateway": [1, 1],"sleepTime": [100, 500]}]`,
		provisioning:         `{"type": "none"}`,
		check:                `{"type": "none"}`,
		deploy:               `{"type": "none"}`,
	},
}

var templateFromFile = `[{"test": %s,"repeatTime": "%s","stopAllLorhammerTime": "%s","sleepBeforeCheckTime": "%s","shutdownAllLorhammerTime": "%s","sleepAtEndTime": "%s","requieredLorhammer":%d,"maxWaitLorhammerTime":"%s","init": %s,"provisioning": %s,"check": %s, "deploy": %s}]`

func TestTransformFile(t *testing.T) {
	for _, ct := range testsFromFile {
		data := []byte(fmt.Sprintf(templateFromFile, ct.test, ct.repeatTime, ct.stopAll, ct.beforeCheck, ct.shutdownAll, ct.sleep, ct.requieredLorhammer, ct.maxWaitLorhammerTime, ct.init, ct.provisioning, ct.check, ct.deploy))
		tests, err := FromFile(data)

		if ct.testValid {
			if err != nil {
				t.Fatalf(`valid scenario should not return err %s for : "%s"`, err, ct.description)
			}
			if len(tests) != 1 {
				t.Fatalf(`1 valid scenario should return 1 valid testSuite for : "%s"`, ct.description)
			}
		} else {
			if err == nil {
				t.Fatalf(`invalid scenario must return an err for : "%s"`, ct.description)
			}
			if tests != nil {
				t.Fatalf(`invalid scenario must return nil as tests for : "%s"`, ct.description)
			}
		}
	}
}
