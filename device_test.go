package elmobd

import (
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
		NewPartSupported(1),
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

func TestToCommand(t *testing.T) {
	assertEqual(t, NewEngineLoad().ToCommand(), "01041")
	assertEqual(t, NewVehicleSpeed().ToCommand(), "010D1")
	assertEqual(t, NewDistSinceDTCClear().ToCommand(), "01311")
}

func TestIsSupported(t *testing.T) {
	sc, err := NewSupportedCommands([]uint32{0x0, 0x0, 0x0, 0x0, 0x0})

	assert(t, err == nil, "Supported commands was created successfully")

	cmd1 := NewPartSupported(1)

	assertEqual(t, sc.IsSupported(cmd1), true)

	cmd2 := NewEngineLoad()

	part, err := sc.GetPart(0)

	assert(t, err == nil, "New part supported was created successfully")

	part.SetRawValue(0x10000000)

	assertEqual(t, sc.IsSupported(cmd2), true)
}

func TestGetPart(t *testing.T) {
	sc, err := NewSupportedCommands([]uint32{0x0, 0x0, 0x0, 0x0})

	assert(t, err == nil, "Supported commands was created successfully")

	part, err := sc.GetPart(0)

	assert(t, part != nil && err == nil, "Getting first part is successful")

	part, err = sc.GetPart(3)

	assert(t, part != nil && err == nil, "Getting last part is successful")

	part, err = sc.GetPart(4)

	assert(t, part == nil && err != nil, "Getting out of bounds path fails")
}

func TestGetPartByPID(t *testing.T) {
	sc, err := NewSupportedCommands([]uint32{0x0, 0x0, 0x0})

	assert(t, err == nil, "Supported commands was created successfully")

	part, err := sc.GetPartByPID(0x0)

	assert(t, part != nil && err == nil, "Getting PID 0x0 is successful")
	assert(t, part.Index() == 1, "Part 1 is return for PID 0")

	part, err = sc.GetPartByPID(0x20)

	assert(t, part != nil && err == nil, "Getting PID 0x20 is successful")
	assert(t, part.Index() == 1, "Part 1 is return for PID 0x20")

	part, err = sc.GetPartByPID(0x21)

	assert(t, part != nil && err == nil, "Getting PID 0x21 is successful")
	assert(t, part.Index() == 2, "Part 2 is return for PID 0x21")

	part, err = sc.GetPartByPID(0x40)

	assert(t, part != nil && err == nil, "Getting PID 0x30 is successful")
	assert(t, part.Index() == 2, "Part 2 is return for PID 0x40")

	part, err = sc.GetPartByPID(0x41)

	assert(t, part != nil && err == nil, "Getting PID 0x41 is successful")
	assert(t, part.Index() == 3, "Part 3 is return for PID 0x41")

	part, err = sc.GetPartByPID(0x60)

	assert(t, part != nil && err == nil, "Getting PID 0x60 is successful")
	assert(t, part.Index() == 3, "Part 3 is return for PID 0x60")

	part, err = sc.GetPartByPID(0x61)

	assert(t, part == nil && err != nil, "Getting PID 0x61 fails, since part 4 doesn't exist")
}

type DummyCommand struct {
	baseCommand
}

func (cmd *DummyCommand) ValueAsLit() string {
	return "0x0"
}

func (cmd *DummyCommand) SetValue(result *Result) error {
	return nil
}

// TestIsSupportedWikipediaExample checks that the IsSupported function returns
// the correct results when using the example on the wikipedia-page about
// decoding Service 01 PID 00:
//
// > For example, if the car response is BE1FA813, it can be decoded like this:
// > So, supported PIDs are: 01, 03, 04, 05, 06, 07, 0C, 0D, 0E, 0F, 10, 11, 13,
// > 15, 1C, 1F and 20
//
// Source: https://en.wikipedia.org/wiki/OBD-II_PIDs#Service_01_PID_00
func TestIsSupportedWikipediaExample(t *testing.T) {
	sc, err := NewSupportedCommands([]uint32{0xBE1FA813, 0x0, 0x0, 0x0, 0x0})

	assert(t, err == nil, "Supported commands was created successfully")

	supported := []OBDParameterID{
		0x01, 0x03, 0x04, 0x05, 0x06, 0x07, 0x0D, 0x0E,
		0x0F, 0x10, 0x11, 0x13, 0x15, 0x1C, 0x1F, 0x20,
	}

	unsupported := []OBDParameterID{
		0x02, 0x08, 0x09, 0x0A, 0x0B, 0x12, 0x14, 0x16,
		0x17, 0x18, 0x19, 0x1A, 0x1B, 0x1D, 0x1E,
	}

	for _, pid := range supported {
		cmd := &DummyCommand{
			baseCommand{SERVICE_01_ID, pid, 1, "dummy"},
		}

		assertEqual(t, sc.IsSupported(cmd), true)
	}

	for _, pid := range unsupported {
		cmd := &DummyCommand{
			baseCommand{SERVICE_01_ID, pid, 1, "dummy"},
		}

		assertEqual(t, sc.IsSupported(cmd), false)
	}
}

// TestIssue27Regression verifies that the IsSupported function returns the correct
// results when checking PIDs outside of part 1.
//
// The test simulates that we have received the result of 5 parts, where
// each part has the first and last bit active, and all other bits inactive.
//
// This means that the first PID of the part should be supported, as well as
// the last PID, which is the one you use to check which PIDs of
// the next part are supported.
//
// Source: https://github.com/rzetterberg/elmobd/issues/27
func TestIssue27Regression(t *testing.T) {
	sc, err := NewSupportedCommands([]uint32{
		// Result from calling:
		// 0x00 "PIDs supported [01 - 20]"
		// ----------------------------------
		// > Bit encoded [A7..D0] == [PID $01..PID $20]
		//
		// A7 = 0x01 "Monitor status since DTCs cleared"
		// D0 = 0x20 "PIDs supported [21 - 40]"
		//
		// A7       B7       C7       D7     D0
		// |        |        |        |      |
		// 10000000 00000000 00000000 00000001,
		0x80000001,

		// Result from calling:
		// 0x20 "PIDs supported [21 - 40]"
		// ----------------------------------
		// > Bit encoded [A7..D0] == [PID $21..PID $40]
		//
		// A7 = 0x21 "Distance traveled with malfunction indicator lamp (MIL) on"
		// D0 = 0x40 "PIDs supported [41 - 60]"
		//
		// A7       B7       C7       D7     D0
		// |        |        |        |      |
		// 10000000 00000000 00000000 00000001,
		0x80000001,

		// Result from calling:
		// 0x40 "PIDs supported [41 - 60]"
		// ----------------------------------
		// > Bit encoded [A7..D0] == [PID $41..PID $60]
		//
		// A7 = 0x41 "Monitor status this drive cycle"
		// D0 = 0x60 "PIDs supported [61 - 80]"
		//
		// A7       B7       C7       D7     D0
		// |        |        |        |      |
		// 10000000 00000000 00000000 00000001,
		0x80000001,

		// Result from calling:
		// 0x60 "PIDs supported [61 - 80]"
		// ----------------------------------
		// > Bit encoded [A7..D0] == [PID $61..PID $80]
		//
		// A7 = 0x61 "Driver's demand engine - percent torque"
		// D0 = 0x80 "PIDs supported [81 - A0]"
		//
		// A7       B7       C7       D7     D0
		// |        |        |        |      |
		// 10000000 00000000 00000000 00000001,
		0x80000001,

		// Result from calling:
		// 0x80 "PIDs supported [81 - A0]"
		// ----------------------------------
		// > Bit encoded [A7..D0] == [PID $81..PID $A0]
		//
		// A7 = 0x81 "Engine run time for Auxiliary Emissions Control Device(AECD)"
		// D0 = 0xA0 "PIDs supported [A1 - C0]"
		//
		// A7       B7       C7       D7     D0
		// |        |        |        |      |
		// 10000000 00000000 00000000 00000001,
		0x80000001,
	})

	assert(t, err == nil, "Supported commands was created successfully")

	supported := []OBDParameterID{
		// Part 1 PIDs
		0x01, 0x20,

		// Part 2 PIDs
		0x21, 0x40,

		// Part 3 PIDs
		0x41, 0x60,

		// Part 4 PIDs
		0x61, 0x80,

		// Part 5 PIDs
		0x81, 0xA0,
	}

	for _, pid := range supported {
		cmd := &DummyCommand{
			baseCommand{SERVICE_01_ID, pid, 1, "dummy"},
		}

		assertEqual(t, sc.IsSupported(cmd), true)
	}
}

func TestParseOBDResponse(t *testing.T) {
	type scenario struct {
		command OBDCommand
		outputs []string
	}

	scenarios := []scenario{
		{
			NewPartSupported(1),
			[]string{"41 00 02 03 04 05"},
		},
		{
			NewMonitorStatus(),
			[]string{"41 01 FF 00 00 00"},
		},
		{
			NewEngineLoad(),
			[]string{"41 04 02"},
		},
		{
			NewFuel(),
			[]string{"41 2F 10"},
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
			NewDistSinceDTCClear(),
			[]string{"41 31 02 33"},
		},
		{
			NewClearTroubleCodes(),
			nil,
		},
	}

	for _, curr := range scenarios {
		assertOBDParseSuccess(t, curr.command, curr.outputs)
	}
}
