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
	go cdx.BuildAndWatch()
	go cdx.Serve(":8000")
	select {}
}
