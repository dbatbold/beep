// +build windows

package main

/*
#cgo LDFLAGS: -lwinmm

#include <windows.h>
*/
import "C"

import (
	"bytes"
	"fmt"
	"os"
	"time"
	"unsafe"
)

var (
	hwaveout    C.HWAVEOUT
	wavehdrLast *C.WAVEHDR
)

func openSoundDevice(device string) {
	var wfx C.WAVEFORMATEX
	var dwCallback C.DWORD
	var dwCallbackInstance C.DWORD
	var fdwOpen C.DWORD = C.CALLBACK_NULL

	wfx.wFormatTag = C.WAVE_FORMAT_PCM
	wfx.nChannels = C.WORD(2)
	wfx.nSamplesPerSec = C.DWORD(44100)
	wfx.nAvgBytesPerSec = C.DWORD(44100)
	wfx.nBlockAlign = C.WORD(2)
	wfx.wBitsPerSample = C.WORD(8)

	res := C.waveOutOpen(&hwaveout, C.WAVE_MAPPER, &wfx, dwCallback, dwCallbackInstance, fdwOpen)
	if res != C.MMSYSERR_NOERROR {
		fmt.Fprintln(os.Stderr, "Error: waveOutOpen:", winmmErrorText(res))
		os.Exit(1)
	}
}

func initSoundDevice() {
}

func playback(buf1 []byte, buf2 []byte, notes string) {
	var buf [2]byte
	var bufChannel bytes.Buffer
	for i, bar := range buf1 {
		buf[0] = bar
		buf[1] = buf2[i]
		bufChannel.Write(buf[:])
	}
	bufWave := bufChannel.Bytes()

	var wavehdr C.WAVEHDR

	wdrsize := C.UINT(unsafe.Sizeof(wavehdr))

	wavehdr.lpData = C.LPSTR(unsafe.Pointer(&bufWave[0]))
	wavehdr.dwBufferLength = C.DWORD(len(bufWave))

	res := C.waveOutPrepareHeader(hwaveout, &wavehdr, wdrsize)
	if res != C.MMSYSERR_NOERROR {
		fmt.Fprintln(os.Stderr, "Error: waveOutPrepareHeader:", winmmErrorText(res))
		os.Exit(1)
	}

	res = C.waveOutWrite(hwaveout, &wavehdr, wdrsize)
	if res != C.MMSYSERR_NOERROR {
		fmt.Fprintln(os.Stderr, "Error: waveOutWrite:", winmmErrorText(res))
	}

	if wavehdrLast != nil {
		wdrsize := C.UINT(unsafe.Sizeof(*wavehdrLast))
		for wavehdrLast.dwFlags&C.WHDR_DONE == 0 {
			// still playing last buffer
			time.Sleep(time.Millisecond)
		}
		res = C.waveOutUnprepareHeader(hwaveout, wavehdrLast, wdrsize)
		if res != C.MMSYSERR_NOERROR {
			fmt.Fprintln(os.Stderr, "Error: waveOutUnprepareHeader:", winmmErrorText(res))
		}
	}

	if !*flagQuiet && len(notes) > 0 {
		fmt.Println(notes)
	}

	wavehdrLast = &wavehdr
}

func flushSoundBuffer() {
	if wavehdrLast != nil {
		wdrsize := C.UINT(unsafe.Sizeof(*wavehdrLast))
		for wavehdrLast.dwFlags&C.WHDR_DONE == 0 {
			// still playing last buffer
			time.Sleep(time.Millisecond)
		}
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
