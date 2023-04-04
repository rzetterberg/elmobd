package elmobd

import (
	"fmt"
	"strings"
	"time"
)

/*==============================================================================
 * External
 */

// MockResult represents the raw text output of running a raw command,
// including information used in debugging to show what input caused what
// error, how long the command took, etc.
type MockResult struct {
	input     string
	outputs   []string
	error     error
	writeTime time.Duration
	readTime  time.Duration
	totalTime time.Duration
}

// Failed checks if the result is successful or not
func (res *MockResult) Failed() bool {
	return res.error != nil
}

// GetError returns the results current error
func (res *MockResult) GetError() error {
	return res.error
}

// GetOutputs returns the outputs of the result
func (res *MockResult) GetOutputs() []string {
	return res.outputs
}

// FormatOverview formats a result as an overview of what command was run and
// how long it took.
func (res *MockResult) FormatOverview() string {
	lines := []string{
		"=======================================",
		" Mocked command \"%s\"",
		"=======================================",
	}

	return fmt.Sprintf(
		strings.Join(lines, "\n"),
		res.input,
	)
}

// MockDevice represent a mocked serial connection
type MockDevice struct {
}

// RunCommand mocks the given AT/OBD command by just returning a result for the
// mocked outputs set earlier.
func (dev *MockDevice) RunCommand(command string) RawResult {
	return &MockResult{
		input:     command,
		outputs:   mockOutputs(command),
		writeTime: 0,
		readTime:  0,
		totalTime: 0,
	}
}

/*==============================================================================
 * Internal
 */

func mockMode1Outputs(subcmd string) []string {
	if strings.HasPrefix(subcmd, "00") {
		// PIDs supported part 1
		return []string{
			"41 00 0C 10 00 00", // Means PIDs supported: 05, 06, 0C
		}
	} else if strings.HasPrefix(subcmd, "20") {
		// PIDs supported part 2
		return []string{
			"41 20 00 00 00 00",
		}
	} else if strings.HasPrefix(subcmd, "40") {
		// PIDs supported part 3
		return []string{
			"41 40 00 00 00 00",
		}
	} else if strings.HasPrefix(subcmd, "60") {
		// PIDs supported part 4
		return []string{
			"41 60 00 00 00 00",
		}
	} else if strings.HasPrefix(subcmd, "80") {
		// PIDs supported part 5
		return []string{
			"41 80 00 00 00 00",
		}
	} else if strings.HasPrefix(subcmd, "01") {
		return []string{
			"41 01 FF 00 00 00",
		}
	} else if strings.HasPrefix(subcmd, "05") { // Engine coolant temperature
		return []string{
			"41 05 4F", // 39 C
		}
	} else if strings.HasPrefix(subcmd, "06") { // Short term fuel trim - Bank 1
		return []string{
			"41 06 02", // -98.4375%
		}
	} else if strings.HasPrefix(subcmd, "0C") { // Engine speed
		return []string{
			"41 0C 03 00", // 192 rpm
		}
	} else if strings.HasPrefix(subcmd, "2F") { // Fuel tank level input
		return []string{
			"41 2F 6B", // 41.96%
		}
	} else if strings.HasPrefix(subcmd, "0D") { // Vehicle speed
		return []string{
			"41 0D 4B", // 75 km/h
		}
	} else if strings.HasPrefix(subcmd, "31") { // Distance traveled since codes cleared
		return []string{
			"41 31 02 0C", // 524 km
		}
	} else if strings.HasPrefix(subcmd, "42") { // Control Module Voltage
		return []string{
			"41 42 33 90", // 13.2 volts
		}
	} else if strings.HasPrefix(subcmd, "A6") { // Odometer
		return []string{
			"41 A6 00 06 68 a0", // 42,000.00 km
		}
	} else if strings.HasPrefix(subcmd, "46") { // Ambient Air Temperature
		return []string{
			"41 46 3A", // 18.0 C
		}
	} else if strings.HasPrefix(subcmd, "A4") { // Transmission Actual Gear
		return []string{
			"41 A4 27 10 00 00", // 10.0:1
		}
	}

	return []string{"NOT SUPPORTED"}
}

func mockOutputs(cmd string) []string {
	if cmd == "ATSP0" {
		return []string{"OK"}
	} else if cmd == "AT@1" {
		return []string{"OBDII by elm329@gmail.com"}
	} else if cmd == "AT RV" {
		return []string{"12.1234"}
	} else if strings.HasPrefix(cmd, "01") {
		return mockMode1Outputs(cmd[2:])
	}

	return []string{"NOT SUPPORTED"}
}
