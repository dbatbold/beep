package main

import (
	"testing"
)

func TestMusicNotes(t *testing.T) {
	t.Skip()
	// print waveform
	playMusicNotes(100, "T6ic")
}

func TestHertzToFreq(t *testing.T) {
	t.Skip()
	freq := hertzToFreq(261.6)
	t.Log(freq)
}
