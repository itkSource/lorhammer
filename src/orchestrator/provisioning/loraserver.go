package provisioning

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"lorhammer/src/model"
	"net"
	"net/http"
	"time"

	"github.com/sirupsen/logrus"
)

var logLoraserver = logrus.WithField("logger", "orchestrator/provisioning/loraserver")

const loraserverType = Type("loraserver")

type loraserver struct {
	APIURL                string `json:"apiUrl"`
	Abp                   bool   `json:"abp"`
	Login                 string `json:"login"`
	Password              string `json:"password"`
	AppID                 string `json:"appId"`
	NbProvisionerParallel int    `json:"nbProvisionerParallel"`

	doRequest func(url string, method string, marshalledObject []byte, jwtToken string) ([]byte, error)
	jwtToKen  string
}

func newLoraserver(rawConfig json.RawMessage) (provisioner, error) {
	config := &loraserver{
		doRequest: doRequest,
	}
	if err := json.Unmarshal(rawConfig, config); err != nil {
		return nil, err
	}

	return config, nil
}

func (loraserver *loraserver) Provision(sensorsToRegister model.Register) error {

	if loraserver.jwtToKen == "" {
		loginReq := loginRequest{
			Login:    loraserver.Login,
			Password: loraserver.Password,
		}

		marshalledLogin, err := json.Marshal(loginReq)
		if err != nil {
			return err
		}

		responseBody, err := loraserver.doRequest(loraserver.APIURL+"/api/internal/login", "POST", marshalledLogin, "")
		if err != nil {
			return err
		}
		var loginResp = new(loginResponse)
		err = json.Unmarshal(responseBody, &loginResp)
		if err != nil {
			return err
		}
		loraserver.jwtToKen = loginResp.Jwt
	}

	if loraserver.AppID == "" {
		//TODO : create organization before the app for the test to be totally stateless
		asApp := asApp{
			Name:           "stress-app",
			Description:    "stress-app",
			Rx1DROffset:    0,
			Rx2DR:          0,
			RxDelay:        0,
			RxWindow:       "RX1",
			IsABP:          loraserver.Abp,
			AdrInterval:    0,
			OrganizationID: "1",
		}

		logLoraserver.WithField("appName", asApp.Name).Info("Creating app in loraserver AS")

		marshalledApp, err := json.Marshal(asApp)
		if err != nil {
			return err
		}

		responseBody, err := loraserver.doRequest(loraserver.APIURL+"/api/applications", "POST", marshalledApp, loraserver.jwtToKen)
		if err != nil {
			return err
		}
		var creationResponse = new(creationResponse)
		err = json.Unmarshal(responseBody, &creationResponse)
		if err != nil {
			return err
		}
		loraserver.AppID = creationResponse.ID
	}

	nbNodeToProvision := 0
	for _, gateway := range sensorsToRegister.Gateways {
		for range gateway.Nodes {
			nbNodeToProvision++
		}
	}

	sensorChan := make(chan *model.Node, nbNodeToProvision)
	defer close(sensorChan)
	poison := make(chan bool, loraserver.NbProvisionerParallel)
	defer close(poison)
	errorChan := make(chan error)
	defer close(errorChan)
	sensorFinishChan := make(chan *model.Node)
	defer close(sensorFinishChan)

	for i := 0; i < loraserver.NbProvisionerParallel; i++ {
		go loraserver.provisionSensorAsync(sensorChan, poison, errorChan, sensorFinishChan)
	}

	go func() {
		for _, gateway := range sensorsToRegister.Gateways {
			gateway := gateway
			for _, sensor := range gateway.Nodes {
				sensor := sensor
				sensorChan <- sensor
			}
		}
	}()

	for i := 0; i < nbNodeToProvision; i++ {
		select {
		case err := <-errorChan:
			logLoraserver.WithError(err).Error("Node not provisioned")
		case sensor := <-sensorFinishChan:
			logLoraserver.WithField("node", sensor).Debug("Node provisioned")
		}
	}

	for i := 0; i < loraserver.NbProvisionerParallel; i++ {
		poison <- true
	}

	return nil
}

