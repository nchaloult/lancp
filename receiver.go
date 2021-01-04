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
	"log"
	"math/big"
	"net"
	"time"
)

func receive() error {
	log.Println("lancp running in receive mode...")

	// TODO: Generate passphrase that the sender will need to present.
	generatedPassphrase := "receiver"

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

	// Compare payload with expected payload.
	if payload != generatedPassphrase {
		return fmt.Errorf("got %q from %s, want %q",
			payload, senderAddr.String(), generatedPassphrase)
	}
	log.Printf("got %q from %s, matched expected passphrase",
		payload, senderAddr.String())

	// TODO: Capture user input for the passphrase the sender is presenting.
	input := "sender"

	// Send response message to sender.
	_, err = conn.WriteTo([]byte(input), senderAddr)
	if err != nil {
		return fmt.Errorf("failed to send response message to sender: %v", err)
	}

	return nil
}

type selfSignedCert struct {
	cert []byte
	sk   []byte
}

// generateSelfSignedCert creates a self-signed x509 certificate to be used when
// establishing a TLS connection with the sender.
//
// It generates a public/private key pair, uses those keys to build an x509
// certificate, self-signs that certificate so the sender will trust it, and
// PEM-encodes that certificate and private key.
//
// Inspired by https://golang.org/src/crypto/tls/generate_cert.go
func generateSelfSignedCert() (*selfSignedCert, error) {
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
		IPAddresses: []net.IP{net.IPv4(127, 0, 0, 1)}, // TODO: Set dynamically.

		SerialNumber: serialNumber,
		Subject: pkix.Name{
			Organization: []string{"lancp"}, // TODO: Don't hard-code this.
		},
		NotBefore: time.Now(),
		NotAfter:  time.Now(),

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
		cert: certPEM.Bytes(),
		sk:   skPEM.Bytes(),
	}, nil
}

// getReceiverTLSConfig builds a tls.Config object for the receiver to use when
// establishing a TLS connection with the sender. It adds the receiver's public/
// private key pair to the config's list of certificates.
func getReceiverTLSConfig(certPEM, privateKeyPEM []byte) (*tls.Config, error) {
	keyPair, err := tls.X509KeyPair(certPEM, privateKeyPEM)
	if err != nil {
		return nil, fmt.Errorf("failed to create x509 public/private key pair"+
			" from the provided self-signed certificate: %v", err)
	}

	return &tls.Config{
		Certificates: []tls.Certificate{keyPair},
	}, nil
}
