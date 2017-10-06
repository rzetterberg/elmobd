package elmobd

import (
	"fmt"
	"strings"
	"strconv"
)

/*==============================================================================
 * External
 */

type Result struct {
	value []byte
}

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

func (res *Result) Validate(cmd OBDCommand) error {
	valueLen := len(res.value)
	expLen   := int(cmd.DataWidth()+2)

	if valueLen != expLen {
		return fmt.Errorf(
			"Expected %d bytes, found %d",
			expLen,
			valueLen,
		)
	}

	modeResp := cmd.ModeId()+0x40

	if res.value[0] != modeResp {
		return fmt.Errorf(
			"Expected mode echo %02X, got %02X",
			modeResp,
			res.value[0],
		)
	}

	if OBDParameterId(res.value[1]) != cmd.ParameterId() {
		return fmt.Errorf(
			"Expected parameter echo %02X got %02X",
			cmd.ParameterId(),
			res.value[1],
		)
	}

	return nil
}

func (res *Result) payloadAsUInt(expAmount int) (uint64, error) {
	var result uint64

	payload := res.value[2:]
	amount  := len(payload)

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

func (res *Result) PayloadAsUInt64() (uint64, error) {
	result, err := res.payloadAsUInt(8)

	if err != nil {
		return 0, err
	}

	return uint64(result), nil
}

func (res *Result) PayloadAsUInt32() (uint32, error) {
	result, err := res.payloadAsUInt(4)

	if err != nil {
		return 0, err
	}

	return uint32(result), nil
}

func (res *Result) PayloadAsUInt16() (uint16, error) {
	result, err := res.payloadAsUInt(2)

	if err != nil {
		return 0, err
	}

	return uint16(result), nil
}

func (res *Result) PayloadAsByte() (byte, error) {
	result, err := res.payloadAsUInt(1)

	if err != nil {
		return 0, err
	}

	return uint8(result), nil
}

type Device struct {
	rawDevice   *RawDevice
	outputDebug bool
}

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

type SupportedCommands struct {
	part1 uint32
	part2 uint32
	part3 uint32
	part4 uint32
	part5 uint32
}

func (sc *SupportedCommands) IsSupported(cmd OBDCommand) bool {
	if cmd.ParameterId() == 0 {
		return true
	}

	pid := cmd.ParameterId()

	inPart1 := (sc.part1 >> uint32(32-pid)) & 1
	inPart2 := (sc.part2 >> uint32(32-pid)) & 1
	inPart3 := (sc.part3 >> uint32(32-pid)) & 1
	inPart4 := (sc.part4 >> uint32(32-pid)) & 1
	inPart5 := (sc.part5 >> uint32(32-pid)) & 1

	return (inPart1 == 1) || (inPart2 == 1) ||
		(inPart3 == 1) || (inPart4 == 1) || (inPart5 == 1)
}

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
