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
	code := C.snd_pcm_set_params(
		pcm_handle,
		C.SND_PCM_FORMAT_U8,
		C.SND_PCM_ACCESS_RW_NONINTERLEAVED,
		2,
		44100,
		1,
		500000)
	if code < 0 {
		fmt.Println("snd_pcm_set_params:", strerror(code))
		os.Exit(1)
	}
}

func playback(buf1 []byte, buf2 []byte) {
	var base unsafe.Pointer
	if buf2 != nil {
		base = unsafe.Pointer(&buf2[0])
	}
	buffers := []unsafe.Pointer{
		unsafe.Pointer(&buf1[0]),
		base,
	}
	bufsize := len(buf1)
	pos := &buffers[0]
	for {
		n := C.snd_pcm_writen(pcm_handle, pos, C.snd_pcm_uframes_t(bufsize))
		written := int(n)
		if written < 0 {
			// error
			fmt.Println("snd_pcm_writen:", written, strerror(C.int(written)))
			code := C.snd_pcm_recover(pcm_handle, C.int(written), 0)
			if code < 0 {
				fmt.Println("snd_pcm_recover:", strerror(code))
				break
			}
			written = 0
			fmt.Printf("snd_pcm_writen: buffer underrun: %d/%d\n", int(n), bufsize)
		}
		if written == bufsize {
			break
		}
		pos = &buffers[written]
		bufsize -= written
	}
}

func flushSoundBuffer() {
	C.snd_pcm_drain(pcm_handle)
}

func strerror(code C.int) string {
	return C.GoString(C.snd_strerror(code))
}

func closeSoundDevice() {
	C.snd_pcm_close(pcm_handle)
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
