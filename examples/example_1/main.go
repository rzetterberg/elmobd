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

	version, err := dev.GetVersion()

	if err != nil {
		fmt.Println("Failed to get version", err)
		return
	}

	fmt.Println("Device has version", version)
}
