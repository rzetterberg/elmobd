// Package elmobd provides communication with motor vehicles OBD system.
//
// Using this library and a ELM327-based USB-device you can communicate
// with your cars on-board diagnostics system (OBD) to:
//
// - Read sensor data in real time
// - Check trouble codes
// - Reset the check engine light
//
// All assumptions this library makes are based on the official Elm Electronics
// datasheet of the ELM327 IC:
// https://www.elmelectronics.com/wp-content/uploads/2017/01/ELM327DS.pdf
package elmobd
