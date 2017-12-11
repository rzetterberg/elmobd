// Package elmobd provides communication with cars OBD-II system using ELM327
// based USB-devices.
//
// Using this library and a ELM327-based USB-device you can communicate
// with your cars on-board diagnostics system to read sensor data. Reading
// trouble codes and resetting them is not yet implemented.
//
// All assumptions this library makes are based on the official Elm Electronics
// datasheet of the ELM327 IC:
// https://www.elmelectronics.com/wp-content/uploads/2017/01/ELM327DS.pdf
//
// After that introduction - Welcome! I hope you'll find this library useful
// and its documentation easy to digest. If that's not the case, please create
// a GitHub-issue: https://github.com/rzetterberg/elmobd/issues
//
// You'll note that this package has the majority its types and functions
// exported. The reason for that is that there are many different commands
// you can send to a ELM327 device, and this library will never we able to
// define all commands, so most functionality is exported so that you can easily
// extend it.
//
// You'll also note that there are A LOT of types. The reason for this is
// each command that can be sent to the ELM327 device is defined as it's own
// type. The approach I wanted to take with this library was to get as much
// type safety as possible.
//
// With that said, as an end user of this library you'll only need to know
// about two kinds types: the Device type and types that implement the OBDCommand
// interface.
//
// The Device type represents an active connection with an ELM327 device. You
// plug in your device into your computer, get the path to the device and
// initlize a new device:
//
//     package main
//
//     import (
//         "flag"
//         "fmt"
//         "github.com/rzetterberg/elmobd"
//     )
//
//     func main() {
//         serialPath := flag.String(
//             "serial",
//             "/dev/ttyUSB0",
//             "Path to the serial device to use",
//         )
//
//         flag.Parse()
//
//         dev, err := elmobd.NewDevice(*serialPath, false)
//
//         if err != nil {
//             fmt.Println("Failed to create new device", err)
//             return
//         }
//     }
//
// This library is design to be as high level as possible, you shouldn't need to
// handle baud rates, setting protocols or anything like that. After the Device
// has been initialized you can just start sending commands.
//
// The function you will be using the most is the Device.RunOBDCommand, which
// accepts a command, sends the command to the device, waits for a response,
// parses the response and gives it back to you. Which means this command is
// blocking execution until it has been finished. The default timeout is
// currently set to 5 seconds.
//
// The RunOBDCommand accepts a single argument, a type that implements the
// OBDCommand interface. That interface has a couple of functions that needs to
// exist in order to be able to generate the low-level request for the device
// and parse the low-level response. Basically you send in an empty OBDCommand
// and get back a OBDCommand with the processed value retrieved from the device.
//
// Suppose that after checking the device connection (in our example above) we
// would like to retrieve the vehicles speed. There's a type called VehicleSpeed
// that implements the OBDCommand interface that we can use. We start by
// creating a new VehicleSpeed using its constructor NewVehicleSpeed that we
// then give to RunOBDCommand:
//
//     speed, err := dev.RunOBDCommand(elmobd.NewVehicleSpeed())
//
//     if err != nil {
//         fmt.Println("Failed to get vehicle speed", err)
//         return
//     }
//
//     fmt.Printf("Vehicle speed: %d km/h\n", speed.Value)
//
// At the moment there are around 15 sensor commands defined in this library.
// They are all defined in the file commands.go of the library
// (https://github.com/rzetterberg/elmobd/blob/master/commands.go). If you look
// in the end of that file you'll find an internal static variable with all the
// sensor commands called sensorCommands.
//
// That's the basics of this library - you create a device connection and then
// run commands. Documentation of more advanced use-cases is in the works, so
// don't worry! If there's something that you wish would be documented, or
// something that is not clear in the current documentation, please create a
// GitHub-issue.
package elmobd
