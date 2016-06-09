package main

import (
	"log"
	"time"

	"github.com/ecc1/gpio"
)

func main() {
	g, err := gpio.Input(14, "none", true)
	if err != nil {
		log.Fatal(err)
	}
	for {
		b, err := g.Read()
		if err != nil {
			log.Fatal(err)
		}
		if b {
			log.Printf("button pressed")
		}
		time.Sleep(500 * time.Millisecond)
	}
}
