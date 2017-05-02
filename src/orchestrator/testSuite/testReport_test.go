package testSuite

import (
	"github.com/google/uuid"
	"io/ioutil"
	"regexp"
	"testing"
	"time"
)

func buildReport(filename string, t *testing.T) {
	err := WriteFile(&TestReport{
		StartDate:          time.Now(),
		EndDate:            time.Now().Add(1 * time.Minute),
		Input:              nil,
		ChecksSuccess:      nil,
		ChecksError:        nil,
		GrafanaSnapshotUrl: "",
	}, filename)

	if err != nil {
		t.Fatalf("Good test report should not give err %s", err)
	}
}

func readAndCleanData(filename string, t *testing.T) []byte {
	data, err := ioutil.ReadFile(filename)

	if err != nil {
		t.Fatalf("Error reading report file : %s", err)
	}

	re := regexp.MustCompile(`\r?\n`)
	input := re.ReplaceAll(data, []byte(""))
	re = regexp.MustCompile(`\s`)
	input = re.ReplaceAll(data, []byte(""))

	return input
}

func TestCreatingReport(t *testing.T) {
	filename := "/tmp/" + uuid.New().String() + "-report_test_lorhammer.json"
	buildReport(filename, t)
	input := readAndCleanData(filename, t)

	var validRe = regexp.MustCompile(`^\{"startDate":"[^\"]+","endDate":"[^\"]+","input":null,"checksSuccess":null,"checksError":null,"grafanaSnapshotUrl":""\}$`)

	if !validRe.Match(input) {
		t.Log(string(input))
		t.Fatal("Bad formatted report")
	}
}

func TestMultipleEntry(t *testing.T) {
	filename := "/tmp/" + uuid.New().String() + "-report_test_lorhammer.json"
	buildReport(filename, t)
	buildReport(filename, t)
	input := readAndCleanData(filename, t)

	var validRe = regexp.MustCompile(`^(:?\{"startDate":"[^\"]+","endDate":"[^\"]+","input":null,"checksSuccess":null,"checksError":null,"grafanaSnapshotUrl":""\}){2}$`)

	if !validRe.Match(input) {
		t.Log(string(input))
		t.Fatal("Bad formatted report")
	}
}
