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

// Validate checks that the result is for the given OBDCommand by checking the
// length of the data, comparing the mode ID and the parameter ID.
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

// Device represents the connection to a ELM327 device. This is the data type
// you use to run commands on the connected ELM327 device, see NewDevice for
// creating a Device and RunOBDCommand for running commands.
type Device struct {
	rawDevice   *RawDevice
	outputDebug bool
}

// NewDevice constructs a Device by initilizing the serial connection and
// setting the protocol to talk with the car to "automatic".
func NewDevice(devicePath string, debug bool) (*Device, error) {
	rawDev, err := NewRawDevice(devicePath)

	if err != nil {
		return nil, err
	}

	dev := Device{
		rawDev,
		debug,
	}

	err = dev.SetAutomaticProtocol()

	if err != nil {
		return nil, err
	}

	return &dev, nil
}

// SetAutomaticProtocol tells the ELM327 device to automatically discover what
// protocol to talk to the car with. How the protocol is chhosen is something
// that the ELM327 does internally. If you're interested in how this works you
// can look in the data sheet linked in the beginning of the package description.
func (dev *Device) SetAutomaticProtocol() error {
	res := dev.rawDevice.RunCommand("ATSP0")

	if res.Error != nil {
		return res.Error
	}

	if dev.outputDebug {
		fmt.Println(res.FormatOverview())
	}

	if res.Outputs[0] != "OK" {
		return fmt.Errorf(
			"Expected OK response, got: %q",
			res.Outputs[0],
		)
	}

	return nil
}

// GetVersion gets the version of the connected ELM327 device. The latest
// version being v2.2.
func (dev *Device) GetVersion() (string, error) {
	res := dev.rawDevice.RunCommand("AT@1")

	if res.Error != nil {
		return "", res.Error
	}

	if dev.outputDebug {
		fmt.Println(res.FormatOverview())
	}

	version := res.Outputs[0][:]

	return strings.Trim(version, " "), nil
}

// CheckSupportedCommands check which commands are supported (PID 1 to PID 160)
// by the car connected to the ELM327 device.
//
// Since a single command can only contain 32-bits of information 5 commands
// are run in series by this function to get the whole 160-bits of information.
func (dev *Device) CheckSupportedCommands() (*SupportedCommands, error) {
	part1, err := dev.CheckSupportedPart(NewPart1Supported())

	if err != nil {
		return nil, err
	}

	part2, err := dev.CheckSupportedPart(NewPart2Supported())

	if err != nil {
		return nil, err
	}

	part3, err := dev.CheckSupportedPart(NewPart3Supported())

	if err != nil {
		return nil, err
	}

	part4, err := dev.CheckSupportedPart(NewPart4Supported())

	if err != nil {
		return nil, err
	}

	part5, err := dev.CheckSupportedPart(NewPart5Supported())

	if err != nil {
		return nil, err
	}

	result := SupportedCommands{
		uint32(part1.(*Part1Supported).Value),
		uint32(part2.(*Part2Supported).Value),
		uint32(part3.(*Part3Supported).Value),
		uint32(part4.(*Part4Supported).Value),
		uint32(part5.(*Part5Supported).Value),
	}

	return &result, nil
}

// CheckSupportedPart checks the availability of a range of 32 commands.
func (dev *Device) CheckSupportedPart(cmd OBDCommand) (OBDCommand, error) {
	rawResult := dev.rawDevice.RunCommand(cmd.ToCommand())

	if rawResult.Error != nil {
		return cmd, rawResult.Error
	}

	if dev.outputDebug {
		fmt.Println(rawResult.FormatOverview())
	}

	result, err := parseSupportedResponse(cmd, rawResult.Outputs)

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

// RunOBDCommand runs the given OBDCommand on the connected ELM327 device and
// populates the OBDCommand with the parsed output from the device.
func (dev *Device) RunOBDCommand(cmd OBDCommand) (OBDCommand, error) {
	rawResult := dev.rawDevice.RunCommand(cmd.ToCommand())

	if rawResult.Error != nil {
		return cmd, rawResult.Error
	}

	if dev.outputDebug {
		fmt.Println(rawResult.FormatOverview())
	}

	result, err := parseOBDResponse(cmd, rawResult.Outputs)

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
		res, err := dev.RunOBDCommand(cmd)

		if err != nil {
			return []OBDCommand{}, err
		}

		result = append(result, res)
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

// parseSupportedResponse parses the raw output produced from running the given
// OBDCommand. Is specific to OBDCommands that check the availability of a
// command range.
//
// The output from these commands have a special structure that needs to be
// handled before the generic parseOBDResponse function can be run on the result.
func parseSupportedResponse(cmd OBDCommand, outputs []string) (*Result, error) {
	if len(outputs) < 2 {
		return nil, fmt.Errorf(
			"Expected more than one output, got: %q",
			outputs,
		)
	}

	if outputs[1] == "UNABLE TO CONNECT" {
		return nil, fmt.Errorf(
			"Unable to connect to car, is the ignition on?",
		)
	}

	return parseOBDResponse(cmd, outputs[1:])
}

// parseOBDResponse parses the raw output produced from running the given
// OBDCommand on the connected ELM327 device.
//
// The output is checked so that it doesn't represent a failed command run
// before return a new Result.
func parseOBDResponse(cmd OBDCommand, outputs []string) (*Result, error) {
	if len(outputs) < 1 {
		return nil, fmt.Errorf(
			"Expected atleast one output, got: %q",
			outputs,
		)
	}

	payload := outputs[0]

	if payload == "UNABLE TO CONNECT" {
		return nil, fmt.Errorf(
			"Unable to connect to car, is the ignition on?",
		)
	}

	if payload == "NO DATA" {
		return nil, fmt.Errorf(
			"No data from car, time out from elm device?",
		)
	}

	return NewResult(payload)
}
