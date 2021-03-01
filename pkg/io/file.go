package io

import (
	"crypto/tls"
	"io"
	"os"

	"github.com/nchaloult/lancp/pkg/net"
)

// TODO: make these user-configurable.
const defaultFilePayloadBufSize = 8192

// TODO: implement timeout and retry logic.
// TODO: draw progress with ioprogress pkg.
func SendFileAlongConn(f *os.File, size int64, conn *tls.Conn) error {
	payloadSize := min(size, defaultFilePayloadBufSize)
	payloadBuf := make([]byte, payloadSize)
	for {
		_, err := f.Read(payloadBuf)
		if err != nil {
			if err == io.EOF {
				break
			}
			return err
		}

		if err = net.SendTLSMessage(payloadBuf, conn); err != nil {
			return err
		}
	}

	return nil
}

// This ought to be in the standard library imo.
func min(x, y int64) int64 {
	if x < y {
		return x
	}
	return y
}
