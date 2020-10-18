package main

import (
	"bufio"
	"bytes"
	"flag"
	"fmt"
	"io"
	"math"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/dbatbold/beep"
)

const (
	noteC5 = 523.25
)

var (
	flagHelp      = flag.Bool("h", false, "help")
	flagCount     = flag.Int("c", 1, "beep count")
	flagFreq      = flag.Float64("f", noteC5, "frequency in Hertz (1-22050)")
	flagVolume    = flag.Int("v", 100, "volume (1-100)")
	flagDuration  = flag.Int("t", 250, "beep time duration in millisecond (1-60000)")
	flagDevice    = flag.String("d", "default", "audio device, Linux example: hw:0,0")
	flagLine      = flag.Bool("l", false, "beep per line from stdin")
	flagMusic     = flag.Bool("m", false, "play music notes from stdin (see beep notation)")
	flagPrintDemo = flag.Bool("p", false, "print demo music sheet (Mozart K33b)")
	flagBell      = flag.Bool("b", false, "send bell to PC speaker")
	flagQuiet     = flag.Bool("q", false, "quiet stdout while playing music")
	flagNotes     = flag.Bool("n", false, "print notes while playing music")
	flagOutput    = flag.String("o", "", "output music waveform to file. Use '-' for stdout")
	flagWeb       = flag.Bool("w", false, "start beep web server")
	flagWebIP     = flag.String("a", "127.0.0.1:4444", "web server address")
	flagVoiceDl   = flag.Bool("vd", false, "download voice files, by default downloads all voices")
	flagMidiPlay  = flag.String("mp", "", "play MIDI file")
	flagMidiURL   = flag.String("mu", "", "play MIDI from URL")
	flagMidiNote  = flag.String("mn", "", "parses MIDI file and print notes")
	flagPlayNotes = flag.String("play", "", "play notes from command argument")
	flagPlayURL   = flag.String("url", "", "play notes from URL")
	flagBattery   = flag.Bool("battery", false, "monitor battery and alert low charge level")

	music *beep.Music
)

const (
	demoHelp = `To play a demo music, run:

$ beep -vd
$ beep -m demo
`
	intro = `beep - Sound notifier with music note engine

Batbold Dashzeveg 2014-12-31`
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
	webServer := *flagWeb
	downloadVoices := *flagVoiceDl
	midiPlay := *flagMidiPlay
	midiURL := *flagMidiURL
	midiNote := *flagMidiNote
	musicNotes := *flagPlayNotes
	musicURL := *flagPlayURL

	beep.PrintSheet = !*flagQuiet
	beep.PrintNotes = *flagNotes

	if help {
		flag.Usage()
		fmt.Fprintf(os.Stderr, "\n%s\n%s\n%s\n%s\n",
			intro,
			beep.BeepNotation,
			beep.BuiltinMusic[0].Notation, //demoMusic,
			demoHelp,
		)
		return
	}
	if printDemo {
		fmt.Print(beep.DemoMusic)
		return
	}
	if volume < 1 || volume > 100 {
		volume = 100
	}
	if duration < 1 || duration > 1000*60 {
		duration = 250
	}
	if freqHertz < 1 || freqHertz > beep.SampleRate64/2 {
		fmt.Fprintf(os.Stderr, "Invalid frequency. Must be 1-22050")
		os.Exit(1)
	}
	freq := beep.HertzToFreq(freqHertz)

	music = beep.NewMusic(*flagOutput)

	if err := beep.OpenSoundDevice(device); err != nil {
		fmt.Println("failed to open sound device:", err)
		os.Exit(1)
	}
	if err := beep.InitSoundDevice(); err != nil {
		fmt.Println("failed to initialize sound device:", err)
		os.Exit(1)
	}
	defer beep.CloseSoundDevice()

	if lineBeep {
		playPerLine(music, volume, freq)
		return
	}

	if playMusic {
		playMusicScore(music, volume)
		return
	}

	if writeBell {
		beep.SendBell()
		return
	}

	if webServer {
		beep.StartWebServer(music, *flagWebIP)
		return
	}

	if downloadVoices {
		var names []string
		for i, arg := range os.Args {
			if i == 0 || strings.HasPrefix(arg, "-") {
				continue
			}
			names = append(names, arg)
		}
		beep.DownloadVoiceFiles(music, os.Stdout, names)
		return
	}

	if len(midiPlay) > 0 {
		parseMidiBeep(music, midiPlay, false)
		return
	}
	if len(midiNote) > 0 {
		parseMidiBeep(music, midiNote, true)
		return
	}
	if len(midiURL) > 0 {
		parseMidiURL(music, midiURL)
		return
	}
	if len(musicNotes) > 0 {
		playMusicNotesFromCL(music, musicNotes, volume)
		return
	}
	if len(musicURL) > 0 {
		playMusicNotesFromURL(music, musicURL, volume)
		return
	}

	if *flagBattery {
		for {
			level, err := beep.BatteryLevel()
			if err == io.EOF {
				fmt.Println("OS not supported")
				os.Exit(1)
			}
			fmt.Printf("Battery %d%%\n", level)
			if level < 10 {
				fmt.Println("Battery is low")
				playBeep(music, volume, duration, 3, freq)
				time.Sleep(time.Second * 60)
			} else {
				time.Sleep(time.Second * 300)
			}
		}
	}

	playBeep(music, volume, duration, count, freq)
}

