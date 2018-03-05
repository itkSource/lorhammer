package prometheus

import (
	"time"

	client "github.com/prometheus/client_golang/api"
	api "github.com/prometheus/client_golang/api/prometheus/v1"
	"github.com/prometheus/common/model"
	"golang.org/x/net/context"
)

//APIClient permit communication with prometheus rest api
type APIClient interface {
	ExecQuery(query string) (float64, error)
}

type apiClientImpl struct {
	queryAPI api.API
}

//NewAPIClient return an APIClient of prometheus
func NewAPIClient(address string) (APIClient, error) {
	c, err := client.NewClient(client.Config{Address: address})
	if err != nil {
		return nil, err
	}
	return &apiClientImpl{
		queryAPI: api.NewAPI(c),
	}, nil
}

func (p *apiClientImpl) ExecQuery(query string) (float64, error) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	val, err := p.queryAPI.Query(ctx, query, time.Now())
	if err != nil {
		return 0, err
	}

	vectorVal := val.(model.Vector)
	if len(vectorVal) == 0 {
		return 0, nil
	}
	return float64(vectorVal[0].Value), nil
}
