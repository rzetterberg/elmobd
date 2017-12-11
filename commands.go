package elmobd

import (
	"fmt"
)

/*==============================================================================
 * Generic types
 */

// OBDParameterID is an alias to give meaning to this particular byte.
type OBDParameterID byte

// OBDCommand is an interface that all OBD commands needs to implement to be
// able to be used with the Device.
type OBDCommand interface {
	ModeID() byte
	ParameterID() OBDParameterID
	DataWidth() byte
	Key() string
	SetValue(*Result) error
	ValueAsLit() string
	ToCommand() string
}

// BaseCommand is a simple struct with the 3 members that all OBDCommands
// will have in common.
type BaseCommand struct {
	parameterID byte
	dataWidth   byte
	key         string
}

// ModeID retrieves the mode ID of the command.
func (cmd *BaseCommand) ModeID() byte {
	return 0x01
}

// ParameterID retrieves the Parameter ID (also called PID) of the command.
func (cmd *BaseCommand) ParameterID() OBDParameterID {
	return OBDParameterID(cmd.parameterID)
}

// DataWidth retrieves the amount of bytes the command expects from the ELM327
// devices.
func (cmd *BaseCommand) DataWidth() byte {
	return cmd.dataWidth
}

// Key retrieves the unique literal key of the command, used when exporting
// commands.
func (cmd *BaseCommand) Key() string {
	return cmd.key
}

// ToCommand retrieves the raw command that can be sent to the ELM327 device.
func (cmd *BaseCommand) ToCommand() string {
	return fmt.Sprintf("%02X%02X", cmd.ModeID(), cmd.ParameterID())
}

// FloatCommand is just a shortcut for commands that retrieve floating point
// values from the ELM327 device.
type FloatCommand struct {
	Value float32
}

// ValueAsLit retrieves the value as a literal representation.
func (cmd *FloatCommand) ValueAsLit() string {
	return fmt.Sprintf("%f", cmd.Value)
}

// IntCommand is just a shortcut for commands that retrieve integer
// values from the ELM327 device.
type IntCommand struct {
	Value int
}

// ValueAsLit retrieves the value as a literal representation.
func (cmd *IntCommand) ValueAsLit() string {
	return fmt.Sprintf("%d", cmd.Value)
}

// UIntCommand is just a shortcut for commands that retrieve unsigned
// integer values from the ELM327 device.
type UIntCommand struct {
	Value uint32
}

// ValueAsLit retrieves the value as a literal representation.
func (cmd *UIntCommand) ValueAsLit() string {
	return fmt.Sprintf("%d", cmd.Value)
}

/*==============================================================================
 * Specific types
 */

// Part1Supported represents a command that checks the supported PIDs 0 to 20
type Part1Supported struct {
	BaseCommand
	UIntCommand
}

// NewPart1Supported creates a new Part1Supported.
func NewPart1Supported() *Part1Supported {
	return &Part1Supported{
		BaseCommand{0, 4, "supported_commands_part1"},
		UIntCommand{},
	}
}

// SetValue processes the byte array value into the right unsigned
// integer value.
func (cmd *Part1Supported) SetValue(result *Result) error {
	payload, err := result.PayloadAsUInt32()

	if err != nil {
		return err
	}

	cmd.Value = uint32(payload)

	return nil
}

// EngineLoad represents a command that checks the engine load in percent
//
// Min: 0.0
// Max: 1.0
type EngineLoad struct {
	BaseCommand
	FloatCommand
}

// NewEngineLoad creates a new EngineLoad with the correct parameters.
func NewEngineLoad() *EngineLoad {
	return &EngineLoad{
		BaseCommand{4, 1, "engine_load"},
		FloatCommand{},
	}
}

// SetValue processes the byte array value into the right float value.
func (cmd *EngineLoad) SetValue(result *Result) error {
	payload, err := result.PayloadAsByte()

	if err != nil {
		return err
	}

	cmd.Value = float32(payload) / 255

	return nil
}

