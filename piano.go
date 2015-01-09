package main

import (
	"bytes"
	"math"
)

type PianoString struct {
	Freq float64
	Length float64
}

// Simulates waveform of piano string
func (p *PianoString) GenerateWave(force byte) []byte {
	samples := int(sampleRate64/p.Freq)
	tick := 2.0*math.Pi/float64(samples)
	bufString := make([]byte, samples)
	bufWaveLimit := 1024*1024*1 
	var bufWave bytes.Buffer

	timer := 0.0
	for i := 0; i<500; i++ {
		for s := 0; s < samples; s++ {
			vib := 5.0*math.Sin(timer/10)
			amp := 10.0*math.Sin(timer/20)
			bar := 127 + byte(vib*100.0*math.Sin(timer) + amp)
			bufString[s] = bar
			timer += tick
		}
		bufWave.Write(bufString)
		if bufWave.Len() > bufWaveLimit {
			// too big
			break
		}
	}
	return bufWave.Bytes()
}
