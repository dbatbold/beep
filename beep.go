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
	flagCount     = flag.Int("c", 1, "count")
	flagFreq      = flag.Float64("f", 0.07459, "frequency")
	flagVolume    = flag.Int("v", 100, "volume (1-100)")
	flagDuration  = flag.Int("t", 1, "time duration (1-100)")
	flagDevice    = flag.String("d", "default", "audio device (hw:0,0)")
	flagLine      = flag.Bool("l", false, "beep per line from stdin")
	flagMusic     = flag.Bool("m", false, "play music notes from stdin (see beep notation)")
	flagPrintDemo = flag.Bool("p", false, "print a demo music by Mozart")
	flagBell      = flag.Bool("b", false, "send bell to PC speaker")
)

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
		playMusicNotes(volume, "")
		return
	}

	if writeBell {
		sendBell()
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
		playback(buf, buf)
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
			if last > bar {
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
			playback(buf, buf)
		}
	}
	flushSoundBuffer()
}
