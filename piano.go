package beep

import (
	"archive/zip"
	"fmt"
	"io/ioutil"
	"math"
	"os"
	"path/filepath"
	"strings"
)

// Piano voice
type Piano struct {
	naturalVoice      bool
	naturalVoiceFound bool
	keyDefMap         map[rune][]int16 // default voice
	keyNatMap         map[rune][]int16 // natural voice
	keyFreqMap        map[rune]float64
	keyNoteMap        map[rune]string
	noteKeyMap        map[string]rune
}

// NewPiano returns new piano voice
func NewPiano() *Piano {
	p := &Piano{
		keyDefMap:  make(map[rune][]int16),
		keyNatMap:  make(map[rune][]int16),
		keyFreqMap: make(map[rune]float64),
		keyNoteMap: make(map[rune]string),
		noteKeyMap: make(map[string]rune),
	}

	keys := "q2w3er5t6y7ui9o0p[=]azsxcfvgbnjmk,l."

	octaveFreq0 := []float64{
		// A0, B0b, B0
		27.50, 29.13, 30.86, // octave 0
	}
	octaveFreqLeft := []float64{
		// C1, Db1, D1, Eb1, E1, F1, Gb1, G1, Ab1, A1, Bb1, B1
		32.70, 34.64, 36.70, 38.89, 41.20, 43.65, 46.24, 48.99, 51.91, 55.00, 58.27, 61.73, // 1
		65.40, 69.29, 73.41, 77.78, 82.40, 87.30, 92.49, 97.99, 103.8, 110.0, 116.5, 123.4, // 2
		130.8, 138.5, 146.8, 155.5, 164.8, 174.6, 185.0, 196.0, 207.6, 220.0, 233.0, 246.9, // 3
	}
	octaveFreqRight := []float64{
		// C4, Db4, D4, Eb4, E4, F4, Gb4, G4, Ab4, A4, Bb4, B4
		261.6, 277.1, 293.6, 311.1, 329.6, 349.2, 369.9, 392.0, 415.3, 440.0, 466.1, 493.8, // 4
		523.2, 554.3, 587.3, 622.2, 659.2, 698.4, 739.9, 783.9, 830.6, 880.0, 932.3, 987.7, // 5
		1046.5, 1108.7, 1174.6, 1244.5, 1318.5, 1396.9, 1479.9, 1567, 1661, 1760, 1864, 1975, // 6
	}
	octaveFreq78 := []float64{
		// C7, Db7, D7, Eb7, E7, F7, Gb7, G7, Ab7, A7, Bb7, B7
		2093, 2217.5, 2349.3, 2489, 2637, 2793, 2960, 3136, 3322.4, 3520, 3729.3, 3951.1, // 7
		4186.00, // 8
	}

	noteNames := []string{
		"A0", "Bb0", "B0",
		"C1", "Db1", "D1", "Eb1", "E1", "F1", "Gb1", "G1", "Ab1", "A1", "Bb1", "B1",
		"C2", "Db2", "D2", "Eb2", "E2", "F2", "Gb2", "G2", "Ab2", "A2", "Bb2", "B2",
		"C3", "Db3", "D3", "Eb3", "E3", "F3", "Gb3", "G3", "Ab3", "A3", "Bb3", "B3",
		"C4", "Db4", "D4", "Eb4", "E4", "F4", "Gb4", "G4", "Ab4", "A4", "Bb4", "B4",
		"C5", "Db5", "D5", "Eb5", "E5", "F5", "Gb5", "G5", "Ab5", "A5", "Bb5", "B5",
		"C6", "Db6", "D6", "Eb6", "E6", "F6", "Gb6", "G6", "Ab6", "A6", "Bb6", "B6",
		"C7", "Db7", "D7", "Eb7", "E7", "F7", "Gb7", "G7", "Ab7", "A7", "Bb7", "B7",
		"C8",
	}

	// initialize maps
	ni := 0
	for i, key := range keys[33:] { // actave 0
		keyID := 1000 + key
		note := noteNames[ni]
		p.keyFreqMap[keyID] = octaveFreq0[i]
		p.keyNoteMap[keyID] = note
		p.noteKeyMap[note] = keyID
		ni++
	}
	for i, key := range keys { // actave 1, 2, 3
		keyID := 2000 + key
		note := noteNames[ni]
		p.keyFreqMap[keyID] = octaveFreqLeft[i]
		p.keyNoteMap[keyID] = note
		p.noteKeyMap[note] = keyID
		ni++
	}
	for i, key := range keys { // actave 4, 5, 6
		keyID := 3000 + key
		note := noteNames[ni]
		p.keyFreqMap[keyID] = octaveFreqRight[i]
		p.keyNoteMap[keyID] = note
		p.noteKeyMap[note] = keyID
		ni++
	}
	for i, key := range keys[:13] { // actave 7, 8
		keyID := 4000 + key
		note := noteNames[ni]
		p.keyFreqMap[keyID] = octaveFreq78[i]
		p.keyNoteMap[keyID] = note
		p.noteKeyMap[note] = keyID
		ni++
	}
	for key := range p.keyFreqMap {
		// generate default piano voice
		p.keyDefMap[key] = p.generateNote(key, wholeNote)
	}

	// load natural voice file, if exists
	filename := filepath.Join(HomeDir(), "voices", "piano.zip")
	voiceFile, err := zip.OpenReader(filename)
	if err == nil {
		// voice file exists
		defer voiceFile.Close()
		p.naturalVoice = true
		for _, zfile := range voiceFile.File {
			file, err := zfile.Open()
			if err != nil {
				fmt.Fprintln(os.Stderr, "Unable to open file from zip:", zfile.Name)
				continue
			}
			defer file.Close()
			if !strings.HasSuffix(zfile.Name, ".wav") {
				continue
			}
			p.naturalVoiceFound = true
			noteName := strings.Split(filepath.Base(zfile.Name), ".")[0]
			if key, found := p.noteKeyMap[noteName]; found {
				var header WaveHeader
				header.ReadHeader(file)
				if header.SampleRate != 44100 || header.BitsPerSample != 16 {
					fmt.Fprintln(os.Stderr, "Unsupported sample file:", zfile.Name)
					continue
				}
				buf, err := ioutil.ReadAll(file)
				if err != nil {
					fmt.Fprintln(os.Stderr, "Unable to read file from zip:", zfile.Name)
					continue
				}
				rest := wholeNote - len(buf)/2
				if rest > 0 {
					// too short, sample should be a whole note
					bufRest := make([]byte, rest*2)
					buf = append(buf, bufRest...)
				}
				buf16 := byteToInt16Buf(buf)
				if len(buf16) < wholeNote {
					fmt.Fprintln(os.Stderr, "Sample note duration must be 90112 samples long.")
				}
				trimWave(buf16)
				p.keyNatMap[key] = buf16
			} else {
				fmt.Fprintln(os.Stderr, "Unknown note name in voice file:", noteName)
			}
		}
	}

	return p
}

