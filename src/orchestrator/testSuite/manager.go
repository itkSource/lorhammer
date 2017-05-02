package testSuite

import (
	"github.com/Sirupsen/logrus"
	"lorhammer/src/orchestrator/checker"
	"lorhammer/src/orchestrator/command"
	"lorhammer/src/orchestrator/deploy"
	"lorhammer/src/orchestrator/prometheus"
	"lorhammer/src/orchestrator/provisioning"
	"lorhammer/src/orchestrator/testType"
	"lorhammer/src/tools"
	"os"
	"time"
)

var LOG = logrus.WithField("logger", "orchestrator/testSuite/test")

func LaunchTest(consulClient tools.Consul, mqttClient tools.Mqtt, test *TestSuite, prometheusApiClient prometheus.ApiClient, grafanaClient *tools.GrafanaClient) (*TestReport, error) {
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

		if err := provisioning.DeProvision(test.Uuid); err != nil {
			LOG.WithError(err).Error("Couldn't unprovision")
			return nil, err
		}
	}

	//wait until shutdown minus time we have already passed in stop (0 or negative value means no shutdown)
	time.Sleep(test.ShutdownAllLorhammerTime - test.StopAllLorhammerTime)
	success, errors := check(prometheusApiClient, test)
	if test.ShutdownAllLorhammerTime > 0 {
		command.ShutdownLorhammers(mqttClient)

		if err := provisioning.DeProvision(test.Uuid); err != nil {
			LOG.WithError(err).Error("Couldn't unprovision")
			return nil, err
		}

		go func() {
			//TODO add boolean in scenario to choose if we want kill also orchestrator (needed for ci)
			time.Sleep(100 * time.Millisecond)
			os.Exit(len(errors))
		}()
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

func check(prometheusApiClient prometheus.ApiClient, test *TestSuite) ([]checker.PrometheusCheckOk, []checker.PrometheusCheckError) {
	if len(test.PrometheusCheck) > 0 {
		ok, errs := checker.Check(prometheusApiClient, test.PrometheusCheck)

		if len(errs) > 0 {
			LOG.WithField("nb", len(errs)).Error("Check results errors")
			for _, err := range errs {
				LOG.WithFields(logrus.Fields{
					"query":  err.Query,
					"reason": err.Reason,
					"val":    err.Val,
				}).Error("Check result error")
			}
		}

		if len(ok) > 0 {
			LOG.WithField("nb", len(ok)).Info("Check results good")
			for _, o := range ok {
				LOG.WithFields(logrus.Fields{
					"query": o.Query,
					"val":   o.Val,
				}).Info("Check result good")
			}
		}

		return ok, errs
	}
	return make([]checker.PrometheusCheckOk, 0), make([]checker.PrometheusCheckError, 0)
}
