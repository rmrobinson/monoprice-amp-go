package main

import (
	"flag"
	"fmt"
	"github.com/rmrobinson/monoprice-amp-go"
	"github.com/tarm/serial"
)

func main() {
	var (
		port   = flag.String("port", "", "The port to control")
		zone   = flag.Int("zone", 1, "The zone to control")
		on     = flag.Bool("on", true, "Whether to turn this output on or off")
		volume = flag.Int("volume", 0, "Volume to set")
	)

	flag.Parse()

	if len(*port) < 1 {
		fmt.Printf("Must supply a port name\n")
		return
	}

	fmt.Printf("Using %s to turn zone %d %t\n", *port, *zone, *on)

	c := &serial.Config{
		Name: *port,
		Baud: 9600,
	}
	s, err := serial.OpenPort(c)

	if err != nil {
		fmt.Printf("Unable to open %s: %s\n", *port, err.Error())
		return
	}

	amp, err := monoprice.NewSerialAmplifier(s)

	if err != nil {
		fmt.Printf("Unable to create: %s\n", err.Error())
		return
	}

	for i := 1; i <= 6; i++ {
		z := amp.Zone(i)

		if z == nil {
			fmt.Printf("Supplied an unsupported zone ID %d\n", i)
			return
		}

		state := z.State()
		fmt.Printf("State of zone %s: %+v\n", z.ID(), state)
	}

	z := amp.Zone(*zone)
	if z == nil {
		fmt.Printf("Supplied an unsupported zone ID: %d\n", *zone)
	}

	err = z.SetPower(*on)
	if err != nil {
		fmt.Printf("Error setting zone %d %t: %s\n", *zone, *on, err.Error())
		return
	}

	err = z.SetVolume(*volume)
	if err != nil {
		fmt.Printf("Error setting zone %d %d: %s\n", *zone, *volume, err.Error())
		return
	}

	state := z.State()
	if err != nil {
		fmt.Printf("Unable to get state of zone %s: %s\n", z.ID(), err.Error())
	} else {
		fmt.Printf("State of zone %s: %+v\n", z.ID(), state)
	}

	z.Refresh()

	state = z.State()
	if err != nil {
		fmt.Printf("Unable to get state of zone %s: %s\n", z.ID(), err.Error())
	} else {
		fmt.Printf("State of zone %s: %+v\n", z.ID(), state)
	}
}
