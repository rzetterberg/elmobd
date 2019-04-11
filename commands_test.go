package elmobd

import (
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

/* TODO fix test assert
func TestFreezeFrameResult(t *testing.T) {
	command := NewFreezeFrame()
	outputs := []string{"41 02 00 00"}
	command = assertOBDParseSuccess(t, command, outputs).(*FreezeFrame)
	
	assert(t, command.Value == 0x0000, "00 00")
}
*/
