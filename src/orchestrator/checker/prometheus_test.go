package checker

import (
	"encoding/json"
	"errors"
	"lorhammer/src/orchestrator/prometheus"
	"testing"
)

var nbCall int

type prometheusAPIClientFake struct {
	returnFloat []float64
	returnError error
}

func (p prometheusAPIClientFake) ExecQuery(query string) (float64, error) {
	nbCall++
	return p.returnFloat[nbCall-1], p.returnError
}

func initPrometheusClientAPI(res []float64, err error) prometheus.APIClient {
	nbCall = 0
	return prometheusAPIClientFake{
		returnFloat: res,
		returnError: err,
	}
}

func TestNewPrometheus(t *testing.T) {
	check, err := newPrometheus(json.RawMessage([]byte(`{}`)), nil)
	if err != nil {
		t.Fatal("Valid config for prometheus should not return error")
	}
	if check == nil {
		t.Fatal("Valid config for prometheus should return checker")
	}
}

func TestNewPrometheusError(t *testing.T) {
	check, err := newPrometheus(json.RawMessage([]byte(`{`)), nil)
	if err == nil {
		t.Fatal("Invalid config for prometheus should return error")
	}
	if check != nil {
		t.Fatal("Invalid config for prometheus should not return checker")
	}
}

func TestNewPrometheusErrorConsul(t *testing.T) {
	check, _ := newPrometheus(json.RawMessage([]byte(`{"Address":""}`)), nil)
	check.(*prometheusChecker).prometheusClientFactory = func(string) (prometheus.APIClient, error) {
		return nil, errors.New("error fake prometheus client")
	}
	if err := check.Start(); err == nil {
		t.Fatal("Prometheus checker should return error on start when address consul is bad")
	}
}

func TestPrometheusChecker_Start(t *testing.T) {
	check, _ := newPrometheus(json.RawMessage([]byte(`{"Address":"http://127.0.0.1:9090"}`)), nil)
	if err := check.Start(); err != nil {
		t.Fatal("Prometheus checker should not return error on start")
	}
}

func TestReturnNothingIfNoTest(t *testing.T) {
	prometheusAPIClient := initPrometheusClientAPI([]float64{float64(0)}, nil)
	checks := make([]prometheusCheck, 0)
	check := &prometheusChecker{
		apiClient: prometheusAPIClient,
		Checks:    checks,
	}
	ok, err := check.Check()

	if len(ok) != 0 {
		t.Fatal("No test should return no ok")
	}
	if len(err) != 0 {
		t.Fatal("No test should return no error")
	}
}

func TestReturnOkIfValueIsGood(t *testing.T) {
	res := float64(1256.3598)
	prometheusAPIClient := initPrometheusClientAPI([]float64{res}, nil)
	checks := make([]prometheusCheck, 1)
	checks[0] = prometheusCheck{
		Description: "description",
		Query:       "query",
		ResultMin:   float64(1256),
		ResultMax:   float64(1257),
	}
	check := &prometheusChecker{
		apiClient: prometheusAPIClient,
		Checks:    checks,
	}

	ok, err := check.Check()

	if len(ok) != 1 {
		t.Fatal("1 good test should return 1 ok")
	}
	if ok[0].(prometheusCheckOk).Query != checks[0] || ok[0].Details()["Query"] != checks[0].Query {
		t.Fatal("1 good test should return same query")
	}
	if ok[0].(prometheusCheckOk).Val != res || ok[0].Details()["Val"] != res {
		t.Fatal("1 good test should return good val")
	}
	if len(err) != 0 {
		t.Fatal("No test should return no error")
	}
}

func TestMultipleGood(t *testing.T) {
	res := float64(1256.3598)
	prometheusAPIClient := initPrometheusClientAPI([]float64{res, res}, nil)
	checks := make([]prometheusCheck, 2)
	checks[0] = prometheusCheck{
		Description: "description",
		Query:       "query",
		ResultMin:   float64(1256.3598),
		ResultMax:   float64(1256.3598),
	}
	checks[1] = prometheusCheck{
		Description: "description",
		Query:       "query",
		ResultMin:   float64(1256),
		ResultMax:   float64(1257),
	}
	check := &prometheusChecker{
		apiClient: prometheusAPIClient,
		Checks:    checks,
	}

	ok, err := check.Check()

	if len(ok) != 2 {
		t.Fatal("2 good test should return 2 ok")
	}
	if len(err) != 0 {
		t.Fatal("No test should return no error")
	}
}

func TestReturnErrorsIfValueIsNotInGap(t *testing.T) {
	res := float64(1256.3598)
	prometheusAPIClient := initPrometheusClientAPI([]float64{res}, nil)
	checks := make([]prometheusCheck, 1)
	checks[0] = prometheusCheck{
		Description: "description",
		Query:       "query",
		ResultMin:   float64(1255),
		ResultMax:   float64(1256),
	}
	check := &prometheusChecker{
		apiClient: prometheusAPIClient,
		Checks:    checks,
	}

	ok, err := check.Check()

	if len(ok) != 0 {
		t.Fatal("1 bad test should return 0 ok")
	}
	if len(err) != 1 {
		t.Fatal("1 bad test should return 1 err")
	}
	if err[0].(prometheusCheckError).Query != checks[0] || err[0].Details()["Query"] != checks[0].Query {
		t.Fatal("1 bad test should return same query")
	}
	if err[0].(prometheusCheckError).Val != res || err[0].Details()["Val"] != res {
		t.Fatal("1 bad test should return good val")
	}
	if len(err[0].(prometheusCheckError).Reason) == 0 {
		t.Fatal("1 bad test should return a complete reason")
	}
}

func TestMultipleGoodBad(t *testing.T) {
	resGood := float64(1256.3598)
	resBad := float64(1)
	prometheusAPIClient := initPrometheusClientAPI([]float64{resGood, resBad}, nil)
	checks := make([]prometheusCheck, 2)
	checks[0] = prometheusCheck{
		Description: "description",
		Query:       "query1",
		ResultMin:   float64(1256),
		ResultMax:   float64(1257),
	}
	checks[1] = prometheusCheck{
		Description: "description",
		Query:       "query2",
		ResultMin:   float64(1256),
		ResultMax:   float64(1257),
	}
	check := &prometheusChecker{
		apiClient: prometheusAPIClient,
		Checks:    checks,
	}

	ok, err := check.Check()

	if len(ok) != 1 {
		t.Fatal("1 good test should return 1 ok")
	}
	if ok[0].(prometheusCheckOk).Val != resGood {
		t.Fatal("Good test should return good res")
	}
	if len(err) != 1 {
		t.Fatal("1 bad test should return 1 err")
	}
	if err[0].(prometheusCheckError).Val != resBad {
		t.Fatal("Bad test should return bad res")
	}
}

func TestPrometheusFail(t *testing.T) {
	prometheusAPIClient := initPrometheusClientAPI([]float64{0}, errors.New("Prometheus fail"))
	checks := make([]prometheusCheck, 1)
	checks[0] = prometheusCheck{
		Description: "description",
		Query:       "query1",
		ResultMin:   float64(1256),
		ResultMax:   float64(1257),
	}
	check := &prometheusChecker{
		apiClient: prometheusAPIClient,
		Checks:    checks,
	}
	ok, err := check.Check()
	if len(ok) != 0 {
		t.Fatal("Prometheus fail should not return ok")
	}
	if len(err) != 1 {
		t.Fatal("Prometheus fail should return an error")
	}
}
