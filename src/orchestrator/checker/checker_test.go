package checker

import (
	"errors"
	"lorhammer/src/orchestrator/prometheus"
	"testing"
)

var nbCall int = 0

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

func TestReturnNothingIfNoTest(t *testing.T) {
	prometheusApiClient := initPrometheusClientApi([]float64{float64(0)}, nil)
	checks := make([]PrometheusCheck, 0)

	ok, err := Check(prometheusApiClient, checks)

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
	checks := make([]PrometheusCheck, 1)
	checks[0] = PrometheusCheck{
		Description: "description",
		Query:       "query",
		ResultMin:   float64(1256),
		ResultMax:   float64(1257),
	}

	ok, err := Check(prometheusApiClient, checks)

	if len(ok) != 1 {
		t.Fatal("1 good test should return 1 ok")
	}
	if ok[0].Query != checks[0] {
		t.Fatal("1 good test should return same query")
	}
	if ok[0].Val != res {
		t.Fatal("1 good test should return good val")
	}
	if len(err) != 0 {
		t.Fatal("No test should return no error")
	}
}

func TestMultipleGood(t *testing.T) {
	res := float64(1256.3598)
	prometheusApiClient := initPrometheusClientApi([]float64{res, res}, nil)
	checks := make([]PrometheusCheck, 2)
	checks[0] = PrometheusCheck{
		Description: "description",
		Query:       "query",
		ResultMin:   float64(1256.3598),
		ResultMax:   float64(1256.3598),
	}
	checks[1] = PrometheusCheck{
		Description: "description",
		Query:       "query",
		ResultMin:   float64(1256),
		ResultMax:   float64(1257),
	}

	ok, err := Check(prometheusApiClient, checks)

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
	checks := make([]PrometheusCheck, 1)
	checks[0] = PrometheusCheck{
		Description: "description",
		Query:       "query",
		ResultMin:   float64(1255),
		ResultMax:   float64(1256),
	}

	ok, err := Check(prometheusApiClient, checks)

	if len(ok) != 0 {
		t.Fatal("1 bad test should return 0 ok")
	}
	if len(err) != 1 {
		t.Fatal("1 bad test should return 1 err")
	}
	if err[0].Query != checks[0] {
		t.Fatal("1 bad test should return same query")
	}
	if err[0].Val != res {
		t.Fatal("1 bad test should return good val")
	}
	if len(err[0].Reason) == 0 {
		t.Fatal("1 bad test should return a complete reason")
	}
}

func TestMultipleGoodBad(t *testing.T) {
	resGood := float64(1256.3598)
	resBad := float64(1)
	prometheusApiClient := initPrometheusClientApi([]float64{resGood, resBad}, nil)
	checks := make([]PrometheusCheck, 2)
	checks[0] = PrometheusCheck{
		Description: "description",
		Query:       "query1",
		ResultMin:   float64(1256),
		ResultMax:   float64(1257),
	}
	checks[1] = PrometheusCheck{
		Description: "description",
		Query:       "query2",
		ResultMin:   float64(1256),
		ResultMax:   float64(1257),
	}

	ok, err := Check(prometheusApiClient, checks)

	if len(ok) != 1 {
		t.Fatal("1 good test should return 1 ok")
	}
	if ok[0].Val != resGood {
		t.Fatal("Good test should return good res")
	}
	if len(err) != 1 {
		t.Fatal("1 bad test should return 1 err")
	}
	if err[0].Val != resBad {
		t.Fatal("Bad test should return bad res")
	}
}

func TestPrometheusFail(t *testing.T) {
	prometheusApiClient := initPrometheusClientApi([]float64{0}, errors.New("Prometheus fail"))
	checks := make([]PrometheusCheck, 1)
	checks[0] = PrometheusCheck{
		Description: "description",
		Query:       "query1",
		ResultMin:   float64(1256),
		ResultMax:   float64(1257),
	}
	ok, err := Check(prometheusApiClient, checks)
	if len(ok) != 0 {
		t.Fatal("Prometheus fail should not return ok")
	}
	if len(err) != 1 {
		t.Fatal("Prometheus fail should return an error")
	}
}
