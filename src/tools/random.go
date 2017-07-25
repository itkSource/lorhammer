package tools

import (
	"crypto/rand"
	"github.com/Sirupsen/logrus"
	"math"
	rmath "math/rand"
	"time"
)

/*
 * Deprecated use Random64 instead
 */
func Random(min, max int) int {
	if min == max {
		return min
	}
	s1 := rmath.NewSource(time.Now().UnixNano())
	r1 := rmath.New(s1)
	res := math.Floor(r1.Float64()*float64(max-min+1)) + float64(min)
	return int(res)
}

func Random64(min, max int64) int64 {
	if min == max {
		return min
	}
	s1 := rmath.NewSource(time.Now().UnixNano())
	r1 := rmath.New(s1)
	res := math.Floor(r1.Float64()*float64(max-min+1)) + float64(min)
	return int64(res)
}

func RandomBytes(nb int) []byte {
	b := make([]byte, nb)
	_, err := rand.Read(b)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"ref": "mqtt/mqtt:Init()",
			"err": err,
		}).Info("Can't random bytes")
	}
	return b
}

func RandomDuration(min, max time.Duration) time.Duration {
	return time.Duration(Random64(min.Nanoseconds(), max.Nanoseconds()))
}

func Random2Bytes() [2]byte {
	return [2]byte{RandomBytes(1)[0], RandomBytes(1)[0]}
}

func Random4Bytes() [4]byte {
	return [4]byte{RandomBytes(1)[0], RandomBytes(1)[0], RandomBytes(1)[0], RandomBytes(1)[0]}
}

func Random8Bytes() [8]byte {
	return [8]byte{RandomBytes(1)[0], RandomBytes(1)[0], RandomBytes(1)[0], RandomBytes(1)[0],
		RandomBytes(1)[0], RandomBytes(1)[0], RandomBytes(1)[0], RandomBytes(1)[0]}
}

func Random16Bytes() [16]byte {
	return [16]byte{RandomBytes(1)[0], RandomBytes(1)[0], RandomBytes(1)[0], RandomBytes(1)[0],
		RandomBytes(1)[0], RandomBytes(1)[0], RandomBytes(1)[0], RandomBytes(1)[0],
		RandomBytes(1)[0], RandomBytes(1)[0], RandomBytes(1)[0], RandomBytes(1)[0],
		RandomBytes(1)[0], RandomBytes(1)[0], RandomBytes(1)[0], RandomBytes(1)[0]}
}
