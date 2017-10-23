package provisioning

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"lorhammer/src/model"
	"net/http"
	"time"

	"github.com/sirupsen/logrus"
)

var logHTTPProvisioner = logrus.WithField("logger", "orchestrator/provisioning/http")

const httpType = Type("http")

type httpProvisoner struct {
	CreationAPIURL    string `json:"creationApiUrl"`
	DeletionAPIURL    string `json:"deletionApiUrl"`
	post              func(url string, contentType string, body io.Reader) (resp *http.Response, err error)
	sensorsRegistered []model.Register
}

func newHTTPProvisioner(rawConfig json.RawMessage) (provisioner, error) {
	timeout := time.Duration(5 * time.Second)
	client := http.Client{
		Timeout: timeout,
	}
	config := &httpProvisoner{
		post: client.Post,
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
	resp, err := httpProvisioner.post(httpProvisioner.CreationAPIURL, "application/json", bytes.NewReader(byteData))
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if !(resp.StatusCode >= http.StatusOK && resp.StatusCode < http.StatusMultipleChoices) {
		logHTTPProvisioner.WithFields(logrus.Fields{
			"errorCode": resp.StatusCode,
			"errorName": resp.Status,
			"dataPost":  string(byteData),
		}).Error("Wrong return code in HTTP provisioning")
		err = errors.New("Wrong return code")
		return err
	}

	httpProvisioner.sensorsRegistered = append(httpProvisioner.sensorsRegistered, sensorsToRegister)
	return nil
}

func (httpProvisioner *httpProvisoner) DeProvision() error {
	for _, sensorsRegistered := range httpProvisioner.sensorsRegistered {
		byteData, err := json.Marshal(sensorsRegistered)
		if err != nil {
			return err
		}
		resp, err := httpProvisioner.post(httpProvisioner.DeletionAPIURL, "application/json", bytes.NewReader(byteData))
		if err != nil {
			return err
		}
		defer resp.Body.Close()
		if !(resp.StatusCode >= http.StatusOK && resp.StatusCode < http.StatusMultipleChoices) {
			logHTTPProvisioner.WithFields(logrus.Fields{
				"errorCode": resp.StatusCode,
				"errorName": resp.Status,
				"dataPost":  string(byteData),
			}).Error("Wrong return code in HTTP de-provisioning")
			err = errors.New("Wrong return code")
			return err
		}
	}
	return nil
}