// CoolantTemperature represents a command that checks the engine coolant
// temperature in Celsius.
//
// Min: -40
// Max: 215
type CoolantTemperature struct {
	BaseCommand
	IntCommand
}

// NewCoolantTemperature creates a new CoolantTemperature with the right
// parameters.
func NewCoolantTemperature() *CoolantTemperature {
	return &CoolantTemperature{
		BaseCommand{5, 1, "coolant_temperature"},
		IntCommand{},
	}
}

// SetValue processes the byte array value into the right integer value.
func (cmd *CoolantTemperature) SetValue(result *Result) error {
	payload, err := result.PayloadAsByte()

	if err != nil {
		return err
	}

	cmd.Value = int(payload) - 40

	return nil
}

// fuelTrim is an abstract type for fuel trim, both for short term and long term.
// Min: -100 (too rich)
// Max: 99.2 (too lean)
type fuelTrim struct {
	BaseCommand
	FloatCommand
}

// SetValue processes the byte array value into the right float value.
func (cmd *fuelTrim) SetValue(result *Result) error {
	payload, err := result.PayloadAsByte()

	if err != nil {
		return err
	}

	cmd.Value = (float32(payload) / 1.28) - 100

	return nil
}

// ShortFuelTrim1 represents a command that checks the short term fuel trim for
// bank 1.
type ShortFuelTrim1 struct {
	fuelTrim
}

// NewShortFuelTrim1 creates a new ShortFuelTrim1 with the right parameters.
func NewShortFuelTrim1() *ShortFuelTrim1 {
	return &ShortFuelTrim1{
		fuelTrim{
			BaseCommand{6, 1, "short_term_fuel_trim_bank1"},
			FloatCommand{},
		},
	}
}

// LongFuelTrim1 represents a command that checks the long term fuel trim for
// bank 1.
type LongFuelTrim1 struct {
	fuelTrim
}

// NewLongFuelTrim1 creates a new LongFuelTrim1 with the right parameters.
func NewLongFuelTrim1() *LongFuelTrim1 {
	return &LongFuelTrim1{
		fuelTrim{
			BaseCommand{7, 1, "long_term_fuel_trim_bank1"},
			FloatCommand{},
		},
	}
}

// ShortFuelTrim2 represents a command that checks the short term fuel trim for
// bank 2.
type ShortFuelTrim2 struct {
	fuelTrim
}

// NewShortFuelTrim2 creates a new ShortFuelTrim2 with the right parameters.
func NewShortFuelTrim2() *ShortFuelTrim2 {
	return &ShortFuelTrim2{
		fuelTrim{
			BaseCommand{8, 1, "short_term_fuel_trim_bank2"},
			FloatCommand{},
		},
	}
}

// LongFuelTrim2 represents a command that checks the long term fuel trim for
// bank 2.
type LongFuelTrim2 struct {
	fuelTrim
}

// NewLongFuelTrim2 creates a new LongFuelTrim2 with the right parameters.
func NewLongFuelTrim2() *LongFuelTrim2 {
	return &LongFuelTrim2{
		fuelTrim{
			BaseCommand{9, 1, "long_term_fuel_trim_bank2"},
			FloatCommand{},
		},
	}
}

// FuelPressure represents a command that checks the fuel pressure in kPa.
//
// Min: 0
// Max: 765
type FuelPressure struct {
	BaseCommand
	UIntCommand
}

// NewFuelPressure creates a new FuelPressure with the right parameters.
func NewFuelPressure() *FuelPressure {
	return &FuelPressure{
		BaseCommand{10, 1, "fuel_pressure"},
		UIntCommand{},
	}
}

// SetValue processes the byte array value into the right unsigned integer value.
func (cmd *FuelPressure) SetValue(result *Result) error {
	payload, err := result.PayloadAsByte()

	if err != nil {
		return err
	}

	cmd.Value = uint32(payload) * 3

	return nil
}

// IntakeManifoldPressure represents a command that checks the intake manifold
// pressure in kPa.
//
// Min: 0
// Max: 255
type IntakeManifoldPressure struct {
	BaseCommand
	UIntCommand
}

