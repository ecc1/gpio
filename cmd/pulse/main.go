package main

import (
	"flag"
	"log"
	"time"

	"github.com/ecc1/gpio"
)

var (
	interval = flag.Int("i", 500000, "pulse interval in microseconds")
	pin      = flag.Int("p", 25, "GPIO pin")
)

func main() {
	flag.Parse()
	g, err := gpio.Output(*pin, true, false)
	if err != nil {
		log.Fatal(err)
	}
	b := true
	for {
		err = g.Write(b)
		if err != nil {
			log.Fatal(err)
		}
		time.Sleep(time.Duration(*interval) * time.Microsecond)
		b = !b
	}
}
