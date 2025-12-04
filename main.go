package main

import (
	"log"

	"github.com/ColonyPM/cpm/cmd/root"
)

func main() {
	if err := root.Execute(); err != nil {
		log.Fatal(err)
	}
}
