//go:build linux
// +build linux

package beep

/*
#cgo LDFLAGS: -lasound

#include <alsa/asoundlib.h>

void *cbuf[2];
*/
import "C"

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"os"
	"os/user"
	"strconv"
	"strings"
	"unsafe"
)

var (
	pcmHandle   *C.snd_pcm_t
	pcmHwParams *C.snd_pcm_hw_params_t
)

// OpenSoundDevice opens hardware sound device
func OpenSoundDevice(device string) error {
	code := C.snd_pcm_open(
		&pcmHandle,
		C.CString(device),
		C.SND_PCM_STREAM_PLAYBACK,
		0)
	if code < 0 {
		err := fmt.Errorf("snd_pcm_open: %v\n", strerror(code))
		fmt.Fprintln(os.Stderr, err)
		return err
	}
	C.snd_pcm_drop(pcmHandle)

	return nil
}

// InitSoundDevice initialize sound device
func InitSoundDevice() error {
	var sampleFormat C.snd_pcm_format_t = C.SND_PCM_FORMAT_S8
	if sample16bit {
		sampleFormat = C.SND_PCM_FORMAT_S16
	}

	if code := C.snd_pcm_hw_params_malloc(&pcmHwParams); code < 0 {
		err := fmt.Errorf("snd_pcm_hw_params_malloc: %v\n", strerror(code))
		fmt.Fprintln(os.Stderr, err)
		return err
	}
	if code := C.snd_pcm_hw_params_any(pcmHandle, pcmHwParams); code < 0 {
		err := fmt.Errorf("snd_pcm_hw_params_any: %v\n", strerror(code))
		fmt.Fprintln(os.Stderr, err)
		return err
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
		err := fmt.Errorf("snd_pcm_set_params: %v\n", strerror(code))
		fmt.Fprintln(os.Stderr, err)
		return err
	}
	code = C.snd_pcm_prepare(pcmHandle)
	if code < 0 {
		err := fmt.Errorf("snd_pcm_prepare: %v\n", strerror(code))
		fmt.Fprintln(os.Stderr, err)
		return err
	}

	return nil
}

// Playback sends stereo wave buffer to sound device
func (m *Music) Playback(buf1, buf2 []int16) {
	bufsize := len(buf1)
	if bufsize < SampleRate {
		// prevent buffer underrun
		rest := make([]int16, SampleRate)
		buf1 = append(buf1, rest...)
		//buf2 = append(buf2, rest...)
	}

	// Changing to single channel interleaved buffer format for PulseAudio
	buf := unsafe.Pointer(&buf1[0])

	for {
		n := C.snd_pcm_writei(pcmHandle, buf, C.snd_pcm_uframes_t(bufsize))
		written := int(n)
		if written < 0 {
			if m.stopping {
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
	m.linePlayed <- true // notify that playback is done
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

// BatteryLevel return battery charge level.
func BatteryLevel() (int, error) {
	const prefix = "POWER_SUPPLY_CAPACITY="
	var uevent *os.File
	for i := 0; i <= 3; i++ {
		fname := fmt.Sprintf("/sys/class/power_supply/BAT%d/uevent", i)
		file, err := os.Open(fname)
		if err != nil {
			log.Println(err)
			continue
		}
		defer file.Close()
		uevent = file
		break
	}
	if uevent == nil {
		return 0, io.EOF
	}
	scanner := bufio.NewScanner(uevent)
	var level int
	err := io.EOF
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, prefix) {
			level, err = strconv.Atoi(strings.Split(line, prefix)[1])
			break
		}
	}
	return level, err
}
