package checker

import (
	"encoding/json"
	"errors"
	"lorhammer/src/orchestrator/prometheus"
	"lorhammer/src/tools"
	"testing"
)

var nbCall int = 0

type fakeConsul struct {
	serviceFirstError error
}

func (_ fakeConsul) GetAddress() string                                      { return "" }
func (_ fakeConsul) Register(ip string, hostname string, httpPort int) error { return nil }
func (f fakeConsul) ServiceFirst(name string, prefix string) (string, error) {
	return "prometheusUrl", f.serviceFirstError
}
func (_ fakeConsul) DeRegister(string) error                     { return nil }
func (_ fakeConsul) AllServices() ([]tools.ConsulService, error) { return nil, nil }

type PrometheusApiClientFake struct {
	returnFloat []float64
	returnError error
}

func (p PrometheusApiClientFake) ExecQuery(query string) (float64, error) {
	nbCall++
	return p.returnFloat[nbCall-1], p.returnError
}

func initPrometheusClientApi(res []float64, err error) prometheus.ApiClient {
	nbCall = 0
	return PrometheusApiClientFake{
		returnFloat: res,
		returnError: err,
	}
}

func TestNewPrometheus(t *testing.T) {
	check, err := newPrometheus(fakeConsul{}, json.RawMessage([]byte(`[]`)))
	if err != nil {
		t.Fatal("Valid config for prometheus should not return error")
	}
	if check == nil {
		t.Fatal("Valid config for prometheus should return checker")
	}
}

func TestNewPrometheusError(t *testing.T) {
	check, err := newPrometheus(fakeConsul{}, json.RawMessage([]byte(`{`)))
	if err == nil {
		t.Fatal("Invalid config for prometheus should return error")
	}
	if check != nil {
		t.Fatal("Invalid config for prometheus should not return checker")
	}
}

func TestNewPrometheusErrorConsul(t *testing.T) {
	check, err := newPrometheus(fakeConsul{serviceFirstError: errors.New("error")}, json.RawMessage([]byte(`[]`)))
	if err == nil {
		t.Fatal("Consul error for prometheus should return error")
	}
	if check != nil {
		t.Fatal("Consul error for prometheus should not return checker")
	}
}

func TestPrometheusChecker_Start(t *testing.T) {
	check, _ := newPrometheus(fakeConsul{}, json.RawMessage([]byte(`[]`)))
	if err := check.Start(); err != nil {
		t.Fatal("Prometheus checker should not return error on start")
	}
}

func TestReturnNothingIfNoTest(t *testing.T) {
	prometheusApiClient := initPrometheusClientApi([]float64{float64(0)}, nil)
	checks := make([]prometheusCheck, 0)
	check := prometheusChecker{
		apiClient: prometheusApiClient,
		checks:    checks,
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
	prometheusApiClient := initPrometheusClientApi([]float64{res}, nil)
	checks := make([]prometheusCheck, 1)
	checks[0] = prometheusCheck{
		Description: "description",
		Query:       "query",
		ResultMin:   float64(1256),
		ResultMax:   float64(1257),
	}
	check := prometheusChecker{
		apiClient: prometheusApiClient,
		checks:    checks,
	}

	ok, err := check.Check()

	if len(ok) != 1 {
		t.Fatal("1 good test should return 1 ok")
	}
	if ok[0].(PrometheusCheckOk).Query != checks[0] || ok[0].Details()["Query"] != checks[0].Query {
		t.Fatal("1 good test should return same query")
	}
	if ok[0].(PrometheusCheckOk).Val != res || ok[0].Details()["Val"] != res {
		t.Fatal("1 good test should return good val")
	}
	if len(err) != 0 {
		t.Fatal("No test should return no error")
	}
}

func TestMultipleGood(t *testing.T) {
	res := float64(1256.3598)
	prometheusApiClient := initPrometheusClientApi([]float64{res, res}, nil)
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
	check := prometheusChecker{
		apiClient: prometheusApiClient,
		checks:    checks,
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
	prometheusApiClient := initPrometheusClientApi([]float64{res}, nil)
	checks := make([]prometheusCheck, 1)
	checks[0] = prometheusCheck{
		Description: "description",
		Query:       "query",
		ResultMin:   float64(1255),
		ResultMax:   float64(1256),
	}
	check := prometheusChecker{
		apiClient: prometheusApiClient,
		checks:    checks,
	}

	ok, err := check.Check()

	if len(ok) != 0 {
		t.Fatal("1 bad test should return 0 ok")
	}
	if len(err) != 1 {
		t.Fatal("1 bad test should return 1 err")
	}
	if err[0].(PrometheusCheckError).Query != checks[0] || err[0].Details()["Query"] != checks[0].Query {
		t.Fatal("1 bad test should return same query")
	}
	if err[0].(PrometheusCheckError).Val != res || err[0].Details()["Val"] != res {
		t.Fatal("1 bad test should return good val")
	}
	if len(err[0].(PrometheusCheckError).Reason) == 0 {
		t.Fatal("1 bad test should return a complete reason")
	}
}

func TestMultipleGoodBad(t *testing.T) {
	resGood := float64(1256.3598)
	resBad := float64(1)
	prometheusApiClient := initPrometheusClientApi([]float64{resGood, resBad}, nil)
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
	check := prometheusChecker{
		apiClient: prometheusApiClient,
		checks:    checks,
	}

	ok, err := check.Check()

	if len(ok) != 1 {
		t.Fatal("1 good test should return 1 ok")
	}
	if ok[0].(PrometheusCheckOk).Val != resGood {
		t.Fatal("Good test should return good res")
	}
	if len(err) != 1 {
		t.Fatal("1 bad test should return 1 err")
	}
	if err[0].(PrometheusCheckError).Val != resBad {
		t.Fatal("Bad test should return bad res")
	}
}

func TestPrometheusFail(t *testing.T) {
	prometheusApiClient := initPrometheusClientApi([]float64{0}, errors.New("Prometheus fail"))
	checks := make([]prometheusCheck, 1)
	checks[0] = prometheusCheck{
		Description: "description",
		Query:       "query1",
		ResultMin:   float64(1256),
		ResultMax:   float64(1257),
	}
	check := prometheusChecker{
		apiClient: prometheusApiClient,
		checks:    checks,
	}
	ok, err := check.Check()
	if len(ok) != 0 {
		t.Fatal("Prometheus fail should not return ok")
	}
	if len(err) != 1 {
		t.Fatal("Prometheus fail should return an error")
	}
}
