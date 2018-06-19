package lora

import (
	"lorhammer/src/model"
	"lorhammer/src/tools"
	"math"
	"net"
	"time"

	loraserver_structs "github.com/brocaar/lora-gateway-bridge/gateway"
	"github.com/brocaar/lorawan"
	"github.com/sirupsen/logrus"
	"lorhammer/src/lorhammer/metrics"
)

var loggerGateway = logrus.WithField("logger", "lorhammer/lora/gateway")

//LorhammerGateway : internal gateway for pointer receiver usage
type LorhammerGateway struct {
	Nodes                 []*model.Node
	NsAddress             string
	MacAddress            lorawan.EUI64
	RxpkDate              int64
	PayloadsReplayMaxLaps int
	AllLapsCompleted      bool
	ReceiveTimeoutTime    time.Duration
}

//NewGateway return a new gateway with node configured
func NewGateway(nbNode int, init model.Init) *LorhammerGateway {
	parsedTime, _ := time.ParseDuration(init.ReceiveTimeoutTime)
	gateway := &LorhammerGateway{
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
func (gateway *LorhammerGateway) Join(prometheus metrics.Prometheus, withJoin bool) error {
	conn, err := net.Dial("udp", gateway.NsAddress)
	if err != nil {
		return err
	}
	defer conn.Close()

	//WRITE
	endPushAckTimer := prometheus.StartPushAckTimer()
	endPullRespTimer := prometheus.StartPullRespTimer()

	/**
	 ** We send pull Data for the gateway to be recognized by the NS
	 ** when sending JoinRequest or any other type of packet
	 **/
	gateway.sendPullData(conn)

	if withJoin {
		gateway.sendJoinRequestPackets(conn)
	}

	threadListenUDP := make(chan []byte, 1)
	defer close(threadListenUDP)
	next := make(chan bool, 1)
	defer close(next)
	poison := make(chan bool, 1)
	defer close(poison)

	go gateway.readPackets(conn, poison, next, threadListenUDP)
	gateway.readLoraJoinPackets(conn, poison, next, threadListenUDP, endPushAckTimer, endPullRespTimer, prometheus, withJoin)
	return nil
}

//Start send push data packet and listen for ack
func (gateway *LorhammerGateway) Start(prometheus metrics.Prometheus, fcnt uint32) error {
	conn, err := net.Dial("udp", gateway.NsAddress)
	if err != nil {
		return err
	}
	defer conn.Close()

	//WRITE
	endPushAckTimer := prometheus.StartPushAckTimer()
	endPullRespTimer := prometheus.StartPullRespTimer()

	gateway.sendPullData(conn)

	//Send pushDataPackets
	gateway.sendPushPackets(conn, fcnt)

	//READ
	threadListenUDP := make(chan []byte, 1)
	defer close(threadListenUDP)
	next := make(chan bool, 1)
	defer close(next)
	poison := make(chan bool, 1)
	defer close(poison)

	go gateway.readPackets(conn, poison, next, threadListenUDP)
	gateway.readLoraPushPackets(conn, poison, next, threadListenUDP, endPushAckTimer, endPullRespTimer, prometheus)
	return nil
}

func (gateway *LorhammerGateway) sendPullData(conn net.Conn) {
	loggerGateway.Info("Sending Pull data message")

	pullDataPacket, err := loraserver_structs.PullDataPacket{
		ProtocolVersion: 2,
		RandomToken:     uint16(tools.Random64(int64(math.MinInt16), int64(math.MaxUint16))),
		GatewayMAC:      gateway.MacAddress,
	}.MarshalBinary()

	if err != nil {
		loggerGateway.WithError(err).Error("can't marshall pull data message")
	}

	if _, err = conn.Write(pullDataPacket); err != nil {
		loggerGateway.WithError(err).Error("Can't write pullDataPacket udp")
	}
}

func (gateway *LorhammerGateway) sendJoinRequestPackets(conn net.Conn) {
	loggerGateway.Info("Sending JoinRequest messages for all the nodes")

	for _, node := range gateway.Nodes {
		if !node.JoinedNetwork {
			packet, err := packet{
				Rxpk: []loraserver_structs.RXPK{
					newRxpk(getJoinRequestDataPayload(node), 0, gateway),
				},
			}.prepare(gateway)

			if err != nil {
				loggerGateway.WithError(err).Error("Can't prepare lora packet in SendJoinRequest")
			}
			if _, err = conn.Write(packet); err != nil {
				loggerGateway.WithError(err).Error("Can't write udp in SendJoinRequest")
			}
		}
	}
}

func (gateway *LorhammerGateway) sendPushPackets(conn net.Conn, fcnt uint32) {
	for _, node := range gateway.Nodes {
		if node.PayloadsReplayLap < gateway.PayloadsReplayMaxLaps || gateway.PayloadsReplayMaxLaps == 0 {
			buf, date, err := GetPushDataPayload(node, fcnt)
			if err != nil {
				loggerGateway.WithError(err).Error("Can't get next lora packet to send")
			}
			packet, err := packet{
				Rxpk: []loraserver_structs.RXPK{
					newRxpk(buf, date, gateway),
				},
			}.prepare(gateway)

			if err != nil {
				loggerGateway.WithError(err).Error("Can't prepare lora packet in sendPushPackets")
			}
			if _, err = conn.Write(packet); err != nil {
				loggerGateway.WithError(err).Error("Can't write udp in sendPushPackets")
			}
		}
	}
	if gateway.isGatewayScenarioCompleted() {
		gateway.AllLapsCompleted = true
		return
	}
}

func (gateway *LorhammerGateway) readPackets(conn net.Conn, poison chan bool, next chan bool, threadListenUDP chan []byte) {
	for {
		quit := false
		select {
		case <-poison:
			quit = true
			break
		case <-next:
			buf := make([]byte, 65507) // max udp data size
			// TODO handle conn.SetReadDeadline with time max - current time to gracefully kill conn
			n, err := conn.Read(buf)

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

func (gateway *LorhammerGateway) readLoraJoinPackets(conn net.Conn, poison chan bool, next chan bool, threadListenUDP chan []byte, endPushAckTimer func(), endPullRespTimer func(), prometheus metrics.Prometheus, withJoin bool) {
	nbReceivedAckMsg, nbReceivedPullRespMsg := gateway.readLoraPackets(conn, poison, next, threadListenUDP, endPushAckTimer, endPullRespTimer)
	nbEmittedMsg := 1 // One PullData request has been sent
	if withJoin {
		nbEmittedMsg += len(gateway.Nodes)
	}
	loggerGateway.WithFields(logrus.Fields{
		"ref":      "lora/gateway:Join()",
		"withJoin": withJoin,
		"nb":       nbEmittedMsg - nbReceivedAckMsg,
		"msgType":  "Push Ack",
	}).Warn("Receive PullData or Join Request ack after 2 seconds")
	prometheus.AddPushAckLongRequest(nbEmittedMsg - nbReceivedAckMsg)

	loggerGateway.WithFields(logrus.Fields{
		"ref":      "lora/gateway:Join()",
		"withJoin": withJoin,
		"nb":       nbEmittedMsg - nbReceivedPullRespMsg,
		"msgType":  "Pull Resp",
	}).Warn("Receive PullData or Join Request ack after 2 seconds")
	prometheus.AddPullRespLongRequest(nbEmittedMsg - nbReceivedPullRespMsg)
}

func (gateway *LorhammerGateway) readLoraPushPackets(conn net.Conn, poison chan bool, next chan bool, threadListenUDP chan []byte, endPushAckTimer func(), endPullRespTimer func(), prometheus metrics.Prometheus) {
	nbReceivedAckMsg, nbReceivedPullRespMsg := gateway.readLoraPackets(conn, poison, next, threadListenUDP, endPushAckTimer, endPullRespTimer)
	if len(gateway.Nodes)-nbReceivedAckMsg > 0 {
		loggerGateway.WithFields(logrus.Fields{
			"ref":     "lora/gateway:Start()",
			"nb":      len(gateway.Nodes) - nbReceivedAckMsg,
			"msgType": "Push Ack",
		}).Warn("Receive data after 2 second")
		prometheus.AddPushAckLongRequest(len(gateway.Nodes) - nbReceivedAckMsg)
	}
	if len(gateway.Nodes)-nbReceivedPullRespMsg > 0 {
		loggerGateway.WithFields(logrus.Fields{
			"ref":     "lora/gateway:Start()",
			"nb":      len(gateway.Nodes) - nbReceivedPullRespMsg,
			"msgType": "Pull Resp",
		}).Warn("Receive data after 2 second")
		prometheus.AddPullRespLongRequest(len(gateway.Nodes) - nbReceivedPullRespMsg)
	}
}

func (gateway *LorhammerGateway) readLoraPackets(conn net.Conn, poison chan bool, next chan bool, threadListenUDP chan []byte, endPushAckTimer func(), endPullRespTimer func()) (int, int) {
	nbReceivedAckMsg, nbReceivedPullRespMsg := 0, 0
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
				err := handlePacket(res)
				if err != nil {
					loggerGateway.WithError(err).Error("Can't handle packet")
				} else if packetType, err := loraserver_structs.GetPacketType(res); err != nil {
					loggerGateway.WithError(err).Error("Can't handle packet type")
				} else {
					if packetType == loraserver_structs.PushACK {
						endPushAckTimer()
						nbReceivedAckMsg++
					} else if packetType == loraserver_structs.PullResp {
						endPullRespTimer()
						nbReceivedPullRespMsg++
						gateway.sendTxAckPacket(conn, res)
					}

					loggerGateway.WithFields(logrus.Fields{
						"ProtocolVersion": res[0],
						"Token":           res[1:2],
						"ack":             res[3],
						"ackOk?":          res[3] == byte(1),
						"time":            gateway.ReceiveTimeoutTime,
					}).Debug("Receive data before time")
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
	return nbReceivedAckMsg, nbReceivedPullRespMsg
}

func (gateway *LorhammerGateway) isGatewayScenarioCompleted() bool {
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
	}).Debug("Gateway scenario is completed")
	return true
}

func (gateway *LorhammerGateway) sendTxAckPacket(conn net.Conn, data []byte) {
	var pullRespPacket loraserver_structs.PullRespPacket
	if err := pullRespPacket.UnmarshalBinary(data); err == nil {
		txAckPacket := loraserver_structs.TXACKPacket{
			ProtocolVersion: 2,
			RandomToken:     pullRespPacket.RandomToken,
			GatewayMAC:      gateway.MacAddress,
			Payload: &loraserver_structs.TXACKPayload{
				TXPKACK: loraserver_structs.TXPKACK{
					Error: "NONE",
				},
			},
		}
		loggerGateway.WithField("TxAckPacket", txAckPacket).Info("Send TxAck packet")
		if dataToSend, err := txAckPacket.MarshalBinary(); err == nil {
			if _, err = conn.Write(dataToSend); err != nil {
				loggerGateway.WithError(err).Debug("Can't send Tx Ack packet")
			}
		} else {
			loggerGateway.WithError(err).Debug("Can't marshal Tx Ack packet")
		}
	} else {
		loggerGateway.WithError(err).Error("Can't unmarshal Pull Resp packet")
	}
}

//ConvertToGateway : convert internal gateway to model gateway
func (gateway LorhammerGateway) ConvertToGateway() model.Gateway {
	return model.Gateway{
		Nodes:                 gateway.Nodes,
		NsAddress:             gateway.NsAddress,
		MacAddress:            gateway.MacAddress,
		RxpkDate:              gateway.RxpkDate,
		PayloadsReplayMaxLaps: gateway.PayloadsReplayMaxLaps,
		AllLapsCompleted:      gateway.AllLapsCompleted,
		ReceiveTimeoutTime:    gateway.ReceiveTimeoutTime,
	}
}
