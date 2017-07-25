package testSuite

import (
	"github.com/Sirupsen/logrus"
	"lorhammer/src/orchestrator/checker"
	"lorhammer/src/orchestrator/command"
	"lorhammer/src/orchestrator/deploy"
	"lorhammer/src/orchestrator/provisioning"
	"lorhammer/src/orchestrator/testType"
	"lorhammer/src/tools"
	"time"
)

var LOG = logrus.WithField("logger", "orchestrator/testSuite/test")

func (test *TestSuite) LaunchTest(consulClient tools.Consul, mqttClient tools.Mqtt, grafanaClient tools.GrafanaClient) (*TestReport, error) {
	check, err := checker.Get(consulClient, test.Check) //build checker here because no need to start test if checker is bad configured
	if err != nil {
		LOG.WithError(err).Error("Error to get checker")
		return nil, err
	}

	if err := deploy.Start(test.Deploy, consulClient); err != nil {
		LOG.WithError(err).Error("Error to deploy")
		return nil, err
	}
	startDate := time.Now()

	if err := testType.Start(test.Test, test.Init, mqttClient); err != nil {
		LOG.WithError(err).Error("Error to start test")
		return nil, err
	}

	// wait until stop (0 or negative value means no stop)
	time.Sleep(test.StopAllLorhammerTime)
	if test.StopAllLorhammerTime > 0 {
		command.StopScenario(mqttClient)
	}

	//wait until check minus time we have already passed in stop
	time.Sleep(test.SleepBeforeCheckTime - test.StopAllLorhammerTime)
	success, errors := checkResults(check)

	//wait until shutdown minus time we have already passed in stop and check (0 or negative value means no shutdown)
	time.Sleep(test.ShutdownAllLorhammerTime - (test.StopAllLorhammerTime + test.SleepBeforeCheckTime))

	if test.StopAllLorhammerTime > 0 || test.ShutdownAllLorhammerTime > 0 {
		if err := provisioning.DeProvision(test.Uuid); err != nil {
			LOG.WithError(err).Error("Couldn't unprovision")
			return nil, err
		}
	}

	if test.ShutdownAllLorhammerTime > 0 {
		command.ShutdownLorhammers(mqttClient)
	}
	endDate := time.Now()
	var snapshotUrl = ""
	// TODO add time for grafana snapshot (idem stop and shutdown)
	if grafanaClient != nil {
		var err error = nil
		snapshotUrl, err = grafanaClient.MakeSnapshot(startDate, endDate)
		if err != nil {
			LOG.WithError(err).Error("Can't snapshot grafana")
		}
	}
	return &TestReport{
		StartDate:          startDate,
		EndDate:            endDate,
		Input:              test,
		ChecksSuccess:      success,
		ChecksError:        errors,
		GrafanaSnapshotUrl: snapshotUrl,
	}, nil
}

func checkResults(check checker.Checker) ([]checker.CheckerSuccess, []checker.CheckerError) {
	ok, errs := check.Check()
	if len(errs) > 0 {
		LOG.WithField("nb", len(errs)).Error("Check results errors")
		for _, err := range errs {
			LOG.WithFields(logrus.Fields(err.Details())).Error("Check result error")
		}
	}

	if len(ok) > 0 {
		LOG.WithField("nb", len(ok)).Info("Check results good")
		for _, o := range ok {
			LOG.WithFields(logrus.Fields(o.Details())).Info("Check result good")
		}
	}

	return ok, errs
}
