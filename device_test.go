package elmobd

import (
	"fmt"
	"testing"
)

/*==============================================================================
 * Benchmarks
 */

var resultParseOBDResponse uint64

func benchParseOBDResponse(cmd OBDCommand, input []string, b *testing.B) {
	var r uint64

	for n := 0; n < b.N; n++ {
		r, _ = parseOBDResponse(cmd, input)
	}

	resultParseOBDResponse = r
}

func BenchmarkParseOBDResponse4(b *testing.B) {
	benchParseOBDResponse(
		NewPart1Supported(),
		[]string{"41 00 02 03 04 05"},
		b,
	)
}

func BenchmarkParseOBDResponse2(b *testing.B) {
	benchParseOBDResponse(
		NewEngineLoad(),
		[]string{"41 04 03"},
		b,
	)
}

/*==============================================================================
 * Tests
 */

func assertEqual(t *testing.T, a interface{}, b interface{}) {
	if a == b {
		return
	}

	t.Fatal(fmt.Sprintf("%v != %v", a, b))
}

func TestToCommand(t *testing.T) {
	assertEqual(t, NewPart1Supported().ToCommand(), "0100")
	assertEqual(t, NewEngineLoad().ToCommand(), "0104")
	assertEqual(t, NewVehicleSpeed().ToCommand(), "010D")
}

func TestIsSupported(t *testing.T) {
	sc := SupportedCommands{0x0}

	cmd1 := NewPart1Supported()

	if !sc.IsSupported(cmd1) {
		t.Error("Expected supported sensors to always be supported")
	}

	cmd2 := NewEngineLoad()

	sc.part1 = 0x10000000

	if !sc.IsSupported(cmd2) {
		t.Errorf("Expected command %v to be supported", cmd2)
	}
}

func TestParseSupportedResponse(t *testing.T) {
	// ---- it accepts valid response

	res, err := parseSupportedResponse(
		NewPart1Supported(),
		[]string{
			"SEARCHING...",
			"41 00 01 02 03 04",
		},
	)

	if err != nil {
		t.Error("Failed parsing", err)
	}

	exp := uint64(0x01020304)

	if res != exp {
		t.Errorf("Expected 0x%02X, got 0x%02X", exp, res)
	}
}

func TestParseOBDResponse(t *testing.T) {
	type scenario struct {
		command OBDCommand
		outputs []string
	}

	scenarios := []scenario{
		scenario{
			NewPart1Supported(),
			[]string{"41 00 02 03 04 05"},
		},
		scenario{
			NewEngineLoad(),
			[]string{"41 04 02"},
		},
		scenario{
			NewCoolantTemperature(),
			[]string{"41 05 FF"},
		},
		scenario{
			NewFuelPressure(),
			[]string{"41 0A BC"},
		},
		scenario{
			NewIntakeManifoldPressure(),
			[]string{"41 0B C2"},
		},
		scenario{
			NewEngineRPM(),
			[]string{"41 0C FF B2"},
		},
		scenario{
			NewVehicleSpeed(),
			[]string{"41 0D A9"},
		},
		scenario{
			NewTimingAdvance(),
			[]string{"41 0E 4F"},
		},
		scenario{
			NewIntakeAirTemperature(),
			[]string{"41 0F EB"},
		},
		scenario{
			NewMafAirFlowRate(),
			[]string{"41 10 C2 8B"},
		},
		scenario{
			NewThrottlePosition(),
			[]string{"41 11 FF"},
		},
	}

	for _, curr := range scenarios {
		_, err := parseOBDResponse(
			curr.command,
			curr.outputs,
		)

		if err != nil {
			t.Error("Failed parsing", err)
		}
	}
}
