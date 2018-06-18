package lora

import (
	"errors"
	"lorhammer/src/model"
	"lorhammer/src/tools"
	"net"
	"sync"
	"testing"
	"time"
)

type fakeConn struct {
	writed bool
	readed bool
}

func (fc *fakeConn) Read(b []byte) (n int, err error) {
	b = []byte{0, 1, 2, 3, 4, 5}
	fc.readed = true
	return len(b), nil
}
func (fc *fakeConn) Write(b []byte) (n int, err error) {
	if len(b) > 0 {
		fc.writed = true
	}
	return 0, nil
}
func (fc *fakeConn) Close() error                       { return nil }
func (fc *fakeConn) LocalAddr() net.Addr                { return nil }
func (fc *fakeConn) RemoteAddr() net.Addr               { return nil }
func (fc *fakeConn) SetDeadline(t time.Time) error      { return nil }
func (fc *fakeConn) SetReadDeadline(t time.Time) error  { return nil }
func (fc *fakeConn) SetWriteDeadline(t time.Time) error { return nil }

type fakeConnErrorReading struct {
	writed bool
	readed bool
	mutex  sync.Mutex
}

func (fc *fakeConnErrorReading) Read(b []byte) (n int, err error) {
	fc.mutex.Lock()
	defer fc.mutex.Unlock()
	fc.readed = true
	return 0, errors.New("Fake error")
}
func (fc *fakeConnErrorReading) Write(b []byte) (n int, err error)  { return 0, nil }
func (fc *fakeConnErrorReading) Close() error                       { return nil }
func (fc *fakeConnErrorReading) LocalAddr() net.Addr                { return nil }
func (fc *fakeConnErrorReading) RemoteAddr() net.Addr               { return nil }
func (fc *fakeConnErrorReading) SetDeadline(t time.Time) error      { return nil }
func (fc *fakeConnErrorReading) SetReadDeadline(t time.Time) error  { return nil }
func (fc *fakeConnErrorReading) SetWriteDeadline(t time.Time) error { return nil }

type fakePrometheus struct {
	nbPushAckLongRequest  int
	nbPullRespLongRequest int
}

func (fp *fakePrometheus) StartPushAckTimer() func()  { return nil }
func (fp *fakePrometheus) StartPullRespTimer() func() { return nil }
func (fp *fakePrometheus) AddGateway(nb int)          {}
func (fp *fakePrometheus) SubGateway(nb int)          {}
func (fp *fakePrometheus) AddNodes(nb int)            {}
func (fp *fakePrometheus) SubNodes(nb int)            {}
func (fp *fakePrometheus) AddPushAckLongRequest(nb int) {
	fp.nbPushAckLongRequest = nb
}
func (fp *fakePrometheus) AddPullRespLongRequest(nb int) {
	fp.nbPullRespLongRequest = nb
}

func TestIsGatewayScenarioCompleted(t *testing.T) {

	gateway := &LorhammerGateway{
		PayloadsReplayMaxLaps: 0,
	}

	isCompleted := gateway.isGatewayScenarioCompleted()
	if isCompleted {
		t.Fatal("Scenario completed with 0 max laps")
	}

	gateway.PayloadsReplayMaxLaps = 1
	isCompleted = gateway.isGatewayScenarioCompleted()
	if !isCompleted {
		t.Fatal("Scenario not completed with 1 max laps")
	}

	gateway.Nodes = []*model.Node{}
	isCompleted = gateway.isGatewayScenarioCompleted()
	if !isCompleted {
		t.Fatal("Scenario not completed with 1 max laps and empty node array")
	}

	gateway.Nodes = []*model.Node{
		{
			PayloadsReplayLap: 0,
		},
	}
	isCompleted = gateway.isGatewayScenarioCompleted()
	if isCompleted {
		t.Fatal("Scenario not completed with 1 max laps and node array not overlaped")
	}

	gateway.Nodes = []*model.Node{
		{
			PayloadsReplayLap: 1,
		},
	}
	isCompleted = gateway.isGatewayScenarioCompleted()
	if !isCompleted {
		t.Fatal("Scenario not completed with 1 max laps and node array overlaped")
	}

	gateway.Nodes = []*model.Node{
		{
			PayloadsReplayLap: 0,
		},
		{
			PayloadsReplayLap: 1,
		},
	}
	isCompleted = gateway.isGatewayScenarioCompleted()
	if isCompleted {
		t.Fatal("Scenario not completed with 1 max laps and node array at least one overlaped")
	}
}

