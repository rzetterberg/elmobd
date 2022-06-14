package main

import (
	"flag"
	"fmt"

	"github.com/rzetterberg/elmobd"
)

func main() {
	addr := flag.String(
		"addr",
		"test:///dev/ttyUSB0",
		"Address of the ELM327 device to use (use either test://, tcp://ip:port or serial:///dev/ttyS0)",
	)
	debug := flag.Bool(
		"debug",
		false,
		"Enable debug outputs",
	)

	flag.Parse()

	dev, err := elmobd.NewDevice(*addr, *debug)

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
