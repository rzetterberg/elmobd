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

	res, err := dev.DirectDeviceCommand("AT@1")

	if err != nil {
		fmt.Println("Failed to get result", err)
		return
	}
	
	
	outputs := res
	//output := outputs[0][:]
	fmt.Println(outputs)
}