func (loraserver *loraserver) provisionSensorAsync(sensorChan chan *model.Node, poison chan bool, errorChan chan error, sensorFinishChan chan *model.Node) {
	exit := false
	for {
		select {
		case sensor := <-sensorChan:
			if sensor != nil { // Why i received nil sometimes !?
				asnode := asNode{
					DevEUI:        sensor.DevEUI.String(),
					AppEUI:        sensor.AppEUI.String(),
					AppKey:        sensor.AppKey.String(),
					ApplicationID: loraserver.AppID,
					Description:   "stresstool node",
					Name:          "STRESSNODE_" + sensor.DevEUI.String(),
					UseApplicationSettings: true,
				}

				logLoraserver.WithField("name", asnode.Name).Debug("Registering sensor")

				if marshalledNode, err := json.Marshal(asnode); err != nil {
					logLoraserver.WithField("asnode", asnode).WithError(err).Error("Can't marshall asnode")
					errorChan <- err
					break
				} else {
					if _, err := loraserver.doRequest(loraserver.APIURL+"/api/nodes", "POST", marshalledNode, loraserver.jwtToKen); err != nil {
						logLoraserver.WithField("marshalledNode", string(marshalledNode)).WithError(err).Error("Can't provision node")
						errorChan <- err
						break
					}
				}

				if loraserver.Abp {
					activation := nodeActivation{
						DevAddr:  sensor.DevAddr.String(),
						AppSKey:  sensor.AppSKey.String(),
						NwkSKey:  sensor.NwSKey.String(),
						FCntUp:   0,
						FCntDown: 0,
						DevEUI:   asnode.DevEUI,
					}
					if marshalledActivation, err := json.Marshal(activation); err != nil {
						logLoraserver.WithError(err).Error("Can't marshal abp node")
						errorChan <- err
						break
					} else {
						url := loraserver.APIURL + "/api/nodes/" + asnode.DevEUI + "/activation"
						if _, errRequest := loraserver.doRequest(url, "POST", marshalledActivation, loraserver.jwtToKen); errRequest != nil {
							logLoraserver.WithError(errRequest).Error("Can't activate abp node")
							errorChan <- errRequest
							break
						}
					}
				}
				sensorFinishChan <- sensor
			}
		case <-poison:
			exit = true
		}
		if exit {
			break
		}
	}

}

func doRequest(url string, method string, marshalledObject []byte, jwtToken string) ([]byte, error) {
	logLoraserver.WithField("url", url).Debug("Will call")
	httpClient := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
			Dial: (&net.Dialer{
				Timeout: 5 * time.Second,
			}).Dial,
			TLSHandshakeTimeout: 5 * time.Second,
		},
		Timeout: 5 * time.Second,
	}
	ctx, cancelCtx := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelCtx()

	req, err := http.NewRequest(method, url, bytes.NewBuffer(marshalledObject))
	if err != nil {
		return nil, err
	}

	if jwtToken != "" {
		req.Header.Set("Grpc-Metadata-Authorization", jwtToken)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Close = true
	req.WithContext(ctx)

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, _ := ioutil.ReadAll(resp.Body)

	switch resp.StatusCode {
	case http.StatusOK:
		logLoraserver.WithField("url", url).Debug("Call succeded")

	default:
		logLoraserver.WithFields(logrus.Fields{
			"respStatus":   resp.StatusCode,
			"responseBody": string(body),
			"url":          url,
		}).Warn("Couldn't proceed with request")
		return nil, errors.New("Couldn't proceed with request")
	}

	return body, nil
}

func (loraserver *loraserver) DeProvision() error {
	if loraserver.APIURL != "" && loraserver.AppID != "" {
		if _, err := loraserver.doRequest(loraserver.APIURL+"/api/applications/"+loraserver.AppID, "DELETE", nil, loraserver.jwtToKen); err != nil {
			return err
		}
	} else {
		return fmt.Errorf("ApiUrl (%s) and appId (%s) can not be empty", loraserver.APIURL, loraserver.AppID)
	}
	return nil
}

type asNode struct {
	DevEUI                 string `json:"devEUI"`
	AppEUI                 string `json:"appEUI"`
	AppKey                 string `json:"appKey"`
	AppsKey                string `json:"appsKey"`
	NWsKey                 string `json:"NWsKey"`
	ApplicationID          string `json:"applicationID"`
	Description            string `json:"description"`
	Name                   string `json:"name"`
	UseApplicationSettings bool   `json:"useApplicationSettings"`
}

type nodeActivation struct {
	AppSKey  string `json:"appSKey"`
	DevAddr  string `json:"devAddr"`
	DevEUI   string `json:"devEUI"`
	FCntDown int    `json:"fCntDown"`
	FCntUp   int    `json:"fCntUp"`
	NwkSKey  string `json:"nwkSKey"`
}

type asApp struct {
	Name           string `json:"name"`
	Description    string `json:"description"`
	IsABP          bool   `json:"isABP"`
	Rx1DROffset    int    `json:"rx1DROffset"`
	Rx2DR          int    `json:"rx2DR"`
	RxDelay        int    `json:"rxDelay"`
	RxWindow       string `json:"rxWindow"`
	AdrInterval    int    `json:"adrInterval"`
	OrganizationID string `json:"organizationID"`
}

type requestHeader struct {
	Alg string `json:"alg"`
	Typ string `json:"typ"`
}

// Claims defines the struct containing the token claims.
type loginResponse struct {
	Jwt string `json:"jwt"`
}

type loginRequest struct {
	Login    string `json:"username"`
	Password string `json:"password"`
}

type creationResponse struct {
	ID string `json:"id"`
}
