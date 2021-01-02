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

	// Verify that a subcommand was provided.
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	subcommand := os.Args[1]
	switch subcommand {
	case "send":
		if len(os.Args) != 3 {
			printUsage()
			os.Exit(1)
		}

		log.Println("lancp running in send mode...")

		filePath := os.Args[2]
		log.Printf("sending file: %s\n", filePath)
	case "receive":
		if len(os.Args) != 2 {
			printUsage()
			os.Exit(1)
		}

		log.Println("lancp running in receive mode...")
	default:
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Fprintf(os.Stderr, "%s\n", usage)
}
