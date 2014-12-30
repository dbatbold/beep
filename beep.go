package main

import (
	"flag"
	"fmt"
	"os"
	"unsafe"
	"math"
	"time"
)

/*
#cgo LDFLAGS: -lasound

#include <alsa/asoundlib.h>
*/
import "C"

var (
	flagHelp  = flag.Bool("h", false, "help")
	flagCount = flag.Int("c", 1, "count")
)

func main() {
	var handle *C.snd_pcm_t
	var device = C.CString("default")

	flag.Parse()

	if *flagHelp {
		flag.PrintDefaults()
		return
	}

	code := C.snd_pcm_open(&handle, device, C.SND_PCM_STREAM_PLAYBACK, 0)
	if code < 0 {
		fmt.Println("snd_pcm_open:", C.GoString(C.snd_strerror(code)))
		os.Exit(1)
	}
	defer C.snd_pcm_close(handle)

	code = C.snd_pcm_set_params(
			handle,
			C.SND_PCM_FORMAT_U8,
			C.SND_PCM_ACCESS_RW_INTERLEAVED,
			1,
			48000,
			1,
			0)
	if code < 0 {
		fmt.Println("snd_pcm_set_params:", C.GoString(C.snd_strerror(code)))
		os.Exit(1)
	}

	buf := make([]byte, 1024*10)
	space := make([]byte, 1024*5)
	for i, _ := range buf[:] {
		buf[i] = byte(255 * math.Sin(float64(i)*0.055))
	}
	for i := 0; i < *flagCount; i++ {
		n := C.snd_pcm_writei(handle, unsafe.Pointer(&buf[0]), C.snd_pcm_uframes_t(len(buf)))
		if n < 0 {
			fmt.Println("snd_pcm_writei:", C.GoString(C.snd_strerror(code)))
			os.Exit(1)
		}
		C.snd_pcm_writei(handle, unsafe.Pointer(&space[0]), C.snd_pcm_uframes_t(len(space)))
	}
	time.Sleep(time.Millisecond * 200)
}