func TestConvertToGateway(t *testing.T) {
	gateway := &LorhammerGateway{
		Nodes: []*model.Node{
			{
				PayloadsReplayLap: 0,
			},
		},
		NsAddress:             "addressNS",
		MacAddress:            tools.Random8Bytes(),
		AllLapsCompleted:      false,
		PayloadsReplayMaxLaps: 1,
		RxpkDate:              10,
		ReceiveTimeoutTime:    time.Second,
	}

	convertedGateway := gateway.ConvertToGateway()

	if !(convertedGateway.ReceiveTimeoutTime == gateway.ReceiveTimeoutTime &&
		convertedGateway.RxpkDate == gateway.RxpkDate &&
		convertedGateway.PayloadsReplayMaxLaps == gateway.PayloadsReplayMaxLaps &&
		convertedGateway.MacAddress == gateway.MacAddress &&
		convertedGateway.NsAddress == gateway.NsAddress &&
		len(convertedGateway.Nodes) == len(gateway.Nodes) &&
		convertedGateway.Nodes[0] == gateway.Nodes[0]) {
		t.Fatal("Converted gateway is not identical to the original")
	}
}

func TestSendPullData(t *testing.T) {
	var fakeConnect net.Conn = &fakeConn{}
	gateway := &LorhammerGateway{}

	gateway.sendPullData(fakeConnect)
	if !fakeConnect.(*fakeConn).writed {
		t.Fatal("Nothing writted on the connection")
	}
}

func TestNoSendTxAckPacket(t *testing.T) {
	var fakeConnect net.Conn = &fakeConn{}
	gateway := &LorhammerGateway{}

	gateway.sendTxAckPacket(fakeConnect, []byte{0, 1})
	if fakeConnect.(*fakeConn).writed {
		t.Fatal("Data writted on the connection")
	}
}

func TestSendTxAckPacket(t *testing.T) {
	var fakeConnect net.Conn = &fakeConn{}
	gateway := &LorhammerGateway{}

	gateway.sendTxAckPacket(fakeConnect, []byte{2, 0, 0, 3, 123, 34, 116, 120, 112, 107, 34, 58, 123, 34, 105, 109, 109, 101, 34, 58, 102, 97, 108, 115, 101, 44, 34, 116, 109, 115, 116, 34, 58, 49, 49, 50, 51, 52, 53, 54, 44, 34, 102, 114, 101, 113, 34, 58, 56, 54, 54, 46, 51, 52, 57, 56, 49, 50, 44, 34, 114, 102, 99, 104, 34, 58, 48, 44, 34, 112, 111, 119, 101, 34, 58, 49, 52, 44, 34, 109, 111, 100, 117, 34, 58, 34, 76, 79, 82, 65, 34, 44, 34, 100, 97, 116, 114, 34, 58, 34, 83, 70, 55, 66, 87, 49, 50, 53, 34, 44, 34, 99, 111, 100, 114, 34, 58, 34, 52, 47, 54, 34, 44, 34, 105, 112, 111, 108, 34, 58, 116, 114, 117, 101, 44, 34, 115, 105, 122, 101, 34, 58, 49, 50, 44, 34, 100, 97, 116, 97, 34, 58, 34, 89, 76, 72, 107, 89, 86, 48, 103, 65, 65, 67, 100, 103, 55, 118, 122, 34, 125, 125})
	if !fakeConnect.(*fakeConn).writed {
		t.Fatal("Data not writted on the connection")
	}
}

