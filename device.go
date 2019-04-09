package elmobd

import (
	"fmt"
	"strconv"
	"strings"
)

/*==============================================================================
 * External
 */

// Result represents the results from running a command on the ELM327 device,
// encoded as a byte array. When you run a command on the ELM327 device the
// response is a space-separated string of hex bytes, which looks something
// like this:
//
//   41 0C 1A F8
//
// The first 2 bytes are control bytes, while the rest of the bytes represent
// the actual result. So this data type contains an array of those bytes in
// binary.
//
// This data type is only used internally to produce the processed value of
// the run OBDCommand, so the end user will never use this data type.
type Result struct {
	value []byte
}

// NewResult constructrs a Result by taking care of parsing the hex bytes into
// binary representation.
func NewResult(rawLine string) (*Result, error) {
	literals := strings.Split(rawLine, " ")

	if len(literals) < 3 {
		return nil, fmt.Errorf(
			"Expected at least 3 OBD literals: %s", rawLine,
		)
	}

	result := Result{make([]byte, 0)}

	for i := range literals {
		curr, err := strconv.ParseUint(
			literals[i],
			16,
			8,
		)

		if err != nil {
			return nil, err
		}

		result.value = append(result.value, uint8(curr))
	}

	return &result, nil
}

// Validate checks that the result is for the given OBDCommand by:
// - Comparing the bytes received and the expected amount of bytes to receive
// - Comparing the received mode ID and the expected mode ID
// - Comparing the received parameter ID and the expected parameter ID
func (res *Result) Validate(cmd OBDCommand) error {
	valueLen := len(res.value)
	expLen := int(cmd.DataWidth() + 2)

	if valueLen != expLen {
		return fmt.Errorf(
			"Expected %d bytes, found %d",
			expLen,
			valueLen,
		)
	}

	modeResp := cmd.ModeID() + 0x40

	if res.value[0] != modeResp {
		return fmt.Errorf(
			"Expected mode echo %02X, got %02X",
			modeResp,
			res.value[0],
		)
	}

	if OBDParameterID(res.value[1]) != cmd.ParameterID() {
		return fmt.Errorf(
			"Expected parameter echo %02X got %02X",
			cmd.ParameterID(),
			res.value[1],
		)
	}

	return nil
}

// payloadAsUInt casts the Result as a unisgned 64-bit integer and making sure
// it has the expected amount of bytes.
//
// This function is used by other more specific utility functions that cast the
// result to a specific unsigned integer typer.
//
// By verifying the amount of bytes in the result using this function we can
// safely cast the resulting uint64 into types with less bits.
func (res *Result) payloadAsUInt(expAmount int) (uint64, error) {
	var result uint64

	payload := res.value[2:]
	amount := len(payload)

	if amount != expAmount {
		return 0, fmt.Errorf(
			"Expected %d bytes of payload, got %d", expAmount, amount,
		)
	}

	for i := range payload {
		curr := payload[amount-(i+1)]

		result |= uint64(curr) << uint(i*8)
	}

	return result, nil
}

// PayloadAsUInt64 is a helper for getting payload as uint64.
func (res *Result) PayloadAsUInt64() (uint64, error) {
	result, err := res.payloadAsUInt(8)

	if err != nil {
		return 0, err
	}

	return uint64(result), nil
}

// PayloadAsUInt322 is a helper for getting payload as uint32.
func (res *Result) PayloadAsUInt322() (uint32, error) {
	result, err := res.payloadAsUInt(2)

	if err != nil {
		return 0, err
	}

	return uint32(result), nil
}

// PayloadAsUInt32 is a helper for getting payload as uint32.
func (res *Result) PayloadAsUInt32() (uint32, error) {
	result, err := res.payloadAsUInt(4)

	if err != nil {
		return 0, err
	}

	return uint32(result), nil
}

// PayloadAsUInt16 is a helper for getting payload as uint16.
func (res *Result) PayloadAsUInt16() (uint16, error) {
	result, err := res.payloadAsUInt(2)

	if err != nil {
		return 0, err
	}

	return uint16(result), nil
}

// PayloadAsByte is a helper for getting payload as byte.
func (res *Result) PayloadAsByte() (byte, error) {
	result, err := res.payloadAsUInt(1)

	if err != nil {
		return 0, err
	}

	return uint8(result), nil
}

