package provisioning

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"lorhammer/src/model"
	"lorhammer/src/tools"
	"net/http"
	"regexp"
	"testing"
	"time"
)

type nopCloser struct {
	io.Reader
}

func (nopCloser) Close() error { return nil }

func newDefautlLoraserver() *loraserver {
	j := `{
	  "apiUrl":"",
	  "abp" : true,
      "login" : "admin",
      "password" : "admin",
	  "nbProvisionerParallel": 1,
	  "networkServerAddr": "0.0.0.0:8000"
	}`
	raw := json.RawMessage([]byte(j))
	l, _ := newLoraserver(raw)
	return l.(*loraserver)
}

func newNodes(nb int) []*model.Node {
	nodes := make([]*model.Node, nb)
	for i := 0; i < nb; i++ {
		nodes[i] = &model.Node{
			DevAddr:        tools.Random4Bytes(),
			DevEUI:         tools.Random8Bytes(),
			AppEUI:         tools.Random8Bytes(),
			AppKey:         tools.Random16Bytes(),
			AppSKey:        tools.Random16Bytes(),
			NwSKey:         tools.Random16Bytes(),
			JoinedNetwork:  false,
			Payloads:       []model.Payload{{Value: ""}},
			NextPayload:    0,
			RandomPayloads: false,
		}
	}
	return nodes
}

func newGateways(nbNodes int, nbGateway int) []model.Gateway {
	gateways := make([]model.Gateway, nbGateway)
	for i := 0; i < nbGateway; i++ {
		gateways[i] = model.Gateway{
			Nodes:              newNodes(nbNodes),
			NsAddress:          "",
			MacAddress:         tools.Random8Bytes(),
			RxpkDate:           int64(1),
			ReceiveTimeoutTime: 1 * time.Second,
		}
	}
	return gateways
}

type fakeHTTPClient struct {
	data []fakeHTTPClientData
}

type testLoraserver struct {
	name    string
	inError bool
	data    []fakeHTTPClientData
}

type fakeHTTPClientData struct {
	url    *regexp.Regexp
	method string
	status int
	err    error
	body   string
}

