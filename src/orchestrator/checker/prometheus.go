package checker

import (
	"encoding/json"
	"lorhammer/src/orchestrator/metrics"
	"lorhammer/src/orchestrator/prometheus"
)

const prometheusType = Type("prometheus")

type prometheusChecker struct {
	Address                 string            `json:"address"`
	Checks                  []prometheusCheck `json:"checks"`
	prometheusClientFactory func(string) (prometheus.APIClient, error)
	apiClient               prometheus.APIClient
}

type prometheusCheck struct {
	Description string  `json:"description"`
	Query       string  `json:"query"`
	ResultMin   float64 `json:"resultMin"`
	ResultMax   float64 `json:"resultMax"`
}

type prometheusCheckOk struct {
	Query prometheusCheck
	Val   float64
}

func (ok prometheusCheckOk) Details() map[string]interface{} {
	details := make(map[string]interface{})
	details["Query"] = ok.Query.Query
	details["Description"] = ok.Query.Description
	details["ResultMin"] = ok.Query.ResultMin
	details["ResultMax"] = ok.Query.ResultMax
	details["Val"] = ok.Val
	return details
}

type prometheusCheckError struct {
	Query  prometheusCheck
	Val    float64
	Reason string
}

func (err prometheusCheckError) Details() map[string]interface{} {
	details := make(map[string]interface{})
	details["Query"] = err.Query.Query
	details["Description"] = err.Query.Description
	details["ResultMin"] = err.Query.ResultMin
	details["ResultMax"] = err.Query.ResultMax
	details["Val"] = err.Val
	details["Reason"] = err.Reason
	return details
}

func newPrometheus(rawConfig json.RawMessage, _ metrics.Prometheus) (Checker, error) {
	var checker = &prometheusChecker{
		prometheusClientFactory: prometheus.NewAPIClient,
	}
	if err := json.Unmarshal(rawConfig, checker); err != nil {
		return nil, err
	}

	return checker, nil
}

func (prom *prometheusChecker) Start() error {
	prometheusAPIClient, err := prom.prometheusClientFactory(prom.Address)
	if err != nil {
		return err
	}
	prom.apiClient = prometheusAPIClient
	return nil
}

func (prom *prometheusChecker) Check() ([]Success, []Error) {
	success := make([]Success, 0)
	errs := make([]Error, 0)
	for _, query := range prom.Checks {
		if val, err := prom.apiClient.ExecQuery(query.Query); err != nil {
			errs = append(errs, prometheusCheckError{
				Query:  query,
				Val:    val,
				Reason: "Query to prometheus failed",
			})
		} else if val < query.ResultMin || val > query.ResultMax {
			errs = append(errs, prometheusCheckError{
				Query:  query,
				Val:    val,
				Reason: "Result mismatch",
			})
		} else {
			success = append(success, prometheusCheckOk{
				Query: query,
				Val:   val,
			})
		}
	}
	return success, errs
}
