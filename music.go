// music.go - beep music note engine
// Batbold Dashzeveg
// GPL v2

package main

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"strings"
)

var beepNotation = `
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

 Octave:
 HL     - switch to left hand keys
 HR     - switch to right hand keys
 HF     - switch to far right keys (last octave)

 Tempo:
 T#     - where # is 0-9, default is 4

 Sustain:
 SA#    - attack time, where # is 0-9, default is 4
 SD#    - decay time, 0-9, default 4
 SS#    - sustain level, 0-9, default 4
 SR#    - release time, 0-9, default 4

 Voice:
 VP     - Piano voice
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

Demo music: Mozart K33b:`

var demoMusic = `
# Mozart K33b
HRDE cc DScszs|DEc DQzDE[|cc DScszs|DEc DQz DE[|vv DSvcsc|DEvs ]v|cc DScszs|VN
HLDE [n z,    |cHRq HLz, |[n z,    |cHRq HLz,  |sl z,    |]m   pb|z, ]m    |

HRDE cz [c|ss DSsz]z|DEs] ps|DSsz][ z][p|DEpDQ[ [ || REDE] DS][p[ |VN
HLDE [n ov|]m [n    |  pb ic|  n,   lHRq|HLnc DQ[ || DEcHRq HLvHRw|

HRDS ][p[ ][p[|DE] DQp DEi|REc DScszs|cszs |cszs|DEcDQzDE[|REv DSvcsc|DEvs ]v|VN
HLDE bHRe HLvw|cHRq   HLic|[n  ]m    |z,   |]m  |zn   z,  |sl  [,    |z. DQp |

HRDE REc DScszs|DEcz [c|REs DSsz]z|DEs] ps|DSsz][ z][p|DE[DSitDQr|VN
HLDE z,  ]m    |[n   ov|]m  [n    |pb   ic|nz     sc  |DQn      [|
`

var demoHelp = `To play a demo music, run:
 $ beep -p | beep -m
`

type Voice interface {
	GenerateNote(freq float64, duration int) []int16
	GetNote(key rune) ([]int16, bool)
}

