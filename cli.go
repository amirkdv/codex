package main

import "os"

func main() {
	inputs := os.Args[1:]
	NewServer(inputs, ":8000").Start()
}
