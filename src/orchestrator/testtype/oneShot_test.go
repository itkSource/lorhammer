package testtype

import (
	"lorhammer/src/model"
	"lorhammer/src/orchestrator/command"
	"sync"
	"testing"
	"time"
)

func TestOneShot(t *testing.T) {
	mqtt := &fakeMqtt{mu: sync.Mutex{}}
	command.NewLorhammer(model.NewLorhammer{CallbackTopic: "topic1"})
	go startOneShot(Test{testType: typeOneShot}, []model.Init{{}}, mqtt)

	time.Sleep(100 * time.Millisecond)

	mqtt.mu.Lock()
	defer mqtt.mu.Unlock()
	if mqtt.nbCall != 1 {
		t.Fatalf("OneShot test should call 1 time mqtt instead of %d", mqtt.nbCall)
	}

}
