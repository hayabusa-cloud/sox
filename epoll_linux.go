//go:build linux

package sox

import "golang.org/x/sys/unix"

type epoll struct {
	fd   int
	evts []unix.EpollEvent
	n    int
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

	return &epoll{fd: fd, evts: evts, n: 0}, nil
}

func (ep *epoll) FD() int {
	return ep.fd
}

func (ep *epoll) add(fd int, events uint32) error {
	evt := &unix.EpollEvent{
		Events: events,
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

func (ep *epoll) wait(msec int) error {
	n, err := unix.EpollWait(ep.fd, ep.evts, msec)
	if err != nil {
		return errFromUnixErrno(err)
	}

	ep.n = n

	return nil
}

func (ep *epoll) event(i int) (fd int, events uint32) {
	if i < 0 || i >= ep.n {
		return 0, 0
	}
	evt := ep.evts[i]
	return int(evt.Fd), evt.Events
}
