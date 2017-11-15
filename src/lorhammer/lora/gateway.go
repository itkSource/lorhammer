package lora

import (
	"lorhammer/src/model"
	"lorhammer/src/tools"
	"math"
	"net"
	"time"

	loraserver_structs "github.com/brocaar/lora-gateway-bridge/gateway"
	"github.com/sirupsen/logrus"
)

var loggerGateway = logrus.WithField("logger", "lorhammer/lora/gateway")

//NewGateway return a new gateway with node configured
func NewGateway(nbNode int, init model.Init) *model.Gateway {
	parsedTime, _ := time.ParseDuration(init.ReceiveTimeoutTime)
	gateway := &model.Gateway{
		NsAddress:             init.NsAddress,
		MacAddress:            tools.Random8Bytes(),
		ReceiveTimeoutTime:    parsedTime,
		PayloadsReplayMaxLaps: init.NbScenarioReplayLaps,
	}

	if init.RxpkDate > 0 {
		gateway.RxpkDate = init.RxpkDate
	}
	for i := 0; i < nbNode; i++ {
		gateway.Nodes = append(gateway.Nodes, newNode(init.Nwskey, init.AppsKey, init.Description, init.Payloads, init.RandomPayloads))
	}

	return gateway
}

//Join send first pull datata to be discovered by network server
//Then send a JoinRequest packet if `withJoin` is set in scenario file
func Join(gateway *model.Gateway, prometheus tools.Prometheus, withJoin bool) {
	Conn, err := net.Dial("udp", gateway.NsAddress)
	defer Conn.Close()

	if err != nil {
		loggerGateway.WithError(err).Error("Can't Dial Udp when join")
	}

	//WRITE
	endTimer := prometheus.StartTimer()

	/**
	 ** We send pull Data for the gateway to be recognized by the NS
	 ** when sending JoinRequest or any other type of packet
	 **/
	sendPullData(gateway, Conn)

	if withJoin {
		sendJoinRequestPackets(gateway, Conn)
	}

	threadListenUDP := make(chan []byte, 1)
	defer close(threadListenUDP)
	next := make(chan bool, 1)
	defer close(next)
	poison := make(chan bool, 1)
	defer close(poison)

	go readPackets(Conn, poison, next, threadListenUDP)
	readLoraJoinPackets(gateway, poison, next, threadListenUDP, endTimer, prometheus, withJoin)
}

//Start send push data packet and listen for ack
func Start(gateway *model.Gateway, prometheus tools.Prometheus, fcnt uint32) {
	Conn, err := net.Dial("udp", gateway.NsAddress)
	defer Conn.Close()

	if err != nil {
		loggerGateway.WithError(err).Error("Can't Dial Udp qhen start")
	}

	//WRITE
	endTimer := prometheus.StartTimer()

	//Send pushDataPackets
	sendPushPackets(gateway, Conn, fcnt)

	//READ
	threadListenUDP := make(chan []byte, 1)
	defer close(threadListenUDP)
	next := make(chan bool, 1)
	defer close(next)
	poison := make(chan bool, 1)
	defer close(poison)

	go readPackets(Conn, poison, next, threadListenUDP)
	readLoraPushPackets(gateway, poison, next, threadListenUDP, endTimer, prometheus)
}

func sendPullData(gateway *model.Gateway, Conn net.Conn) {

	pullDataPacket, err := loraserver_structs.PullDataPacket{
		ProtocolVersion: 2,
		RandomToken:     uint16(tools.Random(math.MinInt16, math.MaxUint16)),
		GatewayMAC:      gateway.MacAddress,
	}.MarshalBinary()

	if err != nil {
		loggerGateway.WithError(err).Error("can't marshall pull data message")
	}

	if _, err = Conn.Write(pullDataPacket); err != nil {
		loggerGateway.WithError(err).Error("Can't write pullDataPacket udp")
	}
}

func sendJoinRequestPackets(gateway *model.Gateway, Conn net.Conn) {
	loggerGateway.Info("Sending JoinRequest messages for all the nodes")

	rxpk := make([]loraserver_structs.RXPK, 1)
	for _, node := range gateway.Nodes {
		if !node.JoinedNetwork {

			rxpk[0] = newRxpk(getJoinRequestDataPayload(node), 0, gateway)
			packet, err := packet{Rxpk: rxpk}.prepare(gateway)

			if err != nil {
				loggerGateway.WithError(err).Error("Can't prepare lora packet in SendJoinRequest")
			}
			if _, err = Conn.Write(packet); err != nil {
				loggerGateway.WithError(err).Error("Can't write udp in SendJoinRequest")
			}
		}
	}
}

