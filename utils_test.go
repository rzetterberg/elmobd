package elmobd

import (
	"fmt"
	"testing"
)

/*==============================================================================
 * Benchmarks
 */

var resultBytesToUint64 uint64

func benchBytesToUint64(input []byte, b *testing.B) {
	var r uint64

	for n := 0; n < b.N; n++ {
		r, _ = BytesToUint64(input)
	}

	resultBytesToUint64 = r
}

func BenchmarkBytesToUint648(b *testing.B) {
	benchBytesToUint64(
		[]byte{0x54, 0x01, 0x12, 0xBC, 0xBE, 0xEF, 0x30, 0x4B},
		b,
	)
}

func BenchmarkBytesToUint644(b *testing.B) {
	benchBytesToUint64([]byte{0x54, 0x01, 0x12, 0xBC}, b)
}

var resultHexLitsToBytes []byte

func benchHexLitsToBytes(input []string, b *testing.B) {
	var r []byte

	for n := 0; n < b.N; n++ {
		r, _ = HexLitsToBytes(input)
	}

	resultHexLitsToBytes = r
}

func BenchmarkHexLitsToBytes8(b *testing.B) {
	benchHexLitsToBytes(
		[]string{"54", "01", "12", "BC", "BE", "EF", "30", "4B"},
		b,
	)
}

func BenchmarkHexLitsToBytes4(b *testing.B) {
	benchHexLitsToBytes([]string{"54", "01", "12", "BC"}, b)
}

var resultBytesToBits []bool

func benchBytesToBits(input []byte, b *testing.B) {
	var r []bool

	for n := 0; n < b.N; n++ {
		r = BytesToBits(input)
	}

	resultBytesToBits = r
}

func BenchmarkBytesToBits8(b *testing.B) {
	benchBytesToBits(
		[]byte{0x54, 0x01, 0x12, 0xBC, 0xBE, 0xEF, 0x30, 0x4B},
		b,
	)
}

func BenchmarkBytesToBits4(b *testing.B) {
	benchBytesToBits([]byte{0x54, 0x01, 0x12, 0xBC}, b)
}

/*==============================================================================
 * Examples
 */

func ExampleHexLitsToBytes() {
	res, err := HexLitsToBytes([]string{"01", "02", "03"})

	if err != nil {
		fmt.Println("Failed to parse hex literals", err)
	} else {
		fmt.Println(res)
	}
	// Output: [1 2 3]
}

/*==============================================================================
 * Tests
 */

func TestBytesToUint64(t *testing.T) {
	// ---- it returns 0 on empty input

	res, err := BytesToUint64([]byte{})

	if err != nil {
		t.Error("Got error from empty input:", err)
	}

	assertEqual(t, res, uint64(0))

	// ---- it converts bytes correctly

	type scenario struct {
		input    []byte
		expected uint64
	}

	scenarios := []scenario{
		scenario{
			[]byte{},
			0,
		},
		scenario{
			[]byte{0x00},
			0x00,
		},
		scenario{
			[]byte{0x01},
			0x01,
		},
		// Example from ELM327 data sheet
		scenario{
			[]byte{0x7B},
			123,
		},
		scenario{
			[]byte{0x1F, 0xF1},
			0x1FF1,
		},
		// Example from ELM327 data sheet
		scenario{
			[]byte{0x1A, 0xF8},
			6904,
		},
		scenario{
			[]byte{0x54, 0x01, 0x12, 0xBC},
			0x540112BC,
		},
	}

	for i := range scenarios {
		curr := scenarios[i]

		res, err = BytesToUint64(curr.input)

		if err != nil {
			t.Error("Failed to convert", err)
		}

		if res != curr.expected {
			t.Errorf("Expected 0x%X, got 0x%X", curr.expected, res)
		}
	}
}

func TestHexLitsToBytes(t *testing.T) {
	// ---- it returns empty on empty input

	_, err := HexLitsToBytes([]string{})

	if err != nil {
		t.Error("Got error from empty input:", err)
	}

	// ---- it rejects invalid hex literals

	_, err = HexLitsToBytes([]string{
		"Tv",
		"Mw",
	})

	if err == nil {
		t.Error("Expected error from invalid hex literals")
	}

	// ---- it accepts valid input

	res, err := HexLitsToBytes([]string{
		"0B",
		"F2",
		"00",
	})

	if err != nil {
		t.Error("Got error from valid input", err)
	}

	if res[0] != 11 {
		t.Error("Expected 0B hex to be 11 dec, got", res[0])
	}

	if res[1] != 242 {
		t.Error("Expected F2 hex to be 242 dec, got", res[1])
	}

	if res[2] != 0 {
		t.Error("Expected 0 hex to be 0 dec, got", res[2])
	}
}

func TestByteToBits(t *testing.T) {
	// ---- it returns right amount of bits

	res := ByteToBits(0)

	if len(res) != 8 {
		t.Error(
			"Expected ",
			8,
			" bits from 4 bytes of input, got",
			len(res),
		)
	}

	// ---- it returns right bits from input 255

	res = ByteToBits(255)

	exp := [8]bool{true, true, true, true, true, true, true, true}

	if res != exp {
		t.Error("Expected all true, got", res)
	}

	// ---- it returns right bits from input 128

	res = ByteToBits(128)

	exp = [8]bool{true, false, false, false, false, false, false, false}

	if res != exp {
		t.Error("Expected", exp, "true, got", res)
	}
}
