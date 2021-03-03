package io

import (
	"crypto/tls"
	"fmt"
	"io"
	_net "net"
	"os"
	"path/filepath"
	"strings"

	"github.com/nchaloult/lancp/pkg/net"
)

// TODO: make these user-configurable.
const (
	defaultFilePayloadBufSize = 8192
	progressBarLen            = 40 // # of characters.
)

// CreateNewFileOnDisk attempts to create a new file in the user's current
// directory. If a file in that directory already exists with the same name,
// then it appends " (x)" to the file name, where x is the lowest revision
// number possible.
func CreateNewFileOnDisk(name string) (*os.File, error) {
	file, err := os.OpenFile(
		name,
		os.O_RDWR|os.O_CREATE|os.O_EXCL, 0666,
	)

	// If another file with that name already exists, keep adding a suffix until
	// we get a non-existent file name.
	versionNum := 1
	for os.IsExist(err) {
		ext := filepath.Ext(name)
		basename := strings.TrimSuffix(name, ext)

		file, err = os.OpenFile(
			fmt.Sprintf("%s (%d)%s", basename, versionNum, ext),
			os.O_RDWR|os.O_CREATE|os.O_EXCL, 0666,
		)

		versionNum++
	}

	return file, err
}

// ReceiveFileFromConn reads a payload sent along the provided network
// connection and writes it to a file.
//
// TODO: implement timeout and retry logic.
func ReceiveFileFromConn(file *os.File, size int64, conn _net.Conn) error {
	progressReader := getProgressReader(size, conn, progressBarLen)
	_, err := io.Copy(file, progressReader)
	return err
}

// TODO: implement timeout and retry logic.
// TODO: draw progress with ioprogress pkg.
func SendFileAlongConn(file *os.File, size int64, conn *tls.Conn) error {
	progressReader := getProgressReader(size, file, progressBarLen)

	payloadSize := min(size, defaultFilePayloadBufSize)
	payloadBuf := make([]byte, payloadSize)
	for {
		_, err := progressReader.Read(payloadBuf)
		if err != nil {
			if err == io.EOF {
				break
			}
			return err
		}

		if err = net.SendMessage(payloadBuf, conn); err != nil {
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
