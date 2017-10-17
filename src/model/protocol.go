package model

import (
	"encoding/json"
)

//CMD is the stuct for communication between orchestrator and lorhammer
type CMD struct {
	CmdName CommandName     `json:"cmd"`
	Payload json.RawMessage `json:"payload"`
}

//Init is the struc send by orchestrator to lorhammer
type Init struct {
	NsAddress            string    `json:"nsAddress"`
	NbGateway            int       `json:"nbGatewayPerLorhammer"`
	NbNode               [2]int    `json:"nbNodePerGateway"`
	NbScenarioReplayLaps int       `json:"nbScenarioReplayLaps"`
	ScenarioSleepTime    [2]string `json:"scenarioSleepTime"`
	GatewaySleepTime     [2]string `json:"gatewaySleepTime"`
	AppsKey              string    `json:"appskey"`
	Nwskey               string    `json:"nwskey"`
	WithJoin             bool      `json:"withJoin"`
	Payloads             []Payload `json:"payloads"`
	RxpkDate             int64     `json:"rxpkDate"`
	ReceiveTimeoutTime   string    `json:"receiveTimeoutTime"`
	Description          string    `json:"description"`
	RandomPayloads       bool      `json:"randomPayloads"`
}

// Payload struct define a payload with timestamp date attached
// { "value": "a string", "date": <timestamp>}
type Payload struct {
	Value string `json:"value"`
	Date  int64  `json:"date"`
}

//Register struct is the command send by lorhammer to orchestrator for register gateway and sensors to network-server
type Register struct {
	ScenarioUUID  string    `json:"scenarioid"`
	Gateways      []Gateway `json:"gateways"`
	CallBackTopic string    `json:"callBackTopic"`
}

//Start is the command send by orchestrator to lorhammer because all gateway and sensor have been registered
type Start struct {
	ScenarioUUID string `json:"scenarioid"`
}
