package lora

import (
	"encoding/hex"
	"errors"
	"lorhammer/src/model"
	"lorhammer/src/tools"

	"github.com/Sirupsen/logrus"
	"github.com/brocaar/lorawan"
)

var LOG_NODE = logrus.WithFields(logrus.Fields{"logger": "lorhammer/lora/node"})

func NewNode(nwsKeyStr string, appsKeyStr string, payloads []model.Payload, randomPayloads bool) *model.Node {

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
		DevEUI:         devEui,
		AppEUI:         RandomEUI(),
		AppKey:         GetGenericAES128Key(),
		DevAddr:        GetDevAddrFromDevEUI(devEui),
		AppSKey:        appsKey,
		NwSKey:         nwsKey,
		Payloads:       payloads,
		NextPayload:    0,
		RandomPayloads: randomPayloads,
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

// GetPushDataPayload return the nextbayte arraypush data
func GetPushDataPayload(node *model.Node, fcnt uint32) ([]byte, error) {

	fport := uint8(1)

	if len(node.Payloads) == 0 {

		LOG_NODE.WithFields(logrus.Fields{
			"DevEui": node.DevEUI.String(),
		}).Warn("The payload sent for node is empty, please specify a correct payload on the json scenario file")
	}
	var i int
	if node.RandomPayloads == true {
		i = tools.Random(0, len(node.Payloads)-1)
	} else {
		i = node.NextPayload
		if len(node.Payloads) <= i-1 {
			node.NextPayload = i + 1
		}
		if len(node.Payloads) == i {
			LOG_NODE.WithFields(logrus.Fields{
				"DevEui": node.DevEUI.String(),
			}).Infof("all payloads sended. restart from beginning (%d/%d)", i, len(node.Payloads))
			node.NextPayload = 0
			i = node.NextPayload
		}
	}
	frmPayloadByteArray, _ := hex.DecodeString(node.Payloads[i].Value)

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
		return nil, errors.New("unable to marshal physical payload")
	}
	return b, nil
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
