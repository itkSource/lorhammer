package provisioning

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"io/ioutil"
	"lorhammer/src/model"
	"net"
	"net/http"
	"time"

	"github.com/sirupsen/logrus"
)

var logLoraserver = logrus.WithField("logger", "orchestrator/provisioning/loraserver")

const (
	loraserverType             = Type("loraserver")
	loraserverOrganisationName = "lorhammer"
	loraserverApplicationName  = "Lorhammer"
	httpLoraserverTimeout      = 1 * time.Minute
)

type httpClientSender interface {
	Do(*http.Request) (*http.Response, error)
}

type loraserver struct {
	APIURL                string `json:"apiUrl"`
	Login                 string `json:"login"`
	Password              string `json:"password"`
	jwtToKen              string
	OrganizationID        string `json:"organizationId"`
	NetworkServerID       string `json:"networkServerId"`
	NetworkServerAddr     string `json:"networkServerAddr"`
	ServiceProfileID      string `json:"serviceProfileID"`
	AppID                 string `json:"appId"`
	DeviceProfileID       string `json:"deviceProfileId"`
	Abp                   bool   `json:"abp"`
	NbProvisionerParallel int    `json:"nbProvisionerParallel"`
	DeleteOrganization    bool   `json:"deleteOrganization"`
	DeleteApplication     bool   `json:"deleteApplication"`

	httpClient httpClientSender
}

func newLoraserver(rawConfig json.RawMessage) (provisioner, error) {
	config := &loraserver{
		httpClient: &http.Client{
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
				Dial: (&net.Dialer{
					Timeout: httpLoraserverTimeout,
				}).Dial,
				TLSHandshakeTimeout: httpLoraserverTimeout,
			},
			Timeout: httpLoraserverTimeout,
		},
	}
	if err := json.Unmarshal(rawConfig, config); err != nil {
		return nil, err
	}

	return config, nil
}

