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
	"strconv"
)

var LOG_LORASERVER = logrus.WithField("logger", "orchestrator/provisioning/loraserver")

const LoraserverType = Type("brocaar")

type Loraserver struct {
	ApiUrl              string `json:"apiUrl"`
	Abp                 bool   `json:"abp"`
	AppsKey             string `json:"appskey"`
	Nwskey              string `json:"nwskey"`
	appId               string
	cleanedProvisioning bool
}

func NewLoraserver(rawConfig json.RawMessage) (provisioner, error) {
	config := &Loraserver{
		appId:               "",
		cleanedProvisioning: false,
	}
	if err := json.Unmarshal(rawConfig, config); err != nil {
		return nil, err
	}

	return config, nil
}

func (loraserver *Loraserver) Provision(sensorsToRegister model.Register) error {

	//TODO add is Abp field to be compliant with last Appserver version
	asApp := AsApp{
		Name:        "stress-app",
		Description: "stress-app",
	}

	LOG_LORASERVER.WithFields(logrus.Fields{
		"ref":     "orchestrator/brocaarProvisionner:provisionBrocaarApplicationServer()",
		"appName": asApp.Name,
	}).Info("Creating app in brocaar AS")

	marshalledApp, err := json.Marshal(asApp)
	if err != nil {
		return err
	}

	responseBody, err := doRequest(loraserver.ApiUrl+"/api/applications", "POST", marshalledApp)
	if err != nil {
		return err
	}
	var creationResponse = new(CreationResponse)
	err = json.Unmarshal(responseBody, &creationResponse)
	if err != nil {
		return err
	}
	loraserver.appId = creationResponse.Id

	idNode := 0
	for _, gateway := range sensorsToRegister.Gateways {

		for _, sensor := range gateway.Nodes {

			LOG_LORASERVER.WithFields(logrus.Fields{
				"name": "STRESSNODE" + strconv.Itoa(idNode),
			}).Info("Registering sensor")

			//TODO add useApplicationSettings field to be compliant with last Appserver version
			asnode := AsNode{
				DevEUI:        sensor.DevEUI.String(),
				AppEUI:        sensor.AppEUI.String(),
				AppKey:        sensor.AppKey.String(),
				AdrInterval:   0,
				ApplicationID: creationResponse.Id,
				Description:   "stresstool node",
				Name:          "STRESSNODE" + strconv.Itoa(idNode),
				Rx1DROffset:   0,
				Rx2DR:         0,
				RxDelay:       0,
				RxWindow:      "RX1",
				IsABP:         loraserver.Abp,
			}
			if marshalledNode, err := json.Marshal(asnode); err != nil {
				return err
			} else {
				if _, err := doRequest(loraserver.ApiUrl+"/api/nodes", "POST", marshalledNode); err != nil {
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
				doRequest(loraserver.ApiUrl+"/api/nodes/"+asnode.DevEUI+"/activation", "POST", marshalledActivation)
			}

			idNode++
		}
	}
	return nil
}

func doRequest(url string, method string, marshalledObject []byte) ([]byte, error) {
	httpClient := &http.Client{Transport: &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}}}

	req, err := http.NewRequest(method, url, bytes.NewBuffer(marshalledObject))
	if err != nil {
		return nil, err
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
		if loraserver.ApiUrl != "" && loraserver.appId != "" {
			if _, err := doRequest(loraserver.ApiUrl+"/api/applications/"+loraserver.appId, "DELETE", nil); err != nil {
				return err
			}
			loraserver.cleanedProvisioning = true
		} else {
			return fmt.Errorf("ApiUrl (%s) and appId (%s) can not be empty", loraserver.ApiUrl, loraserver.appId)
		}
	}
	return nil
}

type AsNode struct {
	DevEUI        string `json:"devEUI"`
	AppEUI        string `json:"appEUI"`
	AppKey        string `json:"appKey"`
	AppsKey       string `json:"appsKey"`
	NWsKey        string `json:"NWsKey"`
	AdrInterval   int    `json:"adrInterval"`
	ApplicationID string `json:"applicationID"`
	Description   string `json:"description"`
	Name          string `json:"name"`
	Rx1DROffset   int    `json:"rx1DROffset"`
	Rx2DR         int    `json:"rx2DR"`
	RxDelay       int    `json:"rxDelay"`
	RxWindow      string `json:"rxWindow"`
	IsABP         bool   `json:"isABP"`
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
	Name        string `json:"name"`
	Description string `json:"description"`
}

type CreationResponse struct {
	Id string `json:"id"`
}
