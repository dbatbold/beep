package beep

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
)

// MidiChunk - MIDI chunk
type MidiChunk struct {
	Type string
	Size int
	Data []byte
}

// Midi - MIDI file
type Midi struct {
	Chunks     []*MidiChunk
	Tracks     []*MidiChunk // voice tracks
	Format     int
	Ntracks    int // number of tracks
	TickDiv    int // if 15th bit is 0 - (h.m.s.frames) resolution of a quarter note, 1 - metric (bar.beat)
	Playing    bool
	OutputFile *os.File
	OutputBuf  []int16

	music *Music
}

// MIDI events
const (
	MidiEventMidi = iota
	MidiEventSysEx
	MidiEventMeta
)

var (
	trackNum int
)

// MidiEvent - MIDI event
type MidiEvent struct {
	Type       int
	Delta      int
	Start      int
	Note       *Note // beep note
	NoteNumber byte  // MIDI note number
}

// CalcDuration calculates duration for ticks
func (m *MidiEvent) CalcDuration(duration int, tickDiv int) {
	switch {
	case duration <= tickDiv/16:
		m.Note.duration = 'I'
	case duration <= tickDiv/8:
		m.Note.duration = 'T'
	case duration <= tickDiv/4:
		m.Note.duration = 'S'
	case duration <= tickDiv/2:
		m.Note.duration = 'E'
	case duration <= tickDiv:
		m.Note.duration = 'Q'
	case duration <= tickDiv*2:
		m.Note.duration = 'H'
	default:
		m.Note.duration = 'W'
	}
}

// Map between MIDI note number and beep notation
var midiNoteMap = map[byte]string{
	21: "H0,", 22: "H0l", 23: "H0.",
	24: "HLq", 25: "HL2", 26: "HLw", 27: "HL3", 28: "HLe", 29: "HLr", 30: "HL5", 31: "HLt", 32: "HL6",
	33: "HLy", 34: "HL7", 35: "HLu", 36: "HLi", 37: "HL9", 38: "HLo", 39: "HL0", 40: "HLp", 41: "HL[",
	42: "HL=", 43: "HL]", 44: "HLa", 45: "HLz", 46: "HLs", 47: "HLx", 48: "HLc", 49: "HLf", 50: "HLv",
	51: "HLg", 52: "HLb", 53: "HLn", 54: "HLj", 55: "HLm", 56: "HLk", 57: "HL,", 58: "HLl", 59: "HL.",
	60: "HRq", 61: "HR2", 62: "HRw", 63: "HR3", 64: "HRe", 65: "HRr", 66: "HR5", 67: "HRt", 68: "HR6",
	69: "HRy", 70: "HR7", 71: "HRu", 72: "HRi", 73: "HR9", 74: "HRo", 75: "HR0", 76: "HRp", 77: "HR[",
	78: "HR=", 79: "HR]", 80: "HRa", 81: "HRz", 82: "HRs", 83: "HRx", 84: "HRc", 85: "HRf", 86: "HRv",
	87: "HRg", 88: "HRb", 89: "HRn", 90: "HRj", 91: "HRm", 92: "HRk", 93: "HR,", 94: "HRl", 95: "HR.",
	96: "H7q", 97: "H72", 98: "H7w", 99: "H73", 100: "H7e", 101: "H7r", 102: "H75", 103: "H7t",
	104: "H76", 105: "H7y", 106: "H77", 107: "H7u", 108: "H7i",
}

var (
	midiOctave       string
	midiNoteCount    int
	midiSaveWaveFile = false
	midiNoteOnMap    = make(map[byte]*MidiEvent)
)

