/*
 beep - A simple sound notifier with a music note engine
 Batbold Dashzeveg
 2014-12-31 GPL v2
*/

package main

import (
	"bufio"
	"bytes"
	"flag"
	"fmt"
	"math"
	"os"
	"strings"
)

var (
	flagHelp      = flag.Bool("h", false, "help")
	flagCount     = flag.Int("c", 1, "count")
	flagFreq      = flag.Float64("f", 0.07459, "frequency")
	flagVolume    = flag.Int("v", 100, "volume (1-100)")
	flagDuration  = flag.Int("t", 1, "time duration (1-100)")
	flagDevice    = flag.String("d", "default", "audio device (hw:0,0)")
	flagLine      = flag.Bool("l", false, "beep per line from stdin")
	flagMusic     = flag.Bool("m", false, "play music notes from stdin (see beep notation)")
	flagPrintDemo = flag.Bool("p", false, "print a demo music by Mozart")

	middleC     = 0.0373
	boundary    = make([]byte, 256)
	quarterNote = 1024*18
	wholeNote   = quarterNote * 4
)

var beepNotation = `
Beep notation:
  | | | | | | | | | | | | | | | | | | | | | | 
  |2|3| |5|6|7| |9|0| |=|a|s| |f|g| |j|k|l| |
 | | | | | | | | | | | | | | | | | | | | | | 
 |q|w|e|r|t|y|u|i|o|p|[|]|z|x|c|v|b|n|m|,|.|

 q - middle C (261.6 hertz)

 Left and right hand keys are same. Uppercase 
 letters are control keys. Lower case letters
 are music notes. Space bar is currect duration
 rest. Spaces after first space are ignored.

 Control keys:

 Rest:
 RW     - whole rest
 RH     - half rest
 RQ     - quarter rest
 RE     - eighth rest
 RS     - sixteenth rest
 RT     - thirty-second rest
 Space  - duration rest

 Durations:
 DW     - whole note
 DH     - half note
 DQ     - quarter note
 DE     - eighth note
 DS     - sixteenth note
 DT     - thirty-second note

 Octave: (not implemented yet)
 HL     - swich to left hand keys
 HR     - swich to right hand keys

 Clef: (not implemented yet)
 CB     - G and F clef partition (Base)

 Measures:
 |      - bar (ignored)

Demo music Mozart K33b:`

var demoMusic = `
# Mozart K33b
 DEc c DSc s z s |DEc DQz DE[
 DEc c DSc s z s |DEc DQz DE[
 DEv v DSv c s c |DEv s ] v
 DEc c DSc s z s |DEc z [ c
 DEs s DSs z ] z |DEs ] p s
 DSs z ] [ z ] [ p |DE[ DSi y DQr
`

var demoHelp = `To play a demo music, run:
 $ beep -m | beep -p
`

func main() {
	flag.Parse()

	help := *flagHelp
	freq := *flagFreq
	count := *flagCount
	volume := *flagVolume
	duration := *flagDuration
	device := *flagDevice
	lineBeep := *flagLine
	playMusic := *flagMusic
	printDemo := *flagPrintDemo

	if help {
		fmt.Println("beep [options]")
		flag.PrintDefaults()
		fmt.Println("\nPiano key map:")
		fmt.Print(beepNotation)
		fmt.Println(demoMusic)
		fmt.Println(demoHelp)
		return
	}
	if printDemo {
		fmt.Print(demoMusic)
		return
	}
	if volume < 0 || volume > 100 {
		volume = 100
	}
	if duration < 1 || duration > 100 {
		duration = 1
	}

	openSoundDevice(device)
	initSoundDevice()
	defer closeSoundDevice()

	if lineBeep {
		beepPerLine(volume, freq, duration)
		return
	}

	if playMusic {
		playMusicNotes(volume)
		return
	}

	// beep
	buf := make([]byte, 1024*15*duration)
	bar := byte(127.0 * (float64(volume) / 100.0))
	gap := 1024 * 10 * duration
	var last byte
	for i, _ := range buf {
		if i < gap {
			buf[i] = bar + byte(float64(bar)*math.Sin(float64(i)*freq))
			last = buf[i]
		} else {
			if last != bar {
				if last > bar {
					last--
				} else {
					last++
				}
			}
			buf[i] = last
		}
	}
	for i := 0; i < count; i++ {
		playback(buf)
	}
	flushSoundBuffer()
}

func beepPerLine(volume int, freq float64, duration int) {
	buf := make([]byte, 1024*7*duration)
	bar := byte(127.0 * (float64(volume) / 100.0))
	gap := 1024 * 4 * duration
	var last byte
	for i, _ := range buf {
		if i < gap {
			buf[i] = bar + byte(float64(bar)*math.Sin(float64(i)*freq))
			last = buf[i]
		} else {
			buf[i] = last
			if last > 127 {
				last--
			}
		}
	}
	reader := bufio.NewReader(os.Stdin)
	for {
		line, isPrefix, err := reader.ReadLine()
		if err != nil {
			break
		}
		fmt.Print(string(line))
		if !isPrefix {
			fmt.Println()
			playback(buf)
		}
	}
	flushSoundBuffer()
}