func (loraserver *loraserver) Provision(sensorsToRegister model.Register) error {
	if loraserver.jwtToKen == "" {
		req := struct {
			Login    string `json:"username"`
			Password string `json:"password"`
		}{
			Login:    loraserver.Login,
			Password: loraserver.Password,
		}

		resp := struct {
			Jwt string `json:"jwt"`
		}{}

		err := loraserver.doRequest(loraserver.APIURL+"/api/internal/login", "POST", req, &resp)
		if err != nil {
			return err
		}

		loraserver.jwtToKen = resp.Jwt
	}

	if err := loraserver.initOrganizationID(); err != nil {
		return err
	}

	if err := loraserver.initNetworkServer(); err != nil {
		return err
	}

	if err := loraserver.initServiceProfile(); err != nil {
		return err
	}

	if err := loraserver.initApplication(); err != nil {
		return err
	}

	if err := loraserver.initDeviceProfile(); err != nil {
		return err
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

func (loraserver *loraserver) initOrganizationID() error {
	if loraserver.OrganizationID == "" {
		// Check if already exist
		type Organization struct {
			ID   string `json:"id"`
			Name string `json:"name"`
		}

		respExist := struct {
			Result []Organization `json:"result"`
		}{}

		err := loraserver.doRequest(loraserver.APIURL+"/api/organizations?limit=100", "GET", nil, &respExist)
		if err != nil {
			return err
		}
		for _, orga := range respExist.Result {
			if orga.Name == loraserverOrganisationName {
				loraserver.OrganizationID = orga.ID
				break
			}
		}

		// Create Organization
		if loraserver.OrganizationID == "" {
			req := struct {
				Organization struct {
					CanHaveGateways bool   `json:"canHaveGateways"`
					DisplayName     string `json:"displayName"`
					Name            string `json:"name"`
				} `json:"name"`
			}{
				Organization: struct {
					CanHaveGateways bool   `json:"canHaveGateways"`
					DisplayName     string `json:"displayName"`
					Name            string `json:"name"`
				}{
					CanHaveGateways: true,
					DisplayName:     "Lorhammer",
					Name:            loraserverOrganisationName,
				},
			}

			resp := struct {
				ID string `json:"id"`
			}{}

			err := loraserver.doRequest(loraserver.APIURL+"/api/organizations", "POST", req, &resp)
			if err != nil {
				return err
			}

			loraserver.OrganizationID = resp.ID
		}
	}
	return nil
}

func (loraserver *loraserver) initNetworkServer() error {
	if loraserver.NetworkServerID == "" {
		// Check if already exist
		type NetworkServer struct {
			ID     string `json:"id"`
			Server string `json:"server"`
		}

		respExist := struct {
			Result []NetworkServer `json:"result"`
		}{}

		err := loraserver.doRequest(loraserver.APIURL+"/api/network-servers?limit=100", "GET", nil, &respExist)
		if err != nil {
			return err
		}
		for _, ns := range respExist.Result {
			if ns.Server == loraserver.NetworkServerAddr {
				loraserver.NetworkServerID = ns.ID
				break
			}
		}
		// Create NS
		if loraserver.NetworkServerID == "" {
			req := struct {
				NetworkServer struct {
					Server string `json:"server"`
					Name   string `json:"name"`
				} `json:"name"`
			}{
				NetworkServer: struct {
					Server string `json:"server"`
					Name   string `json:"name"`
				}{
					Server: loraserver.NetworkServerAddr,
					Name:   loraserver.NetworkServerAddr,
				},
			}

			resp := struct {
				ID string `json:"id"`
			}{}

			err = loraserver.doRequest(loraserver.APIURL+"/api/network-servers", "POST", req, &resp)
			if err != nil {
				return err
			}

			loraserver.NetworkServerID = resp.ID
		}
	}
	return nil
}

func (loraserver *loraserver) initServiceProfile() error {
	if loraserver.ServiceProfileID == "" {
		req := struct {
			ServiceProfile struct {
				Name            string `json:"name"`
				NetworkServerID string `json:"networkServerID"`
				OrganizationID  string `json:"organizationID"`
				AddGWMetadata   bool   `json:"addGWMetadata"`
				// ChannelMask            string `json:"channelMask"`
				// DevStatusReqFreq       int    `json:"devStatusReqFreq"`
				// DlBucketSize           int    `json:"dlBucketSize"`
				// DlRate                 int    `json:"dlRate"`
				// DlRatePolicy           string `json:"dlRatePolicy"`
				// DrMax                  int    `json:"drMax"`
				// DrMin                  int    `json:"drMin"`
				// HrAllowed              bool   `json:"hrAllowed"`
				// MinGWDiversity         int    `json:"minGWDiversity"`
				// NwkGeoLoc              bool   `json:"nwkGeoLoc"`
				// PrAllowed              bool   `json:"prAllowed"`
				// RaAllowed              bool   `json:"raAllowed"`
				// ReportDevStatusBattery bool   `json:"reportDevStatusBattery"`
				// ReportDevStatusMargin  bool   `json:"reportDevStatusMargin"`
				// ServiceProfileID       string `json:"serviceProfileID"`
				// TargetPER              int    `json:"targetPER"`
				// UlBucketSize           int    `json:"ulBucketSize"`
				// UlRate                 int    `json:"ulRate"`
				// UlRatePolicy           string `json:"ulRatePolicy"`
				// TODO find description and meaning of all fields
			} `json:"serviceProfile"`
		}{
			ServiceProfile: struct {
				Name            string `json:"name"`
				NetworkServerID string `json:"networkServerID"`
				OrganizationID  string `json:"organizationID"`
				AddGWMetadata   bool   `json:"addGWMetadata"`
			}{
				Name:            "LorhammerServiceProfile",
				NetworkServerID: loraserver.NetworkServerID,
				OrganizationID:  loraserver.OrganizationID,
				AddGWMetadata:   true,
			},
		}

		resp := struct {
			ID string `json:"id"`
		}{}

		err := loraserver.doRequest(loraserver.APIURL+"/api/service-profiles", "POST", req, &resp)
		if err != nil {
			return err
		}

		loraserver.ServiceProfileID = resp.ID
	}
	return nil
}

func (loraserver *loraserver) initApplication() error {
	if loraserver.AppID == "" {
		// Check if already exist
		type Application struct {
			ID   string `json:"id"`
			Name string `json:"name"`
		}

		respExist := struct {
			Result []Application `json:"result"`
		}{}

		err := loraserver.doRequest(loraserver.APIURL+"/api/applications?limit=100", "GET", nil, &respExist)
		if err != nil {
			return err
		}
		for _, app := range respExist.Result {
			if app.Name == loraserverApplicationName {
				loraserver.AppID = app.ID
				break
			}
		}

		// Create Application
		if loraserver.AppID == "" {
			req := struct {
				Application struct {
					Name             string      `json:"name"`
					Description      string      `json:"description"`
					OrganizationID   string      `json:"organizationID"`
					ServiceProfileID interface{} `json:"serviceProfileID"`
				} `json:"application"`
			}{
				Application: struct {
					Name             string      `json:"name"`
					Description      string      `json:"description"`
					OrganizationID   string      `json:"organizationID"`
					ServiceProfileID interface{} `json:"serviceProfileID"`
				}{
					Name:             loraserverApplicationName,
					Description:      "Lorhammer",
					OrganizationID:   loraserver.OrganizationID,
					ServiceProfileID: loraserver.ServiceProfileID,
				},
			}

			resp := struct {
				ID string `json:"id"`
			}{}

			err := loraserver.doRequest(loraserver.APIURL+"/api/applications", "POST", req, &resp)
			if err != nil {
				return err
			}

			loraserver.AppID = resp.ID
		}
	}
	return nil
}

func (loraserver *loraserver) initDeviceProfile() error {
	if loraserver.DeviceProfileID == "" {
		req := struct {
			DeviceProfile struct {
				Name            string `json:"name"`
				OrganizationID  string `json:"organizationID"`
				NetworkServerID string `json:"networkServerID"`
				RxDROffset1     int    `json:"rxDROffset1"`
				RxDataRate2     int    `json:"rxDataRate2"`
				RxDelay1        int    `json:"rxDelay1"`
				// ClassBTimeout           int    `json:"classBTimeout"`
				// ClassCTimeout           int    `json:"classCTimeout"`
				// FactoryPresetFreqs      []int  `json:"factoryPresetFreqs"`
				// MacVersion              string `json:"macVersion"`
				// MaxDutyCycle            int    `json:"maxDutyCycle"`
				// MaxEIRP                 int    `json:"maxEIRP"`
				// PingSlotDR              int    `json:"pingSlotDR"`
				// PingSlotFreq            int    `json:"pingSlotFreq"`
				// PingSlotPeriod          int    `json:"pingSlotPeriod"`
				// RegParamsRevisionstring string `json:"regParamsRevisionstring"`
				// RfRegionstring          string `json:"rfRegionstring"`
				// RxFreq2           int  `json:"rxFreq2"`
				// Supports32bitFCnt bool `json:"supports32bitFCnt"`
				// SupportsClassB    bool `json:"supportsClassB"`
				// SupportsClassC    bool `json:"supportsClassC"`
				// SupportsJoin      bool `json:"supportsJoin"`
				// TODO find description and meaning of all fields
			} `json:"deviceProfile"`
		}{
			DeviceProfile: struct {
				Name            string `json:"name"`
				OrganizationID  string `json:"organizationID"`
				NetworkServerID string `json:"networkServerID"`
				RxDROffset1     int    `json:"rxDROffset1"`
				RxDataRate2     int    `json:"rxDataRate2"`
				RxDelay1        int    `json:"rxDelay1"`
			}{
				Name:            "LorhammerDeviceProfile",
				OrganizationID:  loraserver.OrganizationID,
				NetworkServerID: loraserver.NetworkServerID,
				RxDROffset1:     0,
				RxDataRate2:     0,
				RxDelay1:        0,
			},
		}

		resp := struct {
			ID string `json:"id"`
		}{}

		err := loraserver.doRequest(loraserver.APIURL+"/api/device-profiles", "POST", req, &resp)
		if err != nil {
			return err
		}

		loraserver.DeviceProfileID = resp.ID
	}
	return nil
}

func (loraserver *loraserver) provisionSensorAsync(sensorChan chan *model.Node, poison chan bool, errorChan chan error, sensorFinishChan chan *model.Node) {
	exit := false
	for {
		select {
		case sensor := <-sensorChan:
			if sensor != nil { // Why sensor is nil sometimes !?
				req := struct {
					Device struct {
						Name            string `json:"name"`
						Description     string `json:"description"`
						ApplicationID   string `json:"applicationID"`
						DeviceProfileID string `json:"deviceProfileID"`
						DevEUI          string `json:"devEUI"`
					} `json:"device"`
				}{
					Device: struct {
						Name            string `json:"name"`
						Description     string `json:"description"`
						ApplicationID   string `json:"applicationID"`
						DeviceProfileID string `json:"deviceProfileID"`
						DevEUI          string `json:"devEUI"`
					}{
						Name:            "STRESSNODE_" + sensor.DevEUI.String(),
						Description:     sensor.Description,
						ApplicationID:   loraserver.AppID,
						DeviceProfileID: loraserver.DeviceProfileID,
						DevEUI:          sensor.DevEUI.String(),
					},
				}
				if req.Device.Description == "" { // device description is required
					req.Device.Description = req.Device.Name
				}

				err := loraserver.doRequest(loraserver.APIURL+"/api/devices", "POST", req, nil)
				if err != nil {
					logLoraserver.WithField("req", req).WithError(err).Error("Can't register device")
					errorChan <- err
					break
				}

				reqKeys := struct {
					DeviceKeys struct {
						DevEUI string `json:"devEUI"`
						AppKey string `json:"appKey"`
						NwkKey string `json:"nwkKey"`
					} `json:"deviceKeys"`
				}{
					DeviceKeys: struct {
						DevEUI string `json:"devEUI"`
						AppKey string `json:"appKey"`
						NwkKey string `json:"nwkKey"`
					}{
						DevEUI: sensor.DevEUI.String(),
						AppKey: sensor.AppKey.String(),
						NwkKey: sensor.AppKey.String(),
					},
				}

				err = loraserver.doRequest(loraserver.APIURL+"/api/devices/"+sensor.DevEUI.String()+"/keys", "POST", reqKeys, nil)
				if err != nil {
					logLoraserver.WithField("reqKeys", reqKeys).WithError(err).Error("Can't register keys device")
					errorChan <- err
					break
				}

				if loraserver.Abp {
					req := struct {
						DeviceActivation struct {
							AppSKeystring string `json:"appSKey"`
							DevAddrstring string `json:"devAddr"`
							DevEUIstring  string `json:"devEUI"`
							FCntDown      int    `json:"fCntDown"`
							FCntUp        int    `json:"fCntUp"`
							NwkSKeystring string `json:"nwkSKey"`
							NwkSEncKey    string `json:"nwkSEncKey"`
							SNwkSEncKey   string `json:"sNwkSIntKey"`
							FNwkSEncKey   string `json:"fNwkSIntKey"`
							SkipFCntCheck bool   `json:"skipFCntCheck"`
						} `json:"deviceActivation"`
					}{
						DeviceActivation: struct {
							AppSKeystring string `json:"appSKey"`
							DevAddrstring string `json:"devAddr"`
							DevEUIstring  string `json:"devEUI"`
							FCntDown      int    `json:"fCntDown"`
							FCntUp        int    `json:"fCntUp"`
							NwkSKeystring string `json:"nwkSKey"`
							NwkSEncKey    string `json:"nwkSEncKey"`
							SNwkSEncKey   string `json:"sNwkSIntKey"`
							FNwkSEncKey   string `json:"fNwkSIntKey"`
							SkipFCntCheck bool   `json:"skipFCntCheck"`
						}{
							AppSKeystring: sensor.AppSKey.String(),
							DevAddrstring: sensor.DevAddr.String(),
							DevEUIstring:  sensor.DevEUI.String(),
							FCntDown:      0,
							FCntUp:        0,
							NwkSKeystring: sensor.NwSKey.String(),
							NwkSEncKey:    sensor.NwSKey.String(),
							SNwkSEncKey:   sensor.NwSKey.String(),
							FNwkSEncKey:   sensor.NwSKey.String(),
							SkipFCntCheck: false,
						},
					}
					url := loraserver.APIURL + "/api/devices/" + sensor.DevEUI.String() + "/activate"
					err := loraserver.doRequest(url, "POST", req, nil)
					if err != nil {
						logLoraserver.WithError(err).Error("Can't activate abp device")
						errorChan <- err
						break
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

func (loraserver *loraserver) DeProvision() error {
	if loraserver.DeleteApplication && loraserver.AppID != "" {
		if err := loraserver.doRequest(loraserver.APIURL+"/api/applications/"+loraserver.AppID, "DELETE", nil, nil); err != nil {
			return err
		}
	}

	if loraserver.DeleteOrganization && loraserver.OrganizationID != "" {
		if err := loraserver.doRequest(loraserver.APIURL+"/api/organizations/"+loraserver.OrganizationID, "DELETE", nil, nil); err != nil {
			return err
		}
	}

	return nil
}

func (loraserver *loraserver) doRequest(url string, method string, bodyRequest interface{}, bodyResult interface{}) error {
	logLoraserver.WithField("url", url).Debug("Will call")
	ctx, cancelCtx := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelCtx()

	marshalledBodyRequest, err := json.Marshal(bodyRequest)
	if err != nil {
		return err
	}

	req, err := http.NewRequest(method, url, bytes.NewBuffer(marshalledBodyRequest))
	if err != nil {
		return err
	}

	if loraserver.jwtToKen != "" {
		req.Header.Set("Grpc-Metadata-Authorization", loraserver.jwtToKen)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Close = true
	req.WithContext(ctx)

	resp, err := loraserver.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, _ := ioutil.ReadAll(resp.Body)

	switch resp.StatusCode {
	case http.StatusOK:
		logLoraserver.WithField("url", url).Debug("Call succeeded")

	default:
		logLoraserver.WithFields(logrus.Fields{
			"respStatus":   resp.StatusCode,
			"responseBody": string(body),
			"requestBody":  string(marshalledBodyRequest),
			"url":          url,
		}).Warn("Couldn't proceed with request")
		return errors.New("Couldn't proceed with request")
	}

	if body != nil && bodyResult != nil {
		return json.Unmarshal(body, bodyResult)
	}
	return nil
}
