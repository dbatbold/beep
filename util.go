package main

import (
	"math"
)

func stringToBytes(s string) []byte {
	var buf []byte
	for _, c := range s {
		buf = append(buf, byte(c))
	}
	return buf
}

func int32ToBytes(i int) []byte {
	var buf [4]byte
	n := int32(i)
	buf[0] = byte(n)
	buf[1] = byte(n >> 8)
	buf[2] = byte(n >> 16)
	buf[3] = byte(n >> 24)
	return buf[:]
}

func int16ToBytes(i int) []byte {
	n := int32(i)
	var buf [2]byte
	buf[0] = byte(n)
	buf[1] = byte(n >> 8)
	return buf[:]
}

// Converts Hertz to frequency unit
func hertzToFreq(hertz float64) float64 {
	// 1 second = 44100 samples
	// 1 hertz = freq * 2Pi
	// freq = 2Pi / 44100 * hertz
	freq := 2.0 * math.Pi / sampleRate64 * hertz
	return freq
}
