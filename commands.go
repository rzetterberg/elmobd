package elmobd

import (
	"fmt"
	"math"
)

const SERVICE_01_ID = 0x01
const SERVICE_04_ID = 0x04

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

// baseCommand is a simple struct with the 3 members that all OBDCommands
// will have in common.
type baseCommand struct {
	modeId      byte
	parameterID OBDParameterID
	dataWidth   byte
	key         string
}

// ResultLessCommand is a command implementing empty functions for setting values (commands without results)
type ResultLessCommand struct {
}

// ModeID retrieves the mode ID of the command.
func (cmd *baseCommand) ModeID() byte {
	return cmd.modeId
}

// ParameterID retrieves the Parameter ID (also called PID) of the command.
func (cmd *baseCommand) ParameterID() OBDParameterID {
	return OBDParameterID(cmd.parameterID)
}

// DataWidth retrieves the amount of bytes the command expects from the ELM327
// devices.
func (cmd *baseCommand) DataWidth() byte {
	return cmd.dataWidth
}

// Key retrieves the unique literal key of the command, used when exporting
// commands.
func (cmd *baseCommand) Key() string {
	return cmd.key
}

// ToCommand retrieves the raw command that can be sent to the ELM327 device.
//
// The command is sent without spaces between the parts, the amount of data
// lines is added to the end of the command to speed up the communication.
// See page 33 of the ELM327 data sheet for details on why we do this.
func (cmd *baseCommand) ToCommand() string {
	dataLines := float64(cmd.DataWidth()) / 4.0

	return fmt.Sprintf(
		"%02X%02X%1X",
		cmd.ModeID(),
		cmd.ParameterID(),
		byte(math.Ceil(dataLines)),
	)
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

// PartSupported represents a command that checks which 31 PIDs are supported
// of a part.
//
// All PIDs are divided into parts 7 parts with the following PIDs:
//
// - Part 1 (0x00): 0x01 to 0x20
// - Part 2 (0x20): 0x21 to 0x40
// - Part 3 (0x40): 0x41 to 0x60
// - Part 4 (0x60): 0x61 to 0x80
// - Part 5 (0x80): 0x81 to 0xA0
// - Part 6 (0xA0): 0xA1 to 0xC0
// - Part 7 (0xC0): 0xC1 to 0xE0
//
// PID 0x00 checks which PIDs that are supported of part 1, after that, the
// last PID of each part checks the whether the next part is supported.
//
// So PID 0x20 of Part 1 checks which PIDs in part 2 are supported, PID 0x40 of
// part 2 checks which PIDs in part 3 are supported, etc etc.
type PartSupported struct {
	baseCommand
	UIntCommand
	index byte
}

// PartRange represents how many PIDs there are in one part
const PartRange = 0x20

// NewPartSupported creates a new PartSupported.
func NewPartSupported(index byte) *PartSupported {
	if index < 1 {
		index = 1
	} else if index > 7 {
		index = 7
	}

	pid := OBDParameterID((index - 1) * PartRange)

	return &PartSupported{
		baseCommand{SERVICE_01_ID, pid, 4, fmt.Sprintf("supported_commands_part%d", index)},
		UIntCommand{},
		index,
	}
}

// SetValue processes the byte array value into the right unsigned
// integer value.
func (part *PartSupported) SetValue(result *Result) error {
	payload, err := result.PayloadAsUInt32()

	if err != nil {
		return err
	}

	part.Value = uint32(payload)

	return nil
}

// SetRawValue sets the raw value directly without any validation or parsing.
func (part *PartSupported) SetRawValue(val uint32) {
	part.Value = val
}

// PIDInRange checks if the given is in range of the current part.
//
// For example, if the current part is 1 (0x01 to 0x20) and the given command
// has a PID of 0x10, then this function returns true.
//
// If the given command has a PID of 0x31 then this function returns false.
func (part *PartSupported) PIDInRange(comparePID OBDParameterID) bool {
	endPID := OBDParameterID(part.index * PartRange)
	startPID := OBDParameterID(endPID-PartRange) + 1

	return startPID <= comparePID && comparePID <= endPID
}

// CommandInRange checks if the PID of the given command is in range of the
// current part.
//
// For example, if the current part is 1 (0x01 to 0x20) and the given command
// has a PID of 0x10, then this function returns true.
//
// If the given command has a PID of 0x31 then this function returns false.
func (part *PartSupported) CommandInRange(cmd OBDCommand) bool {
	return part.PIDInRange(cmd.ParameterID())
}

// SupportsPID checks if the given command is supported in the current part.
//
// To figure out if a PID is supported we need to understand what the Value
// of a PartSupported represents.
//
// It represents the supported/not supported state of 32 PIDs.
//
// It does this by encoding this information as 32 bits, where each bit
// represents the state of a PID:
//
// - When the bit is set, it represents the PID being supported
// - When the bit is unset, it represents the PID being unsupported
//
// To make it easier to map PID values to actual bits, Bit-Encoded-Notation
// is used, where each bit has a name. Each name as a letter representing
// the byte and a number representing the bit in the byte.
//
// Here's how Bit-Encoded-Notation maps against the bits:
//
// A7    A0 B7    B0 C7    C0 D7     D0
// |      | |      | |      | |      |
// v      v v      v v      v v      v
// 00000000 00000000 00000000 00000000
//
// The state of the first PID is kept at bit A7, the second PID at A6, all
// the way until we get to PID 0x20 (32) which is kept in bit D0.
//
// In order to check if a bit is active, we can either:
//
// - Shift the bits of the value to the right until the bit we want to check
//   has the position D0 and then use a AND bitwise conditional with the mask 0x1
// - Shift the bits of the mask 0x1 to the left until it has the same position as
//   the bit we want to check and then use a AND bitwise conditional with value
//
// This function uses the first method of checking if the bit is active.
//
// In order to figure out which bit position has, we take the position of the
// first PID, which is 32 and subtract the PID number to get the bit position.
// This means that 32 - PID 1 = 31 (A7) and 32 - PID 32 = 0 (D0).
//
// Now we know how to figure out what bit holds the information and how to read
// the information from the bit.
//
// This works well for checking if part 1 supported PIDs between 0x1 and 0x20,
// but it fails for parts 2,3,4,5,6,7 and PIDs about 0x20.
//
// In order to make this work for other parts besides 1, we simply normalize the
// PID number by removing 32 times the part index, instead of hard coding the value
// to subtract to 32.
func (part *PartSupported) SupportsPID(comparePID OBDParameterID) bool {
	if !part.PIDInRange(comparePID) {
		return false
	}

	offset := uint32(32 * part.index)
	bitsToShift := offset - uint32(comparePID)
	result := (part.Value >> bitsToShift) & 1

	return result == 1
}

// SupportsNextPart checks if the PID that is used to check the next part is
// supported. This PID is always the last PID of the current part, which means
// we can simply check if the D0 bit is set.
func (part *PartSupported) SupportsNextPart() bool {
	result := part.Value & 1

	return result == 1
}

// SupportsCommand checks if the given command is supported in the current part.
//
// Note: returns false if the given PID is not handled in the current part
func (part *PartSupported) SupportsCommand(cmd OBDCommand) bool {
	return part.SupportsPID(cmd.ParameterID())
}

// Index returns the part index.
func (part *PartSupported) Index() byte {
	return part.index
}

// MonitorStatus represents a command that checks the status since DTCs
// were cleared last time. This includes the MIL status and the amount of
// DTCs.
type MonitorStatus struct {
	baseCommand
	MilActive bool
	DtcAmount byte
}

// ValueAsLit retrieves the value as a literal representation.
func (cmd *MonitorStatus) ValueAsLit() string {
	return fmt.Sprintf(
		"{\"mil_active\": %t, \"dts_amount\": %d",
		cmd.MilActive,
		cmd.DtcAmount,
	)
}

// NewMonitorStatus creates a new MonitorStatus.
func NewMonitorStatus() *MonitorStatus {
	return &MonitorStatus{
		baseCommand{SERVICE_01_ID, 1, 4, "monitor_status"},
		false,
		0,
	}
}

// SetValue processes the byte array value into the right unsigned
// integer value.
func (cmd *MonitorStatus) SetValue(result *Result) error {
	expAmount := 4
	payload := result.value[2:]
	amount := len(payload)

	if amount != expAmount {
		return fmt.Errorf(
			"Expected %d bytes of payload, got %d", expAmount, amount,
		)
	}

	// 0x80 is the MSB: 0b10000000
	cmd.MilActive = (payload[0] & 0x80) == 0x80
	// 0x7F everything but the MSB: 0b01111111
	cmd.DtcAmount = byte(payload[0] & 0x7F)

	return nil
}

// EngineLoad represents a command that checks the engine load in percent
//
// Min: 0.0
// Max: 1.0
type EngineLoad struct {
	baseCommand
	FloatCommand
}

// NewEngineLoad creates a new EngineLoad with the correct parameters.
func NewEngineLoad() *EngineLoad {
	return &EngineLoad{
		baseCommand{SERVICE_01_ID, 4, 1, "engine_load"},
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

// Fuel represents a command that checks the fuel quantity in percent
//
// Min: 0.0
// Max: 1.0
type Fuel struct {
	baseCommand
	FloatCommand
}

// NewFuel creates a new Fuel with the correct parameters.
func NewFuel() *Fuel {
	return &Fuel{
		baseCommand{SERVICE_01_ID, 0x2f, 1, "fuel"},
		FloatCommand{},
	}
}

// SetValue processes the byte array value into the right float value.
func (cmd *Fuel) SetValue(result *Result) error {
	payload, err := result.PayloadAsByte()

	if err != nil {
		return err
	}

	cmd.Value = float32(payload) / 255

	return nil
}

// DistSinceDTCClear represents a command that checks distance since last DTC clear
//
// Min: 0
// Max: 65535
type DistSinceDTCClear struct {
	baseCommand
	UIntCommand
}

// NewDistSinceDTCClear creates a new commend distance since DTC clear with the correct parameters.
func NewDistSinceDTCClear() *DistSinceDTCClear {
	return &DistSinceDTCClear{
		baseCommand{SERVICE_01_ID, 0x31, 2, "dist_since_dtc_clean"},
		UIntCommand{},
	}
}

// SetValue processes the byte array value into the right uint value.
func (cmd *DistSinceDTCClear) SetValue(result *Result) error {
	payload, err := result.PayloadAsUInt16()

	if err != nil {
		return err
	}

	cmd.Value = uint32(payload)

	return nil
}

// Odometer represents the distance travelled in kilometers
//
// Min: 0
// Max: 429,496,729.5
type Odometer struct {
	baseCommand
	FloatCommand
}

// NewOdometer creates a new odometer value with the correct parameters.
func NewOdometer() *Odometer {
	return &Odometer{
		baseCommand{SERVICE_01_ID, 0xa6, 4, "odometer"},
		FloatCommand{},
	}
}

// SetValue processes the byte array value into the right uint value.
func (cmd *Odometer) SetValue(result *Result) error {
	payload, err := result.PayloadAsUInt32()

	if err != nil {
		return err
	}

	cmd.Value = float32(payload) / 10

	return nil
}

// TransmissionActualGear represents the gear ratio
//
// Min: 0
// Max: 65.535
type TransmissionActualGear struct {
	baseCommand
	FloatCommand
}

// NewTransmissionActualGear creates a new transmission actual gear ratio with the correct parameters.
func NewTransmissionActualGear() *TransmissionActualGear {
	return &TransmissionActualGear{
		baseCommand{SERVICE_01_ID, 0xa4, 4, "transmission_actual_gear"},
		FloatCommand{},
	}
}

// SetValue processes the byte array value into the right uint value.
func (cmd *TransmissionActualGear) SetValue(result *Result) error {
	payload, err := result.PayloadAsUInt32()

	if err != nil {
		return err
	}
	// A & B are not used in the calculation
	cmd.Value = float32(payload>>16) / 1000

	return nil
}

// CoolantTemperature represents a command that checks the engine coolant
// temperature in Celsius.
//
// Min: -40
// Max: 215
type CoolantTemperature struct {
	baseCommand
	IntCommand
}

// NewCoolantTemperature creates a new CoolantTemperature with the right
// parameters.
func NewCoolantTemperature() *CoolantTemperature {
	return &CoolantTemperature{
		baseCommand{SERVICE_01_ID, 5, 1, "coolant_temperature"},
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
	baseCommand
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
			baseCommand{SERVICE_01_ID, 6, 1, "short_term_fuel_trim_bank1"},
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
			baseCommand{SERVICE_01_ID, 7, 1, "long_term_fuel_trim_bank1"},
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
			baseCommand{SERVICE_01_ID, 8, 1, "short_term_fuel_trim_bank2"},
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
			baseCommand{SERVICE_01_ID, 9, 1, "long_term_fuel_trim_bank2"},
			FloatCommand{},
		},
	}
}

// FuelPressure represents a command that checks the fuel pressure in kPa.
//
// Min: 0
// Max: 765
type FuelPressure struct {
	baseCommand
	UIntCommand
}

// NewFuelPressure creates a new FuelPressure with the right parameters.
func NewFuelPressure() *FuelPressure {
	return &FuelPressure{
		baseCommand{SERVICE_01_ID, 10, 1, "fuel_pressure"},
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
	baseCommand
	UIntCommand
}

// NewIntakeManifoldPressure creates a new IntakeManifoldPressure with the
// right parameters.
func NewIntakeManifoldPressure() *IntakeManifoldPressure {
	return &IntakeManifoldPressure{
		baseCommand{SERVICE_01_ID, 11, 1, "intake_manifold_pressure"},
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
	baseCommand
	FloatCommand
}

// NewEngineRPM creates a new EngineRPM with the right parameters.
func NewEngineRPM() *EngineRPM {
	return &EngineRPM{
		baseCommand{SERVICE_01_ID, 12, 2, "engine_rpm"},
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
	baseCommand
	UIntCommand
}

// NewVehicleSpeed creates a new VehicleSpeed with the right parameters
func NewVehicleSpeed() *VehicleSpeed {
	return &VehicleSpeed{
		baseCommand{SERVICE_01_ID, 13, 1, "vehicle_speed"},
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
	baseCommand
	FloatCommand
}

// NewTimingAdvance creates a new TimingAdvance with the right parameters.
func NewTimingAdvance() *TimingAdvance {
	return &TimingAdvance{
		baseCommand{SERVICE_01_ID, 14, 1, "timing_advance"},
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
	baseCommand
	IntCommand
}

// NewIntakeAirTemperature creates a new IntakeAirTemperature with the right parameters.
func NewIntakeAirTemperature() *IntakeAirTemperature {
	return &IntakeAirTemperature{
		baseCommand{SERVICE_01_ID, 15, 1, "intake_air_temperature"},
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
	baseCommand
	FloatCommand
}

// NewMafAirFlowRate creates a new MafAirFlowRate with the right parameters.
func NewMafAirFlowRate() *MafAirFlowRate {
	return &MafAirFlowRate{
		baseCommand{SERVICE_01_ID, 16, 2, "maf_air_flow_rate"},
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
	baseCommand
	FloatCommand
}

// NewThrottlePosition creates a new ThrottlePosition with the right parameters.
func NewThrottlePosition() *ThrottlePosition {
	return &ThrottlePosition{
		baseCommand{SERVICE_01_ID, 17, 1, "throttle_position"},
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
	baseCommand
	UIntCommand
}

// NewOBDStandards creates a new OBDStandards with the right parameters.
func NewOBDStandards() *OBDStandards {
	return &OBDStandards{
		baseCommand{SERVICE_01_ID, 28, 1, "obd_standards"},
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
	baseCommand
	UIntCommand
}

// NewRuntimeSinceStart creates a new RuntimeSinceStart with the right
// parameters.
func NewRuntimeSinceStart() *RuntimeSinceStart {
	return &RuntimeSinceStart{
		baseCommand{SERVICE_01_ID, 31, 2, "runtime_since_engine_start"},
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

type ClearTroubleCodes struct {
	baseCommand
	ResultLessCommand
}

func (cmd *ResultLessCommand) SetValue(result *Result) error {
	return nil
}
func (cmd *ResultLessCommand) ValueAsLit() string {
	return ""
}

// NewClearTroubleCodes creates a new ClearTroubleCodes with the right parameters..
func NewClearTroubleCodes() *ClearTroubleCodes {
	return &ClearTroubleCodes{
		baseCommand{SERVICE_04_ID, 0, 0, "clear_trouble_codes"},
		ResultLessCommand{},
	}
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

// Control module voltage
type ControlModuleVoltage struct {
	baseCommand
	FloatCommand
}

// NewThrottlePosition creates a new ThrottlePosition with the right parameters.
func NewControlModuleVoltage() *ControlModuleVoltage {
	return &ControlModuleVoltage{
		baseCommand{SERVICE_01_ID, 0x42, 2, "control_module_voltage"},
		FloatCommand{},
	}
}

// SetValue processes the byte array value into the right float value.
func (cmd *ControlModuleVoltage) SetValue(result *Result) error {
	payload, err := result.PayloadAsUInt16()

	if err != nil {
		return err
	}

	cmd.Value = float32(payload) / 1000

	return nil
}

// AmbientTemperature represents a command that checks the engine coolant
// temperature in Celsius.
//
// Min: -40
// Max: 215
type AmbientTemperature struct {
	baseCommand
	IntCommand
}

// NewCoolantTemperature creates a new CoolantTemperature with the right
// parameters.
func NewAmbientTemperature() *AmbientTemperature {
	return &AmbientTemperature{
		baseCommand{SERVICE_01_ID, 0x46, 1, "ambient_temperature"},
		IntCommand{},
	}
}

// SetValue processes the byte array value into the right integer value.
func (cmd *AmbientTemperature) SetValue(result *Result) error {
	payload, err := result.PayloadAsByte()

	if err != nil {
		return err
	}

	cmd.Value = int(payload) - 40

	return nil
}

// EngineOilTemperature represents a command that checks the engine oil
// temperature in Celsius.
//
// Min: -40
// Max: 215
type EngineOilTemperature struct {
	baseCommand
	IntCommand
}

// NewCoolantTemperature creates a new CoolantTemperature with the right
// parameters.
func NewEngineOilTemperature() *EngineOilTemperature {
	return &EngineOilTemperature{
		baseCommand{SERVICE_01_ID, 0x5c, 1, "engine_oil_temperature"},
		IntCommand{},
	}
}

// SetValue processes the byte array value into the right integer value.
func (cmd *EngineOilTemperature) SetValue(result *Result) error {
	payload, err := result.PayloadAsByte()

	if err != nil {
		return err
	}

	cmd.Value = int(payload) - 40

	return nil
}

// AbsoluteBarometricPressure
type AbsoluteBarometricPressure struct {
	baseCommand
	IntCommand
}

// NewCoolantTemperature creates a new CoolantTemperature with the right
// parameters.
func NewAbsoluteBarometricPressure() *AbsoluteBarometricPressure {
	return &AbsoluteBarometricPressure{
		baseCommand{SERVICE_01_ID, 0x33, 1, "absolute_barometric_pressure"},
		IntCommand{},
	}
}

// SetValue processes the byte array value into the right integer value.
func (cmd *AbsoluteBarometricPressure) SetValue(result *Result) error {
	payload, err := result.PayloadAsByte()

	if err != nil {
		return err
	}

	cmd.Value = int(payload)

	return nil
}