func TestReadPackets(t *testing.T) {
	var fakeConnect net.Conn = &fakeConn{}
	gateway := &LorhammerGateway{}

	var poison, next, threadListenUDP = make(chan bool), make(chan bool), make(chan []byte, 1)
	go gateway.readPackets(fakeConnect, poison, next, threadListenUDP)
	next <- true

	time.Sleep(time.Second * 1)
	select {
	case data := <-threadListenUDP:
		if len(data) != 6 {
			t.Fatal("Wrong read data")
		}
		if !fakeConnect.(*fakeConn).readed {
			t.Fatal("No data readed")
		}
	default:
		t.Fatal("No data received")
	}

	poison <- true
	close(poison)
	close(next)
	close(threadListenUDP)
}

func TestReadPacketsNoNext(t *testing.T) {
	var fakeConnect net.Conn = &fakeConn{}
	gateway := &LorhammerGateway{}

	var poison, next, threadListenUDP = make(chan bool), make(chan bool), make(chan []byte)
	go gateway.readPackets(fakeConnect, poison, next, threadListenUDP)
	poison <- true

	time.Sleep(time.Second)
	select {
	case <-threadListenUDP:
		t.Fatal("Data received not expected")
	default:
		break
	}

	if fakeConnect.(*fakeConn).readed {
		t.Fatal("Data readed not expected")
	}

	close(poison)
	close(next)
	close(threadListenUDP)
}

func TestReadPacketsErrorReading(t *testing.T) {
	var fakeConn net.Conn = &fakeConnErrorReading{
		mutex: sync.Mutex{},
	}
	gateway := &LorhammerGateway{}

	var poison, next, threadListenUDP = make(chan bool), make(chan bool), make(chan []byte)
	go gateway.readPackets(fakeConn, poison, next, threadListenUDP)
	next <- true

	select {
	case <-threadListenUDP:
		t.Fatal("Data received not expected")
	default:
		break
	}

	time.Sleep(time.Second)
	fakeConn.(*fakeConnErrorReading).mutex.Lock()
	if !fakeConn.(*fakeConnErrorReading).readed {
		t.Fatal("Data not readed")
	}
	fakeConn.(*fakeConnErrorReading).mutex.Unlock()

	close(poison)
	close(next)
	close(threadListenUDP)
}

func TestSendJoinRequestPackets(t *testing.T) {
	var fakeConnect net.Conn = &fakeConn{}
	gateway := &LorhammerGateway{
		Nodes: []*model.Node{
			{
				PayloadsReplayLap: 0,
				AppEUI:            tools.Random8Bytes(),
				DevEUI:            tools.Random8Bytes(),
			},
		},
	}

	gateway.sendJoinRequestPackets(fakeConnect)

	if !fakeConnect.(*fakeConn).writed {
		t.Fatal("No data writed")
	}
}

func TestSendJoinRequestPacketsNoNode(t *testing.T) {
	var fakeConnect net.Conn = &fakeConn{}
	gateway := &LorhammerGateway{}

	gateway.sendJoinRequestPackets(fakeConnect)

	if fakeConnect.(*fakeConn).writed {
		t.Fatal("Data writed")
	}
}

func TestSendPushPacketNoNode(t *testing.T) {
	var fakeConnect net.Conn = &fakeConn{}
	gateway := &LorhammerGateway{
		PayloadsReplayMaxLaps: 1,
	}

	gateway.sendPushPackets(fakeConnect, 10)

	if fakeConnect.(*fakeConn).writed {
		t.Fatal("Data writed")
	}
	if !gateway.AllLapsCompleted {
		t.Fatal("All laps not completed")
	}
}

