package elmobd

import (
	"bytes"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/tarm/serial"
)

/*==============================================================================
 * External
 */

// RealResult represents the raw text output of running a raw command,
// including information used in debugging to show what input caused what
// error, how long the command took, etc.
type RealResult struct {
	input     string
	outputs   []string
	error     error
	writeTime time.Duration
	readTime  time.Duration
	totalTime time.Duration
}

// Failed checks if the result is successful or not
func (res *RealResult) Failed() bool {
	return res.error != nil
}

// GetError returns the results current error
func (res *RealResult) GetError() error {
	return res.error
}

// GetOutputs returns the outputs of the result
func (res *RealResult) GetOutputs() []string {
	return res.outputs
}

// FormatOverview formats a result as an overview of what command was run and
// how long it took.
func (res *RealResult) FormatOverview() string {
	lines := []string{
		"=======================================",
		" Ran command \"%s\" in %s",
		" Spent %s writing",
		" Spent %s reading",
		"=======================================",
	}

	return fmt.Sprintf(
		strings.Join(lines, "\n"),
		res.input,
		res.totalTime,
		res.writeTime,
		res.readTime,
	)
}

// RealDevice represent the low level serial connection.
type RealDevice struct {
	mutex      sync.Mutex
	state      deviceState
	input      string
	outputs    []string
	serialPort *serial.Port
}

// NewRealDevice creates a new low-level ELM327 device manager by connecting to
// the device at given path.
//
// After a connection has been established the device is reset, and a minimum of
// 800 ms blocking wait will occur. This makes sure the device does not have
// any custom settings that could make this library handle the device
// incorrectly.
func NewRealDevice(devicePath string) (*RealDevice, error) {
	config := &serial.Config{
		Name:        devicePath,
		Baud:        38400,
		ReadTimeout: time.Second * 5,
		Size:        8,
		Parity:      serial.ParityNone,
		StopBits:    serial.Stop1,
	}

	port, err := serial.OpenPort(config)

	if err != nil {
		return nil, err
	}

	dev := &RealDevice{
		state:      deviceReady,
		mutex:      sync.Mutex{},
		serialPort: port,
	}

	err = dev.Reset()

	if err != nil {
		return nil, err
	}

	return dev, nil
}

// Reset restarts the device, resets all the settings to factory defaults and
// makes sure it actually is a ELM327 device we are talking to.
//
// In case this doesn't work, you should turn off/on the device.
func (dev *RealDevice) Reset() error {
	var err error

	dev.mutex.Lock()
	dev.state = deviceBusy

	err = dev.serialPort.Flush()

	if err != nil {
		goto out
	}

	_, err = dev.write("ATZ")

	if err != nil {
		goto out
	}

	err = dev.read()

	if err != nil {
		goto out
	}

	// Device can identified itself in first or second line
	if !(strings.HasPrefix(dev.outputs[0], "ELM327") || (len(dev.outputs) > 1 && strings.HasPrefix(dev.outputs[1], "ELM327"))) {
		output := dev.outputs[0]
		if len(dev.outputs) > 1 {
			output += " " + dev.outputs[1]
		}
		err = fmt.Errorf(
			"Device did not identify itself as ELM327: %s",
			output,
		)
	}
out:
	if err != nil {
		dev.serialPort.Flush()
		dev.state = deviceError
	} else {
		dev.state = deviceReady
	}

	dev.mutex.Unlock()

	return err
}

// RunCommand runs the given AT/OBD command by sending it to the device and
// waiting for the output. There are no restrictions on what commands you can
// run with this function, so be careful.
//
// WARNING: Do not turn off echoing, because the underlying write function
// relies on echo being on so that it can compare the input command and the
// echo from the device.
//
// For more information about AT/OBD commands, see:
// https://en.wikipedia.org/wiki/Hayes_command_set
// https://en.wikipedia.org/wiki/OBD-II_PIDs
func (dev *RealDevice) RunCommand(command string) RawResult {
	var err error
	var startTotal time.Time
	var startRead time.Time
	var startWrite time.Time

	result := RealResult{
		input:     command,
		writeTime: 0,
		readTime:  0,
		totalTime: 0,
	}

	startTotal = time.Now()

	dev.mutex.Lock()
	dev.state = deviceBusy

	startWrite = time.Now()

	_, err = dev.write(command)

	if err != nil {
		goto out
	}

	result.writeTime = time.Since(startWrite)

	startRead = time.Now()

	err = dev.read()

	result.readTime = time.Since(startRead)

	if err != nil {
		goto out
	}
out:
	if err != nil {
		dev.serialPort.Flush()
		dev.state = deviceError
	} else {
		dev.state = deviceReady
	}

	dev.mutex.Unlock()

	result.error = err
	result.outputs = dev.outputs
	result.totalTime = time.Since(startTotal)

	return &result
}

/*==============================================================================
 * Internal
 */

type deviceState int

const (
	deviceReady deviceState = iota
	deviceBusy
	deviceError
)

func (dev *RealDevice) write(input string) (int, error) {
	dev.input = ""

	n, err := dev.serialPort.Write(
		[]byte(input + "\r\n"),
	)

	if err == nil {
		dev.input = input
	}

	return n, err
}

func (dev *RealDevice) read() error {
	var buffer bytes.Buffer

	ticker := time.NewTicker(10 * time.Millisecond)

	for range ticker.C {
		tmp := make([]byte, 128)
		n, err := dev.serialPort.Read(tmp)

		if err != nil {
			dev.outputs = []string{}
			return err
		}

		buffer.Write(tmp[:n])

		if tmp[n-1] == byte('>') {
			buffer.Truncate(buffer.Len() - 1)
			ticker.Stop()

			break
		}
	}

	return dev.processResult(buffer)
}

func (dev *RealDevice) processResult(result bytes.Buffer) error {
	parts := strings.Split(
		string(result.Bytes()),
		"\r",
	)

	if parts[0] != dev.input {
		return fmt.Errorf(
			"Write echo mismatch: %q not suffix of %q",
			dev.input,
			parts[0],
		)
	}

	parts = parts[1:]

	var trimmedParts []string

	for p := range parts {
		tmp := strings.Trim(parts[p], "\r ")

		if tmp == "" {
			continue
		}

		trimmedParts = append(trimmedParts, tmp)
	}

	if len(trimmedParts) < 1 {
		return fmt.Errorf("No payload receieved")
	}

	dev.outputs = trimmedParts

	return nil
}
