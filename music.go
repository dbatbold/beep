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

 q - middle C (261.6 hertz)

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

 Clef:
 CB     - G and F clef partition (Base)

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
HLDEn   z   s   c      n   c     [
`

var demoHelp = `To play a demo music, run:
 $ beep -p | beep -m
`

func playMusicNotes(volume100 int, debug string) {
	octaveRight := []float64{
		// C, C#, D, D#, E, F, F#, G, G#, A, A#, B
		0, 22, 24, 24, 26, 28, 30, 32, 33, 35, 37, 40, // octave 1
		42, 44, 48, 48, 53, 56, 58, 64, 64, 72, 74, 80, // octave 2
		84, 87, 97, 98, 104, 112, 120, 127, 125, 148, 150, 157, // octave 3
	}
	octaveLeft := []float64{
		// B, A#, A, G#, G, F#, F, E, D#, D, C#, C
		22, 19, 18, 18, 17, 15, 15, 14, 13, 12, 12, 12, // octave 4
		10, 10, 9.5, 9, 8, 7.8, 7.5, 7.0, 6.5, 6.3, 5.9, 5.6, // octave 5
		5, 4.1, 3.8, 3.5, 3.1, 2.8, 2.4, 2.0, 1.7, 1.3, 1.0, 0.5, // octave 6  (too low to tune!)
	}
	keysRight := "q2w3er5t6y7ui9o0p[=]azsxcfvgbnjmk,l."
	keysLeft := ".l,kmjnbgvfcxsza]=[p0o9iu7y6t5re3w2q"
	volume := byte(127.0 * (float64(volume100) / 100.0))
	freqMap := make(map[rune]float64)
	boundary := make([]byte, 256)
	quarterNote := 1024 * 18
	wholeNote := quarterNote * 4
	middleC := 0.0373
	DEBUG := false // print wave form
	printSheet := !*flagQuiet
	printNotes := *flagNotes

	var test bool
	if len(debug) > 0 {
		test = true
	}

	// freq map
	freq := middleC
	for i, key := range keysRight {
		freq += octaveRight[i] / 10000.0
		freqMap[key] = freq
	}
	freq = middleC
	for i, key := range keysLeft {
		freq -= octaveLeft[i] / 10000.0
		freqMap[-key] = freq
	}

	// boundary buffer, used for fading
	for i, _ := range boundary {
		boundary[i] = 1
	}
	boundary[128] = 0 // boundary mark

	// wave buffer map
	bufFreqMap := make(map[rune][]byte)
	for key, freq := range freqMap {
		bufFreqMap[key] = keyFreq(freq, quarterNote, volume)
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
	bufFreqMap[32] = bufRE // space

	// read lines
	reader := bufio.NewReader(os.Stdin)
	bufPlayLimit := 1024 * 1024 * 100
	ctrlKeys := "RDH"
	measures := "WHQEST"
	hands := "RL"
	ignored := "|CB"
	var bufPlay bytes.Buffer
	var bufBase bytes.Buffer
	var duration = 'Q' // default note duration
	var rest rune
	var ctrl rune
	var last rune
	var hand rune = 'R' // default: right hand
	var done bool
	var count int   // line counter
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
		if printSheet {
			fmt.Println(line)
		}
		if strings.HasPrefix(line, "#") {
			// ignore comments
			continue
		}
		if strings.HasSuffix(line, "CB") {
			// Base clef, read base line
			hasBase = true
			base, done = nextMusicLine(reader)
			if done {
				break
			}
			if printSheet {
				fmt.Println(base)
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
				if hand == 'L' && key != 32 {
					key = -key
				}
				if buf, found := bufFreqMap[key]; found {
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
					for r := 0; r < repeat; r++ {
						cut := len(buf) / divide
						buf = trimWave(buf[:cut])
						if last != key && last > 0 {
							bufWave.Write(boundary)
							buf[0] = 0
						} else {
							// no boundary
							buf[0] = 1
						}
						bufWave.Write(buf)
						if bufWave.Len() > bufPlayLimit {
							fmt.Println("Line wave buffer exceeds 100MB limit.")
							return
						}
					}
					last = key
					if printNotes {
						fmt.Printf("%s", noteLetter(key))
					}
				} else if !strings.ContainsAny(keystr, ignored) {
					fmt.Printf("invalid note: %s (%d)\n", keystr, key)
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
		mergeNotes(gclef, volume, boundary)
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
			mergeNotes(fclef, volume, boundary)
		} else {
			fclef = gclef
		}
		if test || DEBUG {
			fmt.Println("LINE")
			for i, bar := range gclef {
				fmt.Printf("%d:%03d|%s\n", i, bar, strings.Repeat("=", int(bar/4)))
			}
		} else {
			playback(gclef, fclef)
		}
		count++
	}
	if !test {
		flushSoundBuffer()
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
func mergeNotes(buf []byte, volume byte, boundary []byte) {
	half := len(boundary) / 2
	buflen := len(buf)
	var c int // count
	var index int
	var found bool
	var first int
	DEBUG := false
	for i, bar := range buf {
		if bar == 0 && i > 0 {
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
				buf[t] = fadeWave(buf[s], c, volume)
			}

			// merge right
			s = index + half + c
			t = index + half - c + 1
			if t < buflen {
				if buflen > s {
					buf[t] = fadeWave(buf[s], c, volume)
				}
			}
			c++
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

func fadeWave(bar byte, i int, volume byte) byte {
	i64 := float64(i)
	gap := float64(volume)
	bar64 := float64(bar)
	if i64 < gap && gap > 0 {
		bar = byte(bar64*((gap-i64)/gap))
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
	return string(note) +"=?"
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
				fmt.Printf("%03d %s\n", i, strings.Repeat("=", int(bar/4)))
			}
		}
		fmt.Println()
	}

	return buf
}
