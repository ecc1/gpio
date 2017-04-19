package main

import (
	"log"
	"time"

	"github.com/ecc1/gpio"
)

func main() {
	g, err := gpio.Output(45, true)
	if err != nil {
		log.Fatal(err)
	}
	b := true
	for {
		err = g.Write(b)
		if err != nil {
			log.Fatal(err)
		}
		time.Sleep(500 * time.Microsecond)
		b = !b
	}
}
