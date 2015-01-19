// +build windows

package main

/*
#cgo LDFLAGS: -lwinmm

#include <stdio.h>
#include <windows.h>
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

func openSoundDevice(device string) {
	var wfx C.WAVEFORMATEX

	wfx.wFormatTag = C.WAVE_FORMAT_PCM
	wfx.nChannels = 2
	wfx.nSamplesPerSec = C.DWORD(sampleRate)
	wfx.nAvgBytesPerSec = C.DWORD(sampleRate)*2*2
	wfx.nBlockAlign = 2*2
	wfx.wBitsPerSample = 16

	res := C.waveOutOpen(&hwaveout, C.WAVE_MAPPER, &wfx, 0, 0, C.CALLBACK_NULL);
	if res != C.MMSYSERR_NOERROR {
		fmt.Fprintln(os.Stderr, "Error: waveOutOpen:", winmmErrorText(res))
		os.Exit(1)
	}
}

func initSoundDevice() {
}

func playback(buf1, buf2 []int16, notes string) {
	bufWave := make([]int16, len(buf1)*2)
	for i, bar := range buf1 {
		bufWave[i*2] = bar
		bufWave[i*2+1] = buf2[i]
	}

	var wavehdr C.WAVEHDR
	wdrsize := C.UINT(unsafe.Sizeof(wavehdr))
	wavehdr.lpData = C.LPSTR(unsafe.Pointer(&bufWave[0]))
	wavehdr.dwBufferLength = C.DWORD(len(bufWave)*2)

	res := C.waveOutPrepareHeader(hwaveout, &wavehdr, wdrsize)
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

	res = C.waveOutWrite(hwaveout, &wavehdr, wdrsize)
	if res != C.MMSYSERR_NOERROR {
		fmt.Fprintln(os.Stderr, "Error: waveOutWrite:", winmmErrorText(res))
	}

	if !*flagQuiet && len(notes) > 0 {
		fmt.Println(notes)
	}

	for wavehdr.dwFlags&C.WHDR_DONE == 0 {
		// still playing
		time.Sleep(time.Millisecond)
	}
	res = C.waveOutUnprepareHeader(hwaveout, &wavehdr, wdrsize)
	if res != C.MMSYSERR_NOERROR {
		fmt.Fprintln(os.Stderr, "Error: waveOutUnprepareHeader:", winmmErrorText(res))
	}

	wavehdrLast = &wavehdr
	
	waiter <- 1 // notify that playback is done
}

func flushSoundBuffer() {
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
	}
}

func closeSoundDevice() {
	res := C.waveOutReset(hwaveout)
	if res != C.MMSYSERR_NOERROR {
		fmt.Fprintln(os.Stderr, "Error: waveOutReset:", winmmErrorText(res))
	}
	res = C.waveOutClose(hwaveout)
	if res != C.MMSYSERR_NOERROR {
		fmt.Fprintln(os.Stderr, "Error: waveOutClose:", winmmErrorText(res))
	}
}

func sendBell() {
	bell := []byte{7}
	os.Stdout.Write(bell)
}

func winmmErrorText(res C.MMRESULT) string {
	var buf [1024]byte
	C.waveOutGetErrorText(res, C.LPSTR(unsafe.Pointer(&buf[0])), C.UINT(len(buf)))
	return fmt.Sprintf("%v: %s", res, string(buf[:]))
}

func beepHomeDir() string {
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
