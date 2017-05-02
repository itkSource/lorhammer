package checker

import (
	"lorhammer/src/orchestrator/prometheus"
)

type PrometheusCheck struct {
	Description string  `json:"description"`
	Query       string  `json:"query"`
	ResultMin   float64 `json:"resultMin"`
	ResultMax   float64 `json:"resultMax"`
}

type PrometheusCheckOk struct {
	Query PrometheusCheck
	Val   float64
}

type PrometheusCheckError struct {
	Query  PrometheusCheck
	Val    float64
	Reason string
}

func Check(prometheusApiClient prometheus.ApiClient, checks []PrometheusCheck) ([]PrometheusCheckOk, []PrometheusCheckError) {
	success := make([]PrometheusCheckOk, 0)
	errs := make([]PrometheusCheckError, 0)
	for _, query := range checks {
		if val, err := prometheusApiClient.ExecQuery(query.Query); err != nil {
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