func TestSendPushPacket(t *testing.T) {
	var fakeConnect net.Conn = &fakeConn{}
	gateway := &LorhammerGateway{
		Nodes: []*model.Node{
			{
				AppEUI:  tools.Random8Bytes(),
				DevEUI:  tools.Random8Bytes(),
				DevAddr: getDevAddrFromDevEUI(tools.Random8Bytes()),
				NwSKey:  getGenericAES128Key(),
			},
		},
	}

	gateway.sendPushPackets(fakeConnect, 10)

	if !fakeConnect.(*fakeConn).writed {
		t.Fatal("Data not writed")
	}
	if gateway.AllLapsCompleted {
		t.Fatal("All laps completed")
	}
}

func TestSendPushPacketEmptyNode(t *testing.T) {
	var fakeConnect net.Conn = &fakeConn{}
	gateway := &LorhammerGateway{
		Nodes: []*model.Node{},
	}

	gateway.sendPushPackets(fakeConnect, 10)

	if fakeConnect.(*fakeConn).writed {
		t.Fatal("Data writed")
	}
	if gateway.AllLapsCompleted {
		t.Fatal("All laps completed")
	}
}

func TestReadLoraPacketsNoPacket(t *testing.T) {
	gateway := LorhammerGateway{
		ReceiveTimeoutTime: time.Second,
	}
	poison, next, threadListenUDP := make(chan bool, 1), make(chan bool, 1), make(chan []byte)
	endPushAckTimerFlag, endPullRespTimerFlag := false, false
	endPushAckTimer, endPullRespTimer := func() {
		endPushAckTimerFlag = true
	}, func() {
		endPullRespTimerFlag = true
	}

	nbReceivedAckMsg, nbReceivedPullRespMsg := gateway.readLoraPackets(nil, poison, next, threadListenUDP, endPushAckTimer, endPullRespTimer)

	if nbReceivedAckMsg > 0 || nbReceivedPullRespMsg > 0 {
		t.Fatal("Received messages different from 0")
	}

	if !<-poison {
		t.Fatal("Poison channel not contains true")
	}

	if !<-next {
		t.Fatal("Next channel not contains true")
	}

	if endPushAckTimerFlag || endPullRespTimerFlag {
		t.Fatal("Wrong called end timer funcs")
	}

	close(poison)
	close(next)
	close(threadListenUDP)
}

func TestReadLoraPacketsPushAcks(t *testing.T) {
	gateway := LorhammerGateway{
		ReceiveTimeoutTime: time.Second,
	}
	poison, next, threadListenUDP := make(chan bool, 1), make(chan bool, 2), make(chan []byte, 1)
	endPushAckTimerFlag, endPullRespTimerFlag := false, false
	endPushAckTimer, endPullRespTimer := func() {
		endPushAckTimerFlag = true
	}, func() {
		endPullRespTimerFlag = true
	}

	threadListenUDP <- []byte{2, 165, 210, 1}
	nbReceivedAckMsg, nbReceivedPullRespMsg := gateway.readLoraPackets(nil, poison, next, threadListenUDP, endPushAckTimer, endPullRespTimer)

	if nbReceivedAckMsg != 1 || nbReceivedPullRespMsg != 0 {
		t.Fatal("Wrong number of received messages")
	}

	if !<-poison {
		t.Fatal("Poison channel not contains true")
	}

	if !<-next {
		t.Fatal("Next channel not contains true")
	}

	if !endPushAckTimerFlag || endPullRespTimerFlag {
		t.Fatal("Wrong called end timer funcs")
	}

	close(poison)
	close(next)
	close(threadListenUDP)
}

