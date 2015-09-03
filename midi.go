// MIDI file parser

package main

import (
	"errors"
	"fmt"
	"io/ioutil"
	"strings"
)

type MidiChunk struct {
	Type string
	Size int
	Data []byte
}

type Midi struct {
	Chunk   []*MidiChunk
	Format  int
	Ntracks int
	TickDiv int
}

const (
	MidiEventTypeMidi int = iota
	MidiEventTypeSysEx
	MidiEventTypeMeta
)

var midiNoteMap = map[byte]string{
	21:  "H0,",
	22:  "H0l",
	23:  "H0.",
	24:  "HLq",
	25:  "HL2",
	26:  "HLw",
	27:  "HL3",
	28:  "HLe",
	29:  "HLr",
	30:  "HL5",
	31:  "HLt",
	32:  "HL6",
	33:  "HLy",
	34:  "HL7",
	35:  "HLu",
	36:  "HLi",
	37:  "HL9",
	38:  "HLo",
	39:  "HL0",
	40:  "HLp",
	41:  "HL[",
	42:  "HL=",
	43:  "HL]",
	44:  "HLa",
	45:  "HLz",
	46:  "HLs",
	47:  "HLx",
	48:  "HLc",
	49:  "HLf",
	50:  "HLv",
	51:  "HLg",
	52:  "HLb",
	53:  "HLn",
	54:  "HLj",
	55:  "HLm",
	56:  "HLk",
	57:  "HL,",
	58:  "HLl",
	59:  "HL.",
	60:  "HRq",
	61:  "HR2",
	62:  "HRw",
	63:  "HR3",
	64:  "HRe",
	65:  "HRr",
	66:  "HR5",
	67:  "HRt",
	68:  "HR6",
	69:  "HRy",
	70:  "HR7",
	71:  "HRu",
	72:  "HRi",
	73:  "HR9",
	74:  "HRo",
	75:  "HR0",
	76:  "HRp",
	77:  "HR[",
	78:  "HR=",
	79:  "HR]",
	80:  "HRa",
	81:  "HRz",
	82:  "HRs",
	83:  "HRx",
	84:  "HRc",
	85:  "HRf",
	86:  "HRv",
	87:  "HRg",
	88:  "HRb",
	89:  "HRn",
	90:  "HRj",
	91:  "HRm",
	92:  "HRk",
	93:  "HR,",
	94:  "HRl",
	95:  "HR.",
	96:  "H7q",
	97:  "H72",
	98:  "H7w",
	99:  "H73",
	100: "H7e",
	101: "H7r",
	102: "H75",
	103: "H7t",
	104: "H76",
	105: "H7y",
	106: "H77",
	107: "H7u",
	108: "H7i",
}

var midiOctave string
var midiDuration string
var midiNoteCount int

func ParseMidiNote(filename string, printKeyboard bool) (*Midi, error) {
	midi := &Midi{}
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	dataSize := len(data)
	if dataSize < 8 {
		return nil, errors.New("Invalid MIDI file")
	}
	var pos int
	midi.Chunk = make([]*MidiChunk, 0)
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
		midi.Chunk = append(midi.Chunk, chunk)
		pos += 8 + size
		if pos >= dataSize {
			break
		}
	}

	var keys [88]byte
	var piano [88]byte
	var notes = "CcDEeFfGgAaB"
	var byteSize int
	var deltaTime int32
	var msgLength int32
	var lastKey byte
	piano[0] = 'A'
	piano[1] = 'a'
	piano[2] = 'B'
	for i, _ := range piano {
		if i > 2 {
			piano[i] = byte(notes[(i-3)%12])
		}
	}
	printKeys := func() {
		for i, _ := range piano {
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
	for _, chunk := range midi.Chunk {
		fmt.Printf("================ Chunk=%s Length=%d\n", chunk.Type, chunk.Size)
		switch chunk.Type {
		case "MThd": // header chunk
			midi.Format = int(chunk.Data[0])<<8 + int(chunk.Data[1])
			midi.Ntracks = int(chunk.Data[2])<<8 + int(chunk.Data[3])
			midi.TickDiv = int(chunk.Data[4])<<8 + int(chunk.Data[5])
			fmt.Printf("Format: %d, Ntracks: %d, TickDiv: %d\n", midi.Format, midi.Ntracks, midi.TickDiv)

		case "MTrk": // track chunk
			var isEvent bool
			chunkSize := len(chunk.Data)
			for i := 0; i < chunkSize; i++ {
				cbyte := chunk.Data[i]
				if isEvent {
					isEvent = false

					// MIDI event
					statusByte := cbyte & 0xF0
					channel := cbyte & 0x0F
					_ = channel
					switch statusByte {
					case 0x80: // note off
						note := chunk.Data[i+1]
						keys[note-21] = ' '
						//velocity := chunk.Data[i+2]
						//fmt.Printf("Note off: delta=%d channel=%02X note=%02X velocity=%02X\n", deltaTime, channel, note, velocity)
						i += 2
						continue

					case 0x90: // note on
						note := chunk.Data[i+1]
						keys[lastKey] = ' '
						keys[note-21] = midiNoteMap[note][2]
						lastKey = note - 21
						//velocity := chunk.Data[i+2]
						//fmt.Printf("Note on: delta=%d channel=%02X note=%02X velocity=%02X\n", deltaTime, channel, note, velocity)
						if printKeyboard {
							printKeys()
						} else {
							midi.printBeepNotation(int(deltaTime), note)
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
						i += byteSize
						//fmt.Printf("message=%s length=%d\n", message, msgLength)
						continue

					case 0xFF: // Meta event
						msgType := chunk.Data[i+1]
						msgLength, byteSize = midi.variableLengthValue(chunk.Data[i+2:])
						msg := i + 2 + byteSize
						message := chunk.Data[msg : msg+int(msgLength)]
						_ = message
						i += byteSize
						switch msgType {
						case 0x01: // Text
							//fmt.Printf("Message=%s length=%d\n", message, msgLength)
						}
						//fmt.Printf("Meta event: type=%02X\n", msgType)
						continue

					default:
						//fmt.Printf("Unknown event: code=%02X\n", cbyte)
					}

				} else {
					isEvent = true
					deltaTime, byteSize = midi.variableLengthValue(chunk.Data[i:])
					i += byteSize - 1
					//fmt.Printf("delta-time=%d\n", deltaTime)
				}
			}
		}
	}
	return midi, nil
}

func (m *Midi) printBeepNotation(deltaTime int, noteCode byte) {
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
func (m *Midi) variableLengthValue(data []byte) (value int32, byteSize int) {

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
