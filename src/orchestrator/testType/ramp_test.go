package testType

import (
	"lorhammer/src/model"
	"sync"
	"testing"
	"time"
)

func TestNominal(t *testing.T) {
	mqtt := &fakeMqtt{mu: sync.Mutex{}}
	go startRamp(Test{testType: TypeRamp, rampTime: time.Duration(1 * time.Second)}, model.Init{NbGateway: 100}, mqtt)

	time.Sleep(1 * time.Second)

	mqtt.mu.Lock()
	defer mqtt.mu.Unlock()
	if mqtt.nbCall != 1 {
		t.Fatalf("Ramp test should call 1 time mqtt instead of %s", mqtt.nbCall)
	}
}

func TestDecimal(t *testing.T) {
	mqtt := &fakeMqtt{mu: sync.Mutex{}}
	ramp := &ramp{
		nbGateway:        1,
		timeMinute:       time.Duration(10 * time.Minute).Minutes(),
		currentRest:      float64(0),
		gatewaysLaunched: 0,
		intervalRampTime: time.Duration(1 * time.Millisecond),
	}
	go ramp.start(model.Init{}, mqtt)

	time.Sleep(100 * time.Millisecond) // 100 ticks because each tick has 1 millisecond duration (e.g. intervalRampTime)

	mqtt.mu.Lock()
	defer mqtt.mu.Unlock()
	if mqtt.nbCall != 1 {
		t.Fatal("For 1 gateways in 10 minutes we need to launch 1 gateway")
	}
}

func TestWithoutTime(t *testing.T) {
	mqtt := &fakeMqtt{mu: sync.Mutex{}}
	ramp := &ramp{
		nbGateway:        10,
		timeMinute:       time.Duration(0 * time.Minute).Minutes(),
		currentRest:      float64(0),
		gatewaysLaunched: 0,
		intervalRampTime: time.Duration(1 * time.Millisecond),
	}
	go ramp.start(model.Init{}, mqtt)

	time.Sleep(1 * time.Millisecond)

	mqtt.mu.Lock()
	defer mqtt.mu.Unlock()
	if mqtt.nbCall != 1 {
		t.Fatal("For 10 gateways in 0 minute we need to launch 10 gateways")
	}
}