func TestReadLoraPacketsWrongPacket(t *testing.T) {
	gateway := LorhammerGateway{
		ReceiveTimeoutTime: time.Second,
	}
	poison, next, threadListenUDP := make(chan bool, 1), make(chan bool, 2), make(chan []byte, 1)
	endPushAckTimerFlag, endPullRespTimerFlag := false, false
	endPushAckTimer, endPullRespTimer := func() {
		endPushAckTimerFlag = true
	}, func() {
		endPullRespTimerFlag = true
	}

	threadListenUDP <- []byte{0, 0}
	nbReceivedAckMsg, nbReceivedPullRespMsg := gateway.readLoraPackets(nil, poison, next, threadListenUDP, endPushAckTimer, endPullRespTimer)

	if nbReceivedAckMsg != 0 || nbReceivedPullRespMsg != 0 {
		t.Fatal("Wrong number of received messages")
	}

	if !<-poison {
		t.Fatal("Poison channel not contains true")
	}

	if !<-next {
		t.Fatal("Next channel not contains true")
	}

	if endPushAckTimerFlag || endPullRespTimerFlag {
		t.Fatal("Wrong called end timer funcs")
	}

	close(poison)
	close(next)
	close(threadListenUDP)
}

func TestReadLoraPacketsPullResp(t *testing.T) {
	var fakeConn net.Conn = &fakeConn{}
	gateway := LorhammerGateway{
		ReceiveTimeoutTime: time.Second,
	}
	poison, next, threadListenUDP := make(chan bool, 1), make(chan bool, 2), make(chan []byte, 1)
	endPushAckTimerFlag, endPullRespTimerFlag := false, false
	endPushAckTimer, endPullRespTimer := func() {
		endPushAckTimerFlag = true
	}, func() {
		endPullRespTimerFlag = true
	}

	threadListenUDP <- []byte{2, 0, 0, 3, 123, 34, 116, 120, 112, 107, 34, 58, 123, 34, 105, 109, 109, 101, 34, 58, 102, 97, 108, 115, 101, 44, 34, 116, 109, 115, 116, 34, 58, 49, 49, 50, 51, 52, 53, 54, 44, 34, 102, 114, 101, 113, 34, 58, 56, 54, 54, 46, 51, 52, 57, 56, 49, 50, 44, 34, 114, 102, 99, 104, 34, 58, 48, 44, 34, 112, 111, 119, 101, 34, 58, 49, 52, 44, 34, 109, 111, 100, 117, 34, 58, 34, 76, 79, 82, 65, 34, 44, 34, 100, 97, 116, 114, 34, 58, 34, 83, 70, 55, 66, 87, 49, 50, 53, 34, 44, 34, 99, 111, 100, 114, 34, 58, 34, 52, 47, 54, 34, 44, 34, 105, 112, 111, 108, 34, 58, 116, 114, 117, 101, 44, 34, 115, 105, 122, 101, 34, 58, 49, 50, 44, 34, 100, 97, 116, 97, 34, 58, 34, 89, 76, 72, 107, 89, 86, 48, 103, 65, 65, 67, 100, 103, 55, 118, 122, 34, 125, 125}
	nbReceivedAckMsg, nbReceivedPullRespMsg := gateway.readLoraPackets(fakeConn, poison, next, threadListenUDP, endPushAckTimer, endPullRespTimer)

	if nbReceivedAckMsg != 0 || nbReceivedPullRespMsg != 1 {
		t.Fatal("Wrong number of received messages")
	}

	if !<-poison {
		t.Fatal("Poison channel not contains true")
	}

	if !<-next {
		t.Fatal("Next channel not contains true")
	}

	if endPushAckTimerFlag || !endPullRespTimerFlag {
		t.Fatal("Wrong called end timer funcs")
	}

	close(poison)
	close(next)
	close(threadListenUDP)
}

