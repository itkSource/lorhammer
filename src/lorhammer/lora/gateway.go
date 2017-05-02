package lora

import (
	"github.com/Sirupsen/logrus"
	loraserver_structs "github.com/brocaar/lora-gateway-bridge/gateway"
	"github.com/brocaar/lorawan"
	"lorhammer/src/model"
	"lorhammer/src/tools"
	"math"
	"net"
	"time"
)

var LOG_GATEWAY = logrus.WithFields(logrus.Fields{"logger": "lorhammer/lora/gateway"})

func NewGateway(nbNode int, nsAddress string, appskey string, nwskey string, payloads []string) *model.Gateway {
	gateway := &model.Gateway{
		NsAddress:  nsAddress,
		MacAddress: RandomEUI(),
	}
	for i := 0; i < nbNode; i++ {
		gateway.Nodes = append(gateway.Nodes, NewNode(nwskey, appskey, payloads))
	}

	return gateway
}

func Join(gateway *model.Gateway, prometheus tools.Prometheus, withJoin bool) {
	Conn, err := net.Dial("udp", gateway.NsAddress)
	defer Conn.Close()

	if err != nil {
		LOG_GATEWAY.WithFields(logrus.Fields{
			"ref": "lora/gateway:Join()",
			"err": err,
		}).Error("Can't Dial Udp")
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

	threadListenUdp := make(chan []byte, 1)
	defer close(threadListenUdp)
	next := make(chan bool, 1)
	defer close(next)
	poison := make(chan bool, 1)
	defer close(poison)

	go readPackets(Conn, poison, next, threadListenUdp)
	readLoraPackets(gateway, poison, next, threadListenUdp, endTimer, prometheus)
}

func Start(gateway *model.Gateway, prometheus tools.Prometheus, fcnt uint32) {
	Conn, err := net.Dial("udp", gateway.NsAddress)
	defer Conn.Close()

	if err != nil {
		LOG_GATEWAY.WithFields(logrus.Fields{
			"ref": "lora/gateway:Start()",
			"err": err,
		}).Error("Can't Dial Udp")
	}

	//WRITE
	endTimer := prometheus.StartTimer()

	//Send pushDataPackets
	sendPushPackets(gateway, Conn, fcnt)

	//READ
	threadListenUdp := make(chan []byte, 1)
	defer close(threadListenUdp)
	next := make(chan bool, 1)
	defer close(next)
	poison := make(chan bool, 1)
	defer close(poison)

	go readPackets(Conn, poison, next, threadListenUdp)
	readLoraPackets(gateway, poison, next, threadListenUdp, endTimer, prometheus)
}

func sendPullData(gateway *model.Gateway, Conn net.Conn) {

	pullDataPacket, err := loraserver_structs.PullDataPacket{
		ProtocolVersion: 2,
		RandomToken:     uint16(tools.Random(math.MinInt16, math.MaxUint16)),
		GatewayMAC:      gateway.MacAddress,
	}.MarshalBinary()

	if err != nil {
		LOG_GATEWAY.WithFields(logrus.Fields{
			"ref": "lora/gateway:sendPullData()",
			"err": err,
		}).Error("can't marshall pull data message")
	}

	if _, err = Conn.Write(pullDataPacket); err != nil {
		LOG_GATEWAY.WithFields(logrus.Fields{
			"ref": "lora/gateway:sendPullData()",
			"err": err,
		}).Error("Can't write pullDataPacket udp")
	}
}

func sendJoinRequestPackets(gateway *model.Gateway, Conn net.Conn) {

	LOG_GATEWAY.WithFields(logrus.Fields{
		"ref": "lora/gateway:sendJoinRequestPackets()",
	}).Info("Sending JoinRequest messages for all the nodes")

	rxpk := make([]loraserver_structs.RXPK, 1)
	for _, node := range gateway.Nodes {
		if !node.JoinedNetwork {
			rxpk[0] = NewRxpk(GetJoinRequestDataPayload(node))
			packet, err := Packet{Rxpk: rxpk}.Prepare(gateway)
			if err != nil {
				LOG_GATEWAY.WithFields(logrus.Fields{
					"ref": "lora/gateway:Start()",
					"err": err,
				}).Error("Can't prepare lora packet")
			}
			if _, err = Conn.Write(packet); err != nil {
				LOG_GATEWAY.WithFields(logrus.Fields{
					"ref": "lora/gateway:Start()",
					"err": err,
				}).Error("Can't write udp")
			}
		}
	}
}

func sendPushPackets(gateway *model.Gateway, Conn net.Conn, fcnt uint32) {
	rxpk := make([]loraserver_structs.RXPK, 1)
	for _, node := range gateway.Nodes {
		rxpk[0] = NewRxpk(GetPushDataPayload(node, fcnt))
		packet, err := Packet{Rxpk: rxpk}.Prepare(gateway)
		if err != nil {
			LOG_GATEWAY.WithFields(logrus.Fields{
				"ref": "lora/gateway:Start()",
				"err": err,
			}).Error("Can't prepare lora packet")
		}
		if _, err = Conn.Write(packet); err != nil {
			LOG_GATEWAY.WithFields(logrus.Fields{
				"ref": "lora/gateway:Start()",
				"err": err,
			}).Error("Can't write udp")
		}
	}
}

func readPackets(Conn net.Conn, poison chan bool, next chan bool, threadListenUdp chan []byte) {
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

			logrus.WithFields(logrus.Fields{
				"ref":  "lora/gateway:readPackets()",
				"SIZE": n,
			}).Info("The size of read buffer")

			if err != nil {
				LOG_GATEWAY.WithFields(logrus.Fields{
					"ref": "lora/gateway:Start()",
					"err": err,
				}).Debug("Can't read udp")
				quit = true
				break
			} else {
				threadListenUdp <- buf[0:n]
			}
		}
		if quit {
			break
		}
	}
}