// ParseMidi parses MIDI file
func ParseMidi(music *Music, filename string, printKeyboard bool) (*Midi, error) {
	midi := &Midi{
		music: music,
	}
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	dataSize := len(data)
	if dataSize < 8 {
		return nil, errors.New("Invalid MIDI file")
	}
	var pos int
	midi.Chunks = make([]*MidiChunk, 0)
	for {
		var chunkType []byte
		var chunkSize []byte
		var chunkData []byte
		chunkType = data[pos : pos+4]
		chunkSize = data[pos+4 : pos+8]
		size := int(chunkSize[0])<<24 + int(chunkSize[1])<<16 + int(chunkSize[2])<<8 + int(chunkSize[3])
		chunkData = data[pos+8 : pos+9+size]
		chunk := &MidiChunk{
			Type: fmt.Sprintf("%s", chunkType),
			Size: size,
			Data: chunkData,
		}
		midi.Chunks = append(midi.Chunks, chunk)
		pos += 8 + size
		if pos >= dataSize {
			break
		}
	}

	var (
		keys      [88]byte
		piano     [88]byte
		notes     = "CcDEeFfGgAaB"
		byteSize  int
		deltaTime int32
		msgLength int32
		lastKey   byte
		note      byte
		velocity  byte
	)

	piano[0] = 'A'
	piano[1] = 'a'
	piano[2] = 'B'

	for i := range piano {
		if i > 2 {
			piano[i] = byte(notes[(i-3)%12])
		}
	}
	printKeys := func() {
		for i := range piano {
			fmt.Print(string(piano[i]))
			if i == 2 || (i-3)%12 == 11 {
				fmt.Print(" ")
			}
		}
		fmt.Println()
		for i, k := range keys {
			if k == 0 {
				k = ' '
			}
			fmt.Print(string(k))
			if i == 2 || (i-3)%12 == 11 {
				fmt.Print(" ")
			}
		}
		fmt.Println()
	}
	for _, chunk := range midi.Chunks {
		//fmt.Printf("================ Chunk=%s Length=%d\n", chunk.Type, chunk.Size)
		switch chunk.Type {
		case "MThd": // header chunk
			midi.Format = int(chunk.Data[0])<<8 + int(chunk.Data[1])
			midi.Ntracks = int(chunk.Data[2])<<8 + int(chunk.Data[3])
			midi.TickDiv = int(chunk.Data[4])<<8 + int(chunk.Data[5])
			//fmt.Printf("Format: %d, Ntracks: %d, TickDiv: %d\n", midi.Format, midi.Ntracks, midi.TickDiv)

		case "MTrk": // track chunk
			midi.Tracks = append(midi.Tracks, chunk)
			var isEvent = true
			chunkSize := len(chunk.Data)
			for i := 0; i < chunkSize; i++ {
				cbyte := chunk.Data[i]
				isEvent = !isEvent
				if !isEvent {
					deltaTime, byteSize = midi.variableLengthValue(chunk.Data[i:])
					_ = deltaTime
					i += byteSize - 1
					//fmt.Printf("delta-time=%d\n", deltaTime)
					continue
				}

				// MIDI event
				statusByte := cbyte & 0xF0
				runningStatus := statusByte&0x80 == 0
				if runningStatus {
					//fmt.Printf("runningStatus %02X\n", statusByte)
				}
				channel := cbyte & 0x0F
				_ = channel

				switch statusByte {
				case 0x80: // note off
					if runningStatus {
						note = chunk.Data[i]
						i++
					} else {
						note = chunk.Data[i+1]
						i += 2
					}
					if note >= 21 && note <= 108 {
						keys[note-21] = ' '
					}
					//velocity := chunk.Data[i+2]
					//fmt.Printf("Note off: delta=%d channel=%02X note=%02X velocity=%02X\n", deltaTime, channel, note, velocity)
					i += 2
					continue

				case 0x90: // note on
					if runningStatus {
						note = chunk.Data[i]
						velocity = chunk.Data[i+1]
						i++
					} else {
						note = chunk.Data[i+1]
						velocity = chunk.Data[i+2]
						i += 2
					}
					if lastKey >= 21 && lastKey <= 108 {
						keys[lastKey-21] = ' '
					}
					if note >= 21 && note <= 108 {
						keys[note-21] = midiNoteMap[note][2]
					}
					lastKey = note - 21
					//fmt.Printf("Note on: delta=%d channel=%02X note=%02X velocity=%02X\n", deltaTime, channel, note, velocity)
					if printKeyboard && velocity > 0 {
						printKeys()
					} else {
						//midi.printBeepNotation(int(deltaTime), note)
					}
					i += 2
					continue

				case 0xA0: // aftertouch
					i += 2
					//fmt.Printf("Aftertouch: channel=%02X note=%02X velocity=%02X\n", channel, msg[1], msg[2])
					continue

				case 0xB0: // control change
					i += 2
					//program := chunk.Data[i+1]
					//fmt.Printf("Control change: channel=%02X program=%02X\n", channel, program)
					continue
				}

				// SysEx events
				switch cbyte {
				case 0xF0: // SysEx
					msgLength, byteSize = midi.variableLengthValue(chunk.Data[i+1:])
					msg := i + 1 + byteSize
					message := chunk.Data[msg : msg+int(msgLength)]
					_ = message
					i += byteSize + int(msgLength)
					//fmt.Printf("message=%s length=%d\n", message, msgLength)
					continue

				case 0xFF: // Meta event
					msgType := chunk.Data[i+1]
					msgLength, byteSize = midi.variableLengthValue(chunk.Data[i+2:])
					msg := i + 2 + byteSize
					message := chunk.Data[msg : msg+int(msgLength)]
					_ = message
					i += byteSize + int(msgLength)
					switch msgType {
					case 0x01: // Text
						//fmt.Printf("Message=%s length=%d\n", message, msgLength)
					}
					//fmt.Printf("Meta event: type=%02X\n", msgType)
					continue

				default:
					//fmt.Printf("Unknown event: code=%02X\n", cbyte)
				}
			}
		}
	}
	return midi, nil
}

