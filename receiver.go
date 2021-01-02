package main

import (
	"fmt"
	"log"
	"net"
)

func receive() error {
	log.Println("lancp running in receive mode...")

	// Listen for a broadcast message from the device running in "send mode."
	conn, err := net.ListenPacket("udp4", fmt.Sprintf(":%d", port))
	if err != nil {
		return fmt.Errorf("failed to stand up local UDP packet announcer: %v",
			err)
	}
	// TODO: might wanna do this sooner; don't defer it until the end of this
	// big ass func. Putting this here will make more sense once the logic in
	// this func is split up.
	defer conn.Close()

	// Capture the payload that the sender included in their broadcast message.
	payloadBuf := make([]byte, 1024)
	n, senderAddr, err := conn.ReadFrom(payloadBuf)
	if err != nil {
		return fmt.Errorf("failed to read broadcast message from sender: %v",
			err)
	}
	payload := string(payloadBuf[:n])

	log.Printf("payload from %s: %q\n", senderAddr.String(), payload)

	return nil
}
