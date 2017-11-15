package lora

import (
	"encoding/hex"
	"errors"
	"lorhammer/src/model"
	"lorhammer/src/tools"

	"github.com/brocaar/lorawan"
	"github.com/sirupsen/logrus"
)

var loggerNode = logrus.WithFields(logrus.Fields{"logger": "lorhammer/lora/node"})

func newNode(nwsKeyStr string, appsKeyStr string, description string, payloads []model.Payload, randomPayloads bool) *model.Node {

	devEui := tools.Random8Bytes()

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
		AppEUI:         tools.Random8Bytes(),
		AppKey:         getGenericAES128Key(),
		DevAddr:        getDevAddrFromDevEUI(devEui),
		AppSKey:        appsKey,
		NwSKey:         nwsKey,
		Payloads:       payloads,
		NextPayload:    0,
		RandomPayloads: randomPayloads,
		Description:    description,
	}
}

func getJoinRequestDataPayload(node *model.Node) []byte {

	phyPayload := lorawan.PHYPayload{
		MHDR: lorawan.MHDR{
			MType: lorawan.MType(lorawan.JoinRequest),
			Major: lorawan.Major(byte(0)),
		},

		MACPayload: &lorawan.JoinRequestPayload{
			AppEUI:   node.AppEUI,
			DevEUI:   node.DevEUI,
			DevNonce: tools.Random2Bytes(),
		},
	}

	err := phyPayload.SetMIC(node.AppKey)
	if err != nil {
		loggerNode.WithFields(logrus.Fields{
			"ref": "lorhammer/lora/payloadFactory:NewJoinRequestPHYPayload()",
			"err": err,
		}).Error("Could not calculate MIC")
	}

	b, err := phyPayload.MarshalBinary()
	if err != nil {
		loggerNode.Error("unable to marshal physical payload")
		return []byte{}
	}
	return b
}

// GetPushDataPayload return the nextbyte arraypush data
func GetPushDataPayload(node *model.Node, fcnt uint32) ([]byte, int64, error) {
	fport := uint8(1)

	var frmPayloadByteArray []byte
	var date int64
	if len(node.Payloads) == 0 {
		loggerNode.WithFields(logrus.Fields{
			"DevEui": node.DevEUI.String(),
		}).Warn("empty payload array given. So it send `LorHammer`")
		frmPayloadByteArray, _ = hex.DecodeString("LorHammer")
	} else {
		var i int
		if node.RandomPayloads == true {
			i = tools.Random(0, len(node.Payloads)-1)
		} else {
			i = node.NextPayload
			if len(node.Payloads) >= i-1 {
				node.NextPayload = i + 1
			}
			// if the current payload is the last of the payload set, a complete round has been executed
			if len(node.Payloads) == node.NextPayload {
				node.PayloadsReplayLap++
				loggerNode.WithFields(logrus.Fields{
					"DevEui":            node.DevEUI.String(),
					"PayloadsReplayLap": node.PayloadsReplayLap,
				}).Info("Complete lap executed")
				node.NextPayload = 0
			}
			// only extract timestamp when payloads are consumed in declaration order and not randomly,
			// keep the "0" default value instead
			date = node.Payloads[i].Date
		}
		loggerNode.WithFields(logrus.Fields{
			"DevEui":             node.DevEUI.String(),
			"Valeur de i":        i,
			"len(node.Payloads)": len(node.Payloads),
			"node.NextPayload":   node.NextPayload,
			"Payload : ":         node.Payloads[i].Value,
			"Date : ":            node.Payloads[i].Date,
		}).Debug("Payload sent")
		frmPayloadByteArray, _ = hex.DecodeString(node.Payloads[i].Value)
	}

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
		loggerNode.WithFields(logrus.Fields{
			"ref": "lorhammer/lora/payloadFactory:NewJoinRequestPHYPayload()",
			"err": err,
		}).Fatal("Could not calculate MIC")
	}

	b, err := phyPayload.MarshalBinary()
	if err != nil {
		return nil, 0, errors.New("unable to marshal physical payload")
	}
	return b, date, nil
}

func getDevAddrFromDevEUI(devEUI lorawan.EUI64) lorawan.DevAddr {
	devAddr := lorawan.DevAddr{}
	devEuiStr := devEUI.String()
	devAddr.UnmarshalText([]byte(devEuiStr[len(devEuiStr)-8:]))
	return devAddr
}

func getGenericAES128Key() lorawan.AES128Key {
	return lorawan.AES128Key{
		byte(1), byte(2), byte(3), byte(4),
		byte(5), byte(6), byte(7), byte(8),
		byte(12), byte(11), byte(10), byte(9),
		byte(13), byte(14), byte(15), byte(16),
	}
}
