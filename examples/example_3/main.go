package main

import (
	"flag"
	"fmt"
	"github.com/wkarasz/elmobd"
)

func main() {
	serialPath := flag.String(
		"serial",
		"/dev/ttyUSB0",
		"Path to the serial device to use",
	)

	flag.Parse()

	dev, err := elmobd.NewTestDevice(*serialPath, false)

	if err != nil {
		fmt.Println("Failed to create new device", err)
		return
	}

	supported, err := dev.CheckSupportedCommands()

	if err != nil {
		fmt.Println("Failed to check supported commands", err)
		return
	}

	rpm := elmobd.NewEngineRPM()

	if supported.IsSupported(rpm) {
		fmt.Println("The car supports checking RPM")
	} else {
		fmt.Println("The car does NOT supports checking RPM")
	}
}
