package elmobd

import (
	"fmt"
)

/*==============================================================================
 * Generic types
 */

type OBDParameterId byte

type OBDCommand interface {
	ModeId() byte
	ParameterId() OBDParameterId
	DataWidth() byte
	Key() string
	SetValue(*Result) error
	ValueAsLit() string
	ToCommand() string
}

type BaseCommand struct {
	parameterId byte
	dataWidth   byte
	key         string
}

func (cmd *BaseCommand) ModeId() byte {
	return 0x01
}

func (cmd *BaseCommand) ParameterId() OBDParameterId {
	return OBDParameterId(cmd.parameterId)
}

func (cmd *BaseCommand) DataWidth() byte {
	return cmd.dataWidth
}

func (cmd *BaseCommand) Key() string {
	return cmd.key
}

func (cmd *BaseCommand) ToCommand() string {
	return fmt.Sprintf("%02X%02X", cmd.ModeId(), cmd.ParameterId())
}

type FloatCommand struct {
	Value float32
}

func (cmd *FloatCommand) ValueAsLit() string {
	return fmt.Sprintf("%f", cmd.Value)
}

type IntCommand struct {
	Value int
}

func (cmd *IntCommand) ValueAsLit() string {
	return fmt.Sprintf("%d", cmd.Value)
}

type UIntCommand struct {
	Value uint32
}

func (cmd *UIntCommand) ValueAsLit() string {
	return fmt.Sprintf("%d", cmd.Value)
}

/*==============================================================================
 * Specific types
 */

// Supported PIDs 0 to 20
type Part1Supported struct {
	BaseCommand
	UIntCommand
}

func NewPart1Supported() *Part1Supported {
	return &Part1Supported{
		BaseCommand{0, 4, "supported_commands_part1"},
		UIntCommand{},
	}
}

func (cmd *Part1Supported) SetValue(result *Result) error {
	payload, err := result.PayloadAsUInt32()

	if err != nil {
		return err
	}

	cmd.Value = uint32(payload)

	return nil
}

// Engine load in percent
// Min: 0.0
// Max: 1.0
type EngineLoad struct {
	BaseCommand
	FloatCommand
}

func NewEngineLoad() *EngineLoad {
	return &EngineLoad{
		BaseCommand{4, 1, "engine_load"},
		FloatCommand{},
	}
}

func (cmd *EngineLoad) SetValue(result *Result) error {
	payload, err := result.PayloadAsByte()

	if err != nil {
		return err
	}

	cmd.Value = float32(payload) / 255

	return nil
}

// Engine coolant temperature in Celsius
// Min: -40
// Max: 215
type CoolantTemperature struct {
	BaseCommand
	IntCommand
}

func NewCoolantTemperature() *CoolantTemperature {
	return &CoolantTemperature{
		BaseCommand{5, 1, "coolant_temperature"},
		IntCommand{},
	}
}

func (cmd *CoolantTemperature) SetValue(result *Result) error {
	payload, err := result.PayloadAsByte()

	if err != nil {
		return err
	}

	cmd.Value = int(payload) - 40

	return nil
}

// Abstract type for fuel trim, both for short term and long term.
// Min: -100 (too rich)
// Max: 99.2 (too lean)
type fuelTrim struct {
	BaseCommand
	FloatCommand
}

func (cmd *fuelTrim) SetValue(result *Result) error {
	payload, err := result.PayloadAsByte()

	if err != nil {
		return err
	}

	cmd.Value = (float32(payload) / 1.28) - 100

	return nil
}

// Short term fuel trim for bank 1
type ShortFuelTrim1 struct {
	fuelTrim
}

func NewShortFuelTrim1() *ShortFuelTrim1 {
	return &ShortFuelTrim1{
		fuelTrim{
			BaseCommand{6, 1, "short_term_fuel_trim_bank1"},
			FloatCommand{},
		},
	}
}

// Long term fuel trim for bank 1
type LongFuelTrim1 struct {
	fuelTrim
}

func NewLongFuelTrim1() *LongFuelTrim1 {
	return &LongFuelTrim1{
		fuelTrim{
			BaseCommand{7, 1, "long_term_fuel_trim_bank1"},
			FloatCommand{},
		},
	}
}

