package model

import (
	"time"

	"github.com/brocaar/lorawan"
)

//Gateway represent a lorawan gateway
type Gateway struct {
	Nodes                 []*Node
	NsAddress             string
	MacAddress            lorawan.EUI64
	RxpkDate              int64
	PayloadsReplayMaxLaps int
	AllLapsCompleted      bool
	ReceiveTimeoutTime    time.Duration
}

//Node represent a lorawan sensor
type Node struct {
	DevAddr           lorawan.DevAddr
	DevEUI            lorawan.EUI64
	AppEUI            lorawan.EUI64
	AppKey            lorawan.AES128Key
	AppSKey           lorawan.AES128Key
	NwSKey            lorawan.AES128Key
	JoinedNetwork     bool
	Payloads          []Payload
	NextPayload       int
	PayloadsReplayLap int
	RandomPayloads    bool
	Description       string
}
