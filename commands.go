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
	SetValue(uint64)
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

func (cmd *Part1Supported) SetValue(value uint64) {
	cmd.Value = uint32(value)
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

func (cmd *EngineLoad) SetValue(value uint64) {
	cmd.Value = float32(value) / 255
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

func (cmd *CoolantTemperature) SetValue(value uint64) {
	cmd.Value = int(value) - 40
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

func (cmd *FuelPressure) SetValue(value uint64) {
	cmd.Value = uint32(value) * 3
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

func (cmd *IntakeManifoldPressure) SetValue(value uint64) {
	cmd.Value = uint32(value)
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

func (cmd *EngineRPM) SetValue(value uint64) {
	cmd.Value = float32(value) / 4
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

func (cmd *VehicleSpeed) SetValue(value uint64) {
	cmd.Value = uint32(value)
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

func (cmd *TimingAdvance) SetValue(value uint64) {
	cmd.Value = float32(value/2) - 64
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

func (cmd *IntakeAirTemperature) SetValue(value uint64) {
	cmd.Value = int(value) - 40
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

func (cmd *MafAirFlowRate) SetValue(value uint64) {
	cmd.Value = float32(value) / 100
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

func (cmd *ThrottlePosition) SetValue(value uint64) {
	cmd.Value = float32(value) / 255
}

/*==============================================================================
 * Utilities
 */

var sensorCommands = []OBDCommand{
	NewEngineLoad(),
	NewCoolantTemperature(),
	NewFuelPressure(),
	NewIntakeManifoldPressure(),
	NewEngineRPM(),
	NewVehicleSpeed(),
	NewTimingAdvance(),
	NewMafAirFlowRate(),
	NewThrottlePosition(),
}

func GetSensorCommands() []OBDCommand {
	return sensorCommands
}
