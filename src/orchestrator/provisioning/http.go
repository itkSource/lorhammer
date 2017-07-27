package provisioning

import (
	"bytes"
	"encoding/json"
	"errors"
	"github.com/Sirupsen/logrus"
	"io"
	"lorhammer/src/model"
	"net/http"
	"time"
)

var logHTTPProvisioner = logrus.WithField("logger", "orchestrator/provisioning/http")

const HttpType = Type("http")

type httpProvisoner struct {
	CreationApiURL string `json:"creationApiUrl"`
	DeletionApiURL string `json:"deletionApiUrl"`
	Post           func(url string, contentType string, body io.Reader) (resp *http.Response, err error)
}

var registeredSensorsBytes [][]byte = make([][]byte, 0)

func NewHttpProvisioner(rawConfig json.RawMessage) (provisioner, error) {
	timeout := time.Duration(5 * time.Second)
	client := http.Client{
		Timeout: timeout,
	}
	config := &httpProvisoner{
		Post: client.Post,
	}

	if err := json.Unmarshal(rawConfig, config); err != nil {
		return nil, err
	}
	return config, nil
}

func (httpProvisioner *httpProvisoner) Provision(sensorsToRegister model.Register) error {
	byteData, err := json.Marshal(sensorsToRegister)
	if err != nil {
		return err
	}
	resp, err := httpProvisioner.Post(httpProvisioner.CreationApiURL, "application/json", bytes.NewReader(byteData))
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if !(resp.StatusCode >= http.StatusOK && resp.StatusCode < http.StatusMultipleChoices) {
		logHTTPProvisioner.WithFields(logrus.Fields{
			"errorCode": resp.StatusCode,
			"errorName": resp.Status}).Error("Wrong return code in HTTP provisioning")
		err = errors.New("Wrong return code")
		return err
	}

	registeredSensorsBytes = append(registeredSensorsBytes, byteData)
	return nil
}

func (httpProvisioner *httpProvisoner) DeProvision() error {
	for _, registeredSensorBytes := range registeredSensorsBytes {
		err := httpProvisioner.deleteRequest(registeredSensorBytes)
		if err != nil {
			return err
		}
	}
	return nil
}

func (httpProvisioner *httpProvisoner) deleteRequest(sensorBytes []byte) error {
	resp, err := httpProvisioner.Post(httpProvisioner.DeletionApiURL, "application/json", bytes.NewReader(sensorBytes))
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if !(resp.StatusCode >= http.StatusOK && resp.StatusCode < http.StatusMultipleChoices) {
		logHTTPProvisioner.WithFields(logrus.Fields{
			"errorCode": resp.StatusCode,
			"errorName": resp.Status}).Error("Wrong return code in HTTP de-provisioning")
		err = errors.New("Wrong return code")
		return err
	}
	return nil
}
