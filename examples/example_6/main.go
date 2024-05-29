package main

import (
	"flag"
	"fmt"
	"time"
	obd "github.com/rzetterberg/elmobd"
)

func speed_callback (res obd.OBDCommand, cxt interface{}) {
	fmt.Println("Speed from inside callback", res.ValueAsLit(), *cxt.(*string)) 
}

func rpm_callback (res obd.OBDCommand, cxt interface{}) {
	fmt.Println("RPM from inside callback", res.ValueAsLit(), *cxt.(*string))
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
	asyncDev, err := obd.NewAsyncDevice(*addr, *debug, 200 * time.Millisecond)
	if err != nil {
		fmt.Println("Failed to create new async device ", err)
		return
	}
	context := "context"
	speed_action := *obd.CreateAction(speed_callback, &context)
	asyncDev.Watch(obd.NewVehicleSpeed(), []obd.Action{speed_action})

	rpm_action := *obd.CreateAction(rpm_callback, &context)
	asyncDev.Watch(obd.NewEngineRPM(), []obd.Action{rpm_action})

	asyncDev.Start() //the callback will now be fired upon receipt of new values
	defer asyncDev.Stop()
	time.Sleep(60 * time.Second)

}

