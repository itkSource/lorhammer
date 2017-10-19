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

func TestRandom2Bytes(t *testing.T) {
	if len(Random2Bytes()) != 2 {
		t.Fatalf("Random2Bytes should return a slice of 2 bytes")
	}
	nbSame := 0
	for i := 0; i < 100000; i++ {
		first := Random2Bytes()
		second := Random2Bytes()
		if first[0] == second[0] && first[1] == second[1] {
			nbSame++
		}
	}
	if nbSame > 10 {
		t.Fatalf("Random2Bytes should be random generated")
	}
}

func TestRandom4Bytes(t *testing.T) {
	if len(Random4Bytes()) != 4 {
		t.Fatalf("Random4Bytes should return a slice of 4 bytes")
	}
	nbSame := 0
	for i := 0; i < 100000; i++ {
		first := Random4Bytes()
		second := Random4Bytes()
		if first[0] == second[0] && first[1] == second[1] && first[2] == second[2] && first[3] == second[3] {
			nbSame++
		}
	}
	if nbSame > 10 {
		t.Fatalf("Random4Bytes should be random generated")
	}
}

func TestRandom8Bytes(t *testing.T) {
	if len(Random8Bytes()) != 8 {
		t.Fatalf("Random8Bytes should return a slice of 8 bytes")
	}
	nbSame := 0
	for i := 0; i < 100000; i++ {
		first := Random8Bytes()
		second := Random8Bytes()
		if first[0] == second[0] && first[1] == second[1] && first[2] == second[2] && first[3] == second[3] &&
			first[4] == second[4] && first[5] == second[5] && first[6] == second[6] && first[7] == second[7] {
			nbSame++
		}
	}
	if nbSame > 10 {
		t.Fatalf("Random8Bytes should be random generated")
	}
}

func TestRandom16Bytes(t *testing.T) {
	if len(Random16Bytes()) != 16 {
		t.Fatalf("Random16Bytes should return a slice of 16 bytes")
	}
	nbSame := 0
	for i := 0; i < 100000; i++ {
		first := Random16Bytes()
		second := Random16Bytes()
		if first[0] == second[0] && first[1] == second[1] && first[2] == second[2] && first[3] == second[3] &&
			first[4] == second[4] && first[5] == second[5] && first[6] == second[6] && first[7] == second[7] &&
			first[8] == second[8] && first[9] == second[9] && first[10] == second[10] && first[11] == second[11] &&
			first[12] == second[12] && first[13] == second[13] && first[14] == second[14] && first[15] == second[15] {
			nbSame++
		}
	}
	if nbSame > 10 {
		t.Fatalf("Random16Bytes should be random generated")
	}
}