func sendPushPackets(gateway *model.Gateway, Conn net.Conn, fcnt uint32) {
	rxpk := make([]loraserver_structs.RXPK, 1)
	for _, node := range gateway.Nodes {
		if node.PayloadsReplayLap < gateway.PayloadsReplayMaxLaps || gateway.PayloadsReplayMaxLaps == 0 {
			buf, date, err := GetPushDataPayload(node, fcnt)
			if err != nil {
				loggerGateway.WithError(err).Error("Can't get next lora packet to send")
			}
			rxpk[0] = newRxpk(buf, date, gateway)
			packet, err := packet{Rxpk: rxpk}.prepare(gateway)

			if err != nil {
				loggerGateway.WithError(err).Error("Can't prepare lora packet in sendPushPackets")
			}
			if _, err = Conn.Write(packet); err != nil {
				loggerGateway.WithError(err).Error("Can't write udp in sendPushPackets")
			}
		}
	}
	if isGatewayScenarioCompleted(gateway) {
		gateway.AllLapsCompleted = true
		return
	}
}

func readPackets(Conn net.Conn, poison chan bool, next chan bool, threadListenUDP chan []byte) {
	for {
		quit := false
		select {
		case <-poison:
			quit = true
			break
		case <-next:
			buf := make([]byte, 65507) // max udp data size
			// TODO handle Conn.SetReadDeadline with time max - current time to gracefully kill conn
			n, err := Conn.Read(buf)

			if err != nil {
				loggerGateway.WithError(err).Debug("Can't read udp")
				quit = true
				break
			} else {
				threadListenUDP <- buf[0:n]
			}
		}
		if quit {
			break
		}
	}
}

func readLoraJoinPackets(gateway *model.Gateway, poison chan bool, next chan bool, threadListenUDP chan []byte, endTimer func(), prometheus tools.Prometheus, withJoin bool) {
	nbReceivedAckMsg := readLoraPackets(gateway, poison, next, threadListenUDP, endTimer, prometheus)
	nbEmittedMsg := 1 // One PullData request has been sent
	if withJoin {
		nbEmittedMsg += len(gateway.Nodes)
	}
	loggerGateway.WithFields(logrus.Fields{
		"ref":      "lora/gateway:Join()",
		"withJoin": withJoin,
		"nb":       nbEmittedMsg - nbReceivedAckMsg,
	}).Warn("Receive PullData or Join Request ack after 2 seconds")
	prometheus.AddLongRequest(nbEmittedMsg - nbReceivedAckMsg)
}

func readLoraPushPackets(gateway *model.Gateway, poison chan bool, next chan bool, threadListenUDP chan []byte, endTimer func(), prometheus tools.Prometheus) {
	nbReceivedAckMsg := readLoraPackets(gateway, poison, next, threadListenUDP, endTimer, prometheus)
	if len(gateway.Nodes)-nbReceivedAckMsg > 0 {
		loggerGateway.WithFields(logrus.Fields{
			"ref": "lora/gateway:Start()",
			"nb":  len(gateway.Nodes) - nbReceivedAckMsg,
		}).Warn("Receive data after 2 second")
		prometheus.AddLongRequest(len(gateway.Nodes) - nbReceivedAckMsg)
	}
}

func readLoraPackets(gateway *model.Gateway, poison chan bool, next chan bool, threadListenUDP chan []byte, endTimer func(), prometheus tools.Prometheus) int {
	nbReceivedAckMsg := 0
	localPoison := make(chan bool)

	go func() {
		quit := false
		next <- true
		for {
			select {
			case <-localPoison:
				quit = true
				break
			case res := <-threadListenUDP:
				endTimer()
				err := handlePacket(res)
				if err != nil {
					loggerGateway.WithError(err).Error("Can't handle packet")
				}
				loggerGateway.WithFields(logrus.Fields{
					"ProtocolVersion": res[0],
					"Token":           res[1:2],
					"ack":             res[3],
					"ackOk?":          res[3] == byte(1),
					"time":            gateway.ReceiveTimeoutTime,
				}).Debug("Receive data before time")
				if res[3] == byte(1) {
					nbReceivedAckMsg++
				}
			}
			if quit {
				break
			} else {
				next <- true
			}
		}
	}()

	<-time.After(gateway.ReceiveTimeoutTime)
	poison <- true
	localPoison <- true
	return nbReceivedAckMsg
}

func isGatewayScenarioCompleted(gateway *model.Gateway) bool {
	//infinite case when PayloadsReplayMaxRound is set to 0 or inferior
	if gateway.PayloadsReplayMaxLaps <= 0 {
		return false
	}
	for _, node := range gateway.Nodes {
		// if only one node has not reached the expected number of rounds, break the loop and return false
		if node.PayloadsReplayLap < gateway.PayloadsReplayMaxLaps {
			loggerGateway.WithFields(logrus.Fields{
				"DevEui":                node.DevEUI.String(),
				"PayloadsReplayLap":     node.PayloadsReplayLap,
				"PayloadsReplayMaxLaps": gateway.PayloadsReplayMaxLaps,
			}).Debug("node has not finished yet")
			return false
		}
	}
	loggerGateway.WithFields(logrus.Fields{
		"MacAddress": gateway.MacAddress.String(),
	}).Debugf("Gateway scenario is completed")
	return true
}
