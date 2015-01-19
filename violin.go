package main

import (
	"archive/zip"
	"fmt"
	"io/ioutil"
	"math"
	"os"
	"strings"
	"path/filepath"
)

type Violin struct {
	NaturalVoice bool
	keyDefMap  map[rune][]int16 // default voice
	keyNatMap  map[rune][]int16 // natural voice
	keyFreqMap map[rune]float64
	keyNoteMap map[rune]string
	noteKeyMap map[string]rune
}

func NewViolin() *Violin {
	p := &Violin{
		keyDefMap:  make(map[rune][]int16),
		keyNatMap:  make(map[rune][]int16),
		keyFreqMap: make(map[rune]float64),
		keyNoteMap: make(map[rune]string),
		noteKeyMap: make(map[string]rune),
	}



	keys := "q2w3er5t6y7ui9o0p[=]azsxcfvgbnjmk,l."

	octaveFreq123 := []float64{
		// C1, Db1, D1, Eb1, E1, F1, Gb1, G1, Ab1, A1, Bb1, B1
		32.70, 34.64, 36.70, 38.89, 41.20, 43.65, 46.24, 48.99, 51.91, 55.00, 58.27, 61.73, // 1
		65.40, 69.29, 73.41, 77.78, 82.40, 87.30, 92.49, 97.99, 103.8, 110.0, 116.5, 123.4, // 2
		130.8, 138.5, 146.8, 155.5, 164.8, 174.6, 185.0, 196.0, 207.6, 220.0, 233.0, 246.9, // 3
	}
	octaveFreq456 := []float64{
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
		"G3", "Ab3", "A3", "Bb3", "B3",
		"C4", "Db4", "D4", "Eb4", "E4", "F4", "Gb4", "G4", "Ab4", "A4", "Bb4", "B4",
		"C5", "Db5", "D5", "Eb5", "E5", "F5", "Gb5", "G5", "Ab5", "A5", "Bb5", "B5",
		"C6", "Db6", "D6", "Eb6", "E6", "F6", "Gb6", "G6", "Ab6", "A6", "Bb6", "B6",
		"C7", "Db7", "D7", "Eb7", "E7",
	}

	// initialize maps
	ni := 0
	for i, key := range keys[31:] { // actave 3
		keyId := 1000 + key
		note := noteNames[ni]
		p.keyFreqMap[keyId] = octaveFreq123[i]
		p.keyNoteMap[keyId] = note
		p.noteKeyMap[note] = keyId
		ni++
	}
	for i, key := range keys {  // actave 4, 5, 6
		keyId := 2000 + key
		note := noteNames[ni]
		p.keyFreqMap[keyId] = octaveFreq456[i]
		p.keyNoteMap[keyId] = note
		p.noteKeyMap[note] = keyId
		ni++
	}
	for i, key := range keys[:5] {  // actave 7
		keyId := 3000 + key
		note := noteNames[ni]
		p.keyFreqMap[keyId] = octaveFreq78[i]
		p.keyNoteMap[keyId] = note
		p.noteKeyMap[note] = keyId
		ni++
	}
	for key, _ := range p.keyFreqMap {
		// generate default violin voice
		p.keyDefMap[key] = p.generateNote(key, quarterNote)
	}

	// load natural voice file, if exists
	filename := beepHomeDir() + "/voices/violin.zip"
	voiceFile, err := zip.OpenReader(filename)
	if err == nil {
		// voice file exists
		defer voiceFile.Close()
		p.NaturalVoice = true
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
				rest := quarterNote - len(buf)/2
				if rest > 0 {
					// too short, sample should be a whole note
					bufRest := make([]byte, rest*2)
					buf = append(buf, bufRest...)
				}
				buf16 := byteToInt16Buf(buf)
				p.keyNatMap[key] = trimWave(buf16[:quarterNote])
			} else {
				fmt.Fprintln(os.Stderr, "Unknown note name in voice file:", noteName)
			}
		}
	}

	return p
}

func (p *Violin) generateNote(key rune, duration int) []int16 {
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
	tick0 := 2 * math.Pi / sampleRate64 * freq
	tick1 := tick0 * 2
	tick2 := tick1 * 3
	tick3 := tick2 * 4
	amp := sampleAmp16bit * 0.5
	for i, _ := range buf {
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
	return trimWave(buf)
}

func (p *Violin) GetNote(key rune) ([]int16, bool) {
	if p.NaturalVoice { 
		buf, found := p.keyNatMap[key]
		if found {
			return buf, found
		}
	}
	buf, found := p.keyDefMap[key]
	return buf, found
}

// Changes note amplitude for ADSR phases
// |  /|\            A - attack
// | / | \ _____     D - decay
// |/  |  |    | \   S - sustain
// |--------------   R - release
//   A  D  S    R
func (p *Violin) SustainNote(buf []int16, volume int16, A, D, S, R int) {
	bufLen := len(buf)
	volume64 := float64(volume)

	if p.NaturalVoice {
		releaseNote(buf, 0, 0.6)
		return
	}

	// Sustain default voice
	if volume == 0 {
		return
	}
	attack := int(float64(bufLen/200) * float64(A))
	decay := (bufLen-attack)/10 + ((bufLen - attack) / 20 * D)
	sustain := int16(volume64 / 10.0 * float64(S+1))
	sustainCount := (bufLen - attack - decay) / 2
	release := bufLen - attack - decay - sustainCount
	attack64 := float64(attack)
	decay64 := float64(decay)
	sustain64 := float64(sustain)
	release64 := float64(release)
	countD := 0.0
	countR := 0.0
	for i, bar := range buf {
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
		buf[i] = int16(bar64)
	}
}
