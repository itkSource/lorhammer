package provisioning

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"lorhammer/src/model"
	"net/http"
	"strconv"

	"github.com/Sirupsen/logrus"
)

var LOG_LORASERVER = logrus.WithField("logger", "orchestrator/provisioning/loraserver")

const LoraserverType = Type("loraserver")

type Loraserver struct {
	ApiUrl              string `json:"apiUrl"`
	Abp                 bool   `json:"abp"`
	Login               string `json:"login"`
	Password            string `json:"password"`
	AppId               string `json:"appId"`
	cleanedProvisioning bool
	JwtToKen            string
}

func NewLoraserver(rawConfig json.RawMessage) (provisioner, error) {
	config := &Loraserver{
		cleanedProvisioning: false,
	}
	if err := json.Unmarshal(rawConfig, config); err != nil {
		return nil, err
	}

	return config, nil
}

func (loraserver *Loraserver) Provision(sensorsToRegister model.Register) error {

	loginReq := LoginRequest{
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
	var loginResp = new(LoginResponse)
	err = json.Unmarshal(responseBody, &loginResp)
	if err != nil {
		return err
	}
	loraserver.JwtToKen = loginResp.Jwt

	if loraserver.AppId == "" {
		//TODO : create organization before the app for the test to be totally stateless
		asApp := AsApp{
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

		LOG_LORASERVER.WithField("appName", asApp.Name).Info("Creating app in loraserver AS")

		marshalledApp, err := json.Marshal(asApp)
		if err != nil {
			return err
		}

		responseBody, err = doRequest(loraserver.ApiUrl+"/api/applications", "POST", marshalledApp, loraserver.JwtToKen)
		if err != nil {
			return err
		}
		var creationResponse = new(CreationResponse)
		err = json.Unmarshal(responseBody, &creationResponse)
		if err != nil {
			return err
		}
		loraserver.AppId = creationResponse.Id
	}

	idNode := 0
	for _, gateway := range sensorsToRegister.Gateways {

		for _, sensor := range gateway.Nodes {
			asnode := AsNode{
				DevEUI:        sensor.DevEUI.String(),
				AppEUI:        sensor.AppEUI.String(),
				AppKey:        sensor.AppKey.String(),
				ApplicationID: loraserver.AppId,
				Description:   "stresstool node",
				Name:          "STRESSNODE_" + sensor.DevEUI.String() + "_" + strconv.Itoa(idNode),
				UseApplicationSettings: true,
			}

			LOG_LORASERVER.WithField("name", asnode.Name).Info("Registering sensor")

			if marshalledNode, err := json.Marshal(asnode); err != nil {
				return err
			} else {
				if _, err := doRequest(loraserver.ApiUrl+"/api/nodes", "POST", marshalledNode, loraserver.JwtToKen); err != nil {
					return err
				}
			}

			if loraserver.Abp {
				activation := NodeActivation{
					DevAddr:  sensor.DevAddr.String(),
					AppSKey:  sensor.AppSKey.String(),
					NwkSKey:  sensor.NwSKey.String(),
					FCntUp:   0,
					FCntDown: 0,
					DevEUI:   asnode.DevEUI,
				}
				marshalledActivation, err := json.Marshal(activation)
				if err != nil {
					LOG_LORASERVER.Panic(err)

				}
				doRequest(loraserver.ApiUrl+"/api/nodes/"+asnode.DevEUI+"/activation", "POST", marshalledActivation, loraserver.JwtToKen)
			}

			idNode++
		}
	}
	return nil
}

func doRequest(url string, method string, marshalledObject []byte, jwtToken string) ([]byte, error) {
	LOG_LORASERVER.WithField("url", url).Info("Will call")
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
		LOG_LORASERVER.WithField("url", url).Info("Call succeded")

	default:
		LOG_LORASERVER.WithFields(logrus.Fields{
			"respStatus":   resp.StatusCode,
			"responseBody": string(body),
			"url":          url,
		}).Warn("Couldn't proceed with request")
		return nil, errors.New("Couldn't proceed with request")
	}

	return body, nil
}

func (loraserver *Loraserver) DeProvision() error {
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

type AsNode struct {
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

type NodeActivation struct {
	AppSKey  string `json:"appSKey"`
	DevAddr  string `json:"devAddr"`
	DevEUI   string `json:"devEUI"`
	FCntDown int    `json:"fCntDown"`
	FCntUp   int    `json:"fCntUp"`
	NwkSKey  string `json:"nwkSKey"`
}

type AsApp struct {
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

type RequestHeader struct {
	Alg string `json:"alg"`
	Typ string `json:"typ"`
}

// Claims defines the struct containing the token claims.
type LoginResponse struct {
	Jwt string `json:"jwt"`
}

type LoginRequest struct {
	Login    string `json:"username"`
	Password string `json:"password"`
}

type CreationResponse struct {
	Id string `json:"id"`
}
