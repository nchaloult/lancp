package main

import (
	"fmt"
	"log"
	"os"
)

const usage = `lancp is a simple tool that lets you easily transfer files between two machines on the same network.

Usage:
    lancp send FILE
    lancp receive

FILE is a path to a file that will be sent to the receiver.
`

func main() {
	// Disable timestamps on messages.
	// Why not use fmt instead, then? https://stackoverflow.com/a/19646964
	log.SetFlags(0)

	log.Println("Your character doesn't blink in first person games")
	printUsage()
}

func printUsage() {
	fmt.Fprintf(os.Stderr, "%s\n", usage)
}
