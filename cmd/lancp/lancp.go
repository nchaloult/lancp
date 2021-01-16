package main

import (
	"fmt"
	"log"
	"os"

	"github.com/nchaloult/lancp/pkg/app"
)

const usage = `lancp
A simple tool for easily transferring files between two machines on the same network.

USAGE:
    lancp send <file>
    lancp receive

FLAGS:
    -h, --help       Prints this usage information and exits
    -v, --version    Prints version information and exits

ARGS:
    <file>    The path to a file to send
`

// TODO: temporary! This config const should be read in from a global config,
// or maybe even provided as a command-line arg.
const port = 6969

func main() {
	// Disable timestamps on messages.
	// Why not use fmt instead, then? https://stackoverflow.com/a/19646964
	log.SetFlags(0)

	// Verify that a subcommand was provided.
	if len(os.Args) < 2 {
		printUsageAndExit()
	}

	subcommand := os.Args[1]
	switch subcommand {
	case "send":
		if len(os.Args) != 3 {
			printUsageAndExit()
		}

		// TODO: validate that this looks like a file path.
		//
		// TODO: write a function somewhere which makes sure that file actually
		// exists on disk.
		filePath := os.Args[2]

		cfg, err := app.NewSenderConfig(filePath, port, port+1)
		if err != nil {
			printError(err)
		}
		if err := cfg.Send(); err != nil {
			printError(err)
		}
	case "receive":
		if len(os.Args) != 2 {
			printUsageAndExit()
		}

		cfg, err := app.NewReceiverConfig(port, port+1)
		if err != nil {
			printError(err)
		}
		if err := cfg.Receive(); err != nil {
			printError(err)
		}
	default:
		printUsageAndExit()
	}
}

func printUsageAndExit() {
	fmt.Fprintf(os.Stderr, "%s\n", usage)
	os.Exit(1)
}

func printError(err error) {
	log.Fatalf("ERROR: %v", err)
}
