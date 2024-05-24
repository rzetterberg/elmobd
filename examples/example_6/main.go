package main

import (
	"flag"
	"fmt"
	"time"
	"github.com/rzetterberg/elmobd"
)

func speed_callback (res elmobd.OBDCommand) {
	fmt.Println("Speed from inside callback", res.ValueAsLit())
}

func rpm_callback (res elmobd.OBDCommand) {
	fmt.Println("RPM from inside callback", res.ValueAsLit())
}

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

	//sample
	asyncDev, err := elmobd.NewAsyncDevice(*addr, *debug, 200 * time.Millisecond)
	if err != nil {
		fmt.Println("Failed to create new async device ", err)
		return
	}
	
	asyncDev.Watch(elmobd.NewVehicleSpeed(), []elmobd.Action{speed_callback})
	asyncDev.Watch(elmobd.NewEngineRPM(), []elmobd.Action{rpm_callback})
	asyncDev.Start() //the callback will now be fired upon receipt of new values
	defer asyncDev.Stop()
	time.Sleep(60 * time.Second)
}
