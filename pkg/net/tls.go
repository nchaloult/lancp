package net

import (
	"bytes"
	"crypto/ed25519"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"math/big"
	"net"
	"time"
)

// SelfSignedCert stores an x509 certificate's pieces (the certificate itself as
// well as a private key) as PEM-encoded bytes.
type SelfSignedCert struct {
	// The certificate as PEM-encoded bytes.
	Bytes []byte

	// The private key as PEM-encoded bytes.
	SK []byte
}

// GetReceiverTLSConfig builds a tls.Config object for the receiver to use when
// establishing a TLS connection with the sender. It adds the receiver's public/
// private key pair to the config's list of certificates.
func GetReceiverTLSConfig(cert *SelfSignedCert) (*tls.Config, error) {
	keyPair, err := tls.X509KeyPair(cert.Bytes, cert.SK)
	if err != nil {
		return nil, fmt.Errorf("failed to create x509 public/private key pair"+
			" from the provided self-signed certificate: %v", err)
	}

	return &tls.Config{
		Certificates: []tls.Certificate{keyPair},
	}, nil
}

// GetSenderTLSConfig builds a tls.Config object for the sender to use when
// establishing a TLS connection with the receiver. It adds the public key of
// the certificate authority that the receiver created to the config's
// collection of trusted certificate authorities.
func GetSenderTLSConfig(certPEM []byte) *tls.Config {
	certPool := x509.NewCertPool()
	certPool.AppendCertsFromPEM(certPEM)

	return &tls.Config{
		RootCAs: certPool,
	}
}

// GenerateSelfSignedCert creates a self-signed x509 certificate to be used when
// establishing a TLS connection with the sender. The created certificate is
// valid for the device with the provided IPv4 address.
//
// It generates a public/private key pair, uses those keys to build an x509
// certificate, self-signs that certificate so the sender will trust it, and
// PEM-encodes that certificate and private key.
//
// Inspired by https://golang.org/src/crypto/tls/generate_cert.go
func GenerateSelfSignedCert(ip net.IP) (*SelfSignedCert, error) {
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

	return &SelfSignedCert{
		Bytes: certPEM.Bytes(),
		SK:    skPEM.Bytes(),
	}, nil
}