func TestReadLoraPushPacketsPushAck(t *testing.T) {
	gateway := LorhammerGateway{
		ReceiveTimeoutTime: time.Second,
		Nodes:              []*model.Node{{}, {}},
	}
	poison, next, threadListenUDP := make(chan bool, 1), make(chan bool, 2), make(chan []byte, 1)
	endPushAckTimerFlag, endPullRespTimerFlag := false, false
	endPushAckTimer, endPullRespTimer := func() {
		endPushAckTimerFlag = true
	}, func() {
		endPullRespTimerFlag = true
	}

	fakePrometheus := &fakePrometheus{}

	threadListenUDP <- []byte{2, 165, 210, 1}
	gateway.readLoraPushPackets(nil, poison, next, threadListenUDP, endPushAckTimer, endPullRespTimer, fakePrometheus)

	if fakePrometheus.nbPushAckLongRequest != 1 || fakePrometheus.nbPullRespLongRequest != 2 {
		t.Fatal("Wrong number in Prometheus")
	}

	if !<-poison {
		t.Fatal("Poison channel not contains true")
	}

	if !<-next {
		t.Fatal("Next channel not contains true")
	}

	if !endPushAckTimerFlag || endPullRespTimerFlag {
		t.Fatal("Wrong called end timer funcs")
	}

	close(poison)
	close(next)
	close(threadListenUDP)
}

func TestReadLoraPushPacketsPullResp(t *testing.T) {
	var fakeConn net.Conn = &fakeConn{}
	gateway := LorhammerGateway{
		ReceiveTimeoutTime: time.Second,
		Nodes:              []*model.Node{{}, {}},
	}
	poison, next, threadListenUDP := make(chan bool, 1), make(chan bool, 2), make(chan []byte, 1)
	endPushAckTimerFlag, endPullRespTimerFlag := false, false
	endPushAckTimer, endPullRespTimer := func() {
		endPushAckTimerFlag = true
	}, func() {
		endPullRespTimerFlag = true
	}

	fakePrometheus := &fakePrometheus{}

	threadListenUDP <- []byte{2, 0, 0, 3, 123, 34, 116, 120, 112, 107, 34, 58, 123, 34, 105, 109, 109, 101, 34, 58, 102, 97, 108, 115, 101, 44, 34, 116, 109, 115, 116, 34, 58, 49, 49, 50, 51, 52, 53, 54, 44, 34, 102, 114, 101, 113, 34, 58, 56, 54, 54, 46, 51, 52, 57, 56, 49, 50, 44, 34, 114, 102, 99, 104, 34, 58, 48, 44, 34, 112, 111, 119, 101, 34, 58, 49, 52, 44, 34, 109, 111, 100, 117, 34, 58, 34, 76, 79, 82, 65, 34, 44, 34, 100, 97, 116, 114, 34, 58, 34, 83, 70, 55, 66, 87, 49, 50, 53, 34, 44, 34, 99, 111, 100, 114, 34, 58, 34, 52, 47, 54, 34, 44, 34, 105, 112, 111, 108, 34, 58, 116, 114, 117, 101, 44, 34, 115, 105, 122, 101, 34, 58, 49, 50, 44, 34, 100, 97, 116, 97, 34, 58, 34, 89, 76, 72, 107, 89, 86, 48, 103, 65, 65, 67, 100, 103, 55, 118, 122, 34, 125, 125}
	gateway.readLoraPushPackets(fakeConn, poison, next, threadListenUDP, endPushAckTimer, endPullRespTimer, fakePrometheus)

	if fakePrometheus.nbPushAckLongRequest != 2 || fakePrometheus.nbPullRespLongRequest != 1 {
		t.Fatal("Wrong number in Prometheus")
	}

	if !<-poison {
		t.Fatal("Poison channel not contains true")
	}

	if !<-next {
		t.Fatal("Next channel not contains true")
	}

	if endPushAckTimerFlag || !endPullRespTimerFlag {
		t.Fatal("Wrong called end timer funcs")
	}

	close(poison)
	close(next)
	close(threadListenUDP)
}

