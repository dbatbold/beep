package beep

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"strings"
)

// BeepNotation description
const BeepNotation = `
Beep notation:
  | | | | | | | | | | | | | | | | | | | | | | 
  |2|3| |5|6|7| |9|0| |=|a|s| |f|g| |j|k|l| |
 | | | | | | | | | | | | | | | | | | | | | | 
 |q|w|e|r|t|y|u|i|o|p|[|]|z|x|c|v|b|n|m|,|.|

 q - middle C (261.63 hertz)

 Left and right hand keys are same. Uppercase 
 letters are control keys. Lowercase letters
 are music notes. Space bar is current duration
 rest. Spaces after first space are ignored.
 Lines start with '#' are ignored.

 Control keys:

 Rest:
 RW     - whole rest
 RH     - half rest
 RQ     - quarter rest
 RE     - eighth rest
 RS     - sixteenth rest
 RT     - thirty-second rest
 RI     - sixty-fourth rest

 Durations:
 DW     - whole note
 DH     - half note
 DQ     - quarter note
 DE     - eighth note
 DS     - sixteenth note
 DT     - thirty-second note
 DI     - sixty-fourth note
 DD     - dotted note (adds half duration)

 Octave:
 H0     - octave 0 keys
 HL     - octave 1, 2, 3 (left hand keys)
 HR     - octave 4, 5, 6 (right hand keys)
 H7     - octave 7, 8 keys

 Tempo:
 T#     - where # is 0-9, default is 4 (1 unit speeds up/down by 4%)

 Sustain:
 SA#    - attack level, where # is 0-9, default is 8
 SD#    - decay level, 0-9, default 4
 SS#    - sustain level, 0-9, default 4
 SR#    - release level, 0-9, default 9

 Voice:
 VD     - Computer generated default voice
 VP     - Piano voice
 VV     - Violin voice
 VN     - If a line ends with 'VN', the next line will be
          played harmony with the line.

 Chord:
 C#     - Play next # notes as a chord, where # is 2-9.
          For example C major chord is "C3qet"

 Amplitude:
 A#     - Changes current amplitude, where # is 1-9, default is 9

 Measures:
 |      - bar, ignored
 ' '    - space, ignored
 Tab    - tab, ignored

 Comments:
 #      - a line comment
 ##     - start or end of a block comment

Demo music: Mozart K33b:`

const (
	// SampleAmp16bit - 16-bit sample amplitude
	SampleAmp16bit = 32767.0

	// SampleRate - sample rate
	SampleRate = 44100

	// SampleRate64 - float64 sample rate
	SampleRate64 = float64(SampleRate)

	bitsPerSample = 16
	sample16bit   = bitsPerSample == 16
	sampleAmp8bit = 127.0
	quarterNote   = 1024 * 22
	wholeNote     = quarterNote * 4
	halfNote      = wholeNote / 2
)

var (
	// DemoMusic notation
	DemoMusic = BuiltinMusic[0].Notation

	// PrintSheet enables printing beep notation while playing music
	PrintSheet bool

	// PrintNotes enables printing each notes while playing music
	PrintNotes bool

	wholeRest = make([]int16, wholeNote)
)

// Music player
type Music struct {
	playing    bool
	stopping   bool
	quietMode  bool
	played     chan bool // for syncing player
	stopped    chan bool
	linePlayed chan bool // for syncing lines
	piano      *Piano
	violin     *Violin
	output     string // output file name
}

// Note data
type Note struct {
	key       rune
	duration  rune
	dotted    bool
	volume    int
	amplitude int
	tempo     int
	buf       []int16
	velocity  int
	samples   int
}

// Sustain params
type Sustain struct {
	attack  int
	decay   int
	sustain int
	release int
	buf     []int16
}

// Chord params
type Chord struct {
	number int
	count  int
	buf    []int16
}