func (midi *Midi) printBeepNotation(deltaTime int, noteCode byte) {
	note, ok := midiNoteMap[noteCode]
	if !ok {
		fmt.Println("\n# invalid MIDI note code: ", noteCode)
		return
	}
	if strings.HasPrefix(note, midiOctave) {
		fmt.Print(note[2:])
	} else {
		fmt.Print(note)
	}
	midiOctave = note[:2]
	midiNoteCount++
	if midiNoteCount == 50 {
		fmt.Println()
		midiNoteCount = 0
	}
}

// Variable length value parser
func (midi *Midi) variableLengthValue(data []byte) (value int32, byteSize int) {

	// byte1    byte2    byte3    byte4
	// 00000000 00000000 00000000 00000000
	// |        |        |        |
	// marker   marker   marker   marker   (marker bit = 1 - include next byte)
	//  0000000  0000000  0000000  0000000
	//  7.....1  7.....1  7.....1  7.....1
	//        0000000000000000000000000000
	//        28.........................1 (max value 7x4 = 28 bit)

	byteSize = 1
	for _, d := range data {
		if d&0x80 == 0x80 {
			byteSize++
		} else {
			break
		}
	}
	if byteSize > 4 {
		// invalid encoding
		return 0, 4
	}
	switch byteSize {
	case 1:
		value = int32(data[0])
	case 2:
		value = int32(data[0]&0x7F)<<7 + int32(data[1])
	case 3:
		value = int32(data[0]&0x7F)<<14 + int32(data[1]&0x7F)<<7 + int32(data[2])
	case 4:
		value = int32(data[0]&0x7F)<<21 + int32(data[1]&0x7F)<<14 + int32(data[2]&0x7F)<<7 + int32(data[3])
	}
	return value, byteSize
}

func (midi *Midi) playTracks() {
	if midi.Playing {
		midi.music.WaitLine()
	}
	go midi.music.Playback(midi.OutputBuf, midi.OutputBuf)
	midi.Playing = true
}

func (midi *Midi) mixTracks(events []*MidiEvent) {
	var bufsize int
	//var count = len(events)
	var voice Voice = midi.music.piano

	sustain := &Sustain{
		attack:  8,
		decay:   4,
		sustain: 4,
		release: 9,
		buf:     make([]int16, quarterNote),
	}

	for _, event := range events {
		bufsize += event.Delta
		if event.Note.velocity > 0 {
			if event.Note.duration == 0 {
				event.Note.duration = 'E'
			}
			event.Note.volume = int(float32(SampleAmp16bit) * (float32(event.Note.velocity) / 127))
			if voice.GetNote(event.Note, sustain) {
				voice.SustainNote(event.Note, sustain)
			} else {
				fmt.Println("Invalid note:", event.Note.key)
			}
		}
	}

	bufWave := make([]int16, bufsize)
	var start int
	for _, event := range events {
		start += event.Delta
		if event.Note.velocity > 0 {
			mixSoundWave(bufWave[start:], event.Note.buf)
		}
	}

	if trackNum++; trackNum > 1 {
		// play other tracks with computer voice
		midi.music.piano.ComputerVoice(true)
	}
	if midi.OutputBuf == nil {
		// first track
		midi.OutputBuf = bufWave
	} else {
		mixSoundWave(midi.OutputBuf, bufWave)
	}
}

