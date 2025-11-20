//go:build unit
// +build unit

package dwc

import (
	"bytes"
	"io/fs"
	"reflect"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

type readerWriterCloserMock struct {
	closed bool
}

func (w *readerWriterCloserMock) Read(p []byte) (n int, err error) {
	return 0, nil
}

func (w *readerWriterCloserMock) Write(p []byte) (n int, err error) {
	return 0, nil
}

func (w *readerWriterCloserMock) Close() error {
	w.closed = true
	return nil
}

type zipFileWriterMock struct {
	buffer *bytes.Buffer
}

func (z *zipFileWriterMock) Write(p []byte) (n int, err error) {
	return z.buffer.Write(p)
}

type DirEntryMock struct {
	fileInfo *fileInfoMock
}

func (d *DirEntryMock) Name() string {
	return d.fileInfo.fileName
}

func (d *DirEntryMock) IsDir() bool {
	return d.fileInfo.isDir
}

func (d *DirEntryMock) Info() (fs.FileInfo, error) {
	return d.fileInfo, nil
}

func (d *DirEntryMock) Type() fs.FileMode {
	return d.fileInfo.Mode()
}

type fileInfoMock struct {
	fileName string
	isDir    bool
}

func (f *fileInfoMock) Name() string {
	return f.fileName
}

func (f *fileInfoMock) Size() int64 {
	return 88
}

func (f *fileInfoMock) Mode() fs.FileMode {
	return 0777
}

func (f *fileInfoMock) ModTime() time.Time {
	return time.Now()
}

func (f *fileInfoMock) IsDir() bool {
	return f.isDir
}

func (f *fileInfoMock) Sys() any {
	return nil
}

func verifyDwCCommand(t *testing.T, got dwcCommand, wanted dwcCommand, cmdBaseLength int) {
	t.Helper()
	if !assert.ElementsMatch(t, wanted, got) { // check if we have the same command parts in any order. This primarily verifies that all flags are set.
		t.Fatalf("Expected command to contain elements %v, got %v", wanted, got)
	}
	if got != nil { // check the base cmd via reflection: The base command is order-sensitive. If got is nil here wanted is also nil: which passes the test.
		baseCmdEmitted := got[0:cmdBaseLength]
		baseCmdWanted := wanted[0:cmdBaseLength]
		if !reflect.DeepEqual(baseCmdWanted, baseCmdEmitted) {
			t.Fatalf("Expected base cmd to be %v, got %v", baseCmdWanted, baseCmdEmitted)
		}
	}
}