// Voice interface
// GetNote: Gets a whole note for the key
// SustainNote: Used for sustaining computer generated voice
// Sustain: Indicates whether the instrument sustain note
// NaturalVoice: Indicates whether natural voice file is loaded
// ComputerVoice: Enable or disable computer voice
type Voice interface {
	GetNote(note *Note, sustain *Sustain) bool
	SustainNote(note *Note, sustain *Sustain)
	Sustain() bool
	NaturalVoice() bool
	NaturalVoiceFound() bool
	ComputerVoice(enable bool)
}

// NewMusic returns new music for output file name, or stdout if output is empty
func NewMusic(output string) *Music {
	music := &Music{
		played:     make(chan bool),
		stopped:    make(chan bool),
		linePlayed: make(chan bool),
		output:     output,
	}
	return music
}

// Wait until sheet is played
func (m *Music) Wait() {
	<-m.played
}

// WaitLine waits until line is played
func (m *Music) WaitLine() {
	<-m.linePlayed
}

// Ratio returns sustain ratio
func (s *Sustain) Ratio() float64 {
	return float64(s.sustain) / 10.0
}

// Reset chord
func (c *Chord) Reset() {
	c.number = 0
	c.count = 0
	c.buf = nil
}

// Play music score from reader
func (m *Music) Play(reader *bufio.Reader, volume100 int) {
	m.playing = true
	defer func() {
		m.played <- true
		if m.stopping {
			m.stopped <- true
		}
		m.playing = false
	}()

	volume := int(SampleAmp16bit * (float64(volume100) / 100.0))
	outputFileName := m.output

	if m.piano == nil {
		m.piano = NewPiano()
	}

	var outputFile *os.File
	var err error

	// output file
	if len(outputFileName) > 0 {
		if outputFileName == "-" {
			outputFile = os.Stdout
			m.quietMode = true
		} else {
			opt := os.O_WRONLY | os.O_TRUNC | os.O_CREATE
			outputFile, err = os.OpenFile(outputFileName, opt, 0644)
			if err != nil {
				fmt.Fprintln(os.Stderr, "Error opening output file:", err)
				os.Exit(1)
			}
		}
		defer outputFile.Close()
	}

	if m.quietMode {
		PrintSheet = false
		PrintNotes = false
	}

	// sustain state
	sustain := &Sustain{
		attack:  8,
		decay:   4,
		sustain: 4,
		release: 9,
		buf:     make([]int16, quarterNote),
	}

	// read lines
	chord := &Chord{}
	bufWaveLimit := 1024 * 1024 * 100
	controlKeys := "RDHTSAVC"
	measures := "WHQESTI"
	hands := "0LR7"
	zeroToNine := "0123456789"
	tempos := zeroToNine
	amplitudes := zeroToNine
	chordNumbers := zeroToNine
	ignoredKeys := "\t |"
	sustainTypes := "ADSR"
	sustainLevels := zeroToNine
	voiceControls := "DPVN"

	var (
		bufOutput    []int16
		duration     = 'Q' // default note duration
		dotted       bool
		rest         rune
		ctrl         rune
		voice        Voice = m.piano // default voice is piano
		sustainType  rune
		hand         = 'R' // default is middle C octave
		handLevel    rune
		count        int // line counter
		tempo        = 4 // normal speed
		amplitude    = 9 // max volume
		mixNextLine  bool
		bufMix       []int16
		lineMix      string
		waitNext     bool
		blockComment bool
	)

	for {
		line, done := nextMusicLine(reader)
		if done {
			break
		}
		if strings.HasPrefix(line, "#") {
			if strings.HasPrefix(line, "##") {
				// ignore block comment
				if blockComment {
					blockComment = false
				} else {
					blockComment = true
				}
			} else {
				// ignore comments
				if PrintSheet {
					fmt.Println(line)
				}
			}
			continue
		}
		if blockComment {
			continue
		}
		if strings.HasSuffix(line, "VN") {
			// include next line to mixer
			mixNextLine = true
		} else {
			mixNextLine = false
		}
		var bufWave []int16
		for _, key := range line {
			keystr := string(key)
			if strings.ContainsAny(keystr, ignoredKeys) {
				continue
			}
			if ctrl == 0 && strings.ContainsAny(keystr, controlKeys) {
				ctrl = key
				continue
			}
			if ctrl > 0 {
				switch ctrl {
				case 'D': // duration
					if strings.ContainsAny(keystr, measures) {
						duration = key
					}
					if key == 'D' {
						dotted = true
					}
				case 'R': // reset
					if strings.ContainsAny(keystr, measures) {
						rest = key
					}
				case 'H': // hand
					if strings.ContainsAny(keystr, hands) {
						hand = key
					}
				case 'T': // tempo
					if strings.ContainsAny(keystr, tempos) {
						tempo = strings.Index(tempos, keystr)
					}
				case 'S': // sustain
					if strings.ContainsAny(keystr, sustainTypes) {
						sustainType = key
						continue
					}
					if strings.ContainsAny(keystr, sustainLevels) {
						level := strings.Index(sustainLevels, keystr)
						switch sustainType {
						case 'A':
							sustain.attack = level
						case 'D':
							sustain.decay = level
						case 'S':
							sustain.sustain = level
						case 'R':
							sustain.release = level
						}
					}
				case 'A': // amplitude
					if strings.ContainsAny(keystr, amplitudes) {
						amplitude = strings.Index(amplitudes, keystr)
					}
				case 'V': // voice
					if strings.ContainsAny(keystr, voiceControls) {
						switch key {
						case 'D': // default voice
							voice.ComputerVoice(true)
						case 'P':
							voice = m.piano
							if voice.NaturalVoiceFound() {
								voice.ComputerVoice(false)
							}
						case 'V':
							if m.violin == nil {
								m.violin = NewViolin()
							}
							voice = m.violin
							voice.ComputerVoice(false)
						}
					}
				case 'C': // chord
					if strings.ContainsAny(keystr, chordNumbers) {
						chord.count = 0
						chord.number = strings.Index(chordNumbers, keystr)
					}
				}
				if rest > 0 {
					bufRest := restNote(rest, dotted, tempo)
					if bufRest != nil {
						if voice.NaturalVoice() {
							releaseNote(sustain.buf, 0, sustain.Ratio())
							mixSoundWave(bufRest, sustain.buf)
							clearBuffer(sustain.buf)
						}
						bufWave = append(bufWave, bufRest...)
					}
					rest = 0
				}
				ctrl = 0
				continue
			}
			switch hand {
			case '0': // octave 0
				handLevel = 1000
			case 'L': // octave 1, 2, 3
				handLevel = 2000
			case 'R': // octave 4, 5, 6
				handLevel = 3000
			case '7', '8': // octave 7, 8
				handLevel = 4000
			}
			note := &Note{
				key:       handLevel + key,
				volume:    volume,
				amplitude: amplitude,
				duration:  duration,
				dotted:    dotted,
				tempo:     tempo,
				samples:   0,
			}
			note.measure()
			if voice.GetNote(note, sustain) {
				dotted = false
				if chord.number > 0 {
					// playing a chord
					chord.count++
					if chord.buf == nil {
						chord.buf = make([]int16, len(note.buf))
						copy(chord.buf, note.buf)
						if voice.NaturalVoice() {
							//copyBuffer(sustain.buf, chord.buf)
						}
					} else {
						mixSoundWave(chord.buf, note.buf)
						if voice.NaturalVoice() {
							//mixSoundWave(sustain.buf, note.buf)
						}
					}
					if chord.count == chord.number {
						if voice.NaturalVoice() {
							release := len(note.buf) / 10 * sustain.sustain
							ratio := sustain.Ratio()
							releaseNote(sustain.buf, release, ratio)
						}
						note.buf = chord.buf
						chord.Reset()
					} else {
						if PrintNotes {
							fmt.Printf("%v-", m.piano.keyNoteMap[note.key])
						}
						continue
					}
				}
				voice.SustainNote(note, sustain)
				bufWave = append(bufWave, note.buf...)
				if len(bufWave) > bufWaveLimit {
					fmt.Fprintln(os.Stderr, "Line wave buffer exceeds 100MB limit.")
					os.Exit(1)
				}
				if PrintNotes {
					fmt.Printf("%v ", m.piano.keyNoteMap[note.key])
				}
			} else {
				voiceName := strings.Split(fmt.Sprintf("%T", voice), ".")[1]
				noteName := m.piano.keyNoteMap[note.key]
				fmt.Printf("%s: Invalid note: %s (%s)\n", voiceName, keystr, noteName)
			}
		}
		if mixNextLine {
			if bufMix == nil {
				bufMix = make([]int16, len(bufWave))
				copy(bufMix, bufWave)
				lineMix = line
			} else {
				lineMix += "\n" + line
				mixSoundWave(bufMix, bufWave)
			}
			count++
			clearBuffer(sustain.buf)
			continue
		}
		if bufMix != nil {
			mixSoundWave(bufMix, bufWave)
			bufWave = bufMix
			bufMix = nil
			line = lineMix + "\n" + line
		}
		if PrintNotes {
			fmt.Println()
		}
		if outputFile == nil {
			if len(bufWave) > 0 {
				if waitNext {
					m.WaitLine() // wait until previous line is done playing
				}
				// prepare next line while playing
				go m.Playback(bufWave, bufWave)
				if PrintSheet {
					fmt.Println(line)
				}
				waitNext = true
			} else if PrintSheet {
				fmt.Println(line)
			}
		} else {
			// saving to file
			buf := make([]int16, 2*len(bufWave))
			for i, bar := range bufWave {
				buf[i*2] = bar
				buf[i*2+1] = bar
			}
			bufOutput = append(bufOutput, buf...)
			if PrintSheet {
				fmt.Println(line)
			}
		}
		clearBuffer(sustain.buf)
		count++
		if m.stopping {
			break
		}
	}
	if waitNext {
		m.WaitLine()
	}

	if outputFile != nil {
		// save wave to file
		buflen := len(bufOutput)
		header := NewWaveHeader(2, SampleRate, 16, buflen*2)
		_, err = header.WriteHeader(outputFile)
		if err != nil {
			fmt.Fprintln(os.Stderr, "Error writing to output file:", err)
			os.Exit(1)
		}
		buf := int16ToByteBuf(bufOutput)
		_, err := outputFile.Write(buf)
		if err != nil {
			fmt.Fprintln(os.Stderr, "Error writing to output file:", err)
			os.Exit(1)
		}
		if outputFileName != "-" {
			fmt.Printf("wrote %s bytes to '%s'\n", numberComma(int64(len(buf))), outputFileName)
		}
	}
}

