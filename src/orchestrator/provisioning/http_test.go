package provisioning

import (
	"encoding/json"
	"errors"
	"io"
	"io/ioutil"
	"lorhammer/src/model"
	"net/http"
	"testing"
)

func TestNilConfig(t *testing.T) {
	httpProvisioner, err := newHTTPProvisioner(nil)

	if err == nil {
		t.Fatal("Error expected not to be nil")
	}

	if httpProvisioner != nil {
		t.Fatal("HTTP provisioner expected to be nil")
	}
}

func TestBadConfig(t *testing.T) {
	badConfig := json.RawMessage("Wrong JSON string")
	httpProvisioner, err := newHTTPProvisioner(badConfig)

	if err == nil {
		t.Fatal("Error expected not to be nil")
	}

	if httpProvisioner != nil {
		t.Fatal("HTTP provisioner expected to be nil")
	}
}

func TestGoodConfig(t *testing.T) {
	goodConfig := json.RawMessage(`{"creationApiUrl": "createURL", "deletionApiUrl": "deleteURL"}`)
	provisioner, err := newHTTPProvisioner(goodConfig)

	if err != nil {
		t.Fatal("Error expected to be nil")
	}

	if provisioner == nil {
		t.Fatal("HTTP provisioner expected to be not nil")
	}

	httpProv := provisioner.(*httpProvisoner)

	if httpProv.CreationAPIURL != "createURL" {
		t.Log(httpProv.CreationAPIURL)
		t.Fatal("Wrong returned creation API URL")
	}

	if httpProv.DeletionAPIURL != "deleteURL" {
		t.Log(httpProv.DeletionAPIURL)
		t.Fatal("Wrong returned deletion API URL")
	}
}

func TestNilProvision(t *testing.T) {
	httpProv := httpProvisoner{
		CreationAPIURL: "testURL",
		post: func(url string, contentType string, body io.Reader) (resp *http.Response, err error) {
			if url != "testURL" {
				t.Log(url)
				t.Fatal("Wrong creation URL API")
			}
			return nil, errors.New("Test error")
		}}
	var sensorsToRegister model.Register
	//sensorsToRegister = nil
	err := httpProv.Provision(sensorsToRegister)
	if err == nil {
		t.Fatal("Error expected not to be nil")
	}
}

func TestWrongStatusCodeProvision(t *testing.T) {
	httpProv := httpProvisoner{CreationAPIURL: "testURL",
		post: func(url string, contentType string, body io.Reader) (resp *http.Response, err error) {
			if url != "testURL" {
				t.Log(url)
				t.Fatal("Wrong creation URL API")
			}
			return &http.Response{
				StatusCode: http.StatusInternalServerError,
				Body:       ioutil.NopCloser(nil),
			}, nil
		}}
	var sensorsToRegister model.Register
	err := httpProv.Provision(sensorsToRegister)
	if err == nil {
		t.Fatal("Error expected not to be nil")
	}
	if len(httpProv.sensorsRegistered) != 0 {
		t.Fatal("Registered sensors expected to be 0")
	}
}

func TestGoodProvision(t *testing.T) {
	httpProv := httpProvisoner{
		CreationAPIURL: "testURL",
		post: func(url string, contentType string, body io.Reader) (resp *http.Response, err error) {
			if url != "testURL" {
				t.Log(url)
				t.Fatal("Wrong creation URL API")
			}
			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       ioutil.NopCloser(nil),
			}, nil
		}}
	sensorsToRegister := model.Register{}
	err := httpProv.Provision(sensorsToRegister)
	if err != nil {
		t.Fatal("Error expected not to be nil")
	}
	if len(httpProv.sensorsRegistered) != 1 {
		t.Fatal("Registered sensors expected to be 1")
	}
}

func TestEmptyDeProvision(t *testing.T) {
	httpProv := httpProvisoner{}
	err := httpProv.DeProvision()
	if err != nil {
		t.Fatal("Error expected to be nil")
	}
}

func TestWrongStatusCodeDeProvision(t *testing.T) {
	httpProv := httpProvisoner{
		DeletionAPIURL:    "testURL",
		sensorsRegistered: []model.Register{{ScenarioUUID: "1"}},
		post: func(url string, contentType string, body io.Reader) (resp *http.Response, err error) {
			if url != "testURL" {
				t.Log(url)
				t.Fatal("Wrong deletion URL API")
			}
			return &http.Response{
				StatusCode: http.StatusInternalServerError,
				Body:       ioutil.NopCloser(nil),
			}, nil
		},
	}
	err := httpProv.DeProvision()
	if err == nil {
		t.Fatal("Error expected not to be nil")
	}
}

func TestDeProvisionOK(t *testing.T) {
	httpProv := httpProvisoner{
		DeletionAPIURL:    "testURL",
		sensorsRegistered: []model.Register{{ScenarioUUID: "1"}},
		post: func(url string, contentType string, body io.Reader) (resp *http.Response, err error) {
			if url != "testURL" {
				t.Log(url)
				t.Fatal("Wrong deletion URL API")
			}
			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       ioutil.NopCloser(nil),
			}, nil
		},
	}
	err := httpProv.DeProvision()
	if err != nil {
		t.Fatal("Error expected to be nil")
	}
}
