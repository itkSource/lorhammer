package model

import (
	"encoding/json"
)

type CMD struct {
	CmdName CommandName     `json:"cmd"`
	Payload json.RawMessage `json:"payload"`
}

type Init struct {
	NsAddress          string    `json:"nsAddress"`
	NbGateway          int       `json:"nbGatewayPerLorhammer"`
	NbNode             [2]int    `json:"nbNodePerGateway"`
	ScenarioSleepTime  [2]string `json:"scenarioSleepTime"`
	GatewaySleepTime   [2]string `json:"gatewaySleepTime"`
	AppsKey            string    `json:"appskey"`
	Nwskey             string    `json:"nwskey"`
	WithJoin           bool      `json:"withJoin"`
	Payloads           []Payload `json:"payloads"`
	RxpkDate           int64     `json:"rxpkDate"`
	ReceiveTimeoutTime string    `json:"receiveTimeoutTime"`
	Description        string    `json:"description"`
	RandomPayloads     bool      `json:"randomPayloads"`
}

// Payload struc define a payload with timestamp date attached
// { "value": "a string", "date": <timestamp>}
type Payload struct {
	Value string `json:"value"`
	Date  int64  `json:"date"`
}

type Register struct {
	ScenarioUUID  string    `json:"scenarioid"`
	Gateways      []Gateway `json:"gateways"`
	CallBackTopic string    `json:"callBackTopic"`
}

type Start struct {
	ScenarioUUID string `json:"scenarioid"`
}
