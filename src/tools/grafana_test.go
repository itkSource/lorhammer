package tools

import (
	"bytes"
	"errors"
	"io"
	"io/ioutil"
	"net/http"
	"testing"
	"time"
)

type fakeConsul struct {
	serviceFirstError error
}

func (fakeConsul) GetAddress() string                                      { return "" }
func (fakeConsul) Register(ip string, hostname string, httpPort int) error { return nil }
func (f fakeConsul) ServiceFirst(name string, prefix string) (string, error) {
	return "grafanaUrl", f.serviceFirstError
}
func (fakeConsul) DeRegister(string) error               { return nil }
func (fakeConsul) AllServices() ([]ConsulService, error) { return nil, nil }

func newGrafana(t *testing.T, bodyGet string, errorGet error, bodyPost string, errorPost error) *grafanaClientImpl {
	g, err := NewGrafana(fakeConsul{})
	if err != nil {
		t.Fatal("Good grafana config should not throw err")
	}
	if g == nil {
		t.Fatal("Good grafana config should return grafana client")
	}
	if g.(*grafanaClientImpl).url != "grafanaUrl" {
		t.Fatal("Grafana client should retain url")
	}
	g.(*grafanaClientImpl).httpGetter = func(url string) (resp *http.Response, err error) {
		return &http.Response{Body: ioutil.NopCloser(bytes.NewReader([]byte(bodyGet)))}, errorGet
	}
	g.(*grafanaClientImpl).httpPoster = func(url string, contentType string, body io.Reader) (resp *http.Response, err error) {
		return &http.Response{Body: ioutil.NopCloser(bytes.NewReader([]byte(bodyPost)))}, errorPost
	}
	return g.(*grafanaClientImpl)
}

func TestNewGrafana(t *testing.T) {
	newGrafana(t, `{"dashboard": {}}`, nil, `{}`, nil)
}

func TestNewGrafanaError(t *testing.T) {
	g, err := NewGrafana(fakeConsul{serviceFirstError: errors.New("error")})
	if err == nil {
		t.Fatal("Bad grafana config should throw err")
	}
	if g != nil {
		t.Fatal("Bad grafana config should not return grafana client")
	}
}

func TestGrafanaClient_MakeSnapshot(t *testing.T) {
	g := newGrafana(t, `{"dashboard": {}}`, nil, `{}`, nil)
	url, err := g.MakeSnapshot(time.Now(), time.Now().Add(1*time.Minute))
	if err != nil {
		t.Fatal("Good grafana dashboard should not throw err")
	}
	if url != "grafanaUrl/dashboard/snapshot/" {
		t.Fatalf("Should return good url instead of %s", url)
	}
}

func TestGrafanaClient_MakeSnapshotErrorGet(t *testing.T) {
	g := newGrafana(t, `{"dashboard": {}}`, errors.New("error get"), `{}`, nil)
	url, err := g.MakeSnapshot(time.Now(), time.Now().Add(1*time.Minute))
	if err == nil {
		t.Fatal("Grafana should throw err when http get return err")
	}
	if url != "" {
		t.Fatalf("Grafana should return empty url instead of %s when http get return err", url)
	}
}

func TestGrafanaClient_MakeSnapshotErrorGetNoJson(t *testing.T) {
	g := newGrafana(t, `{`, nil, `{}`, nil)
	url, err := g.MakeSnapshot(time.Now(), time.Now().Add(1*time.Minute))
	if err == nil {
		t.Fatal("Grafana should throw err when http get return err")
	}
	if url != "" {
		t.Fatalf("Grafana should return empty url instead of %s when http get return err", url)
	}
}

func TestGrafanaClient_MakeSnapshotErrorGetBadDashboard(t *testing.T) {
	g := newGrafana(t, `{"dashboard": ""}`, nil, `{}`, nil)
	url, err := g.MakeSnapshot(time.Now(), time.Now().Add(1*time.Minute))
	if err == nil {
		t.Fatal("Grafana should throw err when http get return err")
	}
	if url != "" {
		t.Fatalf("Grafana should return empty url instead of %s when http get return err", url)
	}
}

func TestGrafanaClient_MakeSnapshotErrorPost(t *testing.T) {
	g := newGrafana(t, `{"dashboard": {}}`, nil, `{}`, errors.New("error post"))
	url, err := g.MakeSnapshot(time.Now(), time.Now().Add(1*time.Minute))
	if err == nil {
		t.Fatal("Grafana should throw err when http post return err")
	}
	if url != "" {
		t.Fatalf("Grafana should return empty url instead of %s when http post return err", url)
	}
}

func TestGrafanaClient_MakeSnapshotErrorPostNoJson(t *testing.T) {
	g := newGrafana(t, `{"dashboard": {}}`, nil, `{`, nil)
	url, err := g.MakeSnapshot(time.Now(), time.Now().Add(1*time.Minute))
	if err == nil {
		t.Fatal("Grafana should throw err when http post return err")
	}
	if url != "" {
		t.Fatalf("Grafana should return empty url instead of %s when http post return err", url)
	}
}