// RawResult represents the raw text output of running a raw command,
// including information used in debugging to show what input caused what
// error, how long the command took, etc.
type RawResult interface {
	Failed() bool
	GetError() error
	GetOutputs() []string
	FormatOverview() string
}

// RawDevice represent the low level device, which can either be the real
// implementation or a mock implementation used for testing.
type RawDevice interface {
	RunCommand(string) RawResult
}

// Device represents the connection to a ELM327 device. This is the data type
// you use to run commands on the connected ELM327 device, see NewDevice for
// creating a Device and RunOBDCommand for running commands.
type Device struct {
	rawDevice   RawDevice
	outputDebug bool
}

// NewDevice constructs a Device by initilizing the serial connection and
// setting the protocol to talk with the car to "automatic".
func NewDevice(devicePath string, debug bool) (*Device, error) {
	rawDev, err := NewRealDevice(devicePath)

	if err != nil {
		return nil, err
	}

	dev := Device{rawDev, debug}

	err = dev.SetAutomaticProtocol()

	if err != nil {
		return nil, err
	}

	return &dev, nil
}

// NewTestDevice constructs a Device which is using a mocked RawDevice.
func NewTestDevice(devicePath string, debug bool) (*Device, error) {
	dev := Device{&MockDevice{}, debug}

	return &dev, nil
}

// SetAutomaticProtocol tells the ELM327 device to automatically discover what
// protocol to talk to the car with. How the protocol is chhosen is something
// that the ELM327 does internally. If you're interested in how this works you
// can look in the data sheet linked in the beginning of the package description.
func (dev *Device) SetAutomaticProtocol() error {
	rawRes := dev.rawDevice.RunCommand("ATSP0")

	if rawRes.Failed() {
		return rawRes.GetError()
	}

	if dev.outputDebug {
		fmt.Println(rawRes.FormatOverview())
	}

	outputs := rawRes.GetOutputs()

	if outputs[0] != "OK" {
		return fmt.Errorf(
			"Expected OK response, got: %q",
			outputs[0],
		)
	}

	return nil
}

// GetVersion gets the version of the connected ELM327 device. The latest
// version being v2.2.
func (dev *Device) GetVersion() (string, error) {
	rawRes := dev.rawDevice.RunCommand("AT@1")

	if rawRes.Failed() {
		return "", rawRes.GetError()
	}

	if dev.outputDebug {
		fmt.Println(rawRes.FormatOverview())
	}

	outputs := rawRes.GetOutputs()
	version := outputs[0][:]

	return strings.Trim(version, " "), nil
}

// CheckSupportedCommands check which commands are supported (PID 1 to PID 160)
// by the car connected to the ELM327 device.
//
// Since a single command can only contain 32-bits of information 5 commands
// are run in series by this function to get the whole 160-bits of information.
//
// Returns error if the first command (PID 1 to 20) fails, the rest of the 5
// commands are ignored if they fail.
func (dev *Device) CheckSupportedCommands() (*SupportedCommands, error) {
	partRes, err := dev.RunOBDCommand(NewPart1Supported())
	part1 := uint32(0)

	if err != nil {
		return nil, err
	}

	part1 = uint32(partRes.(*Part1Supported).Value)

	partRes, err = dev.RunOBDCommand(NewPart2Supported())
	part2 := uint32(0)

	if err == nil {
		part2 = uint32(partRes.(*Part2Supported).Value)
	}

	partRes, err = dev.RunOBDCommand(NewPart3Supported())
	part3 := uint32(0)

	if err == nil {
		part3 = uint32(partRes.(*Part3Supported).Value)
	}

	partRes, err = dev.RunOBDCommand(NewPart4Supported())
	part4 := uint32(0)

	if err == nil {
		part4 = uint32(partRes.(*Part4Supported).Value)
	}

	partRes, err = dev.RunOBDCommand(NewPart5Supported())
	part5 := uint32(0)

	if err == nil {
		part5 = uint32(partRes.(*Part5Supported).Value)
	}

	result := SupportedCommands{part1, part2, part3, part4, part5}

	return &result, nil
}

