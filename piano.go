package main

import (
	"math"
)

type Piano struct {
	Freq float64
	FreqMap map[rune][]int16
}

func (p *Piano) GenerateNote(freq float64, duration int) []int16 {
	buf := make([]int16, duration)
	timer0 := 0.0
	timer1 := 0.0
	timer2 := 0.0
	timer3 := 0.0
	tick0 := 2 * math.Pi / sampleRate64 * freq
	tick1 := tick0 * 2
	tick2 := tick1 * 3
	tick3 := tick2 * 4
	amp := sampleAmp16bit * 0.7
	for i, _ := range buf {
		sin0 := math.Sin(timer0)
		sin1 := sin0 * math.Sin(timer1)
		sin2 := sin1 * math.Sin(timer2)
		sin3 := sin2 * math.Sin(timer3)
		bar0 := amp * sin0
		bar1 := bar0 * sin1/2 * sin0
		bar2 := bar0 * sin2/3 * sin0
		bar3 := bar0 * sin3/4 * sin0
		buf[i] = int16(bar0+bar1+bar2+bar3)
		timer0 += tick0
		timer1 += tick1
		timer2 += tick2
		timer3 += tick3
	}
	return trimWave(buf)
}

func (p *Piano) GetNote(key rune) ([]int16, bool) {
	buf, found := p.FreqMap[key]
	return buf, found
}
