package model

import (
	"github.com/brocaar/lorawan"
)

type Gateway struct {
	Nodes      []*Node
	NsAddress  string
	MacAddress lorawan.EUI64
	RxpkDate   int64
}

type Node struct {
	DevAddr        lorawan.DevAddr
	DevEUI         lorawan.EUI64
	AppEUI         lorawan.EUI64
	AppKey         lorawan.AES128Key
	AppSKey        lorawan.AES128Key
	NwSKey         lorawan.AES128Key
	JoinedNetwork  bool
	Payloads       []Payload
	NextPayload    int
	RandomPayloads bool
}
