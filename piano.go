package main

import (
	"bytes"
	"math"
)

var pianoNoteC2buf []byte

type Piano struct {
	Freq float64
}

// Simulates waveform of piano string
func (p *Piano) GenerateWave(force byte) []byte {
	samples := int(sampleRate64 / p.Freq)
	tick := 2.0 * math.Pi / float64(samples)
	bufString := make([]byte, samples)
	bufWaveLimit := 1024 * 1024 * 1
	var bufWave bytes.Buffer

	timer := 0.0
	for i := 0; i < 500; i++ {
		for s := 0; s < samples; s++ {
			vib := 5.0 * math.Sin(timer/10)
			amp := 10.0 * math.Sin(timer/20)
			bar := 127 + byte(vib*100.0*math.Sin(timer)+amp)
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

func (p *Piano) GenerateNote(freq float64, duration int) []byte {
	buf := make([]byte, duration)
	timer := 0.0
	timer1 := 0.0
	timer2 := 0.0
	timer3 := 0.0
	tick := 2*math.Pi/sampleRate64*freq
	tick1 := 2*math.Pi/sampleRate64*freq*2
	tick2 := 2*math.Pi/sampleRate64*freq*4
	tick3 := 2*math.Pi/sampleRate64*freq*8
	for i, _ := range buf {
		bar := 60.0*math.Sin(timer)
		bar1 := 20*math.Sin(timer1)
		bar2 := 0.0//bar1*math.Sin(timer2)
		bar3 := 0.0//bar2*math.Sin(timer3)
		buf[i] = 127 + byte(bar-bar1-bar2-bar3)
		timer += tick
		timer1 += tick1
		timer2 += tick2
		timer3 += tick3
	}

	/*
	opt := os.O_WRONLY|os.O_CREATE|os.O_TRUNC
	sampleFile, err := os.OpenFile("test.wav", opt, 0644)
	if err != nil {
		panic(err)
	}
	bufLen := len(buf)
	defer sampleFile.Close()
	header := NewWaveHeader(1, 44100, 8, bufLen)
	header.WriteHeader(sampleFile)
	sampleFile.Write(buf)
	*/

	return trimWave(buf)
}
