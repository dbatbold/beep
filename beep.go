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
)

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

	if help {
		fmt.Println("beep [options]")
		flag.PrintDefaults()
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

	buf := make([]byte, 1024*10*duration)
	bar := 127.0 * (float64(volume) / 100.0)
	for i, _ := range buf[:] {
		buf[i] = byte(127 + (bar * math.Sin(float64(i)*freq)))
	}
	bufLow := make([]byte, 1024*5)
	for i, _ := range bufLow[:] {
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
	for i, _ := range buf[:] {
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
