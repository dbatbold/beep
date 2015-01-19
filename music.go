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
 H0     - octave 0 keys
 HL     - octave 1, 2, 3 (left hand keys)
 HR     - octave 4, 5, 6 (right hand keys)
 H7     - octave 7, 8 keys

 Tempo:
 T#     - where # is 0-9, default is 4

 Sustain:
 SA#    - attack time, where # is 0-9, default is 4
 SD#    - decay time, 0-9, default 4
 SS#    - sustain level, 0-9, default 2
 SR#    - release time, 0-9, default 4

 Voice:
 VD     - Computer generated default voice
 VP     - Piano voice
 VV     - Violin voice
 VN     - If a line ends with 'VN', the next line will be played harmony with the line.

 Chord:
 C#     - Play next # notes as a chord, where # is 2-9. For example C major chord is "C3qet"

 Amplitude:
 A#     - Changes current amplitude, where # is 1-9, default is 9

 Measures:
 |      - bar, ignored
 ' '    - space, ignored
 Tab    - tab, ignored

Demo music: Mozart K33b:`

var demoMusic = `
# Mozart K33b
A9HRDE cc DScszs|DEc DQzDE[|cc DScszs|DEc DQz DE[|vv DSvcsc|DEvs ]v|cc DScszs|VN
A2HLDE [n z,    |cHRq HLz, |[n z,    |cHRq HLz,  |sl z,    |]m   pb|z, ]m    |

A9HRDE cz [c|ss DSsz]z|DEs] ps|DSsz][ z][p|DEpDQ[ [ || DERE] DS][p[ |VN
A2HLDE [n ov|]m [n    |  pb ic|  n,   lHRq|HLnc DQ[ || DEcHRq HLvHRw|

A9HRDS ][p[ ][p[|DE] DQp DEi|REc DScszs|cszs |cszs|DEcDQzDE[|REv DSvcsc|DEvs ]v|VN
A2HLDE bHRe HLvw|cHRq   HLic|[n  ]m    |z,   |]m  |zn   z,  |sl  [,    |z. DQp |

A9HRDE REc DScszs|DEcz [c|REs DSsz]z|DEs] ps|DSsz][ z][p|DE[DSitDQr|VN
A2HLDE z,  ]m    |[n   ov|]m  [n    |pb   ic|nz     sc  |DQn      [|
`

var demoHelp = `To play a demo music, run:
 $ beep -p | beep -m
`

type Voice interface {
	GetNote(key rune) ([]int16, bool)
	SustainNote(buf []int16, volume int16, A, D, S, R int)
}

var (
	quarterNote = 1024 * 22
	wholeNote   = quarterNote * 4
	waiter      = make(chan int)
)

func playMusicNotes(reader *bufio.Reader, volume100 int) {
	volume := int16(sampleAmp16bit * (float64(volume100) / 100.0))
	printSheet := !*flagQuiet
	printNotes := *flagNotes
	outputFileName := *flagOutput
	piano := NewPiano()

	var violin *Violin
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

	// rest buffer map
	bufRW := make([]int16, wholeNote)
	bufRH := bufRW[:wholeNote/2]
	bufRQ := bufRW[:wholeNote/4]
	bufRE := bufRW[:wholeNote/8]
	bufRS := bufRW[:wholeNote/16]
	bufRT := bufRW[:wholeNote/32]
	bufRI := bufRW[:wholeNote/64]

	// read lines
	bufWaveLimit := 1024 * 1024 * 100
	controlKeys := "RDHTSAVC"
	measures := "WHQESTI"
	hands := "0RL78"
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
	bufSustain := make([]int16, quarterNote)
	voiceControls := "DPVN"
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
	var wait bool
	for {
		line, done := nextMusicLine(reader)
		if done {
			break
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
							amplitude = strings.Index(amplitudes, keystr)
						}
					case 'V': // voice
						if strings.ContainsAny(keystr, voiceControls) {
							switch key {
							case 'D': // default voice
								piano.NaturalVoice = false
							case 'P':
								voice = piano
							case 'V':
								if violin == nil {
									voice = NewViolin()
								}
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
					var bufRest []int16
					switch rest {
					case 'W':
						bufRest = bufRW
					case 'H':
						bufRest = bufRH
					case 'Q':
						bufRest = bufRQ
					case 'E':
						bufRest = bufRE
					case 'S':
						bufRest = bufRS
					case 'T':
						bufRest = bufRT
					case 'I':
						bufRest = bufRI
					}
					if bufRest != nil {
						if piano.NaturalVoice {
							clearBuffer(bufSustain)
						}
						bufWave = append(bufWave, bufRest...)
					}
					rest = 0
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
				keyId := handLevel + key
				if bufFreq, found := voice.GetNote(keyId); found {
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
					copy(buf, bufFreq) // get a copy of the note
					applyNoteVolume(buf, volume, amplitude)
					bufsize := len(buf)
					cut := bufsize
					if tempo < 4 && bufsize > 1024 {
						// slow tempo
						releaseNote(buf, 0, 0.7)
						for t := 0; t < 4-tempo; t++ {
							if piano.NaturalVoice {
								buf = append(buf, bufRW[:1024]...)
							} else {
								buf = append(buf, trimWave(buf[:1024])...)
							}
						}
					}
					if tempo > 4 {
						// fast tempo
						releaseNote(buf, 0, 0.7)
						for t := 0; t < tempo-4 && cut > 1024; t++ {
							if 1024 < len(buf[:cut]) {
								cut -= 1024
							}
						}
						buf = trimWave(buf[:cut])
					}
					sustRatio := float64(sustainS)/10.0
					if divide > 1 {
						cut = len(buf) / divide
						bufDiv := trimWave(buf[:cut])
						if piano.NaturalVoice {
							mixSoundWave(bufDiv, bufSustain, volume)
							copyBuffer(bufSustain, buf[cut-1:])
							release := (quarterNote / divide)/10*sustainS
							releaseNote(bufSustain, release, sustRatio)
						}
						buf = bufDiv
					} else {
						if piano.NaturalVoice {
							mixSoundWave(buf, bufSustain, volume)
							copyBuffer(bufSustain, buf[bufsize/3:])
							release := bufsize/10*sustainS
							releaseNote(bufSustain, release, sustRatio)
						}
					}
					raiseNote(buf, 0.05)
					releaseNote(buf, 0, 0.95)

					// sustain note
					bufsize = len(buf)
					if !piano.NaturalVoice {
						for r := 1; r < repeat; r++ {
							buf = append(buf, buf[:bufsize]...)
						}
						voice.SustainNote(buf, volume, sustainA, sustainD, sustainS, sustainR)
					} else {
						if repeat > 1 {
							// sustain note
							bufSust := make([]int16, quarterNote*(repeat-1))
							releaseNote(buf, 0, 0.8)
							buf = append(buf, bufSust...)
						}
					}
					if chordNumber > 0 {
						// playing a chord
						chordCount++
						if chordBuf == nil {
							chordBuf = make([]int16, len(buf))
							copy(chordBuf, buf)
							if piano.NaturalVoice {
								copyBuffer(bufSustain, chordBuf)
							}
						} else {
							mixSoundWave(chordBuf, buf, volume)
							if piano.NaturalVoice {
								mixSoundWave(bufSustain, buf, volume)
							}
						}
						if chordCount == chordNumber {
							if piano.NaturalVoice {
								release := (quarterNote / divide)/10*sustainS
								releaseNote(bufSustain, release, sustRatio)
							}
							buf = chordBuf
							chordNumber = 0
							chordBuf = nil
						} else {
							if printNotes {
								fmt.Printf("%v-", piano.keyNoteMap[keyId])
							}
							continue
						}
					}
					bufWave = append(bufWave, buf...)
					if len(bufWave) > bufWaveLimit {
						fmt.Fprintln(os.Stderr, "Line wave buffer exceeds 100MB limit.")
						os.Exit(1)
					}
					if printNotes {
						fmt.Printf("%v ", piano.keyNoteMap[keyId])
					}
				} else {
					fmt.Printf("invalid note: %s (%d)\n", keystr, key)
				}
			}
			return bufWave
		}
		bufLine := controller(line)
		if mixNextLine {
			if bufMix == nil {
				bufMix = make([]int16, len(bufLine))
				copy(bufMix, bufLine)
				lineMix = line
			} else {
				lineMix += "\n" + line
				mixSoundWave(bufMix, bufLine, volume)
			}
			count++
			clearBuffer(bufSustain)
			continue
		}
		if bufMix != nil {
			mixSoundWave(bufMix, bufLine, volume)
			bufLine = bufMix
			bufMix = nil
			line = lineMix + "\n" + line
		}
		if outputFile == nil {
			if printNotes {
				fmt.Println()
			}
			if len(bufLine) > 0 {
				if wait {
					<-waiter // wait until previous line is done playing
				}
				if printSheet {
					fmt.Println(line)
				}
				// prepare next line while playing
				go playback(bufLine, bufLine)
				wait = true
			} else {
				fmt.Println()
			}
		} else {
			// saving to file
			var bufCh [2]int16
			for _, bar := range bufLine {
				bufCh[0] = bar
				bufCh[1] = bar
				bufOutput = append(bufOutput, bufCh[:]...)
			}
		}
		clearBuffer(bufSustain)
		count++
	}
	if wait {
		<-waiter // wait until last line
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
func mixSoundWave(buf1, buf2 []int16, volume int16) {
	bufsize := len(buf2)
	gap := 32500.0
	for i, _ := range buf1 {
		if i == bufsize {
			break
		}
		bar1 := float64(buf1[i])
		bar2 := float64(buf2[i])
		bar64 := (bar1 - bar2) / 2 * 1.5
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
	for i, _ := range target {
		if i < bufsize {
			target[i] = src[i]
		} else {
			target[i] = 0
		}
	}
}

func clearBuffer(buf []int16) {
	for i, _ := range buf {
		buf[i] = 0
	}
}

// Trims sharp edge from wave for smooth play
func trimWave(buf []int16) []int16 {
	if len(buf) == 0 {
		return buf
	}
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
	tick := sampleAmp16bit / decay
	volume := sampleAmp16bit
	for i, bar := range buf {
		bar64 := float64(bar)
		if i >= release && volume > 0 {
			bar64 = bar64 * (volume / sampleAmp16bit)
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
	tick := sampleAmp16bit / raise
	volume := 0.0
	for i, bar := range buf {
		bar64 := float64(bar)
		bar64 = bar64 * (volume / sampleAmp16bit)
		buf[i] = int16(bar64)
		volume += tick
		if sampleAmp16bit <= volume {
			break
		}
	}
}