// NewIntakeManifoldPressure creates a new IntakeManifoldPressure with the
// right parameters.
func NewIntakeManifoldPressure() *IntakeManifoldPressure {
	return &IntakeManifoldPressure{
		BaseCommand{11, 1, "intake_manifold_pressure"},
		UIntCommand{},
	}
}

// SetValue processes the byte array value into the right unsigned integer value.
func (cmd *IntakeManifoldPressure) SetValue(result *Result) error {
	payload, err := result.PayloadAsByte()

	if err != nil {
		return err
	}

	cmd.Value = uint32(payload)

	return nil
}

// EngineRPM represents a command that checks eEngine revolutions per minute.
//
// Min: 0.0
// Max: 16383.75
type EngineRPM struct {
	BaseCommand
	FloatCommand
}

// NewEngineRPM creates a new EngineRPM with the right parameters.
func NewEngineRPM() *EngineRPM {
	return &EngineRPM{
		BaseCommand{12, 2, "engine_rpm"},
		FloatCommand{},
	}
}

// SetValue processes the byte array value into the right float value.
func (cmd *EngineRPM) SetValue(result *Result) error {
	payload, err := result.PayloadAsUInt16()

	if err != nil {
		return err
	}

	cmd.Value = float32(payload) / 4

	return nil
}

// VehicleSpeed represents a command that checks the vehicle speed in km/h.
//
// Min: 0
// Max: 255
type VehicleSpeed struct {
	BaseCommand
	UIntCommand
}

// NewVehicleSpeed creates a new VehicleSpeed with the right parameters
func NewVehicleSpeed() *VehicleSpeed {
	return &VehicleSpeed{
		BaseCommand{13, 1, "vehicle_speed"},
		UIntCommand{},
	}
}

// SetValue processes the byte array value into the right unsigned integer value.
func (cmd *VehicleSpeed) SetValue(result *Result) error {
	payload, err := result.PayloadAsByte()

	if err != nil {
		return err
	}

	cmd.Value = uint32(payload)

	return nil
}

// TimingAdvance represents a command that checks the timing advance in degrees
// before TDC.
//
// Min: -64
// Max: 63.5
//
// For more info about TDC:
// https://en.wikipedia.org/wiki/Dead_centre_(engineering)
type TimingAdvance struct {
	BaseCommand
	FloatCommand
}

// NewTimingAdvance creates a new TimingAdvance with the right parameters.
func NewTimingAdvance() *TimingAdvance {
	return &TimingAdvance{
		BaseCommand{14, 1, "timing_advance"},
		FloatCommand{},
	}
}

// SetValue processes the byte array value into the right float value.
func (cmd *TimingAdvance) SetValue(result *Result) error {
	payload, err := result.PayloadAsByte()

	if err != nil {
		return err
	}

	cmd.Value = float32(payload/2) - 64

	return nil
}

// IntakeAirTemperature represents a command that checks the intake air
// temperature in Celsius.
//
// Min: -40
// Max: 215
type IntakeAirTemperature struct {
	BaseCommand
	IntCommand
}

// NewIntakeAirTemperature creates a new IntakeAirTemperature with the right parameters.
func NewIntakeAirTemperature() *IntakeAirTemperature {
	return &IntakeAirTemperature{
		BaseCommand{15, 1, "intake_air_temperature"},
		IntCommand{},
	}
}

// SetValue processes the byte array value into the right integer value.
func (cmd *IntakeAirTemperature) SetValue(result *Result) error {
	payload, err := result.PayloadAsByte()

	if err != nil {
		return err
	}

	cmd.Value = int(payload) - 40

	return nil
}

// MafAirFlowRate represents a command that checks the mass Air Flow sensor
// flow rate grams/second.
//
// Min: 0
// Max: 655.35
//
// More information about MAF:
// https://en.wikipedia.org/wiki/Mass_flow_sensor
type MafAirFlowRate struct {
	BaseCommand
	FloatCommand
}

