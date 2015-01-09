// music.go - beep music note engine
// Batbold Dashzeveg
// GPL v2

package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"math"
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

 Durations:
 DW     - whole note
 DH     - half note
 DQ     - quarter note
 DE     - eighth note
 DS     - sixteenth note
 DT     - thirty-second note

 Octave:
 HL     - switch to left hand keys
 HR     - switch to right hand keys
 HF     - switch to far right keys (last octave)

 Tempo:
 T#     - where # is 0-9, default is 4

 Sustain:
 SA#    - attack time, where # is 0-9, default is 4
 SD#    - decay time, 0-9, default 4
 SS#    - sustain level, 0-9, default 6
 SR#    - release time, 0-9, default 4

 Clef:
 CB     - G and F clef partition (Base). If line ends
          with 'CB', the next line will be played as base.

 Measures:
 |      - bar, ignored
 ' '    - space, ignored
 Tab    - tab, ignored

Demo music: Mozart K33b:`

var demoMusic = `
# Mozart K33b
HRDE|c c DSc s z s |DEc DQz DE[   |CB
HLDE|[ n   z   ,   |  c HRq HLz , |

HRDE|c c DSc s z s |DEc DQz DE[   |CB
HLDE|[ n   z   ,   |  c HRq HLz , |

HRDE|v v DSv c s c |DEv s ] v |CB
HLDE|s l   z   ,   |  ] m p b |

HRDE|c c DSc s z s |DEc z [ c |CB
HLDE|z ,   ]   m   |  [ n o v |

HRDE|s s DSs z ] z |DEs ] p s |CB
HLDE|] m   [   n   |  p b i c |

HRDS|s z ] [ z ] [ p |DE[ DSi y DQr |CB
HLDE|n   z   s   c   |  n    c  DQ[ |
`

var demoHelp = `To play a demo music, run:
 $ beep -p | beep -m
`

type WaveHeader struct {
	header        [44]byte
	ChunkID       [4]byte
	ChunkSize     [4]byte
	Format        [4]byte
	Subchunk1ID   [4]byte
	Subchunk1Size [4]byte
	AudioFormat   [2]byte
	NumChannels   [2]byte
	SampleRate    [4]byte
	ByteRate      [4]byte
	BlockAlign    [2]byte
	BitsPerSample [2]byte
	Subchunk2ID   [4]byte
	Subchunk2Size [4]byte
}

func NewWaveHeader(channels, sampleRate, bitsPerSample int, dataSize int) WaveHeader {
	wh := WaveHeader{}
	copy(wh.ChunkID[0:], stringToBytes("RIFF"))
	copy(wh.ChunkSize[0:], int32ToBytes(36+dataSize))
	copy(wh.Format[0:], stringToBytes("WAVE"))
	copy(wh.Subchunk1ID[0:], stringToBytes("fmt "))
	copy(wh.Subchunk1Size[0:], int32ToBytes(16))
	copy(wh.AudioFormat[0:], int16ToBytes(1))
	copy(wh.NumChannels[0:], int16ToBytes(channels))
	copy(wh.SampleRate[0:], int32ToBytes(sampleRate))
	copy(wh.ByteRate[0:], int32ToBytes(sampleRate*channels*(bitsPerSample/8)))
	copy(wh.BlockAlign[0:], int16ToBytes(1))
	copy(wh.BitsPerSample[0:], int16ToBytes(bitsPerSample))
	copy(wh.Subchunk2ID[0:], stringToBytes("data"))
	copy(wh.Subchunk2Size[0:], int32ToBytes(dataSize))
	return wh
}

func (w *WaveHeader) WriteHeader(wr io.Writer) (int, error) {
	copy(w.header[0:], w.ChunkID[:])
	copy(w.header[4:], w.ChunkSize[:])
	copy(w.header[8:], w.Format[:])
	copy(w.header[12:], w.Subchunk1ID[:])
	copy(w.header[16:], w.Subchunk1Size[:])
	copy(w.header[20:], w.AudioFormat[:])
	copy(w.header[22:], w.NumChannels[:])
	copy(w.header[24:], w.SampleRate[:])
	copy(w.header[28:], w.ByteRate[:])
	copy(w.header[32:], w.BlockAlign[:])
	copy(w.header[34:], w.BitsPerSample[:])
	copy(w.header[36:], w.Subchunk2ID[:])
	copy(w.header[40:], w.Subchunk2Size[:])
	return wr.Write(w.header[:])
}

func playMusicNotes(volume100 int, debug string) {
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
	volume := byte(127.0 * (float64(volume100) / 100.0))
	freqMapLeft := make(map[rune]float64)
	freqMapRight := make(map[rune]float64)
	freqMapFarRight := make(map[rune]float64)
	quarterNote := 1024 * 21
	wholeNote := quarterNote * 4
	printSheet := !*flagQuiet
	printNotes := *flagNotes
	outputFileName := *flagOutput

	var outputFile *os.File
	var test bool
	var err error

	if len(debug) > 0 {
		test = true
	}

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

	// wave buffer map
	keyFreqMap := make(map[rune][]byte)
	for key, freq := range freqMapLeft {
		keyFreqMap[100+key] = generateVoice(freq, quarterNote, volume)
	}
	for key, freq := range freqMapRight {
		keyFreqMap[200+key] = generateVoice(freq, quarterNote, volume)
	}
	for key, freq := range freqMapFarRight {
		keyFreqMap[300+key] = generateVoice(freq, quarterNote, volume)
	}

	// rest buffer map
	bufRW := make([]byte, wholeNote)
	for i, _ := range bufRW {
		bufRW[i] = 127
	}
	bufRH := bufRW[:wholeNote/2]
	bufRQ := bufRW[:wholeNote/4]
	bufRE := bufRW[:wholeNote/8]
	bufRS := bufRW[:wholeNote/16]
	bufRT := bufRW[:wholeNote/32]

	// read lines
	reader := bufio.NewReader(os.Stdin)
	bufPlayLimit := 1024 * 1024 * 100
	ctrlKeys := "RDHTS"
	measures := "WHQEST"
	hands := "RLF"
	tempos := "0123456789"
	ignored := "\t |CB"
	sustainTypes := "ADSR"
	sustainLevels := "0123456789"
	sustainA := 4
	sustainD := 4
	sustainS := 5
	sustainR := 4
	var bufPlay bytes.Buffer
	var bufBase bytes.Buffer
	var bufOutput bytes.Buffer
	var duration = 'Q' // default note duration
	var rest rune
	var ctrl rune
	var sustainType rune
	var hand rune = 'R' // default is middle C octave
	var handLevel rune
	var done bool
	var count int     // line counter
	var tempo int = 4 // normal speed
	var line string   // G clef notes
	var base string   // F clef notes
	var hasBase bool
	for {
		bufPlay.Reset()
		bufBase.Reset()
		if test {
			line = debug
			if count > 0 {
				done = true
			}
		} else {
			line, done = nextMusicLine(reader)
		}
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
		if strings.HasSuffix(line, "CB") {
			// Base clef, read base line
			hasBase = true
			base, done = nextMusicLine(reader)
			if done {
				break
			}
		} else {
			hasBase = false
		}
		controller := func(bufWave *bytes.Buffer, notation string) {
			for _, key := range notation {
				keystr := string(key)
				if strings.ContainsAny(keystr, ignored) {
					continue
				}
				if ctrl == 0 && strings.ContainsAny(keystr, ctrlKeys) {
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
					}
					ctrl = 0
					continue
				}
				if rest > 0 {
					switch rest {
					case 'W':
						bufWave.Write(bufRW)
					case 'H':
						bufWave.Write(bufRH)
					case 'Q':
						bufWave.Write(bufRQ)
					case 'E':
						bufWave.Write(bufRE)
					case 'S':
						bufWave.Write(bufRS)
					case 'T':
						bufWave.Write(bufRT)
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
				if bufFreq, found := keyFreqMap[handLevel+key]; found {
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
					}
					buf := make([]byte, len(bufFreq))
					copy(buf, bufFreq)
					applyNoteVolume(buf, volume)
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
					wrote, err := bufWave.Write(buf)
					if err != nil || wrote != len(buf) {
						fmt.Fprintln(os.Stderr, "Error writing to buffer:", err)
					}
					if bufWave.Len() > bufPlayLimit {
						fmt.Fprintln(os.Stderr, "Line wave buffer exceeds 100MB limit.")
						os.Exit(1)
					}
					if printNotes {
						fmt.Printf("%s", noteLetter(key))
					}
				} else if !strings.ContainsAny(keystr, ignored) {
					fmt.Printf("invalid note: %s (%d)\n", keystr, handLevel-key)
				}
			}
		}
		controller(&bufPlay, line)
		if bufPlay.Len() == 0 {
			continue
		}
		if printNotes {
			fmt.Println()
		}
		gclef := bufPlay.Bytes()
		//mergeNotes(gclef, volume)
		var fclef []byte
		if hasBase {
			controller(&bufBase, base)
			if printNotes {
				fmt.Println()
			}
			fclef = bufBase.Bytes()
			if len(fclef) < len(gclef) {
				for len(fclef) < len(gclef) {
					fclef = append(fclef, bufRW...)
				}
			}
			//mergeNotes(fclef, volume)
		} else {
			fclef = gclef
		}
		if outputFile == nil {
			notes := line
			if hasBase {
				notes += "\n" + base
			}
			playback(gclef, fclef, notes)
		} else {
			// saving to file
			var buf [2]byte
			for i, bar := range gclef {
				buf[0] = bar
				if hasBase {
					buf[1] = fclef[i]
				} else {
					buf[1] = bar
				}
				bufOutput.Write(buf[:])
			}
		}
		count++
	}
	flushSoundBuffer()

	if outputFile != nil {
		// save wave to file
		bufsize := bufOutput.Len()
		header := NewWaveHeader(2, sampleRate, 8, bufsize)
		_, err = header.WriteHeader(outputFile)
		if err != nil {
			fmt.Fprintln(os.Stderr, "Error writing to output file:", err)
			os.Exit(1)
		}
		num, err := outputFile.Write(bufOutput.Bytes())
		if err != nil || num != bufsize {
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
	return buf.String(), false
}

// Changes note amplitude
func applyNoteVolume(buf []byte, volume byte) {
	return
	for i, bar := range buf {
		bar64 := float64(bar)
		volume64 := float64(volume)
		buf[i] = byte(bar64 * (volume64 / 127.0))
	}
}

// Changes note amplitude for ADSR phases
func sustainNote(buf []byte, volume byte, sustainA, sustainD, sustainS, sustainR int) {
	// |  /|\
	// | / | \ _____
	// |/  |  |    | \
	// |--------------
	//   A  D  S    R
	if volume == 0 {
		return
	}
	buflen := len(buf)
	volume64 := float64(volume)
	attack := int(float64(buflen/32) * float64(sustainA))
	decay := (buflen-attack)/10 + ((buflen-attack)/20 * sustainD)
	sustain := byte(volume64 / 10.0 * float64(sustainS+1))
	sustainCount := (buflen - attack - decay) / 2
	release := buflen - attack - decay - sustainCount
	attack64 := float64(attack)
	decay64 := float64(decay)
	sustain64 := float64(sustain)
	release64 := float64(release)
	countD := 0.0
	countR := 0.0
	for i, bar := range buf {
		i64 := float64(i)
		bar64 := float64(bar) - 127.0
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
		if bar64 == 0 {
			bar64 = 1
		}
		buf[i] = byte(127.0 + bar64)
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

// Generates voice waveform for music notes
func generateVoice(freq float64, duration int, volume byte) []byte {
	samples := int(sampleRate64/freq)
	bufString := make([]byte, samples)
	tick := 2.0*math.Pi/float64(samples)
	volume64 := float64(volume)
	if volume > 20 {
		volume64 -= 20 // root to vibrate string
	}
	timer := 0.0
	i := 0
	var bufWave bytes.Buffer
	for i < duration {
		for s := 0; s < samples; s++ {
			bar64 := volume64*math.Sin(timer)
			if bar64 == 0 {
				bar64 = 1
			}
			bufString[s] = 127 + byte(bar64)
			timer += tick
			i++
		}
		bufWave.Write(bufString)
	}

	buf := bufWave.Bytes()
	return trimWave(buf[:duration])
}

// Trims sharp edge from wave for smooth play
func trimWave(buf []byte) []byte {
	cut := len(buf) - 1
	var last byte
	for i, _ := range buf {
		if i == 0 {
			last = buf[cut]
		}
		if buf[cut] < last {
			// falling
			if 127-buf[cut] < 6 {
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