// Play a MIDI file
func parseMidiBeep(music *beep.Music, filename string, print bool) {
	file, err := os.Open(filename)
	if err != nil {
		fmt.Println("failed to open MIDI file:", err)
		os.Exit(1)
	}
	defer file.Close()
	midi, err := beep.ParseMidi(music, file, print)
	if err != nil {
		fmt.Println("failed to play MIDI file:", err)
		os.Exit(1)
	} else {
		midi.Play()
	}
}

// Play a MIDI from URL
func parseMidiURL(music *beep.Music, urlpath string) {
	req, err := http.NewRequest("GET", urlpath, nil)
	if err != nil {
		fmt.Println("failed to create request:", err)
		os.Exit(1)
	}
	client := http.Client{}
	res, err := client.Do(req)
	if err != nil {
		fmt.Printf("failed to fetch file from %q: %v\n", urlpath, err)
		os.Exit(1)
	}
	if res.StatusCode != 200 {
		fmt.Printf("failed to fetch file from %q %s: %v\n",
			urlpath, res.Status, err)
		os.Exit(1)
	}
	defer res.Body.Close()
	midi, err := beep.ParseMidi(music, res.Body, false)
	if err != nil {
		fmt.Printf("failed to play midi from %q: %v\n", urlpath, err)
	} else {
		midi.Play()
	}
}

func playMusicNotesFromCL(music *beep.Music, musicNotes string, volume int) {
	reader := bufio.NewReader(strings.NewReader(musicNotes))
	go music.Play(reader, volume)
	music.Wait()
	beep.FlushSoundBuffer()
}

func playMusicNotesFromURL(music *beep.Music, urlpath string, volume int) {
	req, err := http.NewRequest("GET", urlpath, nil)
	if err != nil {
		fmt.Println("failed to create request:", err)
		os.Exit(1)
	}
	client := http.Client{}
	res, err := client.Do(req)
	if err != nil {
		fmt.Printf("failed to fetch file from %q: %v\n", urlpath, err)
		os.Exit(1)
	}
	if res.StatusCode != 200 {
		fmt.Printf("failed to fetch file from %q %s: %v\n",
			urlpath, res.Status, err)
		os.Exit(1)
	}
	defer res.Body.Close()
	reader := bufio.NewReader(res.Body)
	go music.Play(reader, volume)
	music.Wait()
	beep.FlushSoundBuffer()
}

func playMusicScore(music *beep.Music, volume int) {
	var files []io.Reader
	for _, fname := range flag.Args() {
		if fname == "demo" {
			demo := bytes.NewBuffer([]byte(beep.DemoMusic))
			files = append(files, demo)
			continue
		}
		file, err := os.Open(fname)
		if err != nil {
			fmt.Fprintln(os.Stderr, "Error:", err)
			os.Exit(1)
		}
		files = append(files, file)
	}
	if len(files) == 0 {
		files = append(files, os.Stdin)
	}
	for i, file := range files {
		reader := bufio.NewReader(file)
		if i > 0 {
			fmt.Println()
			time.Sleep(time.Second)
		}
		beep.InitSoundDevice()
		go music.Play(reader, volume)
		music.Wait()
		beep.FlushSoundBuffer()
	}
	for _, file := range files {
		if file != os.Stdin {
			if closer, ok := file.(io.ReadCloser); ok {
				closer.Close()
			}
		}
	}
}

func playBeep(music *beep.Music, volume, duration, count int, freq float64) {
	bar := beep.SampleAmp16bit * (float64(volume) / 100.0)
	samples := int(beep.SampleRate64 * (float64(duration) / 1000.0))
	rest := 0
	if count > 1 {
		rest = (beep.SampleRate / 20) * 4 // 200ms
	}
	buf := make([]int16, samples+rest)
	var last int16
	var fade = 1024
	if samples < fade {
		fade = 1
	}
	for i := range buf {
		if i < samples-fade {
			buf[i] = int16(bar * math.Sin(float64(i)*freq))
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
	beep.InitSoundDevice()
	for i := 0; i < count; i++ {
		go music.Playback(buf, buf)
		music.WaitLine()
	}
	beep.FlushSoundBuffer()
}

func playPerLine(music *beep.Music, volume int, freq float64) {
	buf := make([]int16, beep.SampleRate/5)
	bar := beep.SampleAmp16bit * (float64(volume) / 100.0)
	gap := beep.SampleRate / 6
	var last int16
	for i := range buf {
		if i < gap {
			buf[i] = int16(bar * math.Sin(float64(i)*freq))
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
			go music.Playback(buf, buf)
			music.WaitLine()
		}
	}
	beep.FlushSoundBuffer()
}
