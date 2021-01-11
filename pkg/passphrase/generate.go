package passphrase

import (
	"math/rand"
	"time"
)

// Generate returns a word from a list uniformly at random. The list contains
// phonetically-distinct and relatively short English words.
//
// TODO: I think this is the wordlist that magic-wormhole uses? Maybe pull more
// from there.
// https://github.com/warner/magic-wormhole/blob/master/src/wormhole/_wordlist.py
func Generate() string {
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
