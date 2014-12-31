/*
 beep - A simple sound notifer with a music note player
 Batbold Dashzeveg
 2014-12 GPL v2
*/

package main

import (
	"bufio"
	"flag"
	"fmt"
	"math"
	"os"
	"unsafe"
)

/*
#cgo LDFLAGS: -lasound

#include <alsa/asoundlib.h>
*/
import "C"

var (
	flagHelp     = flag.Bool("h", false, "help")
	flagCount    = flag.Int("c", 1, "count")
	flagFreq     = flag.Float64("f", 0.088, "frequency")
	flagVolume   = flag.Int("v", 100, "volume (1-100)")
	flagDuration = flag.Int("t", 1, "time duration (1-100)")
	flagDevice   = flag.String("d", "default", "audio device (hw:0,0)")
	flagLine     = flag.Bool("l", false, "beep per line via pipe input")
	flagMusic    = flag.Bool("m", false, "play music notes via pipe input (see piano key map)")
	flagPrintDemo = flag.Bool("p", false, "print demo music (Mozart K33b)")
)

var pianoKeys = `
  | | | | | | | | | | | | | | | | | | | | | | |
  |2|3| |5|6|7| |9|0| |=|a|s| |f|g| |j|k|l| |'|
 | | | | | | | | | | | | | | | | | | | | | | |
 |q|w|e|r|t|y|u|i|o|p|[|]|z|x|c|v|b|n|m|,|.|/|

 ' ' - whole rest
 ':' - half rest
 '!' - quarter rest
 'others' - whole rest
`

// Mozart K33b
var demoMusic = `
 c c cszsc z [!
 c c cszsc z [!
 v v vcscv s ] v!
 c c cszsc z [ c!
 s s sz]zs ] p!
 s sz][z][ppp !i:y:rr
`

func main() {
	var handle *C.snd_pcm_t

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
		fmt.Println(pianoKeys)
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

	code := C.snd_pcm_open(&handle, C.CString(device), C.SND_PCM_STREAM_PLAYBACK, 0)
	if code < 0 {
		fmt.Println("snd_pcm_open:", strerror(code))
		os.Exit(1)
	}
	C.snd_pcm_drop(handle)
	defer C.snd_pcm_close(handle)

	code = C.snd_pcm_set_params(
		handle,
		C.SND_PCM_FORMAT_U8,
		C.SND_PCM_ACCESS_RW_INTERLEAVED,
		1,
		44100,
		1,
		500000)
	if code < 0 {
		fmt.Println("snd_pcm_set_params:", strerror(code))
		os.Exit(1)
	}

	if lineBeep {
		beepPerLine(handle, volume, freq, duration)
		return
	}

	if playMusic {
		playMusicNotes(handle, volume)
		return
	}

	buf := make([]byte, 1024*10*duration)
	bar := 127.0 * (float64(volume) / 100.0)
	for i, _ := range buf {
		buf[i] = byte(127 + (bar * math.Sin(float64(i)*freq)))
	}
	bufLow := make([]byte, 1024*5)
	for i, _ := range bufLow {
		bufLow[i] = 127
	}
	for i := 0; i < count; i++ {
		playback(handle, buf)
		playback(handle, bufLow)
	}
	C.snd_pcm_drain(handle)
}

func beepPerLine(handle *C.snd_pcm_t, volume int, freq float64, duration int) {
	buf := make([]byte, 1024*7*duration)
	bar := 127.0 * (float64(volume) / 100.0)
	gap := 1024*4*duration
	var last byte
	for i, _ := range buf {
		if i < gap {
			buf[i] = byte(127 + (bar * math.Sin(float64(i)*freq)))
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
			playback(handle, buf)
		}
	}
	C.snd_pcm_drain(handle)
}

func playMusicNotes(handle *C.snd_pcm_t, volume int) {
	cnote := 0.0373 // C1
	tune := []float64{
		22, // C1#
		24, // D1
		24, // D1#
		26, // E1
		28, // F1
		30, // F1#
		32, // G1
		33, // G1#
		35, // A1
		37, // A1#
		40, // B1
		42, // C2
		44, // C2#
		48, // D2
		48, // D2#
		53, // E2
		56, // F2
		58, // F2#
		64, // G2
		64, // G2#
		72, // A2
		74, // A2#
		80, // B2
		84, // C3
		87, // C3#
		97, // D3
		98, // D3#
		104, // E3
		112, // F3
		120, // F3#
		127, // G3
		125, // G3#
		148, // A3
		150, // A3#
		157, // B3
		163, // C4
		180, // C4#
		0,
	}
	keys := "q2w3er5t6y7ui9o0p[=]azsxcfvgbnjmk,l./'"
	freqMap := make(map[int32]float64)
	node := cnote
	for i, key := range keys {
		freqMap[key] = node
		node += tune[i] / 10000.0
	}
	bufMap := make(map[int32][]byte)

	for key, freq := range freqMap {
		bufMap[key] = keyFreq(freq, volume)
	}

	bufRest := make([]byte, 10240)
	for i, _ := range bufRest {
		bufRest[i] = 127
	}
	bufMap[' '] = bufRest
	bufMap[':'] = bufRest[:5120]
	bufMap['!'] = bufRest[:2560]

	reader := bufio.NewReader(os.Stdin)
	for {
		line, _, err := reader.ReadLine()
		if err != nil {
			break
		}
		for _, key := range string(line) {
			if buf, ok := bufMap[key]; ok {
				playback(handle, buf)
			} else {
				playback(handle, bufRest)
			}
		}
	}
	C.snd_pcm_drain(handle)
}

func keyFreq(freq float64, volume int) []byte {
	buf := make([]byte, 10240)
	bar := 127.0 * (float64(volume) / 100.0)
	fade := bar
	for i, _ := range buf {
		if i > len(buf) - 127 {
			if fade < 127 {
				fade++
			}
		} else {
			if fade > 0 {
				fade--
			}
		}
		buf[i] = byte(127 + ((bar - fade) * math.Sin(float64(i)*freq)))
	}
	return buf
}

func playback(handle *C.snd_pcm_t, buf []byte) {
	n := C.snd_pcm_writei(handle, unsafe.Pointer(&buf[0]), C.snd_pcm_uframes_t(len(buf)))
	if n < 0 {
		code := C.snd_pcm_recover(handle, C.int(n), 0)
		if code < 0 {
			fmt.Println("snd_pcm_recover:", strerror(code))
		}
	}
}

func strerror(code C.int) string {
	return C.GoString(C.snd_strerror(code))
}
