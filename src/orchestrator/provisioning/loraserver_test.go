package provisioning

import (
	"encoding/json"
	"errors"
	"lorhammer/src/model"
	"lorhammer/src/tools"
	"testing"
	"time"
)

func newLoraserverFromJson(j string) *loraserver {
	raw := json.RawMessage([]byte(j))
	l, _ := newLoraserver(raw)
	return l.(*loraserver)
}

func newDefautlLoraserver() *loraserver {
	return newLoraserverFromJson(`{
	  "apiUrl":"",
	  "abp" : true,
      "login" : "admin",
      "password" : "admin",
      "appId": "19",
      "nbProvisionerParallel": 1
    }`)
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

func newDoRequest(loginError bool, nodeCreationError bool, nodeAbpError bool, deprovisionError bool) func(url string, method string, marshalledObject []byte, jwtToken string) ([]byte, error) {
	return func(url string, method string, marshalledObject []byte, jwtToken string) ([]byte, error) {
		if url == "/api/internal/login" && !loginError {
			return []byte(`{"jwt":"1"}`), nil
		} else if url == "/api/internal/login" && loginError {
			return nil, errors.New("Fake error login")
		} else if url == "/api/nodes" && nodeCreationError {
			return nil, errors.New("Fake error nodeCreation")
		} else if url == "/api/nodes" && !nodeCreationError {
			return []byte(`{}`), nil
		} else if nodeAbpError {
			return nil, errors.New("Fake error abp")
		} else if method == "DELETE" && deprovisionError {
			return nil, errors.New("Fake error deprovision")
		} else {
			return []byte(`{}`), nil
		}
	}
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

func TestNewLoraserverLoginError(t *testing.T) {
	l := newDefautlLoraserver()
	l.doRequest = newDoRequest(true, false, false, false)

	for i := 0; i < 5; i++ {
		err := l.Provision(model.Register{ScenarioUUID: "", Gateways: newGateways(5, 2)})
		if err == nil {
			t.Fatal("Login error return err")
		}
	}
}

func TestNewLoraserverProvision(t *testing.T) {
	l := newDefautlLoraserver()
	l.doRequest = newDoRequest(false, false, false, false)

	for i := 0; i < 5; i++ {
		err := l.Provision(model.Register{ScenarioUUID: "", Gateways: newGateways(50, 100)})
		if err != nil {
			t.Fatal("Good scenario should not return err", err)
		}
	}
}

func TestNewLoraserverProvisionErrorNodeCreation(t *testing.T) {
	l := newDefautlLoraserver()
	l.doRequest = newDoRequest(false, true, false, false)

	for i := 0; i < 5; i++ {
		err := l.Provision(model.Register{ScenarioUUID: "", Gateways: newGateways(2, 1)})
		if err != nil {
			t.Fatal("Good scenario should not return err", err)
		}
	}
}

func TestNewLoraserverProvisionErrorNodeAbp(t *testing.T) {
	l := newDefautlLoraserver()
	l.doRequest = newDoRequest(false, false, true, false)

	for i := 0; i < 5; i++ {
		err := l.Provision(model.Register{ScenarioUUID: "", Gateways: newGateways(2, 1)})
		if err != nil {
			t.Fatal("Good scenario should not return err", err)
		}
	}
}

func TestDeprovisioningLoraserverWithtouUrl(t *testing.T) {
	l := newDefautlLoraserver()
	l.doRequest = newDoRequest(false, false, false, false)

	err := l.DeProvision()
	if err == nil {
		t.Fatal("Deprovsion should throw error when url is empty")
	}
}

func TestDeprovisioningLoraserverError(t *testing.T) {
	l := newDefautlLoraserver()
	l.ApiUrl = "/"
	l.doRequest = newDoRequest(false, false, false, true)

	err := l.DeProvision()
	if err == nil {
		t.Fatal("Deprovision should throw error when loraserver respond error")
	}
}

func TestDeprovisioningLoraserver(t *testing.T) {
	l := newDefautlLoraserver()
	l.ApiUrl = "/"
	l.doRequest = newDoRequest(false, false, false, false)

	err := l.DeProvision()
	if err != nil {
		t.Fatal("Deprovsion should not throw error")
	}
}
