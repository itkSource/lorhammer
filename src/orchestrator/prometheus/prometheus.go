package prometheus

import (
	"github.com/prometheus/client_golang/api/prometheus"
	"github.com/prometheus/common/model"
	"golang.org/x/net/context"
	"lorhammer/src/tools"
	"time"
)

type ApiClient interface {
	ExecQuery(query string) (float64, error)
}

type apiClientImpl struct {
	queryApi prometheus.QueryAPI
}

func NewApiClient(consulClient tools.Consul) (ApiClient, error) {
	address, err := consulClient.ServiceFirst("prometheus", "http://")
	if err != nil {
		return nil, err
	}
	client, err := prometheus.New(prometheus.Config{Address: address})
	if err != nil {
		return nil, err
	}
	return &apiClientImpl{
		queryApi: prometheus.NewQueryAPI(client),
	}, nil
}

func (p *apiClientImpl) ExecQuery(query string) (float64, error) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	val, err := p.queryApi.Query(ctx, query, time.Now())
	if err != nil {
		return 0, err
	}

	vectorVal := val.(model.Vector)
	if len(vectorVal) == 0 {
		return 0, nil
	}
	return float64(vectorVal[0].Value), nil
}