// NewMafAirFlowRate creates a new MafAirFlowRate with the right parameters.
func NewMafAirFlowRate() *MafAirFlowRate {
	return &MafAirFlowRate{
		BaseCommand{16, 2, "maf_air_flow_rate"},
		FloatCommand{},
	}
}

// SetValue processes the byte array value into the right float value.
func (cmd *MafAirFlowRate) SetValue(result *Result) error {
	payload, err := result.PayloadAsUInt16()

	if err != nil {
		return err
	}

	cmd.Value = float32(payload) / 100

	return nil
}

// ThrottlePosition represents a command that checks the throttle position in
// percentage.
//
// Min: 0.0
// Max: 100.0
type ThrottlePosition struct {
	BaseCommand
	FloatCommand
}

// NewThrottlePosition creates a new ThrottlePosition with the right parameters.
func NewThrottlePosition() *ThrottlePosition {
	return &ThrottlePosition{
		BaseCommand{17, 1, "throttle_position"},
		FloatCommand{},
	}
}

// SetValue processes the byte array value into the right float value.
func (cmd *ThrottlePosition) SetValue(result *Result) error {
	payload, err := result.PayloadAsByte()

	if err != nil {
		return err
	}

	cmd.Value = float32(payload) / 255

	return nil
}

// OBDStandards represents a command that checks the OBD standards this vehicle
// conforms to as a single decimal value:
//
// - 1       OBD-II as defined by the CARB
// - 2       OBD as defined by the EPA
// - 3       OBD and OBD-II
// - 4       OBD-I
// - 5       Not OBD compliant
// - 6       EOBD (Europe)
// - 7       EOBD and OBD-II
// - 8       EOBD and OBD
// - 9       EOBD, OBD and OBD II
// - 10      JOBD (Japan)
// - 11      JOBD and OBD II
// - 12      JOBD and EOBD
// - 13      JOBD, EOBD, and OBD II
// - 14      Reserved
// - 15      Reserved
// - 16      Reserved
// - 17      Engine Manufacturer Diagnostics (EMD)
// - 18      Engine Manufacturer Diagnostics Enhanced (EMD+)
// - 19      Heavy Duty On-Board Diagnostics (Child/Partial) (HD OBD-C)
// - 20      Heavy Duty On-Board Diagnostics (HD OBD)
// - 21      World Wide Harmonized OBD (WWH OBD)
// - 22      Reserved
// - 23      Heavy Duty Euro OBD Stage I without NOx control (HD EOBD-I)
// - 24      Heavy Duty Euro OBD Stage I with NOx control (HD EOBD-I N)
// - 25      Heavy Duty Euro OBD Stage II without NOx control (HD EOBD-II)
// - 26      Heavy Duty Euro OBD Stage II with NOx control (HD EOBD-II N)
// - 27      Reserved
// - 28      Brazil OBD Phase 1 (OBDBr-1)
// - 29      Brazil OBD Phase 2 (OBDBr-2)
// - 30      Korean OBD (KOBD)
// - 31      India OBD I (IOBD I)
// - 32      India OBD II (IOBD II)
// - 33      Heavy Duty Euro OBD Stage VI (HD EOBD-IV)
// - 34-250  Reserved
// - 251-255 Not available for assignment (SAE J1939 special meaning)
type OBDStandards struct {
	BaseCommand
	UIntCommand
}

// NewOBDStandards creates a new OBDStandards with the right parameters.
func NewOBDStandards() *OBDStandards {
	return &OBDStandards{
		BaseCommand{28, 1, "obd_standards"},
		UIntCommand{},
	}
}

// SetValue processes the byte array value into the right unsigned integer
// value.
func (cmd *OBDStandards) SetValue(result *Result) error {
	payload, err := result.PayloadAsByte()

	if err != nil {
		return err
	}

	cmd.Value = uint32(payload)

	return nil
}

// RuntimeSinceStart represents a command that checks the run time since engine
// start.
//
// Min: 0
// Max: 65535
type RuntimeSinceStart struct {
	BaseCommand
	UIntCommand
}