func TestReadLoraJoinPacketsPushAck(t *testing.T) {
	gateway := LorhammerGateway{
		ReceiveTimeoutTime: time.Second,
		Nodes:              []*model.Node{{}, {}},
	}
	poison, next, threadListenUDP := make(chan bool, 1), make(chan bool, 2), make(chan []byte, 1)
	endPushAckTimerFlag, endPullRespTimerFlag := false, false
	endPushAckTimer, endPullRespTimer := func() {
		endPushAckTimerFlag = true
	}, func() {
		endPullRespTimerFlag = true
	}

	fakePrometheus := &fakePrometheus{}

	threadListenUDP <- []byte{2, 165, 210, 1}
	gateway.readLoraJoinPackets(nil, poison, next, threadListenUDP, endPushAckTimer, endPullRespTimer, fakePrometheus, false)

	if fakePrometheus.nbPushAckLongRequest != 0 || fakePrometheus.nbPullRespLongRequest != 1 {
		t.Fatal("Wrong number in Prometheus")
	}

	if !<-poison {
		t.Fatal("Poison channel not contains true")
	}

	if !<-next {
		t.Fatal("Next channel not contains true")
	}

	if !endPushAckTimerFlag || endPullRespTimerFlag {
		t.Fatal("Wrong called end timer funcs")
	}

	close(poison)
	close(next)
	close(threadListenUDP)
}

func TestReadLoraJoinPacketsPushAckWithJoin(t *testing.T) {
	gateway := LorhammerGateway{
		ReceiveTimeoutTime: time.Second,
		Nodes:              []*model.Node{{}, {}},
	}
	poison, next, threadListenUDP := make(chan bool, 1), make(chan bool, 2), make(chan []byte, 1)
	endPushAckTimerFlag, endPullRespTimerFlag := false, false
	endPushAckTimer, endPullRespTimer := func() {
		endPushAckTimerFlag = true
	}, func() {
		endPullRespTimerFlag = true
	}

	fakePrometheus := &fakePrometheus{}

	threadListenUDP <- []byte{2, 165, 210, 1}
	gateway.readLoraJoinPackets(nil, poison, next, threadListenUDP, endPushAckTimer, endPullRespTimer, fakePrometheus, true)

	if fakePrometheus.nbPushAckLongRequest != 2 || fakePrometheus.nbPullRespLongRequest != 3 {
		t.Fatal("Wrong number in Prometheus")
	}

	if !<-poison {
		t.Fatal("Poison channel not contains true")
	}

	if !<-next {
		t.Fatal("Next channel not contains true")
	}

	if !endPushAckTimerFlag || endPullRespTimerFlag {
		t.Fatal("Wrong called end timer funcs")
	}

	close(poison)
	close(next)
	close(threadListenUDP)
}

func TestReadLoraJoinPacketsPullResp(t *testing.T) {
	var fakeConn net.Conn = &fakeConn{}
	gateway := LorhammerGateway{
		ReceiveTimeoutTime: time.Second,
		Nodes:              []*model.Node{{}, {}},
	}
	poison, next, threadListenUDP := make(chan bool, 1), make(chan bool, 2), make(chan []byte, 1)
	endPushAckTimerFlag, endPullRespTimerFlag := false, false
	endPushAckTimer, endPullRespTimer := func() {
		endPushAckTimerFlag = true
	}, func() {
		endPullRespTimerFlag = true
	}

	fakePrometheus := &fakePrometheus{}

	threadListenUDP <- []byte{2, 0, 0, 3, 123, 34, 116, 120, 112, 107, 34, 58, 123, 34, 105, 109, 109, 101, 34, 58, 102, 97, 108, 115, 101, 44, 34, 116, 109, 115, 116, 34, 58, 49, 49, 50, 51, 52, 53, 54, 44, 34, 102, 114, 101, 113, 34, 58, 56, 54, 54, 46, 51, 52, 57, 56, 49, 50, 44, 34, 114, 102, 99, 104, 34, 58, 48, 44, 34, 112, 111, 119, 101, 34, 58, 49, 52, 44, 34, 109, 111, 100, 117, 34, 58, 34, 76, 79, 82, 65, 34, 44, 34, 100, 97, 116, 114, 34, 58, 34, 83, 70, 55, 66, 87, 49, 50, 53, 34, 44, 34, 99, 111, 100, 114, 34, 58, 34, 52, 47, 54, 34, 44, 34, 105, 112, 111, 108, 34, 58, 116, 114, 117, 101, 44, 34, 115, 105, 122, 101, 34, 58, 49, 50, 44, 34, 100, 97, 116, 97, 34, 58, 34, 89, 76, 72, 107, 89, 86, 48, 103, 65, 65, 67, 100, 103, 55, 118, 122, 34, 125, 125}
	gateway.readLoraJoinPackets(fakeConn, poison, next, threadListenUDP, endPushAckTimer, endPullRespTimer, fakePrometheus, false)

	if fakePrometheus.nbPushAckLongRequest != 1 || fakePrometheus.nbPullRespLongRequest != 0 {
		t.Fatal("Wrong number in Prometheus")
	}

	if !<-poison {
		t.Fatal("Poison channel not contains true")
	}

	if !<-next {
		t.Fatal("Next channel not contains true")
	}

	if endPushAckTimerFlag || !endPullRespTimerFlag {
		t.Fatal("Wrong called end timer funcs")
	}

	close(poison)
	close(next)
	close(threadListenUDP)
}

