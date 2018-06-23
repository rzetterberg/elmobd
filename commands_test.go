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
