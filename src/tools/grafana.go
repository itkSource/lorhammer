package tools

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/Sirupsen/logrus"
)

const (
	grafanaConsulService   = "grafana"
	grafanaDashboardName   = "lora"
	grafanaURLApiDashboard = "/api/dashboards/db/"
	grafanaURLApiSnapshot  = "/api/snapshots"
	grafanaURLSnapshot     = "/dashboard/snapshot/"
)

var logGrafana = logrus.WithField("logger", "tools/grafana")

//GrafanaClient permit to interact with grafana pi
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
	DeleteURL string `json:"deleteUrl"`
	Key       string `json:"key"`
	URL       string `json:"url"`
}

//NewGrafana return a grafana client
func NewGrafana(consulClient Consul) (GrafanaClient, error) {
	address, err := consulClient.ServiceFirst(grafanaConsulService, "http://")
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
	dashboard, err := grafana.getDashboard(grafanaDashboardName, startTime, endTime)
	if err != nil {
		return "", err
	}
	mSnapshot := makeSnapshot{
		Dashboard: dashboard,
		Name:      grafanaDashboardName,
		Expires:   3600,
	}
	body, err := json.Marshal(mSnapshot)
	if err != nil {
		return "", err
	}
	res, err := grafana.httpPoster(grafana.url+grafanaURLApiSnapshot, "application/json", bytes.NewBuffer(body))
	if err != nil {
		return "", err
	}
	defer res.Body.Close()
	s := snapshot{}
	if err := json.NewDecoder(res.Body).Decode(&s); err != nil {
		buf := new(bytes.Buffer)
		buf.ReadFrom(res.Body)
		logGrafana.WithField("body", buf.String()).WithError(err).Error("Grafana response")
		return "", err
	}
	urlSnapshot := grafana.url + grafanaURLSnapshot + s.Key
	logGrafana.WithField("url", urlSnapshot).Info("Snapshot grafana")
	return urlSnapshot, nil
}

func (grafana *grafanaClientImpl) getDashboard(name string, startTime time.Time, endTime time.Time) (json.RawMessage, error) {
	res, err := grafana.httpGetter(grafana.url + grafanaURLApiDashboard + name)
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