// NewRuntimeSinceStart creates a new RuntimeSinceStart with the right
// parameters.
func NewRuntimeSinceStart() *RuntimeSinceStart {
	return &RuntimeSinceStart{
		BaseCommand{31, 1, "runtime_since_engine_start"},
		UIntCommand{},
	}
}

// SetValue processes the byte array value into the right unsigned integer
// value.
func (cmd *RuntimeSinceStart) SetValue(result *Result) error {
	payload, err := result.PayloadAsUInt16()

	if err != nil {
		return err
	}

	cmd.Value = uint32(payload)

	return nil
}

// Part2Supported represents a command that checks the supported PIDs 21 to 40.
type Part2Supported struct {
	BaseCommand
	UIntCommand
}

// NewPart2Supported creates a new Part2Supported with the right parameters.
func NewPart2Supported() *Part2Supported {
	return &Part2Supported{
		BaseCommand{32, 4, "supported_commands_part2"},
		UIntCommand{},
	}
}

// SetValue processes the byte array value into the right unsigned integer
// value.
func (cmd *Part2Supported) SetValue(result *Result) error {
	payload, err := result.PayloadAsUInt32()

	if err != nil {
		return err
	}

	cmd.Value = uint32(payload)

	return nil
}

// Part3Supported represents a command that checks the supported PIDs 41 to 60.
type Part3Supported struct {
	BaseCommand
	UIntCommand
}

// NewPart3Supported creates a new Part3Supported with the right parameters.
func NewPart3Supported() *Part3Supported {
	return &Part3Supported{
		BaseCommand{64, 4, "supported_commands_part3"},
		UIntCommand{},
	}
}

// SetValue processes the byte array value into the right unsigned integer
// value.
func (cmd *Part3Supported) SetValue(result *Result) error {
	payload, err := result.PayloadAsUInt32()

	if err != nil {
		return err
	}

	cmd.Value = uint32(payload)

	return nil
}

// Part4Supported represents a command that checks the supported PIDs 61 to 80.
type Part4Supported struct {
	BaseCommand
	UIntCommand
}

// NewPart4Supported creates a new Part4Supported with the right parameters.
func NewPart4Supported() *Part4Supported {
	return &Part4Supported{
		BaseCommand{96, 4, "supported_commands_part4"},
		UIntCommand{},
	}
}

// SetValue processes the byte array value into the right unsigned integer
// value.
func (cmd *Part4Supported) SetValue(result *Result) error {
	payload, err := result.PayloadAsUInt32()

	if err != nil {
		return err
	}

	cmd.Value = uint32(payload)

	return nil
}

// Part5Supported represents a command that checks the supported PIDs 81 to A0.
type Part5Supported struct {
	BaseCommand
	UIntCommand
}

// NewPart5Supported creates a new Part5Supported with the right parameters..
func NewPart5Supported() *Part5Supported {
	return &Part5Supported{
		BaseCommand{128, 4, "supported_commands_part5"},
		UIntCommand{},
	}
}

// SetValue processes the byte array value into the right unsigned integer
// value.
func (cmd *Part5Supported) SetValue(result *Result) error {
	payload, err := result.PayloadAsUInt32()

	if err != nil {
		return err
	}

	cmd.Value = uint32(payload)

	return nil
}

/*==============================================================================
 * Utilities
 */

var sensorCommands = []OBDCommand{
	NewEngineLoad(),
	NewCoolantTemperature(),
	NewShortFuelTrim1(),
	NewLongFuelTrim1(),
	NewShortFuelTrim2(),
	NewLongFuelTrim2(),
	NewFuelPressure(),
	NewIntakeManifoldPressure(),
	NewEngineRPM(),
	NewVehicleSpeed(),
	NewTimingAdvance(),
	NewMafAirFlowRate(),
	NewThrottlePosition(),
	NewOBDStandards(),
	NewRuntimeSinceStart(),
}

// GetSensorCommands returns all the defined commands that are not commands
// that check command availability on the connected car.
func GetSensorCommands() []OBDCommand {
	return sensorCommands
}