func TestReadLoraJoinPacketsPullRespWithJoin(t *testing.T) {
	var fakeConn net.Conn = &fakeConn{}
	gateway := LorhammerGateway{
		ReceiveTimeoutTime: time.Second,
		Nodes:              []*model.Node{{}, {}},
	}
	poison, next, threadListenUDP := make(chan bool, 1), make(chan bool, 2), make(chan []byte, 1)
	endPushAckTimerFlag, endPullRespTimerFlag := false, false
	endPushAckTimer, endPullRespTimer := func() {
		endPushAckTimerFlag = true
	}, func() {
		endPullRespTimerFlag = true
	}

	fakePrometheus := &fakePrometheus{}

	threadListenUDP <- []byte{2, 0, 0, 3, 123, 34, 116, 120, 112, 107, 34, 58, 123, 34, 105, 109, 109, 101, 34, 58, 102, 97, 108, 115, 101, 44, 34, 116, 109, 115, 116, 34, 58, 49, 49, 50, 51, 52, 53, 54, 44, 34, 102, 114, 101, 113, 34, 58, 56, 54, 54, 46, 51, 52, 57, 56, 49, 50, 44, 34, 114, 102, 99, 104, 34, 58, 48, 44, 34, 112, 111, 119, 101, 34, 58, 49, 52, 44, 34, 109, 111, 100, 117, 34, 58, 34, 76, 79, 82, 65, 34, 44, 34, 100, 97, 116, 114, 34, 58, 34, 83, 70, 55, 66, 87, 49, 50, 53, 34, 44, 34, 99, 111, 100, 114, 34, 58, 34, 52, 47, 54, 34, 44, 34, 105, 112, 111, 108, 34, 58, 116, 114, 117, 101, 44, 34, 115, 105, 122, 101, 34, 58, 49, 50, 44, 34, 100, 97, 116, 97, 34, 58, 34, 89, 76, 72, 107, 89, 86, 48, 103, 65, 65, 67, 100, 103, 55, 118, 122, 34, 125, 125}
	gateway.readLoraJoinPackets(fakeConn, poison, next, threadListenUDP, endPushAckTimer, endPullRespTimer, fakePrometheus, true)

	if fakePrometheus.nbPushAckLongRequest != 3 || fakePrometheus.nbPullRespLongRequest != 2 {
		t.Fatal("Wrong number in Prometheus")
	}

	if !<-poison {
		t.Fatal("Poison channel not contains true")
	}

	if !<-next {
		t.Fatal("Next channel not contains true")
	}

	if endPushAckTimerFlag || !endPullRespTimerFlag {
		t.Fatal("Wrong called end timer funcs")
	}

	close(poison)
	close(next)
	close(threadListenUDP)
}
