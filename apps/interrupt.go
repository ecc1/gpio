package main

import (
	"log"
	"time"

	"github.com/ecc1/gpio"
)

func main() {
	g, err := gpio.Input(14, "rising", true)
	if err != nil {
		log.Fatal(err)
	}
	for {
		err = g.Wait(10 * time.Second)
		if err != nil {
			_, isTimeout := err.(gpio.TimeoutError)
			if isTimeout {
				log.Print(err)
				continue
			}
			log.Fatal(err)
		}
		log.Print("interrupt")
	}
}
