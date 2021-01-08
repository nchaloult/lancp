package main

import (
	"fmt"
	"log"
	"math/rand"
	"os"
	"time"
)

const usage = `lancp is a simple tool that lets you easily transfer files between two machines on the same network.

Usage:
    lancp send FILE
    lancp receive

FILE is a path to a file that will be sent to the receiver.
`

const filePayloadBufSize = 8192

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
		if err := send(filePath); err != nil {
			printError(err)
		}
	case "receive":
		if len(os.Args) != 2 {
			printUsageAndExit()
		}

		if err := receive(); err != nil {
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

// TODO: move someplace else during The Great Refactor (tm).
//
// TODO: I think this is the wordlist that magic-wormhole uses? Maybe pull more
// from there.
// https://github.com/warner/magic-wormhole/blob/master/src/wormhole/_wordlist.py
func generatePassphrase() string {
	// Initialize global pseudo random number generator.
	rand.Seed(time.Now().Unix())

	// TODO: have these options be read in from a file or something? Make that
	// file location configurable?
	//
	// The binary we ship could be massive if we hard-code a bunch of strings
	// in here like this...
	passphrases := []string{
		"absurd", "banjo", "concert", "dashboard", "erase", "framework",
		"goldfish", "hockey", "involve", "jupiter",
	}

	return passphrases[rand.Intn(len(passphrases))]
}
