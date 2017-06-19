// +build linux

package beep

/*
#cgo LDFLAGS: -lasound

#include <alsa/asoundlib.h>

void *cbuf[2];
*/
import "C"

import (
	"fmt"
	"os"
	"os/user"
	"unsafe"
)

var (
	pcmHandle   *C.snd_pcm_t
	pcmHwParams *C.snd_pcm_hw_params_t
)

// OpenSoundDevice opens hardware sound device
func OpenSoundDevice(device string) {
	code := C.snd_pcm_open(
		&pcmHandle,
		C.CString(device),
		C.SND_PCM_STREAM_PLAYBACK,
		0)
	if code < 0 {
		fmt.Println("snd_pcm_open:", strerror(code))
		os.Exit(1)
	}
	C.snd_pcm_drop(pcmHandle)
}

// InitSoundDevice initialize sound device
func InitSoundDevice() {
	var sampleFormat C.snd_pcm_format_t = C.SND_PCM_FORMAT_S8
	if sample16bit {
		sampleFormat = C.SND_PCM_FORMAT_S16
	}

	if code := C.snd_pcm_hw_params_malloc(&pcmHwParams); code < 0 {
		fmt.Println("snd_pcm_hw_params_malloc:", strerror(code))
		os.Exit(1)
	}
	if code := C.snd_pcm_hw_params_any(pcmHandle, pcmHwParams); code < 0 {
		fmt.Println("snd_pcm_hw_params_any:", strerror(code))
		os.Exit(1)
	}

	// C.SND_PCM_ACCESS_RW_NONINTERLEAVED - is not working with PulseAudio
	code := C.snd_pcm_set_params(
		pcmHandle,
		sampleFormat,
		C.SND_PCM_ACCESS_RW_INTERLEAVED,
		1,
		44100,
		1,
		500000)
	if code < 0 {
		fmt.Println("snd_pcm_set_params:", strerror(code))
		os.Exit(1)
	}
	code = C.snd_pcm_prepare(pcmHandle)
	if code < 0 {
		fmt.Println("snd_pcm_prepare:", strerror(code))
		os.Exit(1)
	}
}

// Playback sends wave buffer to sound device
func Playback(buf1, buf2 []int16) {
	bufsize := len(buf1)
	if bufsize < SampleRate {
		// prevent buffer underrun
		rest := make([]int16, SampleRate)
		buf1 = append(buf1, rest...)
		//buf2 = append(buf2, rest...)
	}

	// Go 1.6 cgocheck fix: Can't pass Go pointer to C function
	//C.cbuf[0] = unsafe.Pointer(&buf1[0])
	//C.cbuf[1] = unsafe.Pointer(&buf2[0])

	// Changing to single channel interleaved buffer format for PulseAudio
	buf := unsafe.Pointer(&buf1[0])

	for {
		n := C.snd_pcm_writei(pcmHandle, buf, C.snd_pcm_uframes_t(bufsize))
		written := int(n)
		if written < 0 {
			if music.stopping {
				break
			}
			// error
			code := C.int(written)
			written = 0
			_ = written
			fmt.Fprintln(os.Stderr, "snd_pcm_writei:", code, strerror(code))
			code = C.snd_pcm_recover(pcmHandle, code, 0)
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
				C.snd_pcm_wait(pcmHandle, 1000)
				continue
			}
			fmt.Fprintf(os.Stderr, "snd_pcm_writei: wrote: %d/%d\n", written, bufsize)
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

// FlushSoundBuffer flushes sound buffer
func FlushSoundBuffer() {
	if pcmHandle != nil {
		C.snd_pcm_drain(pcmHandle)
	}
}

// StopPlayBack stops play back
func StopPlayBack() {
	if pcmHandle != nil {
		C.snd_pcm_drop(pcmHandle)
	}
}

func strerror(code C.int) string {
	return C.GoString(C.snd_strerror(code))
}

// CloseSoundDevice closes sound device
func CloseSoundDevice() {
	if pcmHandle != nil {
		C.snd_pcm_close(pcmHandle)
		C.snd_pcm_hw_free(pcmHandle)
	}
}

// SendBell send bell sound to console
func SendBell() {
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

// HomeDir returns user's home directory
func HomeDir() string {
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
