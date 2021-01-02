package main

import "log"

func main() {
	// Disable timestamps on messages.
	// Why not use fmt instead, then? https://stackoverflow.com/a/19646964
	log.SetFlags(0)

	log.Println("Your character doesn't blink in first person games")
}
