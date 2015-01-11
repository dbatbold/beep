/*
 beep - A simple sound notifier with a music note engine
 Batbold Dashzeveg
 2014-12-31 GPL v2
*/

package main

import (
	"bufio"
	"flag"
	"fmt"
	"math"
	"os"
)

var (
	flagHelp      = flag.Bool("h", false, "help")
	flagCount     = flag.Int("c", 1, "beep count")
	flagFreq      = flag.Float64("f", 523.25, "frequency in Hertz (1-22050)")
	flagVolume    = flag.Int("v", 100, "volume (1-100)")
	flagDuration  = flag.Float64("t", 250, "beep time duration in millisecond (1-600000)")
	flagDevice    = flag.String("d", "default", "audio device, Linux example: hw:0,0")
	flagLine      = flag.Bool("l", false, "beep per line from stdin")
	flagMusic     = flag.Bool("m", false, "play music notes from stdin (see beep notation)")
	flagPrintDemo = flag.Bool("p", false, "print a demo music by Mozart")
	flagBell      = flag.Bool("b", false, "send bell to PC speaker")
	flagQuiet     = flag.Bool("q", false, "quiet stdout while playing music")
	flagNotes     = flag.Bool("n", false, "print notes while playing music")
	flagOutput    = flag.String("o", "", "output music waveform to file. Use '-' for stdout")

	sampleRate   = 44100
	sampleRate64 = float64(sampleRate)
)

func main() {
	flag.Parse()

	help := *flagHelp
	freqHertz := *flagFreq
	count := *flagCount
	volume := *flagVolume
	duration := *flagDuration
	device := *flagDevice
	lineBeep := *flagLine
	playMusic := *flagMusic
	printDemo := *flagPrintDemo
	writeBell := *flagBell

	if help {
		fmt.Fprintf(os.Stderr, "Usage: beep [options]\n")
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "%s\n%s\n%s",
			beepNotation,
			demoMusic,
			demoHelp,
		)
		return
	}
	if printDemo {
		fmt.Print(demoMusic)
		return
	}
	if volume < 1 || volume > 100 {
		volume = 100
	}
	if duration < 1 || duration > 1000*60 {
		duration = 250
	}
	if freqHertz < 1 || freqHertz > sampleRate64/2 {
		fmt.Fprintf(os.Stderr, "Invalid frequency. Must be 1-22050")
		os.Exit(1)
	}
	freq := hertzToFreq(freqHertz)

	openSoundDevice(device)
	initSoundDevice()
	defer closeSoundDevice()

	if lineBeep {
		beepPerLine(volume, freq)
		return
	}

	if playMusic {
		playMusicNotes(volume, "")
		return
	}

	if writeBell {
		sendBell()
		return
	}

var piano Piano
piano.GenerateNote(0)
return

	// beep
	bar := byte(127.0 * (float64(volume) / 100.0))
	samples := int(sampleRate64 * (duration / 1000.0))
	rest := 0
	if count > 1 {
		rest = (sampleRate / 20) * 4 // 200ms
	}
	buf := make([]byte, samples+rest)
	var last byte
	var fade = 255
	if samples < fade {
		fade = 0
	}
	for i, _ := range buf {
		if i < samples-fade {
			buf[i] = 127 + byte(float64(bar)*math.Sin(float64(i)*freq))
			last = buf[i]
		} else {
			if last > 127 {
				last--
			} else {
				last++
			}
			buf[i] = last
		}
	}
	for i := 0; i < count; i++ {
		playback(buf, buf, "")
	}
	flushSoundBuffer()
}

func beepPerLine(volume int, freq float64) {
	buf := make([]byte, sampleRate/5)
	bar := byte(127.0 * (float64(volume) / 100.0))
	gap := sampleRate / 6
	var last byte
	for i, _ := range buf {
		if i < gap {
			buf[i] = 127 + byte(float64(bar)*math.Sin(float64(i)*freq))
			last = buf[i]
		} else {
			if last > 127 {
				last--
			} else {
				last++
			}
			buf[i] = last
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
			playback(buf, buf, "")
		}
	}
	flushSoundBuffer()
}
