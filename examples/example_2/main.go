package main

import (
	"flag"
	"fmt"
	"github.com/rzetterberg/elmobd"
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

	rpm, err := dev.RunOBDCommand(elmobd.NewEngineRPM())

	if err != nil {
		fmt.Println("Failed to get rpm", err)
		return
	}

	fmt.Printf("Engine spins at %s RPMs\n", rpm.ValueAsLit())
}
