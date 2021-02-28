package io

import "os"

// IsFileAccessible checks if a file exists and if we have permissions to read
// it.
func IsFileAccessible(path string) error {
	_, err := os.Stat(path)
	return err
}
