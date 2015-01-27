// +build linux

package main

/*
#cgo LDFLAGS: -lasound

#include <alsa/asoundlib.h>
*/
import "C"

import (
	"fmt"
	"os"
	"os/user"
	"unsafe"
)

var (
	pcm_handle *C.snd_pcm_t
)

func openSoundDevice(device string) {
	code := C.snd_pcm_open(&pcm_handle, C.CString(device), C.SND_PCM_STREAM_PLAYBACK, 0)
	if code < 0 {
		fmt.Println("snd_pcm_open:", strerror(code))
		os.Exit(1)
	}
	C.snd_pcm_drop(pcm_handle)
}

func initSoundDevice() {
	var sampleFormat C.snd_pcm_format_t = C.SND_PCM_FORMAT_S8
	if sample16bit {
		sampleFormat = C.SND_PCM_FORMAT_S16
	}
	code := C.snd_pcm_set_params(
		pcm_handle,
		sampleFormat,
		C.SND_PCM_ACCESS_RW_NONINTERLEAVED,
		2,
		44100,
		1,
		500000)
	if code < 0 {
		fmt.Println("snd_pcm_set_params:", strerror(code))
		os.Exit(1)
	}
	code = C.snd_pcm_prepare(pcm_handle)
	if code < 0 {
		fmt.Println("snd_pcm_prepare:", strerror(code))
		os.Exit(1)
	}
}

func playback(buf1, buf2 []int16) {
	bufsize := len(buf1)
	if bufsize < sampleRate {
		// prevent buffer underrun
		rest := make([]int16, sampleRate)
		buf1 = append(buf1, rest...)
		buf2 = append(buf2, rest...)
	}
	buffers := []unsafe.Pointer{
		unsafe.Pointer(&buf1[0]),
		unsafe.Pointer(&buf2[0]),
	}
	pos := &buffers[0]
	for {
		n := C.snd_pcm_writen(pcm_handle, pos, C.snd_pcm_uframes_t(bufsize))
		written := int(n)
		if written < 0 {
			if music.stopping {
				break
			}
			// error
			code := C.int(written)
			written = 0
			fmt.Fprintln(os.Stderr, "snd_pcm_writen:", code, strerror(code))
			code = C.snd_pcm_recover(pcm_handle, code, 0)
			if code < 0 {
				fmt.Fprintln(os.Stderr, "snd_pcm_recover:", strerror(code))
				break
			}
		}
		break // don't retry, breaks timing
		/*
			if written == bufsize {
				break
			}
			if written == 0 {
				C.snd_pcm_wait(pcm_handle, 1000)
				continue
			}
			fmt.Fprintf(os.Stderr, "snd_pcm_writen: wrote: %d/%d\n", written, bufsize)
			buffers = []unsafe.Pointer{
				unsafe.Pointer(&buf1[written]),
				unsafe.Pointer(&buf2[written]),
			}
			pos = &buffers[0]
			bufsize -= written
		*/
	}
	music.linePlayed <- true // notify that playback is done
}

func flushSoundBuffer() {
	if pcm_handle != nil {
		C.snd_pcm_drain(pcm_handle)
	}
}

func stopPlayBack() {
	if pcm_handle != nil {
		C.snd_pcm_drop(pcm_handle)
	}
}

func strerror(code C.int) string {
	return C.GoString(C.snd_strerror(code))
}

func closeSoundDevice() {
	if pcm_handle != nil {
		C.snd_pcm_close(pcm_handle)
		C.snd_pcm_hw_free(pcm_handle)
	}
}

func sendBell() {
	bell := []byte{7}
	os.Stdout.Write(bell)

	console, err := os.OpenFile("/dev/console", os.O_WRONLY, 0644)
	if err != nil {
		console, err = os.OpenFile("/dev/tty0", os.O_WRONLY, 0644)
		if err != nil {
			return
		}
	}
	defer console.Close()
	console.Write(bell)
}

func beepHomeDir() string {
	var home string
	usr, err := user.Current()
	if err != nil {
		fmt.Fprintln(os.Stderr, "Unable to get current user directory.")
		home = "/home"
	} else {
		home = usr.HomeDir
	}
	return home + "/.beep"
}
