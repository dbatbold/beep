// +build windows

package beep

/*
#cgo LDFLAGS: -lwinmm

#include <windows.h>
size_t wavehdrsize = sizeof(WAVEHDR);
*/
import "C"

import (
	"fmt"
	"os"
	"os/user"
	"time"
	"unsafe"
)

var (
	hwaveout    C.HWAVEOUT
	wavehdrLast *C.WAVEHDR
)

func OpenSoundDevice(device string) {
	var wfx C.WAVEFORMATEX

	wfx.wFormatTag = C.WAVE_FORMAT_PCM
	wfx.nChannels = 2
	wfx.nSamplesPerSec = C.DWORD(SampleRate)
	wfx.nAvgBytesPerSec = C.DWORD(SampleRate) * 2 * 2
	wfx.nBlockAlign = 2 * 2
	wfx.wBitsPerSample = 16

	res := C.waveOutOpen(&hwaveout, C.WAVE_MAPPER, &wfx, 0, 0, C.CALLBACK_NULL)
	if res != C.MMSYSERR_NOERROR {
		fmt.Fprintln(os.Stderr, "Error: waveOutOpen:", winmmErrorText(res))
		os.Exit(1)
	}
}

func InitSoundDevice() {
	wavehdrLast = nil
}

// Playback sends stereo wave buffer to sound device
func (m *Music) Playback(buf1, buf2 []int16) {
	bufWave := make([]int16, len(buf1)*2)
	for i, bar := range buf1 {
		bufWave[i*2] = bar
		bufWave[i*2+1] = buf2[i]
	}

	// Go 1.6 cgocheck fix: Can't pass Go pointer to C function
	wavehdr := (*C.WAVEHDR)(C.malloc(C.wavehdrsize))
	C.memset(unsafe.Pointer(wavehdr), 0, C.wavehdrsize)
	wavehdr.lpData = C.LPSTR(unsafe.Pointer(&bufWave[0]))
	wavehdr.dwBufferLength = C.DWORD(len(bufWave) * 2)

	res := C.waveOutPrepareHeader(hwaveout, wavehdr, C.UINT(C.wavehdrsize))
	if res != C.MMSYSERR_NOERROR {
		fmt.Fprintln(os.Stderr, "Error: waveOutPrepareHeader:", winmmErrorText(res))
		os.Exit(1)
	}

	if wavehdrLast != nil {
		for wavehdrLast.dwFlags&C.WHDR_DONE == 0 {
			// still playing
			time.Sleep(time.Millisecond)
		}
	}

	if !m.stopping {
		res = C.waveOutWrite(hwaveout, wavehdr, C.UINT(C.wavehdrsize))
		if res != C.MMSYSERR_NOERROR {
			fmt.Fprintln(os.Stderr, "Error: waveOutWrite:", winmmErrorText(res))
		}

		for wavehdr.dwFlags&C.WHDR_DONE == 0 {
			// still playing
			time.Sleep(time.Millisecond)
		}
		res = C.waveOutUnprepareHeader(hwaveout, wavehdr, C.UINT(C.wavehdrsize))
		if res != C.MMSYSERR_NOERROR {
			fmt.Fprintln(os.Stderr, "Error: waveOutUnprepareHeader:", winmmErrorText(res))
		}

		for wavehdr.dwFlags&C.WHDR_DONE == 0 {
			// still playing
			time.Sleep(time.Millisecond)
		}
		res = C.waveOutUnprepareHeader(hwaveout, wavehdr, C.UINT(C.wavehdrsize))
		if res != C.MMSYSERR_NOERROR {
			fmt.Fprintln(os.Stderr, "Error: waveOutUnprepareHeader:", winmmErrorText(res))
		}
	}

	wavehdrLast = wavehdr

	m.linePlayed <- true // notify that playback is done
}

func FlushSoundBuffer() {
	if wavehdrLast != nil {
		var wavehdr C.WAVEHDR
		for wavehdrLast.dwFlags&C.WHDR_DONE == 0 {
			time.Sleep(time.Millisecond)
		}
		wdrsize := C.UINT(unsafe.Sizeof(wavehdr))
		res := C.waveOutUnprepareHeader(hwaveout, wavehdrLast, wdrsize)
		if res != C.MMSYSERR_NOERROR {
			fmt.Fprintln(os.Stderr, "Error: waveOutUnprepareHeader:", winmmErrorText(res))
		}
		C.free(unsafe.Pointer(wavehdrLast))
	}
}

func CloseSoundDevice() {
	res := C.waveOutReset(hwaveout)
	if res != C.MMSYSERR_NOERROR {
		fmt.Fprintln(os.Stderr, "Error: waveOutReset:", winmmErrorText(res))
	}
	res = C.waveOutClose(hwaveout)
	if res != C.MMSYSERR_NOERROR {
		fmt.Fprintln(os.Stderr, "Error: waveOutClose:", winmmErrorText(res))
	}
}

func StopPlayBack() {
	res := C.waveOutReset(hwaveout)
	if res != C.MMSYSERR_NOERROR {
		fmt.Fprintln(os.Stderr, "Error: waveOutReset:", winmmErrorText(res))
	}
}

func SendBell() {
	bell := []byte{7}
	os.Stdout.Write(bell)
}

func winmmErrorText(res C.MMRESULT) string {
	var buf [1024]byte
	C.waveOutGetErrorText(res, C.LPSTR(unsafe.Pointer(&buf[0])), C.UINT(len(buf)))
	return fmt.Sprintf("%v: %s", res, string(buf[:]))
}

func HomeDir() string {
	var home string
	usr, err := user.Current()
	if err != nil {
		fmt.Fprintln(os.Stderr, "Unable to get current user directory.")
		home = `C:`
	} else {
		home = usr.HomeDir
	}
	return home + `\_beep`
}