// measure sets the number of samples for the node
func (n *Note) measure() {
	var samples int
	length := wholeNote + (wholeNote / 100 * 4 * (4 - n.tempo)) // 4% per tempo unit
	switch n.duration {
	case 'W':
		samples = length
	case 'H':
		samples = length / 2
	case 'Q':
		samples = length / 4
	case 'E':
		samples = length / 8
	case 'S':
		samples = length / 16
	case 'T':
		samples = length / 32
	case 'I':
		samples = length / 64
	}

	if samples > 0 {
		// Apply dot measure
		if n.dotted {
			samples += samples / 2
		}
	}

	n.samples = samples
}

// Reads next line from music sheet
func nextMusicLine(reader *bufio.Reader) (string, bool) {
	var buf bytes.Buffer
	limit := 1024 * 100
	for {
		part, isPrefix, err := reader.ReadLine()
		if err != nil {
			return "", true
		}
		buf.Write(part)
		if buf.Len() > limit {
			fmt.Println("Line exceeds 100KB limit.")
			os.Exit(1)
		}
		if !isPrefix {
			break
		}
	}
	line := buf.String()
	line = strings.Trim(line, " \t")
	return line, false
}

// Changes note amplitude
func applyNoteVolume(buf []int16, volume, amplitude int) {
	volume64 := float64(volume)
	amplitude64 := float64(amplitude)
	for i, bar := range buf {
		bar64 := float64(bar)
		bar64 *= (volume64 / SampleAmp16bit)
		if amplitude64 > 0 {
			bar64 *= (amplitude64 / 9.0)
		}
		buf[i] = int16(bar64)
	}
}