// Short term fuel trim for bank 2
type ShortFuelTrim2 struct {
	fuelTrim
}

func NewShortFuelTrim2() *ShortFuelTrim2 {
	return &ShortFuelTrim2{
		fuelTrim{
			BaseCommand{8, 1, "short_term_fuel_trim_bank2"},
			FloatCommand{},
		},
	}
}

// Long term fuel trim for bank 2
type LongFuelTrim2 struct {
	fuelTrim
}

func NewLongFuelTrim2() *LongFuelTrim2 {
	return &LongFuelTrim2{
		fuelTrim{
			BaseCommand{9, 1, "long_term_fuel_trim_bank2"},
			FloatCommand{},
		},
	}
}

// Fuel pressure in kPa
// Min: 0
// Max: 765
type FuelPressure struct {
	BaseCommand
	UIntCommand
}

func NewFuelPressure() *FuelPressure {
	return &FuelPressure{
		BaseCommand{10, 1, "fuel_pressure"},
		UIntCommand{},
	}
}

func (cmd *FuelPressure) SetValue(result *Result) error {
	payload, err := result.PayloadAsByte()

	if err != nil {
		return err
	}

	cmd.Value = uint32(payload) * 3

	return nil
}

// Intake manifold pressure in kPa
// Min: 0
// Max: 255
type IntakeManifoldPressure struct {
	BaseCommand
	UIntCommand
}

func NewIntakeManifoldPressure() *IntakeManifoldPressure {
	return &IntakeManifoldPressure{
		BaseCommand{11, 1, "intake_manifold_pressure"},
		UIntCommand{},
	}
}

func (cmd *IntakeManifoldPressure) SetValue(result *Result) error {
	payload, err := result.PayloadAsByte()

	if err != nil {
		return err
	}

	cmd.Value = uint32(payload)

	return nil
}

// Engine revolutions per minute
// Min: 0.0
// Max: 16383.75
type EngineRPM struct {
	BaseCommand
	FloatCommand
}

func NewEngineRPM() *EngineRPM {
	return &EngineRPM{
		BaseCommand{12, 2, "engine_rpm"},
		FloatCommand{},
	}
}

func (cmd *EngineRPM) SetValue(result *Result) error {
	payload, err := result.PayloadAsUInt16()

	if err != nil {
		return err
	}

	cmd.Value = float32(payload) / 4

	return nil
}

// Vechile speed in km/h
// Min: 0
// Max: 255
type VehicleSpeed struct {
	BaseCommand
	UIntCommand
}

func NewVehicleSpeed() *VehicleSpeed {
	return &VehicleSpeed{
		BaseCommand{13, 1, "vehicle_speed"},
		UIntCommand{},
	}
}

func (cmd *VehicleSpeed) SetValue(result *Result) error {
	payload, err := result.PayloadAsByte()

	if err != nil {
		return err
	}

	cmd.Value = uint32(payload)

	return nil
}

// Timing advance in degrees before TDC
// Min: -64
// Max: 63.5
//
// https://en.wikipedia.org/wiki/Dead_centre_(engineering)
type TimingAdvance struct {
	BaseCommand
	FloatCommand
}

func NewTimingAdvance() *TimingAdvance {
	return &TimingAdvance{
		BaseCommand{14, 1, "timing_advance"},
		FloatCommand{},
	}
}

func (cmd *TimingAdvance) SetValue(result *Result) error {
	payload, err := result.PayloadAsByte()

	if err != nil {
		return err
	}

	cmd.Value = float32(payload/2) - 64

	return nil
}

// Intake air temperature in Celsius
// Min: -40
// Max: 215
type IntakeAirTemperature struct {
	BaseCommand
	IntCommand
}

func NewIntakeAirTemperature() *IntakeAirTemperature {
	return &IntakeAirTemperature{
		BaseCommand{15, 1, "intake_air_temperature"},
		IntCommand{},
	}
}

func (cmd *IntakeAirTemperature) SetValue(result *Result) error {
	payload, err := result.PayloadAsByte()

	if err != nil {
		return err
	}

	cmd.Value = int(payload) - 40

	return nil
}

