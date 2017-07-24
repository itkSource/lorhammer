package provisioning

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/Sirupsen/logrus"
	"io/ioutil"
	"lorhammer/src/model"
	"net/http"
)

var log_loraserver = logrus.WithField("logger", "orchestrator/provisioning/loraserver")

const loraserverType = Type("loraserver")

type loraserver struct {
	ApiUrl              string `json:"apiUrl"`
	Abp                 bool   `json:"abp"`
	Login               string `json:"login"`
	Password            string `json:"password"`
	AppId               string `json:"appId"`
	cleanedProvisioning bool
	JwtToKen            string
}

func newLoraserver(rawConfig json.RawMessage) (provisioner, error) {
	config := &loraserver{
		cleanedProvisioning: false,
	}
	if err := json.Unmarshal(rawConfig, config); err != nil {
		return nil, err
	}

	return config, nil
}

func (loraserver *loraserver) Provision(sensorsToRegister model.Register) error {

	loginReq := loginRequest{
		Login:    loraserver.Login,
		Password: loraserver.Password,
	}

	marshalledLogin, err := json.Marshal(loginReq)
	if err != nil {
		return err
	}

	responseBody, err := doRequest(loraserver.ApiUrl+"/api/internal/login", "POST", marshalledLogin, "")
	if err != nil {
		return err
	}
	var loginResp = new(loginResponse)
	err = json.Unmarshal(responseBody, &loginResp)
	if err != nil {
		return err
	}
	loraserver.JwtToKen = loginResp.Jwt

	if loraserver.AppId == "" {
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
			OrganizationId: "1",
		}

		log_loraserver.WithField("appName", asApp.Name).Info("Creating app in loraserver AS")

		marshalledApp, err := json.Marshal(asApp)
		if err != nil {
			return err
		}

		responseBody, err = doRequest(loraserver.ApiUrl+"/api/applications", "POST", marshalledApp, loraserver.JwtToKen)
		if err != nil {
			return err
		}
		var creationResponse = new(creationResponse)
		err = json.Unmarshal(responseBody, &creationResponse)
		if err != nil {
			return err
		}
		loraserver.AppId = creationResponse.Id
	}

	for _, gateway := range sensorsToRegister.Gateways {

		for _, sensor := range gateway.Nodes {

			go func(gateway model.Gateway, sensor *model.Node) {
				asnode := asNode{
					DevEUI:        sensor.DevEUI.String(),
					AppEUI:        sensor.AppEUI.String(),
					AppKey:        sensor.AppKey.String(),
					ApplicationID: loraserver.AppId,
					Description:   "stresstool node",
					Name:          "STRESSNODE_" + sensor.DevEUI.String(),
					UseApplicationSettings: true,
				}

				log_loraserver.WithField("name", asnode.Name).Info("Registering sensor")

				if marshalledNode, err := json.Marshal(asnode); err != nil {
					log_loraserver.WithField("asnode", asnode).Error("Can't marshall asnode")
					return
				} else {
					if _, err := doRequest(loraserver.ApiUrl+"/api/nodes", "POST", marshalledNode, loraserver.JwtToKen); err != nil {
						log_loraserver.WithField("marshalledNode", marshalledNode).Error("Can't provision node")
						return
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
					marshalledActivation, err := json.Marshal(activation)
					if err != nil {
						log_loraserver.Panic(err)

					}
					doRequest(loraserver.ApiUrl+"/api/nodes/"+asnode.DevEUI+"/activation", "POST", marshalledActivation, loraserver.JwtToKen)
				}

			}(gateway, sensor)
		}
	}
	return nil
}

func doRequest(url string, method string, marshalledObject []byte, jwtToken string) ([]byte, error) {
	log_loraserver.WithField("url", url).Info("Will call")
	httpClient := &http.Client{Transport: &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}}}

	req, err := http.NewRequest(method, url, bytes.NewBuffer(marshalledObject))
	if err != nil {
		return nil, err
	}

	if jwtToken != "" {
		req.Header.Set("Grpc-Metadata-Authorization", jwtToken)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, _ := ioutil.ReadAll(resp.Body)

	switch resp.StatusCode {
	case http.StatusOK:
		log_loraserver.WithField("url", url).Info("Call succeded")

	default:
		log_loraserver.WithFields(logrus.Fields{
			"respStatus":   resp.StatusCode,
			"responseBody": string(body),
			"url":          url,
		}).Warn("Couldn't proceed with request")
		return nil, errors.New("Couldn't proceed with request")
	}

	return body, nil
}

func (loraserver *loraserver) DeProvision() error {
	if !loraserver.cleanedProvisioning {
		if loraserver.ApiUrl != "" && loraserver.AppId != "" {
			if _, err := doRequest(loraserver.ApiUrl+"/api/applications/"+loraserver.AppId, "DELETE", nil, loraserver.JwtToKen); err != nil {
				return err
			}
			loraserver.cleanedProvisioning = true
		} else {
			return fmt.Errorf("ApiUrl (%s) and appId (%s) can not be empty", loraserver.ApiUrl, loraserver.AppId)
		}
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
	OrganizationId string `json:"organizationID"`
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
	Id string `json:"id"`
}