func playMusicNotes(volume int) {
	octaveLeft := []float64{
		// C, C#, D, D#, E, F, F#, G, G#, A, A#, B
		0, 22, 24, 24, 26, 28, 30, 32, 33, 35, 37, 40, // octave 1
		42, 44, 48, 48, 53, 56, 58, 64, 64, 72, 74, 80, // octave 2
		84, 87, 97, 98, 104, 112, 120, 127, 125, 148, 150, 157, // octave 3
	}
	keys := "q2w3er5t6y7ui9o0p[=]azsxcfvgbnjmk,l."
	freqMap := make(map[rune]float64)
	freq := middleC
	for i, key := range keys {
		freq += octaveLeft[i] / 10000.0
		freqMap[key] = freq
	}

	// wave buffer map
	bufMap := make(map[rune][]byte)
	for key, freq := range freqMap {
		bufMap[key] = keyFreq(freq, volume)
	}

	// rest buffer map
	bufRW := make([]byte, wholeNote)
	for i, _ := range bufRW {
		bufRW[i] = 127
	}
	bufRW[0] = 0 // boundary byte
	bufRH := bufRW[:wholeNote/2]
	bufRQ := bufRW[:wholeNote/4]
	bufRE := bufRW[:wholeNote/8]
	bufRS := bufRW[:wholeNote/16]
	bufRT := bufRW[:wholeNote/32]
	bufMap[' '] = bufRE

	// read lines
	reader := bufio.NewReader(os.Stdin)
	bufPlayLimit := 1024 * 1024 * 100
	bufLineLimit := 1024 * 100
	ctrlKeys := "RDH"
	measures := "WHQEST"
	DEBUG := false
	var bufPlay bytes.Buffer
	var bufLine bytes.Buffer
	var duration = 'Q' // default note duration
	var rest rune
	var ctrl rune
	var last rune
	var done bool
	for {
		bufPlay.Reset()
		bufLine.Reset()
		for {
			part, isPrefix, err := reader.ReadLine()
			if err != nil {
				done = true
				break
			}
			bufLine.Write(part)
			if bufLine.Len() > bufLineLimit {
				fmt.Println("Line exceeds 100KB limit.")
				return
			}
			if !isPrefix {
				break
			}
		}
		if done {
			break
		}
		if bufLine.Len() == 0 {
			continue
		}
		line := bufLine.String()
		fmt.Println(line)
		if strings.HasPrefix(line, "#") {
			// ignore comments
			continue
		}
		for _, key := range line {
			keystr := string(key)
			if key == ' ' && last == ' ' {
				// spaces after first space are ignored
				continue
			}
			if strings.ContainsAny(keystr, ctrlKeys) {
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
				}
				ctrl = 0
				continue
			}
			if rest > 0 {
				switch rest {
				case 'W':
					bufPlay.Write(bufRW)
				case 'H':
					bufPlay.Write(bufRH)
				case 'Q':
					bufPlay.Write(bufRQ)
				case 'E':
					bufPlay.Write(bufRE)
				case 'S':
					bufPlay.Write(bufRS)
				case 'T':
					bufPlay.Write(bufRT)
				}
				rest = 0
			}
			if buf, found := bufMap[key]; found {
				if last == 0 {
					last = key
				}
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
					bufPlay.Write(buf)
					if last != key {
						bufPlay.Write(boundary)
					}
					if bufPlay.Len() > bufPlayLimit {
						fmt.Println("Line wave buffer exceeds 100MB limit.")
						return
					}
				}
				last = key
			}
		}
		if bufPlay.Len() == 0 {
			continue
		}
		bufWave := bufPlay.Bytes()
		mergeNotes(bufWave)
		playback(bufWave)

		if DEBUG {
			fmt.Println("LINE")
			for i, bar := range bufWave {
				fmt.Printf("%d|%s\n", i, strings.Repeat("=", int(bar/4)))
			}
		}
	}
	flushSoundBuffer()
}

// Merges two wave form by fading for playing smooth
func mergeNotes(buf []byte) {
	half := len(boundary) / 2
	buflen := len(buf)
	var c int // count
	var f int // fill
	var middle int
	var found bool
	var first int
	DEBUG := false
	for i, bar := range buf {
		if bar == 0 && i > 0 {
			found = true
		}
		if found {
			if first == 0 {
				first = i
			}
			if c == half {
				middle = i
				buf[i] = 127
				f = 0
			}
			if middle > 0 {
				// fill left
				s := middle - half + f
				t := middle - half - f - 1
				buf[s] = fade(-buf[t], f)

				// fill right
				s = middle + half + f
				t = middle + half - f
				if buflen > s {
					buf[t] = fade(-buf[s], f)
				} else {
					if buflen > t {
						buf[t] = 127
					}
				}
			}
			c++
			f++
			if f > half {
				found = false
				middle = 0
				f = 0
				c = 0
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

func fade(bar byte, i int) byte {
	step := float64(i)
	gap := 127.0
	if step < gap {
		b := byte(step + float64(bar)*((gap-step)/gap))
		if b == 0 {
			b = 1
		}
		return b
	}
	return byte(gap)
}

// Generates sine wave for music notes
func keyFreq(freq float64, volume int) []byte {
	buf := make([]byte, quarterNote)
	bar := byte(127.0 * (float64(volume) / 100.0))
	barfloat := float64(bar)
	for i, _ := range buf {
		b := bar + byte(barfloat*math.Sin(float64(i)*freq))
		if b == 0 {
			b = 1
		}
		buf[i] = b
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
			if buf[cut]-127 < 6 {
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