// Mixes two waveform
func mixSoundWave(buf1, buf2 []int16) {
	buflen2 := len(buf2)
	gap := SampleAmp16bit - 500.0
	for i := range buf1 {
		if i == buflen2 {
			break
		}
		bar1 := float64(buf1[i])
		bar2 := float64(buf2[i])
		bar64 := (bar1 - bar2) / 2 * 1.6
		if bar64 > gap {
			bar64 = gap
		} else if bar64 <= -gap {
			bar64 = -gap
		}
		buf1[i] = int16(bar64)
	}
}

func copyBuffer(target, src []int16) {
	bufsize := len(src)
	for i := range target {
		if i < bufsize {
			target[i] = src[i]
		} else {
			target[i] = 0
		}
	}
}

func clearBuffer(buf []int16) {
	for i := range buf {
		buf[i] = 0
	}
}

// Removes sharp edge at the end of waveform
func trimWave(buf []int16) {
	if len(buf) == 0 {
		return
	}
	bufsize := len(buf)
	cut := bufsize - 1
	var last int16
	for i := range buf {
		if i == 0 {
			last = buf[cut]
		}
		if buf[cut] < last {
			// falling
			if buf[cut] < 0 && buf[cut] < 32 {
				break
			}
		}
		last = buf[cut]
		if i > 1024 {
			// too long
			cut = len(buf) - 1
			break
		}
		cut--
		if cut == 0 {
			// volume must be low
			cut = len(buf) - 1
			break
		}
	}
	for i := cut; i < bufsize; i++ {
		buf[i] = 0
	}
}

