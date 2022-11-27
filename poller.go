package sox

const (
	pollerDefaultEventsNum = 1 << 8
)

const (
	PollerEventIn  = 0x1
	PollerEventOut = 0x4
	PollerEventHup = 0x10
)

type poller interface {
	add(fd int, events uint32) error
	del(fd int) error
	wait(msec int) error
	event(i int) (fd int, events uint32)
}
