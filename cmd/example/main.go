package main

// Usage example of beep library

import (
	"bufio"
	"flag"
	"fmt"
	"strings"

	"github.com/dbatbold/beep"
)

func main() {
	flag.Parse()
	output := flag.Arg(0)

	music := beep.NewMusic(output)
	volume := 100

	if len(output) > 0 {
		fmt.Println("Writing WAV to", output)
	} else {
		beep.PrintSheet = true
	}

	beep.OpenSoundDevice("default")
	beep.InitSoundDevice()
	defer beep.CloseSoundDevice()

	musicScore := `
        VP SA8 SR9
        A9HRDE cc DScszs|DEc DQzDE[|cc DScszs|DEc DQz DE[|vv DSvcsc|DEvs ]v|cc DScszs|VN
        A3HLDE [n z,    |cHRq HLz, |[n z,    |cHRq HLz,  |sl z,    |]m   pb|z, ]m    |
        
        A9HRDE cz [c|ss DSsz]z|DEs] ps|DSsz][ z][p|DEpDQ[ [|VN
        A3HLDE [n ov|]m [n    |  pb ic|  n,   lHRq|HLnc DQ[|
    `

	reader := bufio.NewReader(strings.NewReader(musicScore))
	go beep.PlayMusicNotes(music, reader, volume)
	music.Wait()
	beep.FlushSoundBuffer()
}