// RunOBDCommand runs the given OBDCommand on the connected ELM327 device and
// populates the OBDCommand with the parsed output from the device.
func (dev *Device) RunOBDCommand(cmd OBDCommand) (OBDCommand, error) {
	rawRes := dev.rawDevice.RunCommand(cmd.ToCommand())

	if rawRes.Failed() {
		return cmd, rawRes.GetError()
	}

	if dev.outputDebug {
		fmt.Println(rawRes.FormatOverview())
	}

	result, err := parseOBDResponse(cmd, rawRes.GetOutputs())

	if err != nil {
		return cmd, err
	}

	err = result.Validate(cmd)

	if err != nil {
		return cmd, err
	}

	err = cmd.SetValue(result)

	return cmd, err
}

// RunManyOBDCommands is a helper function to run multiple commands in series.
func (dev *Device) RunManyOBDCommands(commands []OBDCommand) ([]OBDCommand, error) {
	var result []OBDCommand

	for _, cmd := range commands {
		processed, err := dev.RunOBDCommand(cmd)

		if err != nil {
			return []OBDCommand{}, err
		}

		result = append(result, processed)
	}

	return result, nil
}

// SupportedCommands represents the lookup table for which commands
// (PID 1 to PID 160) that are supported by the car connected to the ELM327
// device.
type SupportedCommands struct {
	part1 uint32
	part2 uint32
	part3 uint32
	part4 uint32
	part5 uint32
}

// IsSupported checks if the given OBDCommand is supported.
//
// It does this by comparing the PID of the OBDCommand against the lookup table.
func (sc *SupportedCommands) IsSupported(cmd OBDCommand) bool {
	if cmd.ParameterID() == 0 {
		return true
	}

	pid := cmd.ParameterID()

	inPart1 := (sc.part1 >> uint32(32-pid)) & 1
	inPart2 := (sc.part2 >> uint32(32-pid)) & 1
	inPart3 := (sc.part3 >> uint32(32-pid)) & 1
	inPart4 := (sc.part4 >> uint32(32-pid)) & 1
	inPart5 := (sc.part5 >> uint32(32-pid)) & 1

	return (inPart1 == 1) || (inPart2 == 1) ||
		(inPart3 == 1) || (inPart4 == 1) || (inPart5 == 1)
}

// FilterSupported filters out the OBDCommands that are supported.
func (sc *SupportedCommands) FilterSupported(commands []OBDCommand) []OBDCommand {
	var result []OBDCommand

	for _, cmd := range commands {
		if sc.IsSupported(cmd) {
			result = append(result, cmd)
		}
	}

	return result
}

/*==============================================================================
 * Internal
 */

// parseOBDResponse parses the raw outputs produced from running the given
// OBDCommand on the connected ELM327 device.
//
// A response from the ELM327 device can fail for a variety of reasons,
// such as failing to connect to the car, or not receiving any data from the
// car.
//
// A response can also contain lines that say "SEARCHING..." or "BUS INIT"
// before the actual payload.
//
// This function iterates the outputs, stops if it finds any errors and ignores
// lines containing "SEARCHING..." or "BUS INIT". The first line that passes
// these checks is assumed to be the payload.
//
// This means that this function cannot handle multiline responses
// (such as getting the VIN number, and multiple PID requests baked into one).
// Handling these more advanced responses is something that is going to be
// implemented, but right now has been deprioritized.
func parseOBDResponse(cmd OBDCommand, outputs []string) (*Result, error) {
	payload := ""

	for _, out := range outputs {
		if strings.HasPrefix(out, "UNABLE TO CONNECT") {
			return nil, fmt.Errorf(
				"'UNABLE TO CONNECT' received, is the ignition on?",
			)
		} else if strings.HasPrefix(out, "NO DATA") {
			return nil, fmt.Errorf(
				"'NO DATA' received, timeout from elm device?",
			)
		} else if strings.HasPrefix(out, "SEARCHING") {
			continue
		} else if strings.HasPrefix(out, "BUS INIT") {
			continue
		}

		payload = out

		break
	}

	if payload == "" {
		return nil, fmt.Errorf(
			"Empty payload parsed from outputs: %s",
			outputs,
		)
	}

	return NewResult(payload)
}
