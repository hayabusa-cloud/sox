package sox

import "time"

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
