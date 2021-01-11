package net

import "fmt"

const (
	minValidPort = 1025
	maxValidPort = 65535
)

// GetPortAsString returns the provided port number in the format ":0000(0)".
func GetPortAsString(port int) (string, error) {
	if port < minValidPort || port > maxValidPort {
		return "", fmt.Errorf("port must be be in the range [%d, %d], got: %d",
			minValidPort, maxValidPort, port)
	}

	return fmt.Sprintf(":%d", port), nil
}