func (p *Piano) generateNote(key rune, duration int) []int16 {
	// default voice
	freq, found := p.keyFreqMap[key]
	if !found {
		fmt.Fprintln(os.Stderr, "frequency not found: key", key)
		return []int16{}
	}
	buf := make([]int16, duration)
	timer0 := 0.0
	timer1 := 0.0
	timer2 := 0.0
	timer3 := 0.0
	tick0 := 2 * math.Pi / SampleRate64 * freq
	tick1 := tick0 * 2
	tick2 := tick1 * 3
	tick3 := tick2 * 4
	amp := SampleAmp16bit * 0.5
	for i := range buf {
		sin0 := math.Sin(timer0)
		sin1 := sin0 * math.Sin(timer1)
		sin2 := sin1 * math.Sin(timer2)
		sin3 := sin2 * math.Sin(timer3)
		bar0 := amp * sin0
		bar1 := bar0 * sin1 / 2 * sin0
		bar2 := bar0 * sin2 / 3 * sin0
		bar3 := bar0 * sin3 / 4 * sin0
		buf[i] = int16(bar0 + bar1 + bar2 + bar3)
		timer0 += tick0
		timer1 += tick1
		timer2 += tick2
		timer3 += tick3
	}
	trimWave(buf)
	return buf
}

