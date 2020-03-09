package elmobd

import (
	"fmt"
	"testing"
)

/*==============================================================================
 * Tests
 */

func TestMonitorStatusResult(t *testing.T) {
	command := NewMonitorStatus()
	outputs := []string{"41 01 FF 00 00 00"}
	command = assertOBDParseSuccess(t, command, outputs).(*MonitorStatus)

	assert(t, command.MilActive == true, "MIL was not active")
	assert(t, command.DtcAmount == 127, "DTCs were not 127")
}

func TestPartSupportedCommandInRange(t *testing.T) {
	type scenario struct {
		part        *PartSupported
		insidePIDs  []OBDParameterID
		outsidePIDs []OBDParameterID
	}

	scenarios := []scenario{
		{
			NewPartSupported(1),
			[]OBDParameterID{0x01, 0x10, 0x20},
			[]OBDParameterID{0x00, 0x21, 0x40, 0x41},
		},
		{
			NewPartSupported(2),
			[]OBDParameterID{0x21, 0x30, 0x40},
			[]OBDParameterID{0x00, 0x01, 0x20, 0x41},
		},
		{
			NewPartSupported(3),
			[]OBDParameterID{0x41, 0x50, 0x60},
			[]OBDParameterID{0x00, 0x01, 0x40, 0x61},
		},
		{
			NewPartSupported(4),
			[]OBDParameterID{0x61, 0x70, 0x80},
			[]OBDParameterID{0x00, 0x01, 0x60, 0x81},
		},
	}

	for _, scen := range scenarios {
		for _, pid := range scen.insidePIDs {
			assert(
				t,
				scen.part.PIDInRange(pid) == true,
				fmt.Sprintf(
					"PID 0x%X was in range of part %d",
					pid,
					scen.part.Index(),
				),
			)
		}

		for _, pid := range scen.outsidePIDs {
			assert(
				t,
				scen.part.PIDInRange(pid) == false,
				fmt.Sprintf(
					"PID 0x%X was out of range of part %d",
					pid,
					scen.part.Index(),
				),
			)
		}
	}
}

func TestPartSupportedSupportsCommand(t *testing.T) {
	part := NewPartSupported(1)

	part.Value = 0xBE1FA813

	supported := []OBDParameterID{
		0x01, 0x03, 0x04, 0x05, 0x06, 0x07, 0x0D, 0x0E,
		0x0F, 0x10, 0x11, 0x13, 0x15, 0x1C, 0x1F, 0x20,
	}

	unsupported := []OBDParameterID{
		0x02, 0x08, 0x09, 0x0A, 0x0B, 0x12, 0x14, 0x16,
		0x17, 0x18, 0x19, 0x1A, 0x1B, 0x1D, 0x1E,
	}

	for _, pid := range supported {
		assertEqual(t, part.SupportsPID(pid), true)
	}

	for _, pid := range unsupported {
		assertEqual(t, part.SupportsPID(pid), false)
	}
}

func TestService01Command(t *testing.T) {
	service01cmd := NewPartSupported(1)
	assert(t, service01cmd.ModeID() == SERVICE_01_ID, fmt.Sprintf("Service id \"%d\" is not \"%d\"", service01cmd.ModeID(), SERVICE_01_ID))
}

func TestService04Command(t *testing.T) {
	command := NewClearTroubleCodes()
	assert(t, command.ModeID() == SERVICE_04_ID, "Service id is not "+string(SERVICE_04_ID))
}
