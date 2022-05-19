package elmobd

import (
	"fmt"
	"net/url"
	"os"
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
func NewDevice(addr string, debug bool) (*Device, error) {
	// If addr is an existing file/device we use it as a serial device
	if _, err := os.Stat(addr); err == nil {
		addr = fmt.Sprintf("serial://%s", addr)
	}

	u, err := url.Parse(addr)
	if err != nil {
		return nil, fmt.Errorf("failed to parse device address: %w", err)
	}

	dev := Device{outputDebug: debug}

	switch u.Scheme {
	case "serial":
		dev.rawDevice, err = NewSerialDevice(u)
	case "tcp", "tcp4", "tcp6", "unix":
		dev.rawDevice, err = NewNetDevice(u)
	case "test":
		dev.rawDevice, err = &MockDevice{}, nil
	}

	if err != nil {
		return nil, err
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

// GetVoltage gets the current battery voltage of the vehicle as measured
// by the ELM327 device.
func (dev *Device) GetVoltage() (float32, error) {
	rawRes := dev.rawDevice.RunCommand("AT RV")

	if rawRes.Failed() {
		return -1, rawRes.GetError()
	}

	if dev.outputDebug {
		fmt.Println(rawRes.FormatOverview())
	}

	output := rawRes.GetOutputs()[0]
	voltage, err := strconv.ParseFloat(output[:len(output)-1], 32)

	if err != nil {
		return -1, fmt.Errorf("voltage is not a floating point number: %w", err)
	}

	return float32(voltage), nil
}

// CheckSupportedCommands check which commands are supported by the car connected
// to the ELM327 device.
func (dev *Device) CheckSupportedCommands() (*SupportedCommands, error) {
	result := &SupportedCommands{
		[]*PartSupported{},
	}

	index := byte(1)

	for {
		part := NewPartSupported(index)

		partRes, err := dev.RunOBDCommand(part)

		if err != nil {
			return nil, err
		}

		result.AddPart(partRes.(*PartSupported))

		// Check if the car supports the PID that checks if the next part of PIDs
		// are supported
		if !part.SupportsNextPart() {
			break
		}
	}

	return result, nil
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
	} else {
		if result == nil {
			return cmd, nil
		}
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
	parts []*PartSupported
}

// NewSupportedCommands creates a new PartSupported.
func NewSupportedCommands(partValues []uint32) (*SupportedCommands, error) {
	parts := []*PartSupported{}
	index := byte(1)

	for _, val := range partValues {
		part := NewPartSupported(index)

		part.SetRawValue(val)

		parts = append(parts, part)

		index++
	}

	return &SupportedCommands{parts}, nil
}

// AddPart adds the given part to the slice of parts checked.
func (sc *SupportedCommands) AddPart(part *PartSupported) {
	sc.parts = append(sc.parts, part)
}

// GetPart gets the part at the given index.
func (sc *SupportedCommands) GetPart(index byte) (*PartSupported, error) {
	partsAmount := len(sc.parts)

	if partsAmount == 0 {
		return nil, fmt.Errorf("Cannot get part by index %d, as there are no parts", index)
	}

	if index >= byte(partsAmount) {
		return nil, fmt.Errorf("Cannot get part by index %d, there are only %d parts", index, partsAmount)
	}

	return sc.parts[index], nil
}

// GetPartByPID gets the part at the given index.
func (sc *SupportedCommands) GetPartByPID(pid OBDParameterID) (*PartSupported, error) {
	if pid == 0 {
		return sc.GetPart(0)
	}

	index := byte((pid - 1) / 0x20)

	return sc.GetPart(index)
}

// IsSupported checks if the given OBDCommand is supported.
//
// It does this by comparing the PID of the OBDCommand against the lookup table.
func (sc *SupportedCommands) IsSupported(cmd OBDCommand) bool {
	if cmd.ParameterID() == 0 {
		return true
	}

	pid := cmd.ParameterID()
	part, err := sc.GetPartByPID(pid)

	if err != nil {
		return false
	}

	return part.SupportsPID(pid)
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
		return nil, nil
	}

	return NewResult(payload)
}
