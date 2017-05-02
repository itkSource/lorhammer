package lora

import (
	"github.com/brocaar/lorawan"
	"testing"
)

func TestGetDevAddrFromDevEUI(t *testing.T) {
	eui := lorawan.EUI64{
		byte(1), byte(2), byte(3), byte(4),
		byte(5), byte(6), byte(7), byte(8),
	}

	devAddr := GetDevAddrFromDevEUI(eui)

	if devAddr.String() != "05060708" {
		t.Fatalf("DevAddr string value must equal %d ", "05060708")
	}

}

func TestNode_GetPushDataPayload(t *testing.T) {
	node := NewNode("19842bd94743246b367c2e90942a1f73",
		"19842bd94743246b367c2e90942a1f774",
		[]string{"01B501002919000006018403131313121233", "01B501002919000006018403131313121244"})

	fcnt := uint32(1)
	dataPayload := GetPushDataPayload(node, fcnt)
	phyPayload := lorawan.PHYPayload{}
	err := phyPayload.UnmarshalBinary(dataPayload)
	if err != nil {
		t.Fatal("Couldn't unmarshall PHYPayload Binary")
	}

	if phyPayload.MHDR.MType != lorawan.MType(lorawan.ConfirmedDataUp) {
		t.Fatal("Push data messages should always be of type ConfirmedDataType")
	}

	if phyPayload.MHDR.Major != lorawan.LoRaWANR1 {
		t.Fatal("Push data messages should always be of protocol LoraWan")
	}

	if len(phyPayload.MIC) != 4 {
		t.Fatal("An exactly 4 bytes MIC should be sent with the message")
	}

	_, ok := phyPayload.MACPayload.(*lorawan.MACPayload)
	if !ok {
		t.Fatal("the MacPayload should be of Type MACPayload ")
	}

}

func TestNewJoinRequestPHYPayload(t *testing.T) {
	node := NewNode("",
		"",
		[]string{"01B501002919000006018403131313121233", "01B501002919000006018403131313121244"})

	dataPayload := GetJoinRequestDataPayload(node)
	jrPayload := lorawan.PHYPayload{}
	jrPayload.UnmarshalBinary(dataPayload)

	if jrPayload.MHDR.MType != lorawan.MType(lorawan.JoinRequest) {
		t.Fatal("Join request messages should be of type JoinRequest")
	}

	if jrPayload.MHDR.Major != lorawan.LoRaWANR1 {
		t.Fatal("Join request messages should always be of protocol LoraWan")
	}

	_, ok := jrPayload.MACPayload.(*lorawan.JoinRequestPayload)
	if !ok {
		t.Fatal("the MacPayload should be of Type JoinRequestPayload")
	}

}
