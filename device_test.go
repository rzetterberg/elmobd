package elmobd

import (
	"fmt"
	"testing"
)

/*==============================================================================
 * Benchmarks
 */

var resultParseOBDResponse *Result

func benchParseOBDResponse(cmd OBDCommand, input []string, b *testing.B) {
	var r *Result

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
	sc := SupportedCommands{0x0, 0x0, 0x0, 0x0, 0x0}

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

func TestParseOBDResponse(t *testing.T) {
	type scenario struct {
		command OBDCommand
		outputs []string
	}

	scenarios := []scenario{
		{
			NewPart1Supported(),
			[]string{"41 00 02 03 04 05"},
		},
		{
			NewEngineLoad(),
			[]string{"41 04 02"},
		},
		{
			NewCoolantTemperature(),
			[]string{"41 05 FF"},
		},
		{
			NewShortFuelTrim1(),
			[]string{"41 06 F2"},
		},
		{
			NewLongFuelTrim1(),
			[]string{"41 07 2F"},
		},
		{
			NewShortFuelTrim2(),
			[]string{"41 08 20"},
		},
		{
			NewLongFuelTrim2(),
			[]string{"41 09 01"},
		},
		{
			NewFuelPressure(),
			[]string{"41 0A BC"},
		},
		{
			NewIntakeManifoldPressure(),
			[]string{"41 0B C2"},
		},
		{
			NewEngineRPM(),
			[]string{"41 0C FF B2"},
		},
		{
			NewVehicleSpeed(),
			[]string{"41 0D A9"},
		},
		{
			NewTimingAdvance(),
			[]string{"41 0E 4F"},
		},
		{
			NewIntakeAirTemperature(),
			[]string{"41 0F EB"},
		},
		{
			NewMafAirFlowRate(),
			[]string{"41 10 C2 8B"},
		},
		{
			NewThrottlePosition(),
			[]string{"41 11 FF"},
		},
		// Regression tests for https://github.com/rzetterberg/elmobd/issues/5
		{
			NewEngineRPM(),
			[]string{
				"SEARCHING...",
				"41 0C FF B2",
			},
		},
		{
			NewEngineRPM(),
			[]string{
				"BUS INIT",
				"41 0C FF B2",
			},
		},
		{
			NewThrottlePosition(),
			[]string{
				"SEARCHING...",
				"SEARCHING...",
				"SEARCHING...",
				"41 11 FF",
			},
		},
		{
			NewPart1Supported(),
			[]string{
				"SEARCHING...",
				"41 00 01 02 03 04",
			},
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
