package testType

import (
	"lorhammer/src/model"
	"sync"
	"testing"
	"time"
)

func TestOneShot(t *testing.T) {
	mqtt := &fakeMqtt{mu: sync.Mutex{}}
	go startOneShot(Test{testType: TypeOneShot}, model.Init{}, mqtt)

	time.Sleep(100 * time.Millisecond)

	mqtt.mu.Lock()
	defer mqtt.mu.Unlock()
	if mqtt.nbCall != 1 {
		t.Fatalf("OneShot test should call 1 time mqtt instead of %s", mqtt.nbCall)
	}

}
