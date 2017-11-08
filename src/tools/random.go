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
	r := RandomBytes(2)
	return [2]byte{r[0], r[1]}
}

//Random4Bytes generate random [4]byte
func Random4Bytes() [4]byte {
	r := RandomBytes(4)
	return [4]byte{r[0], r[1], r[2], r[3]}
}

//Random8Bytes generate random [8]byte
func Random8Bytes() [8]byte {
	r := RandomBytes(8)
	return [8]byte{r[0], r[1], r[2], r[3],
		r[4], r[5], r[6], r[7]}
}

//Random16Bytes generate random [16]byte
func Random16Bytes() [16]byte {
	r := RandomBytes(16)
	return [16]byte{r[0], r[1], r[2], r[3],
		r[4], r[5], r[6], r[7],
		r[8], r[9], r[10], r[11],
		r[12], r[13], r[14], r[15]}
}
