package main

import (
	"math"
)

type Violin struct {
	Freq    float64
	FreqMap map[rune][]int16
}

func (v *Violin) GenerateNote(freq float64, duration int) []int16 {
	buf := make([]int16, duration)
	timer0 := 0.0
	timer1 := 0.0
	timer2 := 0.0
	timer3 := 0.0
	timer4 := 0.0
	tick0 := 2 * math.Pi / sampleRate64 * freq
	tick1 := tick0 * 2
	tick2 := tick0 * 3
	tick3 := tick0 * 4
	tick4 := tick0 * 5
	amp := sampleAmp16bit * 0.5
	for i, _ := range buf {
		sin0 := math.Sin(timer0)
		sin1 := math.Sin(timer1)
		bar0 := amp * sin0
		bar1 := bar0 * sin1 * sin0
		bar2 := bar1 * math.Sin(timer2) * sin0
		bar3 := bar2 * math.Sin(timer3) * sin0
		bar4 := 0.0 //bar0/2 * math.Sin(timer4)
		buf[i] = 127 + int16(bar0+bar1+bar2+bar3+bar4)
		timer0 += tick0
		timer1 += tick1
		timer2 += tick2
		timer3 += tick3
		timer4 += tick4
	}
	return trimWave(buf)
}

func (v *Violin) GetNote(key rune) ([]int16, bool) {
	buf, found := v.FreqMap[key]
	return buf, found
}
