package provisioning

import (
	"bufio"
	"encoding/json"
	"fmt"
	"github.com/sirupsen/logrus"
	"github.com/brocaar/lorawan"
	"lorhammer/src/model"
	"net"
)

//deleteAllData yesREALLY
//app add <eui> <name>
//gateway add <eui>
//mote add <eui mote> ota app <eui app> key <key app>

const SemtechV4Type = Type("semtechV4")

type SemtechV4 struct {
	NsAddress string `json:"nsAddress"`
	AsAddress string `json:"asAddress"`
	CsAddress string `json:"csAddress"`
	NcAddress string `json:"ncAddress"`
	ackreq    int
	chanAck   chan bool
}

type Command struct {
	Command string `json:"command"`
	Ackreq  int    `json:"ackreq"`
}

type ack struct {
	ack int
}

var LOG_SEMTECHV4 = logrus.WithFields(logrus.Fields{"logger": "orchestrator/provisioning/semtechv4"})

func NewSemtechV4(rawConfig json.RawMessage) (provisioner, error) {
	config := &SemtechV4{
		ackreq:  0,
		chanAck: make(chan bool),
	}
	if err := json.Unmarshal(rawConfig, config); err != nil {
		return nil, err
	}
	return config, nil
}

func (config *SemtechV4) Provision(sensorsToRegister model.Register) error {
	defer close(config.chanAck)

	ConnNs := startConn(config.NsAddress)
	defer ConnNs.Close()
	go logInput(ConnNs, config.chanAck)

	ConnAs := startConn(config.AsAddress)
	defer ConnAs.Close()
	go logInput(ConnAs, config.chanAck)

	ConnCs := startConn(config.CsAddress)
	defer ConnCs.Close()
	go logInput(ConnCs, config.chanAck)

	ConnNc := startConn(config.NcAddress)
	defer ConnNc.Close()
	go logInput(ConnNc, config.chanAck)

	config.start(ConnNs, sensorsToRegister, true, true)
	config.start(ConnAs, sensorsToRegister, false, true)
	config.start(ConnCs, sensorsToRegister, false, false)
	config.start(ConnNc, sensorsToRegister, false, false)

	return nil
}

func (config *SemtechV4) DeProvision() error {
	ConnNs := startConn(config.NsAddress)
	defer ConnNs.Close()
	ConnAs := startConn(config.AsAddress)
	defer ConnAs.Close()
	ConnCs := startConn(config.CsAddress)
	defer ConnCs.Close()
	ConnNc := startConn(config.NcAddress)
	defer ConnNc.Close()
	cleanAll := Command{
		Command: "deleteAllData yesREALLY",
	}
	if err := config.sendCommand(ConnNs, cleanAll); err != nil {
		return err
	}
	if err := config.sendCommand(ConnAs, cleanAll); err != nil {
		return err
	}
	if err := config.sendCommand(ConnCs, cleanAll); err != nil {
		return err
	}
	if err := config.sendCommand(ConnNc, cleanAll); err != nil {
		return err
	}
	LOG_SEMTECHV4.Info("DeProvisioning ok")
	return nil
}

func startConn(url string) net.Conn {
	Conn, err := net.Dial("udp", url)
	if err != nil {
		LOG_SEMTECHV4.WithError(err).Error("Can't contact semtech ns")
	}
	return Conn
}

func (config *SemtechV4) start(Conn net.Conn, sensorsToRegister model.Register, addGateways bool, sensorWithKey bool) {
	for _, gateway := range sensorsToRegister.Gateways {
		for index, sensor := range gateway.Nodes {
			config.addApp(Conn, sensor.AppEUI, index)
		}
		if addGateways {
			config.addGateway(Conn, gateway.MacAddress)
		}
		//for _, sensor := range gateway.Nodes {
		//	if sensorWithKey {
		//		config.addSensorWithKey(Conn, sensor)
		//	} else {
		//		config.addSensor(Conn, sensor)
		//	}
		//}
	}
}

func (config *SemtechV4) addApp(Conn net.Conn, appEUI lorawan.EUI64, index int) {
	addApp := Command{
		Command: fmt.Sprintf("app add %s lorhammer_%d", appEUI, index),
	}
	if err := config.sendCommand(Conn, addApp); err != nil {
		LOG_SEMTECHV4.WithError(err).Error("Can't send message")
	} else {
		LOG_SEMTECHV4.WithField("appEui", appEUI).Info("App added")
	}
}

func (config *SemtechV4) addGateway(Conn net.Conn, gatewayEUI lorawan.EUI64) {
	addGateway := Command{
		Command: fmt.Sprintf("gateway add %s", gatewayEUI.String()),
	}
	if err := config.sendCommand(Conn, addGateway); err != nil {
		LOG_SEMTECHV4.WithError(err).Error("Can't send message")
	}
}

func (config *SemtechV4) addSensorWithKey(Conn net.Conn, sensor *model.Node) {
	addMote := Command{
		Command: fmt.Sprintf("mote add %s ota app %s key %s", sensor.DevEUI, sensor.AppEUI, sensor.AppKey),
	}
	if err := config.sendCommand(Conn, addMote); err != nil {
		LOG_SEMTECHV4.WithError(err).Error("Can't send message")
	} else {
		LOG_SEMTECHV4.WithFields(logrus.Fields{
			"appEui": sensor.AppEUI,
			"appKey": sensor.AppKey,
			"devEui": sensor.DevEUI,
		}).Info("Sensor registered")
	}
}

func (config *SemtechV4) addSensor(Conn net.Conn, sensor *model.Node) {
	addMote := Command{
		Command: fmt.Sprintf("mote add %s app %s", sensor.DevEUI, sensor.AppEUI),
	}
	if err := config.sendCommand(Conn, addMote); err != nil {
		LOG_SEMTECHV4.WithError(err).Error("Can't send message")
	} else {
		LOG_SEMTECHV4.WithFields(logrus.Fields{
			"appEui": sensor.AppEUI,
			"appKey": sensor.AppKey,
			"devEui": sensor.DevEUI,
		}).Info("Sensor registered")
	}
}

func (config *SemtechV4) sendCommand(Conn net.Conn, command Command) error {
	command.Ackreq = config.ackreq
	serializedAddMote, err := json.Marshal(command)
	if err != nil {
		return err
	}
	_, err2 := Conn.Write(serializedAddMote)
	if err2 != nil {
		return err2
	}
	<-config.chanAck
	config.ackreq++
	return nil
}

func logInput(Conn net.Conn, chanAck chan bool) {
	reader := bufio.NewReader(Conn)
	for {
		msg, err := reader.ReadSlice(byte('}'))
		if err != nil {
			LOG_SEMTECHV4.WithError(err).Error("Can't read tcp")
			break
		} else {
			LOG_SEMTECHV4.WithFields(logrus.Fields{
				"message": string(msg),
				"from":    Conn.RemoteAddr().String(),
			}).Info("Received tcp message")
			chanAck <- true
		}
	}
}
