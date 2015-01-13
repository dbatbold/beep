package main

import (
	"math"
)
type Piano struct {
	Freq float64
}

func (p *Piano) GenerateNote(freq float64, duration int) []byte {
	buf := make([]byte, duration)
	timer := 0.0
	timer1 := 0.0
	timer2 := 0.0
	timer3 := 0.0
	tick := 2 * math.Pi / sampleRate64 * freq
	tick1 := 2 * math.Pi / sampleRate64 * freq * 2
	tick2 := 2 * math.Pi / sampleRate64 * freq * 4
	tick3 := 2 * math.Pi / sampleRate64 * freq * 10
	for i, _ := range buf {
		bar := 60.0 * math.Sin(timer)
		bar1 := 60.0 * math.Sin(timer1)
		bar2 := (bar1 / 3) * math.Sin(timer2)
		bar3 := (bar2) * math.Sin(timer3) * math.Cos(timer)
		buf[i] = 127 + byte(bar-bar1-bar2-bar3)
		timer += tick
		timer1 += tick1
		timer2 += tick2
		timer3 += tick3
	}
	return trimWave(buf)
}
