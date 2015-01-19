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
	flagDuration  = flag.Float64("t", 250, "beep time duration in millisecond (1-60000)")
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
	bitsPerSample = 16
	sample16bit = bitsPerSample == 16
	sampleAmp16bit = 32767.0
	sampleAmp8bit = 127.0
)

var beepOptions = `Usage: beep [options]
  -c=1: beep count
  -d="default": audio device, Linux example: hw:0,0
  -f=523.25: frequency in Hertz (1-22050)
  -h: print help
  -l: beep per line from stdin
  -m: play music notes from stdin (see beep notation)
  -p: print the demo music by Mozart
  -t=1: beep time duration in millisecond (1-60000)
  -v=100: volume (1-100)
  -b: send bell to PC speaker
  -q: quiet stdout while playing music
  -n: print notes while playing music
  -o=file: output music waveform to a WAV file. Use '-' for stdout
`

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
		fmt.Printf("%s%s\n%s\n%s",
			beepOptions,
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
		reader := bufio.NewReader(os.Stdin)
		playMusicNotes(reader, volume)
		return
	}

	if writeBell {
		sendBell()
		return
	}

	// beep
	bar := sampleAmp16bit*(float64(volume) / 100.0)
	samples := int(sampleRate64 * (duration / 1000.0))
	rest := 0
	if count > 1 {
		rest = (sampleRate / 20) * 4 // 200ms
	}
	buf := make([]int16, samples+rest)
	var last int16
	var fade = 1024
	if samples < fade {
		fade = 1
	}
	for i, _ := range buf {
		if i < samples-fade {
			buf[i] = int16(bar*math.Sin(float64(i)*freq))
			last = buf[i]
		} else {
			if last > 0 {
				last -= 31
			} else {
				last += 31
			}
			buf[i] = last
		}
	}
	for i := 0; i < count; i++ {
		playback(buf, buf)
	}
	flushSoundBuffer()
}

func beepPerLine(volume int, freq float64) {
	buf := make([]int16, sampleRate/5)
	bar := sampleAmp16bit*(float64(volume) / 100.0)
	gap := sampleRate / 6
	var last int16
	for i, _ := range buf {
		if i < gap {
			buf[i] = int16(bar*math.Sin(float64(i)*freq))
			last = buf[i]
		} else {
			if last > 0 {
				last -= 31
			} else {
				last += 31
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
			playback(buf, buf)
		}
	}
	flushSoundBuffer()
}