// Play all MIDI tracks at same time
func (midi *Midi) Play() {
	if midi.TickDiv < 0 {
		fmt.Println("Metric TickDiv is not supported.")
		return
	}
	if midi.music.piano == nil {
		midi.music.piano = NewPiano()
	}
	var (
		tickDiv    = midi.TickDiv
		events     []*MidiEvent
		handLevel  rune
		deltaTime  int32
		byteSize   int
		event      *MidiEvent
		lastStatus byte
		noteNumber byte
		velocity   byte
		msgLength  int32
		timer      int
	)

	fmt.Println("TickDiv:", tickDiv)
	fmt.Println("Tracks:", len(midi.Tracks))
	fmt.Println("Format:", midi.Format)

	if len(midi.music.output) > 0 {
		fmt.Print("Saving ... ")
	}

	for _, chunk := range midi.Tracks {
		chunkSize := len(chunk.Data)
		var isEvent = true
		//fmt.Println("\nCHUNK")
		for i := 0; i < chunkSize; i++ {
			//fmt.Printf("%d:%02X ", i, chunk.Data[i])
			isEvent = !isEvent
			if !isEvent {
				deltaTime, byteSize = midi.variableLengthValue(chunk.Data[i:])
				i += byteSize - 1
				timer += int(deltaTime)
				//fmt.Print(deltaTime, " ")
				continue
			}
			statusByte := chunk.Data[i] & 0xF0
			runningStatus := statusByte&0x80 == 0
			if runningStatus {
				statusByte = lastStatus
				//fmt.Printf("runningStatus %02X\n", statusByte)
			}

			switch statusByte {
			case 0x80: // Note Off message
				if runningStatus {
					noteNumber = chunk.Data[i]
					i++
				} else {
					noteNumber = chunk.Data[i+1]
					i += 2
				}
				noteName := midiNoteMap[noteNumber]
				if len(noteName) == 0 {
					continue
				}
				noteOnEvent := midiNoteOnMap[noteNumber]
				if noteOnEvent != nil {
					noteDuration := timer - noteOnEvent.Start
					noteOnEvent.CalcDuration(noteDuration, tickDiv)
					delete(midiNoteOnMap, noteNumber)
				}
				hand := noteName[1]
				key := rune(noteName[2])

				switch hand {
				case '0': // octave 0
					handLevel = 1000
				case 'L': // octave 1, 2, 3
					handLevel = 2000
				case 'R': // octave 4, 5, 6
					handLevel = 3000
				case '7', '8': // octave 7, 8
					handLevel = 4000
				default:
					fmt.Println("Invalid hand level:", handLevel)
				}
				delta := quarterNote / tickDiv * int(deltaTime)

				note := &Note{
					key:       handLevel + key,
					volume:    int(SampleAmp16bit) / 4 * 3,
					amplitude: 9,
					duration:  0,
					dotted:    false,
					tempo:     4,
					velocity:  0,
				}

				event = &MidiEvent{
					Type:       MidiEventMidi,
					Delta:      delta,
					Note:       note,
					NoteNumber: noteNumber,
				}
				events = append(events, event)
				//fmt.Printf("Note Off: %d %s\n", deltaTime, noteName)

			case 0x90: // Note On message
				//fmt.Printf("\ni=%d\n", i)
				if runningStatus {
					noteNumber = chunk.Data[i]
					velocity = chunk.Data[i+1]
					i++
				} else {
					noteNumber = chunk.Data[i+1]
					velocity = chunk.Data[i+2]
					i += 2
				}
				noteName, found := midiNoteMap[noteNumber]
				if !found {
					fmt.Println("Invalid note name:", noteNumber)
					continue
				}
				if velocity == 0 {
					// NoteOn event with 0 velocity is same as NoteOff event. This can determine the note duration.
					noteOnEvent := midiNoteOnMap[noteNumber]
					if noteOnEvent != nil {
						noteDuration := timer - noteOnEvent.Start
						noteOnEvent.CalcDuration(noteDuration, tickDiv)
						delete(midiNoteOnMap, noteNumber)
					}
				}
				hand := noteName[1]
				key := rune(noteName[2])

				switch hand {
				case '0': // octave 0
					handLevel = 1000
				case 'L': // octave 1, 2, 3
					handLevel = 2000
				case 'R': // octave 4, 5, 6
					handLevel = 3000
				case '7', '8': // octave 7, 8
					handLevel = 4000
				default:
					fmt.Println("Invalid hand level:", handLevel)
				}

				note := &Note{
					key:       handLevel + key,
					volume:    int(SampleAmp16bit) / 4 * 3,
					amplitude: 9,
					duration:  0,
					dotted:    false,
					tempo:     4,
					velocity:  int(velocity),
				}

				delta := quarterNote / tickDiv * int(deltaTime)
				event = &MidiEvent{
					Type:       MidiEventMidi,
					Delta:      delta,
					Start:      timer,
					Note:       note,
					NoteNumber: noteNumber,
				}
				//fmt.Printf("Note On: %d %s %s v=%d\n", deltaTime, noteName, string(event.Note.duration), velocity)
				midiNoteOnMap[noteNumber] = event
				events = append(events, event)

			case 0xA0: // aftertouch
				i += 2
				//fmt.Printf("Aftertouch: channel=%02X note=%02X velocity=%02X\n", channel, msg[1], msg[2])

			case 0xB0: // control change
				i += 2
				//program := chunk.Data[i+1]
				//fmt.Printf("Control change: channel=%02X program=%02X\n", channel, program)

			case 0xC0: // program change
				i++

			case 0xD0: // channel pressure
				i++

			case 0xE0: // pitch change
				i += 2
			}

			// SysEx events
			switch chunk.Data[i] {
			case 0xF0: // SysEx
				msgLength, byteSize = midi.variableLengthValue(chunk.Data[i+1:])
				msg := i + 1 + byteSize
				message := chunk.Data[msg : msg+int(msgLength)]
				_ = message
				i += byteSize + int(msgLength)
				//fmt.Printf("\nmessage=%s length=%d\n", message, msgLength)

			case 0xF7: // Escape sequences
				msgLength, byteSize = midi.variableLengthValue(chunk.Data[i+1:])
				msg := i + 1 + byteSize
				message := chunk.Data[msg : msg+int(msgLength)]
				_ = message
				i += byteSize + int(msgLength)
				//fmt.Printf("\nF7:message=%s length=%d\n", message, msgLength)

			case 0xFF: // Meta event
				msgType := chunk.Data[i+1]
				msgLength, byteSize = midi.variableLengthValue(chunk.Data[i+2:])
				msg := i + 2 + byteSize
				message := chunk.Data[msg : msg+int(msgLength)]
				_ = message
				i += 1 + byteSize + int(msgLength)
				switch msgType {
				case 0x01: // Text
					fmt.Println("Message:", string(message))
				}
				//fmt.Printf("\nMeta event: type=%02X length=%d\n", msgType, msgLength)

			default:
				//fmt.Printf("Unknown event: code=%02X\n", cbyte)
			}

			lastStatus = statusByte
		}

		if events != nil {
			midi.mixTracks(events)
			events = nil
		}
	}

	if len(midi.music.output) == 0 {
		midi.playTracks()
		midi.music.WaitLine()
	}

	if len(midi.music.output) > 0 {
		// save to WAVE file
		var err error
		opt := os.O_WRONLY | os.O_TRUNC | os.O_CREATE
		midi.OutputFile, err = os.OpenFile(midi.music.output, opt, 0644)
		if err != nil {
			fmt.Fprintln(os.Stderr, "Error opening to output file:", err)
			os.Exit(1)
		}
		defer midi.OutputFile.Close()
		buf := make([]int16, 2*len(midi.OutputBuf))
		for i, bar := range midi.OutputBuf {
			// stereo buffer
			buf[i*2] = bar
			buf[i*2+1] = bar
		}
		buf16 := int16ToByteBuf(buf)
		header := NewWaveHeader(2, SampleRate, 16, len(buf16))
		_, err = header.WriteHeader(midi.OutputFile)
		if err != nil {
			fmt.Fprintln(os.Stderr, "Error writing to output file:", err)
			os.Exit(1)
		}
		_, err = midi.OutputFile.Write(buf16)
		if err != nil {
			fmt.Fprintln(os.Stderr, "Error writing to output file:", err)
			os.Exit(1)
		}
		fmt.Println(len(buf16), "bytes to", midi.music.output)
	}
}
