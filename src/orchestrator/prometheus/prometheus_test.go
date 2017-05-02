package prometheus

import (
	"context"
	"errors"
	api "github.com/prometheus/client_golang/api/prometheus/v1"
	"github.com/prometheus/common/model"
	"lorhammer/src/tools"
	"testing"
	"time"
)

type fakeConsul struct {
	serviceFirstUrl   string
	serviceFirstError error
}

func (_ fakeConsul) GetAddress() string                                      { return "" }
func (_ fakeConsul) Register(ip string, hostname string, httpPort int) error { return nil }
func (f fakeConsul) ServiceFirst(name string, prefix string) (string, error) {
	return f.serviceFirstUrl, f.serviceFirstError
}
func (_ fakeConsul) DeRegister(string) error                     { return nil }
func (_ fakeConsul) AllServices() ([]tools.ConsulService, error) { return nil, nil }

type fakePrometheusApi struct {
	results      []float64
	resultsError error
}

func (f fakePrometheusApi) Query(ctx context.Context, query string, ts time.Time) (model.Value, error) {
	values := make([]*model.Sample, len(f.results))
	for i, val := range f.results {
		values[i] = &model.Sample{Value: model.SampleValue(val)}
	}
	return model.Vector(values), f.resultsError
}
func (f fakePrometheusApi) QueryRange(ctx context.Context, query string, r api.Range) (model.Value, error) {
	return nil, nil
}

func (f fakePrometheusApi) LabelValues(ctx context.Context, label string) (model.LabelValues, error) {
	return nil, nil
}

func TestNewApiClient(t *testing.T) {
	c, err := NewApiClient(fakeConsul{})
	if err != nil {
		t.Fatal("Valid prometheus config should not return error")
	}
	if c == nil {
		t.Fatal("Valid prometheus config should return api client")
	}
}

func TestNewApiClientErrorConsul(t *testing.T) {
	c, err := NewApiClient(fakeConsul{serviceFirstError: errors.New("error")})
	if err == nil {
		t.Fatal("No prometheus found in consul should return error")
	}
	if c != nil {
		t.Fatal("No prometheus found in consul should not return api client")
	}
}

func TestNewApiClientErrorBadConsulUrl(t *testing.T) {
	c, err := NewApiClient(fakeConsul{serviceFirstUrl: ":"})
	if err == nil {
		t.Fatal("Bad url for prometheus in consul should return error")
	}
	if c != nil {
		t.Fatal("Bad url for prometheus in consul should not return api client")
	}
}

func TestApiClientImpl_ExecQuery(t *testing.T) {
	c, _ := NewApiClient(fakeConsul{})
	c.(*apiClientImpl).queryApi = fakePrometheusApi{results: []float64{255.0}}
	res, err := c.ExecQuery("")
	if err != nil {
		t.Fatal("Valid call to prometheus api should not throw error")
	}
	if res != 255.0 {
		t.Fatal("Valid call to prometheus api should return good value casted in float64")
	}

}

func TestApiClientImpl_ExecQueryZero(t *testing.T) {
	c, _ := NewApiClient(fakeConsul{})
	c.(*apiClientImpl).queryApi = fakePrometheusApi{results: []float64{}}
	res, err := c.ExecQuery("")
	if err != nil {
		t.Fatal("Valid call to prometheus api should not throw error")
	}
	if res != 0 {
		t.Fatal("Valid call to prometheus api without result should return 0")
	}
}

func TestApiClientImpl_ExecQueryError(t *testing.T) {
	c, _ := NewApiClient(fakeConsul{})
	c.(*apiClientImpl).queryApi = fakePrometheusApi{results: []float64{}, resultsError: errors.New("error")}
	_, err := c.ExecQuery("")
	if err == nil {
		t.Fatal("If prometheus api report error, prometheus client should throw error")
	}
}
