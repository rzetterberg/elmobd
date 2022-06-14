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

	cmd, err := dev.RunOBDCommand(elmobd.NewMonitorStatus())

	if err != nil {
		fmt.Println("Failed to get monitor status", err)
		return
	}

	status := cmd.(*elmobd.MonitorStatus)

	fmt.Printf("MIL is on: %t, DTCamount: %d\n", status.MilActive, status.DtcAmount)
}
