// music.go - beep music note engine
// Batbold Dashzeveg
// GPL v2

package main

import (
	"bufio"
	"bytes"
	"fmt"
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

 Control keys:

 Rest:
 RW     - whole rest
 RH     - half rest
 RQ     - quarter rest
 RE     - eighth rest
 RS     - sixteenth rest
 RT     - thirty-second rest

 Space  - eighth rest, depends on current duration

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

 Clef:
 CB     - G and F clef partition (Base). If line ends
          with 'CB', the next line will be played as base.

 Measures:
 |      - bar (ignored)

Demo music: Mozart K33b:`

var demoMusic = `
# Mozart K33b
HRDEc c DSc s z s |DEc DQz DE[ CB
HLDE[ n   z   ,      c HRq HLz ,

HRDEc c DSc s z s |DEc DQz DE[ CB
HLDE[ n   z   ,      c HRq HLz ,

HRDEv v DSv c s c |DEv s ] v CB
HLDEs l   z   ,      ] m p b

HRDEc c DSc s z s |DEc z [ c CB
HLDEz ,   ]   m      [ n o v 

HRDEs s DSs z ] z |DEs ] p s CB
HLDE] m   [   n      p b i c 

HRDSs z ] [ z ] [ p |DE[ DSi y DQr CB
HLDEn   z   s   c      n   c   DQ[ 
`

var demoHelp = `To play a demo music, run:
 $ beep -p | beep -m
`

var waveHeader = []byte{
	0, 0, 0, 0, // 0 ChunkID
	0, 0, 0, 0, // 4 ChunkSize
	0, 0, 0, 0, // 8 Format
	0, 0, 0, 0, // 12 Subchunk1ID
	0, 0, 0, 0, // 16 Subchunk1Size
	0, 0, // 20 AudioFormat
	0, 0, // 22 NumChannels
	0, 0, 0, 0, // 24 SampleRate
	0, 0, 0, 0, // 28 ByteRate
	0, 0, // 32 BlockAlign
	0, 0, // 34 BitsPerSample
	0, 0, 0, 0, // 36 Subchunk2ID
	0, 0, 0, 0, // 40 Subchunk2Size
}

func playMusicNotes(volume100 int, debug string) {
	octaveLeft := []float64{
		// C, C#, D, D#, E, F, F#, G, G#, A, A#, B
		32.70, 34.64, 36.70, 38.89, 41.20, 43.65, 46.24, 48.99, 51.91, 55.00, 58.27, 61.73, // octave 0
		65.40, 69.29, 73.41, 77.78, 82.40, 87.30, 92.49, 97.99, 103.82, 110.00, 116.54, 123.47, // octave 1
		130.82, 138.59, 146.83, 155.56, 164.81, 174.61, 185.00, 196.00, 207.65, 220.00, 233.08, 246.94, // octave 2
	}
	octaveRight := []float64{
		// C, C#, D, D#, E, F, F#, G, G#, A, A#, B
		261.63, 277.18, 293.66, 311.13, 329.63, 349.23, 369.99, 392.00, 415.30, 440.00, 466.16, 493.88, // octave 3 (middle C)
		523.25, 554.37, 587.33, 622.25, 659.26, 698.46, 739.99, 783.99, 830.61, 880.00, 932.33, 987.77, // octave 4
		1046.5, 1108.73, 1174.66, 1244.51, 1318.51, 1396.91, 1479.98, 1567.98, 1661.22, 1760.00, 1864.66, 1975.53, // octave 5
	}
	octaveFarRight := []float64{
		// C, C#, D, D#, E, F, F#, G, G#, A, A#, B
		2093.0, 2217.5, 2349.3, 2489.0, 2637.0, 2793.0, 2960.0, 3136.0, 3322.4, 3520.0, 3729.3, 4186.0, // octave 6 (last octave)
	}
	keys := "q2w3er5t6y7ui9o0p[=]azsxcfvgbnjmk,l."
	volume := byte(127.0 * (float64(volume100) / 100.0))
	freqMapLeft := make(map[rune]float64)
	freqMapRight := make(map[rune]float64)
	freqMapFarRight := make(map[rune]float64)
	bufMerge := make([]byte, 512)
	quarterNote := 1024 * 18
	wholeNote := quarterNote * 4
	//middleC := hertzToFreq(261.63)
	DEBUG := false // print wave form
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
		freqMapLeft[key] = hertzToFreq(octaveLeft[i])
		freqMapRight[key] = hertzToFreq(octaveRight[i])
	}
	for i, key := range keys[:12] {
		freqMapFarRight[key] = hertzToFreq(octaveFarRight[i])
	}

	// boundary buffer, used for fading
	for i, _ := range bufMerge {
		bufMerge[i] = 1
	}
	bufMerge[255] = 0 // merge mark
	bufMerge[256] = 0 // merge mark

	// wave buffer map
	keyFreqMap := make(map[rune][]byte)
	for key, freq := range freqMapLeft {
		keyFreqMap[100+key] = keyFreq(freq, quarterNote, volume)
	}
	for key, freq := range freqMapRight {
		keyFreqMap[200+key] = keyFreq(freq, quarterNote, volume)
	}
	for key, freq := range freqMapFarRight {
		keyFreqMap[300+key] = keyFreq(freq, quarterNote, volume)
	}

	// rest buffer map
	bufRW := make([]byte, wholeNote)
	for i, _ := range bufRW {
		bufRW[i] = 1
	}
	bufRW[0] = 0 // boundary mark
	bufRH := bufRW[:wholeNote/2]
	bufRQ := bufRW[:wholeNote/4]
	bufRE := bufRW[:wholeNote/8]
	bufRS := bufRW[:wholeNote/16]
	bufRT := bufRW[:wholeNote/32]

	// space bar, half of current rest
	keyFreqMap[132] = bufRQ // space
	keyFreqMap[232] = bufRQ // space
	keyFreqMap[332] = bufRQ // space

	// read lines
	reader := bufio.NewReader(os.Stdin)
	bufPlayLimit := 1024 * 1024 * 100
	ctrlKeys := "RDHT"
	measures := "WHQEST"
	hands := "RLF"
	tempos := "0123456789"
	ignored := "|CB"
	var bufPlay bytes.Buffer
	var bufBase bytes.Buffer
	var bufOutput bytes.Buffer
	var duration = 'Q' // default note duration
	var rest rune
	var ctrl rune
	var last rune
	var hand rune = 'R' // default is middle C octave
	var handLevel rune
	var done bool
	var count int   // line counter
	var tempo int = 4  // normal speed
	var line string // G clef notes
	var base string // F clef notes
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
				if key == 32 && last == 32 {
					// spaces after first space are ignored
					continue
				}
				if ctrl == 0 && strings.ContainsAny(keystr, ctrlKeys) {
					ctrl = key
					continue
				}
				if ctrl > 0 {
					switch ctrl {
					case 'D':
						if strings.ContainsAny(keystr, measures) {
							duration = key
						}
					case 'R':
						if strings.ContainsAny(keystr, measures) {
							rest = key
						}
					case 'H':
						if strings.ContainsAny(keystr, hands) {
							hand = key
						}
					case 'T':
						if strings.ContainsAny(keystr, tempos) {
							tempo = strings.Index(tempos, keystr)
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
				if buf, found := keyFreqMap[handLevel+key]; found {
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
					if key == 32 {
						// space rest is half of current duration
						divide *= 2
					}
					for r := 0; r < repeat; r++ {
						bufsize := len(buf)
						cut := bufsize
						if tempo < 4 {
							// slow tempo
							for t := 0; t < 4-tempo; t++ {
								if bufsize > 1024 {
									buf = append(trimWave(buf), trimWave(buf[:1024])...)
								}
							}
						}
						if tempo > 4 {
							// fast tempo
							for t := 0; t < tempo-4; t++ {
								if 1024 < len(buf[:cut]) {
									cut -= 1024
								}
							}
							buf = buf[:cut]
						}
						cut = len(buf) / divide
						buf = trimWave(buf[:cut])
						if last != key && last > 0 {
							bufWave.Write(bufMerge)
						}
						buf[0] = 0 // note boundary
						bufWave.Write(buf)
						if bufWave.Len() > bufPlayLimit {
							fmt.Fprintln(os.Stderr, "Line wave buffer exceeds 100MB limit.")
							os.Exit(1)
						}
					}
					last = key
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
		mergeNotes(gclef, volume, bufMerge)
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
			mergeNotes(fclef, volume, bufMerge)
		} else {
			fclef = gclef
		}
		if test || DEBUG {
			fmt.Println("LINE")
			for i, bar := range gclef {
				fmt.Printf("%d:%03d|%s\n", i, bar, strings.Repeat("=", int(bar/4)))
			}
		} else {
			if outputFile == nil {
				notes := line
				if hasBase {
					notes += "\n"+ base
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
		}
		count++
	}
	flushSoundBuffer()

	if outputFile != nil {
		// save wave to file
		bufsize := bufOutput.Len()
		copy(waveHeader[0:], stringToBytes("RIFF"))    // ChunkID
		copy(waveHeader[4:], int32ToBytes(36+bufsize)) // ChunkSize
		copy(waveHeader[8:], stringToBytes("WAVE"))    // Format
		copy(waveHeader[12:], stringToBytes("fmt "))   // Subchunk1ID
		copy(waveHeader[16:], int32ToBytes(16))        // Subchunk1Size
		copy(waveHeader[20:], int16ToBytes(1))         // AudioFormat
		copy(waveHeader[22:], int16ToBytes(2))         // NumChannels
		copy(waveHeader[24:], int32ToBytes(44100))     // SampleRate
		copy(waveHeader[28:], int32ToBytes(44100))     // ByteRate
		copy(waveHeader[32:], int16ToBytes(1))         // BlockAlign
		copy(waveHeader[34:], int16ToBytes(8))         // BitsPerSample
		copy(waveHeader[36:], stringToBytes("data"))   // Subchunk2ID
		copy(waveHeader[40:], int32ToBytes(bufsize))   // Subchunk2Size
		_, err = outputFile.Write(waveHeader)
		if err != nil {
			fmt.Fprintln(os.Stderr, "Error writing to output file:", err)
			os.Exit(1)
		}
		outputFile.Write(bufOutput.Bytes())
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

// Merges two wave form by fading for playing smooth
func mergeNotes(buf []byte, volume byte, bufMerge []byte) {
	half := len(bufMerge)/2
	buflen := len(buf)
	var c int // count
	var index int
	var found bool
	var first int
	DEBUG := false
	for i, bar := range buf {
		if i > 0 && bar == 0 && buf[i-1] == 0 {
			// found merge buffer found
			found = true
			index = i
			c = 0
		}
		if found {
			if first == 0 {
				first = i
			}

			// merge left
			s := index - half - c
			t := index - half + c
			if s > 0 {
				buf[t] = fadeWaveOut(buf[s], c, volume)
			}

			// merge right
			s = index + half + c
			t = index + half - c + 1
			if t < buflen && buflen > s {
				buf[t] = fadeWaveOut(buf[s], c, volume)
			}
			if i % 2 == 0 {
				c++
			}
			if c >= half {
				found = false
			}
		}
	}

	if DEBUG {
		for i, bar := range buf {
			if i > first-20 && i < first+half*2+20 {
				if i == first || i == first+half*2 {
					fmt.Println("BOUNDRY")
				}
				fmt.Printf("|%s\n", i, strings.Repeat("=", int(bar/4)))
			}
		}
	}
}

func fadeWaveIn(bar byte, i int, volume byte) byte {
	i64 := float64(i)
	gap := float64(255)
	bar64 := float64(bar)
	if i64 < gap && gap > 0 {
		bar = byte(bar64 * (i64 / gap))
		if bar > 0 {
			return bar
		}
	}
	return 1
}

func fadeWaveOut(bar byte, i int, volume byte) byte {
	i64 := float64(i)
	gap := float64(255)
	bar64 := float64(bar)
	if i64 < gap && gap > 0 {
		bar = byte(bar64 * ((gap - i64)/gap))
		if bar > 0 {
			return bar
		}
	}
	return 1
}

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

// Generates sine wave for music notes
func keyFreq(freq float64, duration int, volume byte) []byte {
	buf := make([]byte, duration)
	vol64 := float64(volume)
	for i, _ := range buf {
		bar := volume - byte(vol64*math.Cos(float64(i)*freq))
		if bar == 0 {
			bar = 1
		}
		buf[i] = bar
	}
	return trimWave(buf)
}

// Trims sharp edge from wave for smooth play
func trimWave(buf []byte) []byte {
	cut := len(buf) - 1
	var last byte
	DEBUG := false
	for i, _ := range buf {
		if i == 0 {
			last = buf[cut]
		}
		if buf[cut] < last {
			// falling
			if buf[cut] < 6 {
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

	if DEBUG {
		for i, _ := range buf {
			if i < 50 {
				bar := buf[len(buf)-1-i]
				fmt.Printf("%d:%03d|%s\n", i, bar, strings.Repeat("=", int(bar/4)))
			}
		}
		fmt.Println()
	}

	return buf
}

func stringToBytes(s string) []byte {
	var buf []byte
	for _, c := range s {
		buf = append(buf, byte(c))
	}
	return buf
}

func int32ToBytes(i int) []byte {
	var buf [4]byte
	n := int32(i)
	buf[0] = byte(n)
	buf[1] = byte(n >> 8)
	buf[2] = byte(n >> 16)
	buf[3] = byte(n >> 24)
	return buf[:]
}

func int16ToBytes(i int) []byte {
	n := int32(i)
	var buf [2]byte
	buf[0] = byte(n)
	buf[1] = byte(n >> 8)
	return buf[:]
}

func hertzToFreq(hertz float64) float64 {
	// 1 second = 44100 samples
	// 1 hertz = freq * 2Pi
	// freq = 2Pi / 44100 * hertz
	freq := 2.0 * math.Pi / 44100.0 * hertz
	return freq
}
