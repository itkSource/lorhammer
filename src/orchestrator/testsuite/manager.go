package testsuite

import (
	"errors"
	"lorhammer/src/orchestrator/checker"
	"lorhammer/src/orchestrator/command"
	"lorhammer/src/orchestrator/deploy"
	"lorhammer/src/orchestrator/provisioning"
	"lorhammer/src/orchestrator/testtype"
	"lorhammer/src/tools"
	"time"

	"github.com/sirupsen/logrus"
	"lorhammer/src/orchestrator/metrics"
)

var loggerManager = logrus.WithField("logger", "orchestrator/testsuite/manager")

//LaunchTest manage life cycle of a test (start, stop, check, report...)
func (test *TestSuite) LaunchTest(mqttClient tools.Mqtt, prometheus metrics.Prometheus) (*TestReport, error) {
	check, err := checker.Get(test.Check, prometheus) //build checker here because no need to start test if checker is bad configured
	if err != nil {
		loggerManager.WithError(err).Error("Error to get checker")
		return nil, err
	}

	if err := deploy.Start(test.Deploy, mqttClient); err != nil {
		loggerManager.WithError(err).Error("Error to deploy")
		return nil, err
	}
	startDate := time.Now()

	//wait until all required lorhammers are here
	for {
		if command.NbLorhammer() >= test.RequieredLorhammer {
			break
		}
		if time.Now().Sub(startDate) > test.MaxWaitLorhammerTime {
			loggerManager.WithField("MaxWaitLorhammerTime", test.MaxWaitLorhammerTime).Error("No requiered lorhammer after time")
			return nil, errors.New("no required lorhammer")
		}
		time.Sleep(100 * time.Millisecond)
	}

	if err := testtype.Start(test.Test, test.Init, mqttClient); err != nil {
		loggerManager.WithError(err).Error("Error to start test")
		return nil, err
	}

	// wait until stop (0 or negative value means no stop)
	time.Sleep(test.StopAllLorhammerTime)
	if test.StopAllLorhammerTime > 0 {
		command.StopScenario(mqttClient)
	}

	//wait until check minus time we have already passed in stop
	time.Sleep(test.SleepBeforeCheckTime - test.StopAllLorhammerTime)
	success, errs := checkResults(check)

	//wait until shutdown minus time we have already passed in stop and check (0 or negative value means no shutdown)
	time.Sleep(test.ShutdownAllLorhammerTime - (test.StopAllLorhammerTime + test.SleepBeforeCheckTime))

	if test.StopAllLorhammerTime > 0 || test.ShutdownAllLorhammerTime > 0 {
		if err := provisioning.DeProvision(test.UUID); err != nil {
			loggerManager.WithError(err).Error("Couldn't unprovision")
			return nil, err
		}
	}

	if test.ShutdownAllLorhammerTime > 0 {
		command.ShutdownLorhammers(mqttClient)
	}
	endDate := time.Now()

	return &TestReport{
		StartDate:     startDate,
		EndDate:       endDate,
		Input:         test,
		ChecksSuccess: success,
		ChecksError:   errs,
	}, nil
}

func checkResults(check checker.Checker) ([]checker.Success, []checker.Error) {
	ok, errs := check.Check()
	if len(errs) > 0 {
		loggerManager.WithField("nb", len(errs)).Error("Check results errors")
		for _, err := range errs {
			loggerManager.WithFields(logrus.Fields(err.Details())).Error("Check result error")
		}
	}

	if len(ok) > 0 {
		loggerManager.WithField("nb", len(ok)).Info("Check results good")
		for _, o := range ok {
			loggerManager.WithFields(logrus.Fields(o.Details())).Info("Check result good")
		}
	}

	return ok, errs
}
