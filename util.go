package main

import (
	"fmt"
	"io"
	"math"
	"strconv"
	"strings"
)

func stringToBytes(s string) []byte {
	var buf []byte
	for _, c := range s {
		buf = append(buf, byte(c))
	}
	return buf
}

func int32ToBytes(i int) []byte {
	var buf [4]byte
	n := int32(i)
	buf[0] = byte(n)
	buf[1] = byte(n >> 8)
	buf[2] = byte(n >> 16)
	buf[3] = byte(n >> 24)
	return buf[:]
}

func int16ToBytes(i int) []byte {
	n := int32(i)
	var buf [2]byte
	buf[0] = byte(n)
	buf[1] = byte(n >> 8)
	return buf[:]
}

func bytesToInt32(buf []byte) int32 {
	i := int(buf[3])<<24 + int(buf[2])<<16 + int(buf[1])<<8 + int(buf[0])
	return int32(i)
}

func bytesToInt16(buf []byte) int16 {
	i := int(buf[1])<<8 + int(buf[0])
	return int16(i)
}

// Converts []int16 to []byte
func int16ToByteBuf(buf16 []int16) []byte {
	buf := make([]byte, len(buf16)*2)
	for i, bar := range buf16 {
		buf[i*2] = byte(bar)
		buf[i*2+1] = byte(bar >> 8)
	}
	return buf
}

// Converts []byte to []int16
func byteToInt16Buf(buf []byte) []int16 {
	buf16 := make([]int16, len(buf)/2)
	for i := range buf16 {
		buf16[i] = int16(buf[i*2]) + int16(buf[i*2+1])<<8
	}
	return buf16
}

// Converts Hertz to frequency unit
func hertzToFreq(hertz float64) float64 {
	// 1 second = 44100 samples
	// 1 hertz = freq * 2Pi
	// freq = 2Pi / 44100 * hertz
	freq := 2.0 * math.Pi / sampleRate64 * hertz
	return freq
}

// WaveHeader - WAV file header
type WaveHeader struct {
	header        [44]byte
	ChunkID       string
	ChunkSize     int
	Format        string
	Subchunk1ID   string
	Subchunk1Size int
	AudioFormat   int
	NumChannels   int
	SampleRate    int
	ByteRate      int
	BlockAlign    int
	BitsPerSample int
	Subchunk2ID   string
	Subchunk2Size int
}

// NewWaveHeader returns new WAV header
func NewWaveHeader(channels, sampleRate, bitsPerSample, dataSize int) *WaveHeader {
	wh := &WaveHeader{
		ChunkID:       "RIFF",
		ChunkSize:     36 + dataSize,
		Format:        "WAVE",
		Subchunk1ID:   "fmt ",
		Subchunk1Size: 16,
		AudioFormat:   1,
		NumChannels:   channels,
		SampleRate:    sampleRate,
		ByteRate:      sampleRate * channels * (bitsPerSample / 8),
		BlockAlign:    channels * bitsPerSample / 8,
		BitsPerSample: bitsPerSample,
		Subchunk2ID:   "data",
		Subchunk2Size: dataSize,
	}
	return wh
}

// WriteHeader writer WAV header
func (w *WaveHeader) WriteHeader(writer io.Writer) (int, error) {
	copy(w.header[0:], stringToBytes(w.ChunkID[:4]))
	copy(w.header[4:], int32ToBytes(w.ChunkSize))
	copy(w.header[8:], stringToBytes(w.Format[:4]))
	copy(w.header[12:], stringToBytes(w.Subchunk1ID[:4]))
	copy(w.header[16:], int32ToBytes(w.Subchunk1Size))
	copy(w.header[20:], int16ToBytes(w.AudioFormat))
	copy(w.header[22:], int16ToBytes(w.NumChannels))
	copy(w.header[24:], int32ToBytes(w.SampleRate))
	copy(w.header[28:], int32ToBytes(w.ByteRate))
	copy(w.header[32:], int16ToBytes(w.BlockAlign))
	copy(w.header[34:], int16ToBytes(w.BitsPerSample))
	copy(w.header[36:], stringToBytes(w.Subchunk2ID[:4]))
	copy(w.header[40:], int32ToBytes(w.Subchunk2Size))
	return writer.Write(w.header[:])
}

// ReadHeader reads WAV header
func (w *WaveHeader) ReadHeader(reader io.Reader) (int, error) {
	n, err := reader.Read(w.header[:])
	if err != nil {
		return 0, err
	}
	w.ChunkID = string(w.header[:4])
	w.ChunkSize = int(bytesToInt32(w.header[4:8]))
	w.Format = string(w.header[8:12])
	w.Subchunk1ID = string(w.header[12:16])
	w.Subchunk1Size = int(bytesToInt32(w.header[16:20]))
	w.AudioFormat = int(bytesToInt16(w.header[20:22]))
	w.NumChannels = int(bytesToInt16(w.header[22:24]))
	w.SampleRate = int(bytesToInt32(w.header[24:28]))
	w.ByteRate = int(bytesToInt32(w.header[28:32]))
	w.BlockAlign = int(bytesToInt16(w.header[32:34]))
	w.BitsPerSample = int(bytesToInt16(w.header[34:36]))
	w.Subchunk2ID = string(w.header[36:40])
	w.Subchunk2Size = int(bytesToInt32(w.header[40:44]))
	return n, nil
}

func (w *WaveHeader) String() string {
	return fmt.Sprintf(`ChunkID: '%v'
ChunkSize: %v,
Format: '%v',
Subchunk1ID: '%v'
Subchunk1Size: %v
AudioFormat: %v
NumChannels: %v
SampleRate: %v
ByteRate: %v
BlockAlign: %v
BitsPerSample: %v
Subchunk2ID: '%v'
Subchunk2Size: %v
`,
		w.ChunkID,
		w.ChunkSize,
		w.Format,
		w.Subchunk1ID,
		w.Subchunk1Size,
		w.AudioFormat,
		w.NumChannels,
		w.SampleRate,
		w.ByteRate,
		w.BlockAlign,
		w.BitsPerSample,
		w.Subchunk2ID,
		w.Subchunk2Size,
	)
}

// Formats a number with commas: 1234567 to 1,234,567
func numberComma(number int64) string {
	var parts []int64
	n := number
	minus := n < 0
	if minus {
		n = -n
	}
	for n > 0 {
		parts = append(parts, n%1000)
		n /= 1000
	}
	var s []string
	last := len(parts) - 1
	for i := range parts {
		format := "%0.3d"
		if i == 0 {
			format = "%d"
			if minus {
				format = "-%d"
			}
		}
		s = append(s, fmt.Sprintf(format, parts[last-i]))
	}
	return strings.Join(s, ",")
}

// Converts string to number, returns 0 for error
func stringNumber(s string) int {
	n, err := strconv.Atoi(s)
	if err != nil {
		return 0
	}
	return n
}
