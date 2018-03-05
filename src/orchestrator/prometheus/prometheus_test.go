package prometheus

import (
	"context"
	"errors"
	"testing"
	"time"

	api "github.com/prometheus/client_golang/api/prometheus/v1"
	"github.com/prometheus/common/model"
)

type fakePrometheusAPI struct {
	results      []float64
	resultsError error
}

func (f fakePrometheusAPI) Query(ctx context.Context, query string, ts time.Time) (model.Value, error) {
	values := make([]*model.Sample, len(f.results))
	for i, val := range f.results {
		values[i] = &model.Sample{Value: model.SampleValue(val)}
	}
	return model.Vector(values), f.resultsError
}
func (f fakePrometheusAPI) QueryRange(ctx context.Context, query string, r api.Range) (model.Value, error) {
	return nil, nil
}

func (f fakePrometheusAPI) LabelValues(ctx context.Context, label string) (model.LabelValues, error) {
	return nil, nil
}

func TestNewApiClient(t *testing.T) {
	c, err := NewAPIClient("")
	if err != nil {
		t.Fatal("Valid prometheus config should not return error")
	}
	if c == nil {
		t.Fatal("Valid prometheus config should return api client")
	}
}

func TestNewApiClientErrorBadConsulUrl(t *testing.T) {
	c, err := NewAPIClient(":")
	if err == nil {
		t.Fatal("Bad url for prometheus in consul should return error")
	}
	if c != nil {
		t.Fatal("Bad url for prometheus in consul should not return api client")
	}
}

func TestApiClientImpl_ExecQuery(t *testing.T) {
	c, _ := NewAPIClient("")
	c.(*apiClientImpl).queryAPI = fakePrometheusAPI{results: []float64{255.0}}
	res, err := c.ExecQuery("")
	if err != nil {
		t.Fatal("Valid call to prometheus api should not throw error")
	}
	if res != 255.0 {
		t.Fatal("Valid call to prometheus api should return good value casted in float64")
	}

}

func TestApiClientImpl_ExecQueryZero(t *testing.T) {
	c, _ := NewAPIClient("")
	c.(*apiClientImpl).queryAPI = fakePrometheusAPI{results: []float64{}}
	res, err := c.ExecQuery("")
	if err != nil {
		t.Fatal("Valid call to prometheus api should not throw error")
	}
	if res != 0 {
		t.Fatal("Valid call to prometheus api without result should return 0")
	}
}

func TestApiClientImpl_ExecQueryError(t *testing.T) {
	c, _ := NewAPIClient("")
	c.(*apiClientImpl).queryAPI = fakePrometheusAPI{results: []float64{}, resultsError: errors.New("error")}
	_, err := c.ExecQuery("")
	if err == nil {
		t.Fatal("If prometheus api report error, prometheus client should throw error")
	}
}
