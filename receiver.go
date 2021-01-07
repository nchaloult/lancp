package main

import (
	"bytes"
	"crypto/ed25519"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"io"
	"log"
	"math/big"
	"net"
	"os"
	"strconv"
	"time"
)

func receive() error {
	// TODO: Revisit where these are initialized. Find new homes for these once
	// this big ass func is broken up into pieces.
	portAsStr := fmt.Sprintf(":%d", port)
	tlsPortAsStr := fmt.Sprintf(":%d", port+1)

	log.Println("lancp running in receive mode...")

	// TODO: Generate passphrase that the sender will need to present.
	generatedPassphrase := "receiver"

	// Listen for a broadcast message from the device running in "send mode."
	udpConn, err := net.ListenPacket("udp4", portAsStr)
	if err != nil {
		return fmt.Errorf("failed to stand up local UDP packet announcer: %v",
			err)
	}

	// Capture the payload that the sender included in their broadcast message.
	// TODO: Shrink buffer size once you've written passphrase generation logic.
	passphrasePayloadBuf := make([]byte, 1024)
	n, senderAddr, err := udpConn.ReadFrom(passphrasePayloadBuf)
	if err != nil {
		udpConn.Close()
		return fmt.Errorf("failed to read broadcast message from sender: %v",
			err)
	}
	passphrasePayload := string(passphrasePayloadBuf[:n])

	// Compare payload with expected payload.
	if passphrasePayload != generatedPassphrase {
		udpConn.Close()
		return fmt.Errorf("got %q from %s, want %q",
			passphrasePayload, senderAddr.String(), generatedPassphrase)
	}
	log.Printf("got %q from %s, matched expected passphrase",
		passphrasePayload, senderAddr.String())

	// TODO: Capture user input for the passphrase the sender is presenting.
	input := "sender"

	// Send response message to sender.
	_, err = udpConn.WriteTo([]byte(input), senderAddr)
	if err != nil {
		udpConn.Close()
		return fmt.Errorf("failed to send response message to sender: %v", err)
	}

	// Even though the sender hasn't had time to receive and parse the response
	// UDP datagram we just sent, we can close the PacketConn on our end since
	// UDP is a stateless protocol.
	udpConn.Close()

	// Begin standing up TCP server to exchange cert, and prepare to establish a
	// TLS connection with the sender.

	// TODO: This called func lives in sender.go rn. Move this to some new
	// shared location when you refactor everything.
	localAddr, err := getPreferredOutboundAddr()
	if err != nil {
		return fmt.Errorf("failed to get this device's local IP address: %v",
			err)
	}

	// Generate self-signed TLS cert.
	cert, err := generateSelfSignedCert(localAddr)
	if err != nil {
		return fmt.Errorf("failed to generate certificate: %v", err)
	}

	// Listen for the first part of the TCP handshake from the sender. Send the
	// sender the TLS certificate on that connection.
	tcpLn, err := net.Listen("tcp", portAsStr)
	if err != nil {
		return fmt.Errorf("failed to start a TCP listener: %v", err)
	}
	// Block until the sender initiates the handshake.
	tcpConn, err := tcpLn.Accept()
	if err != nil {
		return fmt.Errorf("failed to establish TCP connection with sender: %v",
			err)
	}

	// Send TLS certificate to the sender.
	tcpConn.Write(cert.bytes)

	// Listen for an attempt to establish a TLS connection from the sender.
	tlsCfg, err := getReceiverTLSConfig(cert)
	if err != nil {
		return fmt.Errorf("failed to build TLS config: %v", err)
	}
	tlsLn, err := tls.Listen("tcp", tlsPortAsStr, tlsCfg)
	if err != nil {
		return fmt.Errorf("failed to start a TLS listener: %v", err)
	}
	defer tlsLn.Close()

	// Close up the TCP connection after the TLS listener has already fired up.
	// This is so the sender can try to establish a TLS connection with the
	// receiver immediately after they get the receiver's public key, and not
	// have to make any guesses or assumptions about how long the receiver will
	// take to shut down their TCP listener and spin up their TLS listener.
	tcpConn.Close()
	tcpLn.Close()

	// Block until the sender initiates the handshake.
	tlsConn, err := tlsLn.Accept()
	if err != nil {
		return fmt.Errorf("failed to establish TLS connection with sender: %v",
			err)
	}
	defer tlsConn.Close()

	// Create a file on disk that will eventually store the payload we receive
	// from the sender.
	//
	// TODO: Read the file name that the sender sends first. Right now, the file
	// name is hard-coded by the receiver.
	file, err := os.Create("from-sender")
	if err != nil {
		return fmt.Errorf("failed to create a new file on disk: %v", err)
	}
	defer file.Close()

	// Receive file's bytes from the sender.
	// TODO: Is this an okay size for this buffer? How big could it ever get?
	fileSizeBuf := make([]byte, 10)
	n, err = tlsConn.Read(fileSizeBuf)
	if err != nil {
		return fmt.Errorf("failed to read file size from sender: %v", err)
	}
	fileSizeAsStr := string(fileSizeBuf[:n])
	fileSize, err := strconv.ParseInt(fileSizeAsStr, 10, 64)
	if err != nil {
		return fmt.Errorf("failed to convert %q to an int: %v", fileSizeAsStr)
	}

	// Write that payload to a file on disk.
	var receivedBytes int64
	for {
		if (fileSize - receivedBytes) < filePayloadBufSize {
			io.CopyN(file, tlsConn, (fileSize - receivedBytes))
			tlsConn.Read(make([]byte, (receivedBytes+filePayloadBufSize)-fileSize))
			break
		}

		io.CopyN(file, tlsConn, filePayloadBufSize)
		receivedBytes += filePayloadBufSize
	}

	log.Printf("received %d bytes from sender\n", receivedBytes)

	return nil
}

