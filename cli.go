package main

import (
	"log"
	"os"
)

func main() {
	inputs := os.Args[1:]
	cdx, err := NewCodex(inputs)
	if err != nil {
		log.Fatal(err)
	}
	cdx.BuildAndWatch()
}