// Mass Air Flow sensor flow rate grams/second
// Min: 0
// Max: 655.35
// https://en.wikipedia.org/wiki/Mass_flow_sensor
type MafAirFlowRate struct {
	BaseCommand
	FloatCommand
}

func NewMafAirFlowRate() *MafAirFlowRate {
	return &MafAirFlowRate{
		BaseCommand{16, 2, "maf_air_flow_rate"},
		FloatCommand{},
	}
}

func (cmd *MafAirFlowRate) SetValue(result *Result) error {
	payload, err := result.PayloadAsUInt16()

	if err != nil {
		return err
	}

	cmd.Value = float32(payload) / 100

	return nil
}

// Throttle position in percentage
// Min: 0.0
// Max: 100.0
type ThrottlePosition struct {
	BaseCommand
	FloatCommand
}

func NewThrottlePosition() *ThrottlePosition {
	return &ThrottlePosition{
		BaseCommand{17, 1, "throttle_position"},
		FloatCommand{},
	}
}

func (cmd *ThrottlePosition) SetValue(result *Result) error {
	payload, err := result.PayloadAsByte()

	if err != nil {
		return err
	}

	cmd.Value = float32(payload) / 255

	return nil
}

// OBD standards this vehicle conforms to as a single decimal value:
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

func NewOBDStandards() *OBDStandards {
	return &OBDStandards{
		BaseCommand{28, 1, "obd_standards"},
		UIntCommand{},
	}
}

func (cmd *OBDStandards) SetValue(result *Result) error {
	payload, err := result.PayloadAsByte()

	if err != nil {
		return err
	}

	cmd.Value = uint32(payload)

	return nil
}

// Run time since engine start
// Min: 0
// Max: 65535
type RuntimeSinceStart struct {
	BaseCommand
	UIntCommand
}

func NewRuntimeSinceStart() *RuntimeSinceStart {
	return &RuntimeSinceStart{
		BaseCommand{31, 1, "runtime_since_engine_start"},
		UIntCommand{},
	}
}

func (cmd *RuntimeSinceStart) SetValue(result *Result) error {
	payload, err := result.PayloadAsUInt16()

	if err != nil {
		return err
	}

	cmd.Value = uint32(payload)

	return nil
}

// Supported PIDs 21 to 40
type Part2Supported struct {
	BaseCommand
	UIntCommand
}

func NewPart2Supported() *Part2Supported {
	return &Part2Supported{
		BaseCommand{32, 4, "supported_commands_part2"},
		UIntCommand{},
	}
}

func (cmd *Part2Supported) SetValue(result *Result) error {
	payload, err := result.PayloadAsUInt32()

	if err != nil {
		return err
	}

	cmd.Value = uint32(payload)

	return nil
}

// Supported PIDs 41 to 60
type Part3Supported struct {
	BaseCommand
	UIntCommand
}

func NewPart3Supported() *Part3Supported {
	return &Part3Supported{
		BaseCommand{64, 4, "supported_commands_part3"},
		UIntCommand{},
	}
}

func (cmd *Part3Supported) SetValue(result *Result) error {
	payload, err := result.PayloadAsUInt32()

	if err != nil {
		return err
	}

	cmd.Value = uint32(payload)

	return nil
}

// Supported PIDs 61 to 80
type Part4Supported struct {
	BaseCommand
	UIntCommand
}

func NewPart4Supported() *Part4Supported {
	return &Part4Supported{
		BaseCommand{96, 4, "supported_commands_part4"},
		UIntCommand{},
	}
}

func (cmd *Part4Supported) SetValue(result *Result) error {
	payload, err := result.PayloadAsUInt32()

	if err != nil {
		return err
	}

	cmd.Value = uint32(payload)

	return nil
}

// Supported PIDs 81 to A0
type Part5Supported struct {
	BaseCommand
	UIntCommand
}

func NewPart5Supported() *Part5Supported {
	return &Part5Supported{
		BaseCommand{128, 4, "supported_commands_part5"},
		UIntCommand{},
	}
}

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

func GetSensorCommands() []OBDCommand {
	return sensorCommands
}