func releaseNote(buf []int16, duration int, ratio float64) {
	// |    release
	// |----|
	// |      \     buflen
	// |----|--|----|
	//      |  duration > 0
	//      ratio

	buflen := len(buf)
	if duration > 0 && duration < buflen {
		buflen = duration
	}
	release := int(float64(buflen) * ratio)
	decay := float64(buflen - release)
	tick := SampleAmp16bit / decay
	volume := SampleAmp16bit
	for i, bar := range buf {
		bar64 := float64(bar)
		if i >= release && volume > 0 {
			bar64 = bar64 * (volume / SampleAmp16bit)
			buf[i] = int16(bar64)
			volume -= tick
		}
		if volume <= 0 {
			buf[i] = 0
		}
	}
}

func raiseNote(buf []int16, ratio float64) {
	buflen := len(buf)
	raise := float64(buflen) * ratio
	tick := SampleAmp16bit / raise
	volume := 0.0
	for i, bar := range buf {
		bar64 := float64(bar)
		bar64 = bar64 * (volume / SampleAmp16bit)
		buf[i] = int16(bar64)
		volume += tick
		if SampleAmp16bit <= volume {
			break
		}
	}
}

// Returns rest note buffer with tempo
func restNote(rest rune, dotted bool, tempo int) []int16 {
	var samples int
	length := wholeNote + (wholeNote / 100 * 4 * (4 - tempo)) // 4% per tempo unit
	switch rest {
	case 'W':
		samples = length
	case 'H':
		samples = length / 2
	case 'Q':
		samples = length / 4
	case 'E':
		samples = length / 8
	case 'S':
		samples = length / 16
	case 'T':
		samples = length / 32
	case 'I':
		samples = length / 64
	}

	if dotted {
		samples += samples / 2
	}

	return make([]int16, samples)
}
