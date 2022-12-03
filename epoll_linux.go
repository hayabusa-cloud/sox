//go:build linux

package sox

import (
	"golang.org/x/sys/unix"
	"time"
	"unsafe"
)

type epoll struct {
	fd   int
	evts []unix.EpollEvent
}

func newPoller(n int) (*epoll, error) {
	if n < 1 {
		return nil, ErrInvalidParam
	}
	evts := make([]unix.EpollEvent, n)

	fd, err := unix.EpollCreate1(unix.EPOLL_CLOEXEC)
	if err != nil {
		return nil, errFromUnixErrno(err)
	}

	return &epoll{fd: fd, evts: evts}, nil
}

func (ep *epoll) FD() int {
	return ep.fd
}

func (ep *epoll) add(fd int, events uint32) error {
	evt := &unix.EpollEvent{
		Events: events | unix.EPOLLET,
		Fd:     int32(fd),
	}
	err := unix.EpollCtl(ep.fd, unix.EPOLL_CTL_ADD, fd, evt)
	if err != nil {
		return errFromUnixErrno(err)
	}

	return nil
}

func (ep *epoll) del(fd int) error {
	err := unix.EpollCtl(ep.fd, unix.EPOLL_CTL_DEL, fd, nil)
	if err != nil {
		return errFromUnixErrno(err)
	}

	return nil
}

func (ep *epoll) wait(d time.Duration) (events []pollerEvent, err error) {
	n, err := unix.EpollWait(ep.fd, ep.evts, int(d.Milliseconds()))
	if err != nil {
		return events, errFromUnixErrno(err)
	}
	ptr := (*pollerEvent)(unsafe.Pointer(&ep.evts[0]))
	events = unsafe.Slice(ptr, n)

	return
}

func (ep *epoll) Close() error {
	return unix.Close(ep.fd)
}