func playMusicNotes(volume100 int) {
	octaveLeft := []float64{
		// C, C#, D, D#, E, F, F#, G, G#, A, A#, B
		32.70, 34.64, 36.70, 38.89, 41.20, 43.65, 46.24, 48.99, 51.91, 55.00, 58.27, 61.73,
		65.40, 69.29, 73.41, 77.78, 82.40, 87.30, 92.49, 97.99, 103.8, 110.0, 116.5, 123.4,
		130.8, 138.5, 146.8, 155.5, 164.8, 174.6, 185.0, 196.0, 207.6, 220.0, 233.0, 246.9,
	}
	octaveRight := []float64{
		// C, C#, D, D#, E, F, F#, G, G#, A, A#, B
		261.6, 277.1, 293.6, 311.1, 329.6, 349.2, 369.9, 392.0, 415.3, 440.0, 466.1, 493.8,
		523.2, 554.3, 587.3, 622.2, 659.2, 698.4, 739.9, 783.9, 830.6, 880.0, 932.3, 987.7,
		1046.5, 1108.7, 1174.6, 1244.5, 1318.5, 1396.9, 1479.9, 1567, 1661, 1760, 1864, 1975,
	}
	octaveFarRight := []float64{
		// C, C#, D, D#, E, F, F#, G, G#, A, A#, B
		2093, 2217.5, 2349.3, 2489, 2637, 2793, 2960, 3136, 3322.4, 3520, 3729.3, 4186,
	}
	keys := "q2w3er5t6y7ui9o0p[=]azsxcfvgbnjmk,l."
	volume := int16(sampleAmp16bit * (float64(volume100) / 100.0))
	freqMapLeft := make(map[rune]float64)
	freqMapRight := make(map[rune]float64)
	freqMapFarRight := make(map[rune]float64)
	quarterNote := 1024 * 22
	wholeNote := quarterNote * 4
	printSheet := !*flagQuiet
	printNotes := *flagNotes
	outputFileName := *flagOutput
	piano := &Piano{}
	violin := &Violin{}

	var outputFile *os.File
	var err error

	// output file
	if len(outputFileName) > 0 {
		if outputFileName == "-" {
			outputFile = os.Stdout
			printSheet = false
			printNotes = false
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

	// frequency map
	for i, key := range keys {
		freqMapLeft[key] = octaveLeft[i]
		freqMapRight[key] = octaveRight[i]
	}
	for i, key := range keys[:12] {
		freqMapFarRight[key] = octaveFarRight[i]
	}

	// generate notes
	piano.FreqMap = make(map[rune][]int16)
	violin.FreqMap = make(map[rune][]int16)
	for key, freq := range freqMapLeft {
		keyLevel := 100 + key
		piano.FreqMap[keyLevel] = piano.GenerateNote(freq, quarterNote)
		violin.FreqMap[keyLevel] = violin.GenerateNote(freq, quarterNote)
	}
	for key, freq := range freqMapRight {
		keyLevel := 200 + key
		piano.FreqMap[keyLevel] = piano.GenerateNote(freq, quarterNote)
		violin.FreqMap[keyLevel] = violin.GenerateNote(freq, quarterNote)
	}
	for key, freq := range freqMapFarRight {
		keyLevel := 300 + key
		piano.FreqMap[keyLevel] = piano.GenerateNote(freq, quarterNote)
		violin.FreqMap[keyLevel] = violin.GenerateNote(freq, quarterNote)
	}

	// rest buffer map
	bufRW := make([]int16, wholeNote)
	bufRH := bufRW[:wholeNote/2]
	bufRQ := bufRW[:wholeNote/4]
	bufRE := bufRW[:wholeNote/8]
	bufRS := bufRW[:wholeNote/16]
	bufRT := bufRW[:wholeNote/32]
	bufRI := bufRW[:wholeNote/64]

	// read lines
	reader := bufio.NewReader(os.Stdin)
	bufWaveLimit := 1024 * 1024 * 100
	controlKeys := "RDHTSAVC"
	measures := "WHQESTI"
	hands := "RLF"
	tempos := "0123456789"
	amplitudes := "0123456789"
	chordNumbers := "0123456789"
	ignored := "\t |"
	sustainTypes := "ADSR"
	sustainLevels := "0123456789"
	sustainA := 4
	sustainD := 4
	sustainS := 4
	sustainR := 4
	voiceControls := "PVN"
	var bufOutput []int16
	var duration = 'Q' // default note duration
	var rest rune
	var ctrl rune
	var voice Voice = piano // default voice is piano
	var sustainType rune
	var hand rune = 'R' // default is middle C octave
	var handLevel rune
	var count int         // line counter
	var tempo int = 4     // normal speed
	var amplitude int = 9 // max volume
	var mixNextLine bool
	var chordBuf []int16
	var chordNumber int
	var chordCount int
	var bufMix []int16
	var lineMix string
	for {
		line, done := nextMusicLine(reader)
		if done {
			break
		}
		if len(line) == 0 {
			if printSheet {
				fmt.Println()
			}
			continue
		}
		if strings.HasPrefix(line, "#") {
			// ignore comments
			if printSheet {
				fmt.Println(line)
			}
			continue
		}
		if strings.HasSuffix(line, "VN") {
			// include next line to mixer
			mixNextLine = true
		} else {
			mixNextLine = false
		}
		controller := func(notation string) []int16 {
			var bufWave []int16
			for _, key := range notation {
				keystr := string(key)
				if strings.ContainsAny(keystr, ignored) {
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
							case 'A': // attack
								sustainA = level
							case 'D': // decay
								sustainD = level
							case 'S': // sustain
								sustainS = level
							case 'R': // release
								sustainR = level
							}
						}
					case 'A': // amplitude
						if strings.ContainsAny(keystr, amplitudes) {
							amplitude = strings.Index(keystr, amplitudes)
						}
					case 'V': // voice
						if strings.ContainsAny(keystr, voiceControls) {
							switch key {
							case 'P':
								voice = piano
							case 'V':
								voice = violin
							}
						}
					case 'C': // chord
						if strings.ContainsAny(keystr, chordNumbers) {
							chordCount = 0
							chordNumber = strings.Index(chordNumbers, keystr)
						}
					}
					ctrl = 0
					continue
				}
				if rest > 0 {
					switch rest {
					case 'W':
						bufWave = append(bufWave, bufRW...)
					case 'H':
						bufWave = append(bufWave, bufRH...)
					case 'Q':
						bufWave = append(bufWave, bufRQ...)
					case 'E':
						bufWave = append(bufWave, bufRE...)
					case 'S':
						bufWave = append(bufWave, bufRS...)
					case 'T':
						bufWave = append(bufWave, bufRT...)
					case 'I':
						bufWave = append(bufWave, bufRI...)
					}
					rest = 0
				}
				switch hand {
				case 'L':
					handLevel = 100
				case 'R':
					handLevel = 200
				case 'F':
					handLevel = 300
				}
				if bufFreq, found := voice.GetNote(handLevel + key); found {
					repeat := 1
					divide := 1
					switch duration {
					case 'W':
						repeat = 4
					case 'H':
						repeat = 2
					case 'E':
						divide = 2
					case 'S':
						divide = 4
					case 'T':
						divide = 8
					case 'I':
						divide = 16
					}
					buf := make([]int16, len(bufFreq))
					copy(buf, bufFreq)
					applyNoteVolume(buf, volume, amplitude)
					bufsize := len(buf)
					cut := bufsize
					if tempo < 4 && bufsize > 1024 {
						// slow tempo
						for t := 0; t < 4-tempo; t++ {
							buf = append(buf, trimWave(buf[:1024])...)
						}
					}
					if tempo > 4 {
						// fast tempo
						for t := 0; t < tempo-4 && cut > 1024; t++ {
							if 1024 < len(buf[:cut]) {
								cut -= 1024
							}
						}
						buf = buf[:cut]
					}
					if divide > 1 {
						cut = len(buf) / divide
						buf = trimWave(buf[:cut])
					}
					bufsize = len(buf)
					for r := 1; r < repeat; r++ {
						buf = append(buf, buf[:bufsize]...)
					}
					sustainNote(buf, volume, sustainA, sustainD, sustainS, sustainR)
					if chordNumber > 0 {
						// playing a chord
						chordCount++
						if chordBuf == nil {
							chordBuf = make([]int16, len(buf))
							copy(chordBuf, buf)
						} else {
							chordBuf = mixWaves(chordBuf, buf, volume)
						}
						if chordCount == chordNumber {
							bufWave = append(bufWave, chordBuf...)
							chordNumber = 0
							chordBuf = nil
						} else {
							if printNotes {
								fmt.Printf("%s-", noteLetter(key))
							}
						}
						continue
					}
					bufWave = append(bufWave, buf...)
					if len(bufWave) > bufWaveLimit {
						fmt.Fprintln(os.Stderr, "Line wave buffer exceeds 100MB limit.")
						os.Exit(1)
					}
					if printNotes {
						fmt.Printf("%s", noteLetter(key))
					}
				} else {
					fmt.Printf("invalid note: %s (%d)\n", keystr, handLevel-key)
				}
			}
			return bufWave
		}
		bufLine := controller(line)
		if len(bufLine) == 0 {
			continue
		}
		if printNotes {
			fmt.Println()
		}
		if mixNextLine {
			if bufMix == nil {
				bufMix = make([]int16, len(bufLine))
				copy(bufMix, bufLine)
				lineMix = line
			} else {
				lineMix += "\n" + line
				bufMix = mixWaves(bufMix, bufLine, volume)
			}
			count++
			continue
		}
		if bufMix != nil {
			bufMix = mixWaves(bufMix, bufLine, volume)
			bufLine = bufMix
			bufMix = nil
			line = lineMix + "\n" + line
		}
		if outputFile == nil {
			playback(bufLine, bufLine, line)
		} else {
			// saving to file
			var bufCh [2]int16
			for _, bar := range bufLine {
				bufCh[0] = bar
				bufCh[1] = bar
				bufOutput = append(bufOutput, bufCh[:]...)
			}
		}
		count++
	}
	flushSoundBuffer()

	if outputFile != nil {
		// save wave to file
		bufLen := len(bufOutput)
		header := NewWaveHeader(2, sampleRate, 16, bufLen*2)
		_, err = header.WriteHeader(outputFile)
		if err != nil {
			fmt.Fprintln(os.Stderr, "Error writing to output file:", err)
			os.Exit(1)
		}
		_, err := outputFile.Write(int16ToByteBuf(bufOutput))
		if err != nil {
			fmt.Fprintln(os.Stderr, "Error writing to output file:", err)
			os.Exit(1)
		}
	}
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
func applyNoteVolume(buf []int16, volume int16, amplitude int) {
	volume64 := float64(volume)
	amplitude64 := float64(amplitude)
	for i, bar := range buf {
		bar64 := float64(bar)
		bar64 *= (volume64 / sampleAmp16bit)
		if amplitude64 > 0 {
			bar64 *= (amplitude64 / 9.0)
		}
		buf[i] = int16(bar64)
	}
}