var data = []testLoraserver{
	{
		name:    "all exist",
		inError: false,
		data: []fakeHTTPClientData{
			{url: regexp.MustCompile(`/api/internal/login`), method: "POST", err: nil, body: `{"jwt":"1"}`},
			{url: regexp.MustCompile(`/api/organizations\?limit=100`), method: "GET", err: nil, body: `{"result":[{"id":"1","name":"` + loraserverOrganisationName + `"}]}`},
			{url: regexp.MustCompile(`/api/network-servers\?limit=100`), method: "GET", err: nil, body: `{"result":[{"id":"1","server":"0.0.0.0:8000"}]}`},
			{url: regexp.MustCompile(`/api/service-profiles`), method: "POST", err: nil, body: `{"serviceProfileID":"1"}`},
			{url: regexp.MustCompile(`/api/applications\?limit=100`), method: "GET", err: nil, body: `{"result":[{"id":"1","name":"` + loraserverApplicationName + `"}]}`},
			{url: regexp.MustCompile(`/api/device-profiles`), method: "POST", err: nil, body: `{"deviceProfileID":"1"}`},
			{url: regexp.MustCompile(`/api/devices`), method: "POST", err: nil, body: ``},
			{url: regexp.MustCompile(`/api/devices/[^/]+/keys`), method: "POST", err: nil, body: ``},
		},
	}, {
		name:    "must create all",
		inError: false,
		data: []fakeHTTPClientData{
			{url: regexp.MustCompile(`/api/internal/login`), method: "POST", err: nil, body: `{"jwt":"1"}`},
			{url: regexp.MustCompile(`/api/organizations\?limit=100`), method: "GET", err: nil, body: `{"result":[]}`},
			{url: regexp.MustCompile(`/api/organizations`), method: "POST", err: nil, body: `{"id":"1"}`},
			{url: regexp.MustCompile(`/api/network-servers\?limit=100`), method: "GET", err: nil, body: `{"result":[]}`},
			{url: regexp.MustCompile(`/api/network-servers`), method: "POST", err: nil, body: `{"id":"1"}`},
			{url: regexp.MustCompile(`/api/service-profiles`), method: "POST", err: nil, body: `{"serviceProfileID":"1"}`},
			{url: regexp.MustCompile(`/api/applications\?limit=100`), method: "GET", err: nil, body: `{"result":[]}`},
			{url: regexp.MustCompile(`/api/applications`), method: "POST", err: nil, body: `{"id":"1"}`},
			{url: regexp.MustCompile(`/api/device-profiles`), method: "POST", err: nil, body: `{"deviceProfileID":"1"}`},
			{url: regexp.MustCompile(`/api/devices`), method: "POST", err: nil, body: ``},
			{url: regexp.MustCompile(`/api/devices/[^/]+/keys`), method: "POST", err: nil, body: ``},
		},
	}, {
		name:    "error login",
		inError: true,
		data: []fakeHTTPClientData{
			{url: regexp.MustCompile(`/api/internal/login`), method: "POST", err: errors.New("fake error login"), body: `{"jwt":"1"}`},
		},
	}, {
		name:    "error organization get",
		inError: true,
		data: []fakeHTTPClientData{
			{url: regexp.MustCompile(`/api/internal/login`), method: "POST", err: nil, body: `{"jwt":"1"}`},
			{url: regexp.MustCompile(`/api/organizations\?limit=100`), method: "GET", err: errors.New("fake error get organization"), body: `{"result":[]}`},
		},
	}, {
		name:    "error organization post",
		inError: true,
		data: []fakeHTTPClientData{
			{url: regexp.MustCompile(`/api/internal/login`), method: "POST", err: nil, body: `{"jwt":"1"}`},
			{url: regexp.MustCompile(`/api/organizations\?limit=100`), method: "GET", err: nil, body: `{"result":[]}`},
			{url: regexp.MustCompile(`/api/organizations`), method: "POST", err: errors.New("fake error post organization"), body: `{"id":"1"}`},
		},
	}, {
		name:    "error network-servers get",
		inError: true,
		data: []fakeHTTPClientData{
			{url: regexp.MustCompile(`/api/internal/login`), method: "POST", err: nil, body: `{"jwt":"1"}`},
			{url: regexp.MustCompile(`/api/organizations\?limit=100`), method: "GET", err: nil, body: `{"result":[]}`},
			{url: regexp.MustCompile(`/api/organizations`), method: "POST", err: nil, body: `{"id":"1"}`},
			{url: regexp.MustCompile(`/api/network-servers\?limit=100`), method: "GET", err: errors.New("fake error get network-servers"), body: `{"result":[]}`},
		},
	}, {
		name:    "error network-servers post",
		inError: true,
		data: []fakeHTTPClientData{
			{url: regexp.MustCompile(`/api/internal/login`), method: "POST", err: nil, body: `{"jwt":"1"}`},
			{url: regexp.MustCompile(`/api/organizations\?limit=100`), method: "GET", err: nil, body: `{"result":[]}`},
			{url: regexp.MustCompile(`/api/organizations`), method: "POST", err: nil, body: `{"id":"1"}`},
			{url: regexp.MustCompile(`/api/network-servers\?limit=100`), method: "GET", err: nil, body: `{"result":[]}`},
			{url: regexp.MustCompile(`/api/network-servers`), method: "POST", err: errors.New("fake error post network-servers"), body: `{"id":"1"}`},
		},
	}, {
		name:    "error service-profiles",
		inError: true,
		data: []fakeHTTPClientData{
			{url: regexp.MustCompile(`/api/internal/login`), method: "POST", err: nil, body: `{"jwt":"1"}`},
			{url: regexp.MustCompile(`/api/organizations\?limit=100`), method: "GET", err: nil, body: `{"result":[]}`},
			{url: regexp.MustCompile(`/api/organizations`), method: "POST", err: nil, body: `{"id":"1"}`},
			{url: regexp.MustCompile(`/api/network-servers\?limit=100`), method: "GET", err: nil, body: `{"result":[]}`},
			{url: regexp.MustCompile(`/api/network-servers`), method: "POST", err: nil, body: `{"id":"1"}`},
			{url: regexp.MustCompile(`/api/service-profiles`), method: "POST", err: errors.New("fake error post service-profile"), body: `{"serviceProfileID":"1"}`},
		},
	}, {
		name:    "error application get",
		inError: true,
		data: []fakeHTTPClientData{
			{url: regexp.MustCompile(`/api/internal/login`), method: "POST", err: nil, body: `{"jwt":"1"}`},
			{url: regexp.MustCompile(`/api/organizations\?limit=100`), method: "GET", err: nil, body: `{"result":[]}`},
			{url: regexp.MustCompile(`/api/organizations`), method: "POST", err: nil, body: `{"id":"1"}`},
			{url: regexp.MustCompile(`/api/network-servers\?limit=100`), method: "GET", err: nil, body: `{"result":[]}`},
			{url: regexp.MustCompile(`/api/network-servers`), method: "POST", err: nil, body: `{"id":"1"}`},
			{url: regexp.MustCompile(`/api/service-profiles`), method: "POST", err: nil, body: `{"serviceProfileID":"1"}`},
			{url: regexp.MustCompile(`/api/applications\?limit=100`), method: "GET", err: errors.New("fake error get applications"), body: `{"result":[]}`},
		},
	}, {
		name:    "error application post",
		inError: true,
		data: []fakeHTTPClientData{
			{url: regexp.MustCompile(`/api/internal/login`), method: "POST", err: nil, body: `{"jwt":"1"}`},
			{url: regexp.MustCompile(`/api/organizations\?limit=100`), method: "GET", err: nil, body: `{"result":[]}`},
			{url: regexp.MustCompile(`/api/organizations`), method: "POST", err: nil, body: `{"id":"1"}`},
			{url: regexp.MustCompile(`/api/network-servers\?limit=100`), method: "GET", err: nil, body: `{"result":[]}`},
			{url: regexp.MustCompile(`/api/network-servers`), method: "POST", err: nil, body: `{"id":"1"}`},
			{url: regexp.MustCompile(`/api/service-profiles`), method: "POST", err: nil, body: `{"serviceProfileID":"1"}`},
			{url: regexp.MustCompile(`/api/applications\?limit=100`), method: "GET", err: nil, body: `{"result":[]}`},
			{url: regexp.MustCompile(`/api/applications`), method: "POST", err: errors.New("fake error post application"), body: `{"id":"1"}`},
		},
	}, {
		name:    "error post devices-profiles",
		inError: true,
		data: []fakeHTTPClientData{
			{url: regexp.MustCompile(`/api/internal/login`), method: "POST", err: nil, body: `{"jwt":"1"}`},
			{url: regexp.MustCompile(`/api/organizations\?limit=100`), method: "GET", err: nil, body: `{"result":[]}`},
			{url: regexp.MustCompile(`/api/organizations`), method: "POST", err: nil, body: `{"id":"1"}`},
			{url: regexp.MustCompile(`/api/network-servers\?limit=100`), method: "GET", err: nil, body: `{"result":[]}`},
			{url: regexp.MustCompile(`/api/network-servers`), method: "POST", err: nil, body: `{"id":"1"}`},
			{url: regexp.MustCompile(`/api/service-profiles`), method: "POST", err: nil, body: `{"serviceProfileID":"1"}`},
			{url: regexp.MustCompile(`/api/applications\?limit=100`), method: "GET", err: nil, body: `{"result":[]}`},
			{url: regexp.MustCompile(`/api/applications`), method: "POST", err: nil, body: `{"id":"1"}`},
			{url: regexp.MustCompile(`/api/device-profiles`), method: "POST", err: errors.New("fake error post device-profiles"), body: `{"deviceProfileID":"1"}`},
		},
	}, {
		name:    "error post device must not return error (because lot of devices)",
		inError: false,
		data: []fakeHTTPClientData{
			{url: regexp.MustCompile(`/api/internal/login`), method: "POST", err: nil, body: `{"jwt":"1"}`},
			{url: regexp.MustCompile(`/api/organizations\?limit=100`), method: "GET", err: nil, body: `{"result":[]}`},
			{url: regexp.MustCompile(`/api/organizations`), method: "POST", err: nil, body: `{"id":"1"}`},
			{url: regexp.MustCompile(`/api/network-servers\?limit=100`), method: "GET", err: nil, body: `{"result":[]}`},
			{url: regexp.MustCompile(`/api/network-servers`), method: "POST", err: nil, body: `{"id":"1"}`},
			{url: regexp.MustCompile(`/api/service-profiles`), method: "POST", err: nil, body: `{"serviceProfileID":"1"}`},
			{url: regexp.MustCompile(`/api/applications\?limit=100`), method: "GET", err: nil, body: `{"result":[]}`},
			{url: regexp.MustCompile(`/api/applications`), method: "POST", err: nil, body: `{"id":"1"}`},
			{url: regexp.MustCompile(`/api/device-profiles`), method: "POST", err: nil, body: `{"deviceProfileID":"1"}`},
			{url: regexp.MustCompile(`/api/devices`), method: "POST", err: errors.New("fake error post devices"), body: ``},
		},
	}, {
		name:    "error post device keys must not return error (because lot of devices)",
		inError: false,
		data: []fakeHTTPClientData{
			{url: regexp.MustCompile(`/api/internal/login`), method: "POST", err: nil, body: `{"jwt":"1"}`},
			{url: regexp.MustCompile(`/api/organizations\?limit=100`), method: "GET", err: nil, body: `{"result":[]}`},
			{url: regexp.MustCompile(`/api/organizations`), method: "POST", err: nil, body: `{"id":"1"}`},
			{url: regexp.MustCompile(`/api/network-servers\?limit=100`), method: "GET", err: nil, body: `{"result":[]}`},
			{url: regexp.MustCompile(`/api/network-servers`), method: "POST", err: nil, body: `{"id":"1"}`},
			{url: regexp.MustCompile(`/api/service-profiles`), method: "POST", err: nil, body: `{"serviceProfileID":"1"}`},
			{url: regexp.MustCompile(`/api/applications\?limit=100`), method: "GET", err: nil, body: `{"result":[]}`},
			{url: regexp.MustCompile(`/api/applications`), method: "POST", err: nil, body: `{"id":"1"}`},
			{url: regexp.MustCompile(`/api/device-profiles`), method: "POST", err: nil, body: `{"deviceProfileID":"1"}`},
			{url: regexp.MustCompile(`/api/devices`), method: "POST", err: nil, body: ``},
			{url: regexp.MustCompile(`/api/devices/[^/]+/keys`), method: "POST", err: errors.New("fake error post devices keys"), body: ``},
		},
	},
}

