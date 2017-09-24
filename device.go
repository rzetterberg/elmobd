package elmobd

import (
	"fmt"
	"strings"
)

/*==============================================================================
 * Device
 */

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

	result := SupportedCommands{
		uint32(part1.(*Part1Supported).Value),
	}

	return &result, nil
}

func (dev *Device) CheckSupportedPart(cmd OBDCommand) (OBDCommand, error) {
	res := dev.rawDevice.RunCommand(cmd.ToCommand())

	if res.Error != nil {
		return cmd, res.Error
	}

	if dev.outputDebug {
		fmt.Println(res.FormatOverview())
	}

	data, err := parseSupportedResponse(cmd, res.Outputs)

	if err != nil {
		return cmd, err
	}

	cmd.SetValue(data)

	return cmd, nil
}

func (dev *Device) RunOBDCommand(cmd OBDCommand) (OBDCommand, error) {
	res := dev.rawDevice.RunCommand(cmd.ToCommand())

	if res.Error != nil {
		return cmd, res.Error
	}

	if dev.outputDebug {
		fmt.Println(res.FormatOverview())
	}

	data, err := parseOBDResponse(cmd, res.Outputs)

	if err != nil {
		return cmd, err
	}

	cmd.SetValue(data)

	return cmd, nil
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

/*==============================================================================
 * Supported commands
 */

type SupportedCommands struct {
	part1 uint32
}

func (sc *SupportedCommands) IsSupported(cmd OBDCommand) bool {
	if cmd.ParameterId() == 0 {
		return true
	}

	pid := cmd.ParameterId()

	inPart1 := (sc.part1 >> uint32(32-pid)) & 1

	return inPart1 == 1
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

func parseSupportedResponse(cmd OBDCommand, outputs []string) (uint64, error) {
	if len(outputs) < 2 {
		return 0, fmt.Errorf(
			"Expected more than one output, got: %q",
			outputs,
		)
	}

	if outputs[1] == "UNABLE TO CONNECT" {
		return 0, fmt.Errorf(
			"Unable to connect to car, is the ignition on?",
		)
	}

	return parseOBDResponse(cmd, outputs[1:])
}

func parseOBDResponse(cmd OBDCommand, outputs []string) (uint64, error) {
	if len(outputs) < 1 {
		return 0, fmt.Errorf(
			"Expected atleast one output, got: %q",
			outputs,
		)
	}

	payload := outputs[0]

	if payload == "UNABLE TO CONNECT" {
		return 0, fmt.Errorf(
			"Unable to connect to car, is the ignition on?",
		)
	}

	if payload == "NO DATA" {
		return 0, fmt.Errorf(
			"No data from car, time out from elm device?",
		)
	}

	hexLiterals := strings.Split(payload, " ")

	if len(hexLiterals) != int(cmd.DataWidth()+2) {
		return 0, fmt.Errorf(
			"Expected %d hex literals, got %d",
			cmd.DataWidth()+2,
			len(hexLiterals),
		)
	}

	modeResp := fmt.Sprintf("%02X", cmd.ModeId()+0x40)

	if hexLiterals[0] != modeResp {
		return 0, fmt.Errorf(
			"Expected mode echo %s, got %s",
			modeResp,
			hexLiterals[0],
		)
	}

	expParamEcho := fmt.Sprintf("%02X", cmd.ParameterId())

	if hexLiterals[1] != expParamEcho {
		return 0, fmt.Errorf(
			"Expected parameter echo %s got %s",
			expParamEcho,
			hexLiterals[1],
		)
	}

	bytes, err := HexLitsToBytes(hexLiterals[2:])

	if err != nil {
		return 0, err
	}

	return BytesToUint64(bytes)
}
