// +build linux darwin

package beep

import (
	"fmt"
	"os"
	"os/user"
	"runtime"
	"time"
	"unsafe"
)

// #cgo LDFLAGS: -lsoundio -lm
// #include "beep_soundio.h"
import "C"

// OpenSoundDevice opens hardware sound device
func OpenSoundDevice(device string) error {
	os := C.CString(runtime.GOOS)
	if s := C.init_sound_device(os); s != nil {
		return fmt.Errorf("unable to initialize device: %s", C.GoString(s))
	}
	return nil
}

// InitSoundDevice initialize sound device
func InitSoundDevice() error {
	if s := C.open_sound_device(); s != nil {
		return fmt.Errorf("unable to open device: %s", C.GoString(s))
	}
	return nil
}

// Playback sends stereo wave buffer to sound device
func (m *Music) Playback(buf1, buf2 []int16) {
	buf := (*C.short)(unsafe.Pointer(&buf1[0]))
	bufLen := len(buf1)
	C.playback(buf, buf, C.int(bufLen))

	//frame := time.Duration(bufLen / SampleRate)
	//timer := time.NewTimer(frame*time.Second + 10*time.Millisecond)
	//select {
	//case <-timer.C:
	//case <-m.stopped:
	//	fmt.Println("S")
	//}
	m.linePlayed <- true
}

// FlushSoundBuffer flushes sound buffer
func FlushSoundBuffer() {
	C.flush_sound_buffer()
	time.Sleep(time.Second / 2)
}

// StopPlayBack stops play back
func StopPlayBack() {
	C.stop_playback()
}

// CloseSoundDevice closes sound device
func CloseSoundDevice() {
	C.close_sound_device()
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
