package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"strings"

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
	flagMidiNote  = flag.String("mn", "", "parses MIDI file and print notes")
	flagPlayNotes = flag.String("play", "", "play notes from command argument")

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
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("Error:", r)
		}
	}()
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
	midiNote := *flagMidiNote
	musicNotes := *flagPlayNotes

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

	beep.OpenSoundDevice(device)
	beep.InitSoundDevice()
	defer beep.CloseSoundDevice()

	if lineBeep {
		beep.PlayPerLine(music, volume, freq)
		return
	}

	if playMusic {
		beep.PlayMusicSheet(music, volume)
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
		parseMidiBeep(music, midiPlay)
		return
	}

	if len(midiNote) > 0 {
		parseMidiNote(midiNote)
		return
	}

	if len(musicNotes) > 0 {
		playMusicNotesFromCL(music, musicNotes, volume)
		return
	}

	beep.PlayBeep(music, volume, duration, count, freq)
}

// Play a MIDI file
func parseMidiBeep(music *beep.Music, filename string) {
	midi, err := beep.ParseMidi(filename, false)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
	} else {
		midi.Play(music)
	}
}

// Parses a MIDI file and print notes
func parseMidiNote(filename string) {
	_, err := beep.ParseMidi(filename, true)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
	}
}

func playMusicNotesFromCL(music *beep.Music, musicNotes string, volume int) {
	reader := bufio.NewReader(strings.NewReader(musicNotes))
	go beep.PlayMusicNotes(music, reader, volume)
	music.Wait()
	beep.FlushSoundBuffer()
}
