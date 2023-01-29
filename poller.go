// ©Hayabusa Cloud Co., Ltd. 2022. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package sox

import (
	"io"
	"os"
	"time"
)

const (
	pollerDefaultEventsNum = 1 << 10
)

const (
	pollerEventIn  = 0x1
	pollerEventOut = 0x4
	pollerEventHup = 0x10
)

type pollerEvent struct {
	Events uint32
	Fd     int32
	pad    [4]byte
}

type poller interface {
	add(fd int, events uint32) error
	del(fd int) error
	wait(d time.Duration) (events []pollerEvent, err error)
	Close() error
}

type pollFd interface {
	// Fd returns the file descriptor
	Fd() int
}

// PollReader is the interface that groups Fd and the basic Read method
type PollReader interface {
	pollFd
	io.Reader
}

// PollUintReader is the interface that groups ReadUint()
// and the methods in interface PollReader
type PollUintReader interface {
	PollReader
	ReadUint() (val uint, err error)
}

// PollWriter is the interface that groups Fd and the basic Write method
type PollWriter interface {
	pollFd
	io.Writer
}

// PollUintWriter is the interface that groups WriteUint()
// and the methods in interface PollWriter
type PollUintWriter interface {
	PollWriter
	WriteUint(val uint) error
}

// PollReadWriter is the interface that groups the methods
// in interface PollReader and PollWriter
type PollReadWriter interface {
	PollReader
	PollWriter
}

// PollUintReadWriter is the interface that groups the methods
// in interface PollUintReader and PollUintWriter
type PollUintReadWriter interface {
	PollUintReader
	PollUintWriter
}

// PollCloser is the interface that groups Fd and the basic Close method
type PollCloser interface {
	pollFd
	io.Closer
}

// PollReadCloser is the interface that groups the methods
// in interface PollReader and PollCloser
type PollReadCloser interface {
	PollReader
	io.Closer
}

// PollWriteCloser is the interface that groups the methods
// in interface PollWriter and PollCloser
type PollWriteCloser interface {
	PollWriter
	io.Closer
}

// PollReadWriteCloser is the interface that groups the methods
// in interface PollReader, PollWriter and PollCloser
type PollReadWriteCloser interface {
	PollReader
	PollWriter
	io.Closer
}

// PollUintReadCloser is the interface that groups the methods
// in interface PollUintReader and PollUintCloser
type PollUintReadCloser interface {
	PollUintReader
	io.Closer
}

// PollUintWriteCloser is the interface that groups the methods
// in interface PollUintWriter and PollCloser
type PollUintWriteCloser interface {
	PollUintWriter
	io.Closer
}

// PollUintReadWriteCloser is the interface that groups the methods
// in interface PollUintReader, PollUintWriter and PollCloser
type PollUintReadWriteCloser interface {
	PollUintReader
	PollUintWriter
	io.Closer
}

type file interface {
	File() (f *os.File, err error)
}
