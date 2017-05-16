package tools

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/Sirupsen/logrus"
	"io"
	"io/ioutil"
	"net/http"
	"time"
)

const (
	_GRAFANA_CONSUL_SERVICE    = "grafana"
	_GRAFANA_DASHBOARD_NAME    = "lora"
	_GRAFANA_URL_API_DASHBOARD = "/api/dashboards/db/"
	_GRAFANA_URL_API_SNAPSHOT  = "/api/snapshots"
	_GRAFANA_URL_SNAPSHOT      = "/dashboard/snapshot/"
)

var LOG_GRAFANA = logrus.WithField("logger", "tools/grafana")

type GrafanaClient interface {
	MakeSnapshot(startTime time.Time, endTime time.Time) (string, error)
}

type grafanaClientImpl struct {
	url        string
	httpGetter func(url string) (resp *http.Response, err error)
	httpPoster func(url string, contentType string, body io.Reader) (resp *http.Response, err error)
}

type makeSnapshot struct {
	Dashboard json.RawMessage `json:"dashboard"`
	Name      string          `json:"name"`
	Expires   int             `json:"expires"`
}

type snapshot struct {
	DeleteKey string `json:"deleteKey"`
	DeleteUrl string `json:"deleteUrl"`
	Key       string `json:"key"`
	Url       string `json:"url"`
}

func NewGrafana(consulClient Consul) (GrafanaClient, error) {
	address, err := consulClient.ServiceFirst(_GRAFANA_CONSUL_SERVICE, "http://")
	if err != nil {
		return nil, err
	}
	return &grafanaClientImpl{
		url:        address,
		httpGetter: http.Get,
		httpPoster: http.Post,
	}, nil
}

func (grafana *grafanaClientImpl) MakeSnapshot(startTime time.Time, endTime time.Time) (string, error) {
	dashboard, err := grafana.getDashboard(_GRAFANA_DASHBOARD_NAME, startTime, endTime)
	if err != nil {
		return "", err
	}
	mSnapshot := makeSnapshot{
		Dashboard: dashboard,
		Name:      _GRAFANA_DASHBOARD_NAME,
		Expires:   3600,
	}
	body, err := json.Marshal(mSnapshot)
	if err != nil {
		return "", err
	}
	res, err := grafana.httpPoster(grafana.url+_GRAFANA_URL_API_SNAPSHOT, "application/json", bytes.NewBuffer(body))
	if err != nil {
		return "", err
	}
	defer res.Body.Close()
	s := snapshot{}
	if err := json.NewDecoder(res.Body).Decode(&s); err != nil {
		buf := new(bytes.Buffer)
		buf.ReadFrom(res.Body)
		LOG_GRAFANA.WithField("body", buf.String()).Error("Grafana response")
		return "", err
	}
	urlSnapshot := grafana.url + _GRAFANA_URL_SNAPSHOT + s.Key
	LOG_GRAFANA.WithField("url", urlSnapshot).Info("Snapshot grafana")
	return urlSnapshot, nil
}

func (grafana *grafanaClientImpl) getDashboard(name string, startTime time.Time, endTime time.Time) (json.RawMessage, error) {
	res, err := grafana.httpGetter(grafana.url + _GRAFANA_URL_API_DASHBOARD + name)
	if err != nil {
		return nil, err
	}
	r := make(map[string]json.RawMessage)
	htmlData, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(htmlData, &r); err != nil {
		return nil, err
	}
	d := make(map[string]json.RawMessage)
	if err := json.Unmarshal(r["dashboard"], &d); err != nil {
		return nil, err
	}
	delete(d, "refresh")                                                                   // don't refresh by default
	d["time"] = []byte(fmt.Sprintf("{\"from\":\"%s\",\"to\":\"%s\"}", startTime, endTime)) // set display time with test time
	serialized, err := json.Marshal(d)
	return serialized, nil
}
