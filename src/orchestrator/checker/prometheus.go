package checker

import (
	"encoding/json"
	"lorhammer/src/orchestrator/prometheus"
	"lorhammer/src/tools"
)

const PrometheusType = Type("prometheus")

type prometheusChecker struct {
	apiClient prometheus.ApiClient
	checks    []prometheusCheck
}

type prometheusCheck struct {
	Description string  `json:"description"`
	Query       string  `json:"query"`
	ResultMin   float64 `json:"resultMin"`
	ResultMax   float64 `json:"resultMax"`
}

type PrometheusCheckOk struct {
	Query prometheusCheck
	Val   float64
}

func (ok PrometheusCheckOk) Details() map[string]interface{} {
	details := make(map[string]interface{})
	details["Query"] = ok.Query.Query
	details["Description"] = ok.Query.Description
	details["ResultMin"] = ok.Query.ResultMin
	details["ResultMax"] = ok.Query.ResultMax
	details["Val"] = ok.Val
	return details
}

type PrometheusCheckError struct {
	Query  prometheusCheck
	Val    float64
	Reason string
}

func (err PrometheusCheckError) Details() map[string]interface{} {
	details := make(map[string]interface{})
	details["Query"] = err.Query.Query
	details["Description"] = err.Query.Description
	details["ResultMin"] = err.Query.ResultMin
	details["ResultMax"] = err.Query.ResultMax
	details["Val"] = err.Val
	details["Reason"] = err.Reason
	return details
}

func newPrometheus(consulClient tools.Consul, rawConfig json.RawMessage) (Checker, error) {
	var checks = make([]prometheusCheck, 0)
	if err := json.Unmarshal(rawConfig, checks); err != nil {
		return nil, err
	}
	prometheusApiClient, err := prometheus.NewApiClient(consulClient)
	if err != nil {
		return nil, err
	}
	return prometheusChecker{
		apiClient: prometheusApiClient,
		checks:    checks,
	}, nil
}

func (prom prometheusChecker) Check() ([]CheckerSuccess, []CheckerError) {
	success := make([]CheckerSuccess, 0)
	errs := make([]CheckerError, 0)
	for _, query := range prom.checks {
		if val, err := prom.apiClient.ExecQuery(query.Query); err != nil {
			errs = append(errs, PrometheusCheckError{
				Query:  query,
				Val:    val,
				Reason: "Query to prometheus failed",
			})
		} else if val < query.ResultMin || val > query.ResultMax {
			errs = append(errs, PrometheusCheckError{
				Query:  query,
				Val:    val,
				Reason: "Result mismatch",
			})
		} else {
			success = append(success, PrometheusCheckOk{
				Query: query,
				Val:   val,
			})
		}
	}
	return success, errs
}
