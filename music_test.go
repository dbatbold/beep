package beep

import (
	"fmt"
)

func ExampleNote_measure_tempo_3() {
	note := &Note{
		key:       3000 + 'q',
		volume:    100,
		amplitude: 9,
		duration:  'W',
		dotted:    false,
		tempo:     3,
		samples:   0,
	}
	note.measure()
	fmt.Println("Temp 3:", note.samples)

	// samples in a whole note: 90112
	// samples in tempo level 3: 90112 + 4% = 93716

	// Output:
	// Temp 3: 93716
}

func ExampleNote_measure_tempo_9() {
	note := &Note{
		key:       3000 + 'q',
		volume:    100,
		amplitude: 9,
		duration:  'Q',
		dotted:    false,
		tempo:     9,
		samples:   0,
	}
	note.measure()
	fmt.Println("Temp 9:", note.samples)

	// samples in quarter note: 22528
	// samples in tempo level 9: 22528 - 20% = 18023

	// Output:
	// Temp 9: 18023
}

func Example_restNote_tempo_6() {
	buf := restNote('W', false, 6)
	fmt.Println("Temp 6:", len(buf))

	// Output:
	// Temp 6: 82904
}

func Example_restNote_tempo_0() {
	buf := restNote('Q', false, 0)
	fmt.Println("Temp 0:", len(buf))

	// Output:
	// Temp 0: 26132
}