// Mix two waves
func mixWaves(buf1 []int16, buf2 []int16, volume int16) []int16 {
	buf1len := len(buf1)
	buf2len := len(buf2)
	lendiff := buf2len - buf1len
	if lendiff > 0 {
		bufdiff := make([]int16, lendiff)
		buf1 = append(buf1, bufdiff...)
	}
	if lendiff < 0 {
		bufdiff := make([]int16, -lendiff)
		buf2 = append(buf2, bufdiff...)
	}
	gap := 32500.0
	for i, _ := range buf1 {
		bar1 := float64(buf1[i])
		bar2 := float64(buf2[i])
		bar64 := (bar1 - bar2) / 2 * 1.4
		if bar64 > gap {
			bar64 = gap
		} else if bar64 <= -gap {
			bar64 = -gap
		}
		buf1[i] = int16(bar64)
	}
	return buf1
}

// Changes note amplitude for ADSR phases
func sustainNote(buf []int16, volume int16, sustainA, sustainD, sustainS, sustainR int) {
	// |  /|\
	// | / | \ _____
	// |/  |  |    | \
	// |--------------
	//   A  D  S    R
	if volume == 0 {
		return
	}
	bufLen := len(buf)
	volume64 := float64(volume)
	attack := int(float64(bufLen/200) * float64(sustainA))
	decay := (bufLen-attack)/10 + ((bufLen - attack) / 20 * sustainD)
	sustain := int16(volume64 / 10.0 * float64(sustainS+1))
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

// Translates beep notation to CDEFGAB notation
func noteLetter(note rune) string {
	if note < 0 {
		note = -note
	}
	switch note {
	case 'q', 'i', 'c':
		return "C "
	case '2', '9', 'f':
		return "C# "
	case 'w', 'o', 'v':
		return "D "
	case '3', '0', 'g':
		return "D# "
	case 'e', 'p', 'b':
		return "E "
	case 'r', '[', 'n':
		return "F "
	case '5', '=', 'j':
		return "F# "
	case 't', ']', 'm':
		return "G "
	case '6', 'a', 'k':
		return "G# "
	case 'y', 'z', ',':
		return "A "
	case '7', 's', 'l':
		return "A# "
	case 'u', 'x', '.':
		return "B "
	case '8', 'd', ';':
		return "B#"
	case ' ':
		return ""
	}
	return string(note) + "=?"
}

// Trims sharp edge from wave for smooth play
func trimWave(buf []int16) []int16 {
	cut := len(buf) - 1
	var last int16
	for i, _ := range buf {
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
	buf = buf[:cut]
	return buf
}
