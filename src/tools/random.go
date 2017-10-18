package tools

import (
	"crypto/rand"
	"math"
	rmath "math/rand"
	"time"
)

//Random generate random int between min and max
//Deprecated: use Random64 instead
func Random(min, max int) int {
	if min == max {
		return min
	}
	s1 := rmath.NewSource(time.Now().UnixNano())
	r1 := rmath.New(s1)
	res := math.Floor(r1.Float64()*float64(max-min+1)) + float64(min)
	return int(res)
}

//Random64 generate random int64 between min and max
func Random64(min, max int64) int64 {
	if min == max {
		return min
	}
	s1 := rmath.NewSource(time.Now().UnixNano())
	r1 := rmath.New(s1)
	res := math.Floor(r1.Float64()*float64(max-min+1)) + float64(min)
	return int64(res)
}

//RandomBytes generate random int64 between min and max
func RandomBytes(nb int) []byte {
	b := make([]byte, nb)
	rand.Read(b)
	return b
}

//RandomDuration generate random time.Duration between min and max
func RandomDuration(min, max time.Duration) time.Duration {
	return time.Duration(Random64(min.Nanoseconds(), max.Nanoseconds()))
}

//Random2Bytes generate random [2]byte
func Random2Bytes() [2]byte {
	return [2]byte{RandomBytes(1)[0], RandomBytes(1)[0]}
}

//Random4Bytes generate random [4]byte
func Random4Bytes() [4]byte {
	return [4]byte{RandomBytes(1)[0], RandomBytes(1)[0], RandomBytes(1)[0], RandomBytes(1)[0]}
}

//Random8Bytes generate random [8]byte
func Random8Bytes() [8]byte {
	return [8]byte{RandomBytes(1)[0], RandomBytes(1)[0], RandomBytes(1)[0], RandomBytes(1)[0],
		RandomBytes(1)[0], RandomBytes(1)[0], RandomBytes(1)[0], RandomBytes(1)[0]}
}

//Random16Bytes generate random [16]byte
func Random16Bytes() [16]byte {
	return [16]byte{RandomBytes(1)[0], RandomBytes(1)[0], RandomBytes(1)[0], RandomBytes(1)[0],
		RandomBytes(1)[0], RandomBytes(1)[0], RandomBytes(1)[0], RandomBytes(1)[0],
		RandomBytes(1)[0], RandomBytes(1)[0], RandomBytes(1)[0], RandomBytes(1)[0],
		RandomBytes(1)[0], RandomBytes(1)[0], RandomBytes(1)[0], RandomBytes(1)[0]}
}
