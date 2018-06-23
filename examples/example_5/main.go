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

	cmd, err := dev.RunOBDCommand(elmobd.NewMonitorStatus())

	if err != nil {
		fmt.Println("Failed to get monitor status", err)
		return
	}

        status := cmd.(*elmobd.MonitorStatus)

	fmt.Printf("MIL is on: %t, DTCamount: %d\n", status.MilActive, status.DtcAmount)
}
