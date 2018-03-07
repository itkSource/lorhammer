package testtype

import (
	"lorhammer/src/model"
	"sync"
	"testing"
	"time"
)

type fakeMqtt struct {
	mu     sync.Mutex
	nbCall int
}

func (*fakeMqtt) GetAddress() string                                          { return "" }
func (*fakeMqtt) Connect() error                                              { return nil }
func (*fakeMqtt) Disconnect()                                                 {}
func (*fakeMqtt) Handle(topics []string, handle func(message []byte)) error   { return nil }
func (*fakeMqtt) HandleCmd(topics []string, handle func(cmd model.CMD)) error { return nil }
func (f *fakeMqtt) PublishCmd(topic string, cmdName model.CommandName) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.nbCall++
	return nil
}
func (f *fakeMqtt) PublishSubCmd(topic string, cmdName model.CommandName, subCmd interface{}) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.nbCall++
	return nil
}

func TestRepeat(t *testing.T) {
	mqtt := &fakeMqtt{mu: sync.Mutex{}}
	go startRepeat(Test{testType: typeRepeat, repeatTime: time.Duration(1 * time.Second)}, []model.Init{{}}, mqtt)

	time.Sleep(3500 * time.Millisecond)

	mqtt.mu.Lock()
	defer mqtt.mu.Unlock()
	if mqtt.nbCall != 3 {
		t.Fatalf("Repeat test should call 3 time mqtt instead of %d", mqtt.nbCall)
	}

}
