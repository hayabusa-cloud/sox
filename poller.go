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
	Fd() int
}

type PollReader interface {
	pollFd
	io.Reader
}

type PollWriter interface {
	pollFd
	io.Writer
}

type file interface {
	File() (f *os.File, err error)
}