func readLoraPackets(gateway *model.Gateway, poison chan bool, next chan bool, threadListenUdp chan []byte, endTimer func(), prometheus tools.Prometheus) {

	nbMsg := 0
	localPoison := make(chan bool)

	go func() {
		quit := false
		next <- true
		for {
			select {
			case <-localPoison:
				quit = true
				break
			case res := <-threadListenUdp:
				endTimer()
				err := HandlePacket(res)
				if err != nil {
					LOG_GATEWAY.WithFields(logrus.Fields{
						"ref": "lora/gateway:readLoraPackets()",
						"err": err,
					}).Error("couldn't handle packet")
				}
				LOG_GATEWAY.WithFields(logrus.Fields{
					"ref":             "lora/gateway:Start()",
					"ProtocolVersion": res[0],
					"Token":           res[1:2],
					"ack":             res[3],
					"ackOk?":          res[3] == byte(1),
				}).Debug("Receive data before 1 second")
				if res[3] == byte(1) {
					nbMsg++
				}
			}
			if quit {
				break
			} else {
				next <- true
			}
		}
	}()

	<-time.After(1 * time.Second)
	poison <- true
	localPoison <- true
	if len(gateway.Nodes)-nbMsg > 0 {
		LOG_GATEWAY.WithFields(logrus.Fields{
			"ref": "lora/gateway:Start()",
			"nb":  len(gateway.Nodes) - nbMsg,
		}).Warn("Receive data after 2 second")
	}
	prometheus.AddLongRequest(len(gateway.Nodes) - nbMsg)
}

func RandomEUI() lorawan.EUI64 {
	return lorawan.EUI64{
		tools.RandomBytes(1)[0], tools.RandomBytes(1)[0], tools.RandomBytes(1)[0], tools.RandomBytes(1)[0],
		tools.RandomBytes(1)[0], tools.RandomBytes(1)[0], tools.RandomBytes(1)[0], tools.RandomBytes(1)[0],
	}
}
