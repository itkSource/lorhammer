package lora

import (
	"encoding/hex"
	"github.com/Sirupsen/logrus"
	"github.com/brocaar/lorawan"
	"lorhammer/src/model"
	"lorhammer/src/tools"
)

var LOG_NODE = logrus.WithFields(logrus.Fields{"logger": "lorhammer/lora/node"})

func NewNode(nwsKeyStr string, appsKeyStr string, payloads []string) *model.Node {
	payload := ""
	if len(payloads) > 0 {
		payload = payloads[tools.Random(0, len(payloads)-1)]
	}

	devEui := RandomEUI()

	nwsKey := lorawan.AES128Key{}
	if nwsKeyStr != "" {
		nwsKey.UnmarshalText([]byte(nwsKeyStr))
	}

	appsKey := lorawan.AES128Key{}
	if appsKeyStr != "" {
		appsKey.UnmarshalText([]byte(appsKeyStr))
	}

	return &model.Node{
		DevEUI:  devEui,
		AppEUI:  RandomEUI(),
		AppKey:  GetGenericAES128Key(),
		DevAddr: GetDevAddrFromDevEUI(devEui),
		AppSKey: appsKey,
		NwSKey:  nwsKey,
		Payload: payload,
	}
}

func GetJoinRequestDataPayload(node *model.Node) []byte {

	phyPayload := lorawan.PHYPayload{
		MHDR: lorawan.MHDR{
			MType: lorawan.MType(lorawan.JoinRequest),
			Major: lorawan.Major(byte(0)),
		},

		MACPayload: &lorawan.JoinRequestPayload{
			AppEUI:   node.AppEUI,
			DevEUI:   node.DevEUI,
			DevNonce: [2]byte{tools.RandomBytes(1)[0], tools.RandomBytes(1)[0]},
		},
	}

	err := phyPayload.SetMIC(node.AppKey)
	if err != nil {
		LOG_NODE.WithFields(logrus.Fields{
			"ref": "lorhammer/lora/payloadFactory:NewJoinRequestPHYPayload()",
			"err": err,
		}).Error("Could not calculate MIC")
	}

	b, err := phyPayload.MarshalBinary()
	if err != nil {
		LOG_NODE.Error("unable to marshal physical payload")
		return []byte{}
	}
	return b
}

func GetPushDataPayload(node *model.Node, fcnt uint32) []byte {

	fport := uint8(1)

	if node.Payload == "" {

		LOG_NODE.WithFields(logrus.Fields{
			"DevEui": node.DevEUI.String(),
		}).Warn("The payload sent for node is empty, please specify a correct payload on the json scenario file")
	}

	frmPayloadByteArray, _ := hex.DecodeString(node.Payload)

	phyPayload := lorawan.PHYPayload{
		MHDR: lorawan.MHDR{
			MType: lorawan.MType(lorawan.ConfirmedDataUp),
			Major: lorawan.LoRaWANR1,
		},

		MACPayload: &lorawan.MACPayload{
			FHDR: lorawan.FHDR{
				DevAddr: node.DevAddr,
				FCtrl: lorawan.FCtrl{
					ADR:       false,
					ADRACKReq: false,
					ACK:       false,
				},
				FCnt: fcnt,
			},
			FPort:      &fport,
			FRMPayload: []lorawan.Payload{&lorawan.DataPayload{Bytes: frmPayloadByteArray}},
		},
	}

	err := phyPayload.SetMIC(node.NwSKey)

	if err != nil {
		LOG_NODE.WithFields(logrus.Fields{
			"ref": "lorhammer/lora/payloadFactory:NewJoinRequestPHYPayload()",
			"err": err,
		}).Fatal("Could not calculate MIC")
	}

	b, err := phyPayload.MarshalBinary()
	if err != nil {
		LOG_NODE.Error("unable to marshal physical payload")
		return []byte{}
	}
	return b
}

func GetDevAddrFromDevEUI(devEUI lorawan.EUI64) lorawan.DevAddr {
	devAddr := lorawan.DevAddr{}
	devEuiStr := devEUI.String()
	devAddr.UnmarshalText([]byte(devEuiStr[len(devEuiStr)-8:]))
	return devAddr
}

func GetGenericAES128Key() lorawan.AES128Key {
	return lorawan.AES128Key{
		byte(1), byte(2), byte(3), byte(4),
		byte(5), byte(6), byte(7), byte(8),
		byte(12), byte(11), byte(10), byte(9),
		byte(13), byte(14), byte(15), byte(16),
	}
}