type selfSignedCert struct {
	// The certificate as PEM-encoded bytes.
	bytes []byte

	// The private key as PEM-encoded bytes.
	sk []byte
}

// generateSelfSignedCert creates a self-signed x509 certificate to be used when
// establishing a TLS connection with the sender. The created certificate is
// valid for the device with the provided IPv4 address.
//
// It generates a public/private key pair, uses those keys to build an x509
// certificate, self-signs that certificate so the sender will trust it, and
// PEM-encodes that certificate and private key.
//
// Inspired by https://golang.org/src/crypto/tls/generate_cert.go
func generateSelfSignedCert(ip net.IP) (*selfSignedCert, error) {
	// Get public/private key pair for certificate.
	_, sk, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return nil, fmt.Errorf("failed to generate public/private key pair: %v",
			err)
	}

	// Get serial number for certificate.
	serialNumberLimit := new(big.Int).Lsh(big.NewInt(1), 128)
	serialNumber, err := rand.Int(rand.Reader, serialNumberLimit)
	if err != nil {
		return nil, fmt.Errorf("failed to generate serial number for"+
			" certificate: %v", err)
	}

	// Build a certificate template.
	certTemplate := x509.Certificate{
		IPAddresses: []net.IP{ip},

		SerialNumber: serialNumber,
		Subject: pkix.Name{
			Organization: []string{"lancp"}, // TODO: Don't hard-code this.
		},
		NotBefore: time.Now(),
		// Would rather not shrink this time gap any further to allow a bit of
		// discrepancy between the system time on the sender's machine vs. the
		// receiver's machine.
		//
		// TODO: Can we recover from cert expiration errors by creating new
		// certs with a larger time gaps until one works? Would that be safe?
		NotAfter: time.Now().Add(time.Minute),

		KeyUsage: x509.KeyUsageDigitalSignature,

		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
	}

	// Turn the certificate template into PEM-encoded bytes.

	// Self-sign the cert by making the parent authority be the same cert.
	certBytes, err := x509.CreateCertificate(
		rand.Reader,
		&certTemplate,
		&certTemplate,
		sk.Public().(ed25519.PublicKey),
		sk,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create cert from template: %v", err)
	}
	certPEM := new(bytes.Buffer)
	err = pem.Encode(certPEM, &pem.Block{
		Type:  "CERTIFICATE",
		Bytes: certBytes,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to PEM-encode certificate: %v", err)
	}

	// Turn the private key into PEM-encoded bytes.
	skBytes, err := x509.MarshalPKCS8PrivateKey(sk)
	if err != nil {
		return nil, fmt.Errorf("failed to convert private key from key pair"+
			" into PKCS#8 form: %v", err)
	}
	skPEM := new(bytes.Buffer)
	err = pem.Encode(skPEM, &pem.Block{
		Type:  "PRIVATE KEY",
		Bytes: skBytes,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to PEM-encode private key: %v", err)
	}

	return &selfSignedCert{
		bytes: certPEM.Bytes(),
		sk:    skPEM.Bytes(),
	}, nil
}

// getReceiverTLSConfig builds a tls.Config object for the receiver to use when
// establishing a TLS connection with the sender. It adds the receiver's public/
// private key pair to the config's list of certificates.
func getReceiverTLSConfig(cert *selfSignedCert) (*tls.Config, error) {
	keyPair, err := tls.X509KeyPair(cert.bytes, cert.sk)
	if err != nil {
		return nil, fmt.Errorf("failed to create x509 public/private key pair"+
			" from the provided self-signed certificate: %v", err)
	}

	return &tls.Config{
		Certificates: []tls.Certificate{keyPair},
	}, nil
}
