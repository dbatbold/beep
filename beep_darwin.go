//go:build darwin
// +build darwin

package beep

/*
#cgo LDFLAGS: -framework CoreAudio -framework AudioToolbox -framework IOKit -framework CoreFoundation
#include <AudioToolbox/AudioToolbox.h>
#include <IOKit/ps/IOPowerSources.h>
#include <IOKit/ps/IOPSKeys.h>

typedef struct {
    AudioQueueRef queue;
    AudioQueueBufferRef buffers[3];
    int currentBuffer;
} AudioContext;

static AudioContext audioCtx = {0};

void audioCallback(void *userData, AudioQueueRef queue, AudioQueueBufferRef buffer) {
    // Buffer has been played, do nothing
}
*/
import "C"

import (
	"fmt"
	"io"
	"os"
	"os/user"
	"unsafe"
)

var (
	audioQueue C.AudioQueueRef
)

// OpenSoundDevice opens hardware sound device
func OpenSoundDevice(device string) error {
	var format C.AudioStreamBasicDescription
	format.mSampleRate = C.Float64(SampleRate)
	format.mFormatID = C.kAudioFormatLinearPCM
	format.mFormatFlags = C.kLinearPCMFormatFlagIsSignedInteger | C.kLinearPCMFormatFlagIsPacked
	format.mBytesPerPacket = 4
	format.mFramesPerPacket = 1
	format.mBytesPerFrame = 4
	format.mChannelsPerFrame = 2
	format.mBitsPerChannel = 16
	format.mReserved = 0

	status := C.AudioQueueNewOutput(
		&format,
		C.AudioQueueOutputCallback(C.audioCallback),
		nil,
		nil,
		nil,
		0,
		&audioQueue,
	)

	if status != 0 {
		err := fmt.Errorf("AudioQueueNewOutput failed with status: %d", status)
		fmt.Fprintln(os.Stderr, err)
		return err
	}

	return nil
}

// InitSoundDevice initialize sound device
func InitSoundDevice() error {
	if audioQueue == nil {
		return fmt.Errorf("audio queue not initialized")
	}

	// Start the audio queue
	status := C.AudioQueueStart(audioQueue, nil)
	if status != 0 {
		err := fmt.Errorf("AudioQueueStart failed with status: %d", status)
		fmt.Fprintln(os.Stderr, err)
		return err
	}

	return nil
}

// Playback sends stereo wave buffer to sound device
func (m *Music) Playback(buf1, buf2 []int16) {
	if audioQueue == nil {
		m.linePlayed <- true
		return
	}

	// Interleave the stereo buffers
	bufWave := make([]int16, len(buf1)*2)
	for i := range buf1 {
		bufWave[i*2] = buf1[i]
		bufWave[i*2+1] = buf2[i]
	}

	// Allocate an audio queue buffer
	var buffer C.AudioQueueBufferRef
	bufferSize := C.UInt32(len(bufWave) * 2)
	status := C.AudioQueueAllocateBuffer(audioQueue, bufferSize, &buffer)
	if status != 0 {
		fmt.Fprintf(os.Stderr, "AudioQueueAllocateBuffer failed: %d\n", status)
		m.linePlayed <- true
		return
	}

	// Copy data to the buffer
	buffer.mAudioDataByteSize = bufferSize
	C.memcpy(buffer.mAudioData, unsafe.Pointer(&bufWave[0]), C.size_t(bufferSize))

	// Enqueue the buffer
	if !m.stopping {
		status = C.AudioQueueEnqueueBuffer(audioQueue, buffer, 0, nil)
		if status != 0 {
			fmt.Fprintf(os.Stderr, "AudioQueueEnqueueBuffer failed: %d\n", status)
			C.AudioQueueFreeBuffer(audioQueue, buffer)
			m.linePlayed <- true
			return
		}
	} else {
		C.AudioQueueFreeBuffer(audioQueue, buffer)
	}

	m.linePlayed <- true // notify that playback is done
}

// FlushSoundBuffer flushes sound buffer
func FlushSoundBuffer() {
	if audioQueue != nil {
		C.AudioQueueFlush(audioQueue)
	}
}

// StopPlayBack stops play back
func StopPlayBack() {
	if audioQueue != nil {
		C.AudioQueueStop(audioQueue, 1) // true = immediate stop
	}
}

// CloseSoundDevice closes sound device
func CloseSoundDevice() {
	if audioQueue != nil {
		C.AudioQueueStop(audioQueue, 1)
		C.AudioQueueDispose(audioQueue, 1)
		audioQueue = nil
	}
}

// SendBell send bell sound to console
func SendBell() {
	bell := []byte{7}
	os.Stdout.Write(bell)

	// On macOS, try to write to /dev/console or /dev/tty
	console, err := os.OpenFile("/dev/console", os.O_WRONLY, 0644)
	if err != nil {
		console, err = os.OpenFile("/dev/tty", os.O_WRONLY, 0644)
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
		home = "/Users"
	} else {
		home = usr.HomeDir
	}
	return home + "/.beep"
}

// BatteryLevel return battery charge level.
func BatteryLevel() (int, error) {
	// Get power source information
	blob := C.IOPSCopyPowerSourcesInfo()
	if blob == nil {
		return 0, io.EOF
	}
	defer C.CFRelease(C.CFTypeRef(blob))

	sources := C.IOPSCopyPowerSourcesList(blob)
	if sources == nil {
		return 0, io.EOF
	}
	defer C.CFRelease(C.CFTypeRef(sources))

	count := C.CFArrayGetCount(sources)
	if count == 0 {
		return 0, io.EOF
	}

	// Get the first power source (typically the battery)
	source := C.CFArrayGetValueAtIndex(sources, 0)
	description := C.IOPSGetPowerSourceDescription(blob, source)
	if description == nil {
		return 0, io.EOF
	}

	// Get current capacity
	key := C.CFStringRef(C.CFStringCreateWithCString(nil, C.kIOPSCurrentCapacityKey, C.kCFStringEncodingUTF8))
	defer C.CFRelease(C.CFTypeRef(key))

	value := C.CFDictionaryGetValue(description, unsafe.Pointer(key))
	if value == nil {
		return 0, io.EOF
	}

	var capacity C.int
	if C.CFNumberGetValue(C.CFNumberRef(value), C.kCFNumberIntType, unsafe.Pointer(&capacity)) {
		return int(capacity), nil
	}

	return 0, io.EOF
}
