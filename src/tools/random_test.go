package tools

import (
	"testing"
	"time"
)

var toTests = [][2]int{
	{0, 0},
	{0, 100},
	{10, 500},
	{1, 2},
	{1, 10000},
	{1, 5},
}

func TestRandom(t *testing.T) {
	for _, test := range toTests {
		res := Random(test[0], test[1])
		if res < test[0] || res > test[1] {
			t.Fatalf("Random give %d but it outside of %d - %d", res, test[0], test[1])
		}
	}
}

var toTests64 = [][2]int64{
	{0, 0},
	{0, 100},
	{10, 500},
	{1, 2},
	{1, 10000},
	{1, 5},
}

func TestRandom64(t *testing.T) {
	for _, test := range toTests64 {
		res := Random64(test[0], test[1])
		if res < test[0] || res > test[1] {
			t.Fatalf("Random give %d but it outside of %d - %d", res, test[0], test[1])
		}
	}
}

var toTestsDurations = [][2]time.Duration{
	{time.Duration(0), time.Duration(0)},
	{time.Duration(0), time.Duration(100)},
	{time.Duration(10 * time.Hour), time.Duration(500 * time.Hour)},
	{time.Duration(1), time.Duration(2)},
	{time.Duration(1), time.Duration(10000)},
	{time.Duration(1), time.Duration(5)},
}

func TestRandomDuration(t *testing.T) {
	for _, test := range toTestsDurations {
		res := RandomDuration(test[0], test[1])
		if res < test[0] || res > test[1] {
			t.Fatalf("Random give %d but it outside of %d - %d", res, test[0], test[1])
		}
	}
}

var toTestsBytes = []int{1, 10, 100}

func TestRandomBytes(t *testing.T) {
	for _, test := range toTestsBytes {
		b := RandomBytes(test)
		if len(b) != test {
			t.Fatalf("Random bytes give %d bytes instead of %d bytes", len(b), test)
		}
	}
}
