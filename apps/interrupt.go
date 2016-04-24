package main

import (
	"log"

	"github.com/ecc1/gpio"
)

func main() {
	g, err := gpio.Input(14, "rising", true)
	if err != nil {
		log.Fatal(err)
	}
	for {
		err = g.Wait()
		if err != nil {
			log.Fatal(err)
		}
		log.Printf("interrupt\n")
	}
}