// GetNote prepares piano note wave buffer
func (p *Piano) GetNote(note *Note, sustain *Sustain) (found bool) {
	var bufNote []int16
	if p.naturalVoice {
		bufNote, found = p.keyNatMap[note.key]
	}
	if !found {
		bufNote, found = p.keyDefMap[note.key]
	}
	if !found {
		return
	}

	buf := make([]int16, len(bufNote))
	copy(buf, bufNote) // get a copy of the note
	applyNoteVolume(buf, note.volume, note.amplitude)

	// Sustain note
	if note.duration == 'W' {
		// Whole note
		if p.NaturalVoice() {
			// mix with previous sustain note
			mixSoundWave(buf, sustain.buf)
			// sustain current note
			copyBuffer(sustain.buf, buf[len(buf)/3:])
		}
	} else {
		if p.NaturalVoice() {
			// mix with previous sustain note
			mixSoundWave(buf[:note.samples], sustain.buf)
			// sustain current note
			copyBuffer(sustain.buf, buf[note.samples:])
		}
	}

	// measure note
	if n := note.samples - len(buf); n > 0 {
		// expand buffer
		buf = append(buf, make([]int16, n)...)
	}
	buf = buf[:note.samples]

	// clean note
	trimWave(buf)

	// release note
	releaseNote(buf, 0, 0.99)

	note.buf = buf
	return
}

// Sustain flag
func (p *Piano) Sustain() bool {
	return true
}

// NaturalVoice flag
func (p *Piano) NaturalVoice() bool {
	return p.naturalVoice
}

// NaturalVoiceFound flag
func (p *Piano) NaturalVoiceFound() bool {
	return p.naturalVoiceFound
}

// ComputerVoice enables or disables computer voice
func (p *Piano) ComputerVoice(enable bool) {
	p.naturalVoice = !enable
}

// SustainNote applies sustain settings to note
func (p *Piano) SustainNote(note *Note, sustain *Sustain) {
	buflen := len(note.buf)
	volume64 := float64(note.volume)

	if p.naturalVoice {
		attack := float64(9-sustain.attack) / 100 * 2
		raiseNote(note.buf, attack)
		raiseNote(sustain.buf, 0.01)
		release := float64(1+sustain.release) / 10
		releaseNote(note.buf, 0, release)
		releaseNote(sustain.buf, 0, 0.5)
		applyNoteVolume(sustain.buf, note.volume, sustain.sustain)
		return
	}

	// Sustain default voice amplitude for ADSR phases
	// |  /|\            A - attack
	// | / | \ _____     D - decay
	// |/  |  |    | \   S - sustain
	// |--------------   R - release
	//   A  D  S    R

	if note.volume == 0 {
		return
	}

	sustain.attack = 9 // overriding
	sustain.sustain = 4
	attack := int(float64(buflen/200) * float64(sustain.attack))
	decay := (buflen-attack)/10 + ((buflen - attack) / 20 * sustain.decay)
	S := int16(volume64 / 10.0 * float64(sustain.sustain+1))
	sustainCount := (buflen - attack - decay) / 2
	R := buflen - attack - decay - sustainCount
	attack64 := float64(attack)
	decay64 := float64(decay)
	sustain64 := float64(S)
	release64 := float64(R)
	countD := 0.0
	countR := 0.0

	for i, bar := range note.buf {
		i64 := float64(i)
		bar64 := float64(bar)
		if i >= attack+decay+sustainCount {
			// Release phase, decay volume to zero
			bar64 = bar64 * ((sustain64 * (release64 - countR) / release64) / volume64)
			countR++
		} else if i >= attack+decay {
			// Sustain phase, hold volume on sustain level
			bar64 = bar64 * (sustain64 / volume64)
		} else if i >= attack && decay > 0 {
			// Decay phase, decay volume to sustain level
			gap := (volume64 - sustain64) * ((decay64 - countD) / decay64)
			bar64 = bar64 * ((sustain64 + gap) / volume64)
			countD++
		} else if i <= attack && attack > 0 {
			// Attack phase, raise volume to max
			bar64 = bar64 * (i64 / attack64)
		}
		note.buf[i] = int16(bar64)
	}
}
