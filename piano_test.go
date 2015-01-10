package main

import (
	"fmt"
	"os"
	"testing"
	//"strings"
	"bytes"
)

// channel:
// byte 0,1 - left audio
// byte 2,3 - right audio
func TestFormatPianoSample(t *testing.T) {
	var header WaveHeader
	var bufWave bytes.Buffer
	opt := os.O_RDONLY
	sampleFile, err := os.OpenFile("piano.wav", opt, 0644)
	if err != nil {
		t.Error(err)
	}
	defer sampleFile.Close()
	n, err := header.ReadHeader(sampleFile)
	if err != nil {
		t.Error(err)
	}
	var buf [1024 * 10]byte
	for {
		n, err = sampleFile.Read(buf[:])
		if err != nil {
			break
		}
		bufWave.Write(buf[:n])
	}
	fmt.Println(header.String())
	//t.Log("SIZE", bufWave.Len())

	var left16 int16
	var right16 int16
	width := 76
	for i, bar := range bufWave.Bytes() {
		switch i % 4 {
		case 0:
			if i%width == 0 {
				//fmt.Print("\"")
			}
			left16 = int16(bar)
		case 1:
			left16 += int16(bar) << 8
		case 2:
			right16 = int16(bar)
		case 3:
			right16 += int16(bar) << 8
			fmt.Print(left16, " ")
			//fmt.Printf("|%s\n", strings.Repeat("-", 40+int(left16/900+1)))
			//fmt.Printf("%.4x", 0xffff/2+int(left16))
			//if i % width == width - 1 {
			//	fmt.Println("\",")
			//}
		}
		if i > 50 {
			fmt.Println()
			var piano Piano
			piano.GenerateNote(0)
			break
		}
	}
}