func (f fakeHTTPClient) Do(req *http.Request) (*http.Response, error) {
	for _, d := range f.data {
		if d.url.Match([]byte(req.URL.String())) && d.method == req.Method {
			return &http.Response{
				Body:       nopCloser{bytes.NewReader([]byte(d.body))},
				StatusCode: 200,
			}, d.err
		}
	}
	return nil, fmt.Errorf("Unknown response for url %s", req.URL.String())
}

func TestNewLoraserverError(t *testing.T) {
	raw := json.RawMessage([]byte("{"))
	l, err := newLoraserver(raw)
	if err == nil {
		t.Fatal("bad json should throw error")
	}
	if l != nil {
		t.Fatal("bad json should not return loraserver client")
	}
}

func TestNewLoraserver(t *testing.T) {
	l := newDefautlLoraserver()
	if l == nil {
		t.Fatal("good json should return loraserver client")
	}
}

func TestProvisioning(t *testing.T) {
	t.Parallel()
	for _, d := range data {
		t.Run(d.name, func(t *testing.T) {
			l := newDefautlLoraserver()
			l.httpClient = fakeHTTPClient{data: d.data}
			err := l.Provision(model.Register{ScenarioUUID: "", Gateways: newGateways(5, 2)})
			if !d.inError && err != nil {
				t.Fatalf("Provisioining return %s", err.Error())
			} else if d.inError && err == nil {
				t.Fatal("Should return error")
			}
		})
	}
}

func TestDeprovisioningLoraserver(t *testing.T) {
	l := newDefautlLoraserver()
	l.APIURL = "/"
	l.httpClient = fakeHTTPClient{}

	err := l.DeProvision()
	if err != nil {
		t.Fatal("Deprovsion should not throw error")
	}
}
