package testSuite

import (
	"encoding/json"
	"lorhammer/src/orchestrator/checker"
	"os"
	"time"
)

type TestReport struct {
	StartDate          time.Time         `json:"startDate"`
	EndDate            time.Time         `json:"endDate"`
	Input              *TestSuite        `json:"input"`
	ChecksSuccess      []checker.Success `json:"checksSuccess"`
	ChecksError        []checker.Error   `json:"checksError"`
	GrafanaSnapshotUrl string            `json:"grafanaSnapshotUrl"`
}

func WriteFile(testReport *TestReport, pathReportFile string) error {
	serialized, err := json.MarshalIndent(testReport, "", "    ")
	if err != nil {
		return err
	}
	createFileIfNotExist(pathReportFile)
	f, err := os.OpenFile(pathReportFile, os.O_APPEND|os.O_WRONLY, 0600)
	if err != nil {
		return err
	}
	defer f.Close()

	if _, err = f.Write(serialized); err != nil {
		return err
	}
	return nil
}

func createFileIfNotExist(path string) error {
	// detect if file exists
	var _, err = os.Stat(path)

	// create file if not exists
	if os.IsNotExist(err) {
		var file, err = os.Create(path)
		file.Close()
		return err
	}
	return nil
}
