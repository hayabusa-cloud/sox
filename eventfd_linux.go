//go:build linux

package sox

import (
	"encoding/binary"
	"golang.org/x/sys/unix"
	"runtime"
	"sync/atomic"
)

type eventfd struct {
	buf  []byte
	lock *atomic.Bool

	fd int
}

func newEventfd() (*eventfd, error) {
	fd, err := unix.Eventfd(0, unix.EFD_NONBLOCK|unix.EFD_CLOEXEC)
	if err != nil {
		return nil, errFromUnixErrno(err)
	}

	return &eventfd{buf: make([]byte, 8), lock: &atomic.Bool{}, fd: fd}, nil
}

func (e *eventfd) FD() int {
	return e.fd
}

func (e *eventfd) Read(p []byte) (n int, err error) {
	n, err = unix.Read(e.fd, p)
	if err != nil {
		return n, errFromUnixErrno(err)
	}

	return n, nil
}

func (e *eventfd) ReadUint() (val uint, err error) {
	for !e.lock.CompareAndSwap(false, true) {
		runtime.Gosched()
	}
	defer e.lock.Store(false)
	_, err = e.Read(e.buf)
	if err != nil {
		return 0, err
	}
	val = uint(binary.LittleEndian.Uint64(e.buf))

	return val, nil
}

func (e *eventfd) Write(p []byte) (n int, err error) {
	n, err = unix.Write(e.fd, p)
	if err != nil {
		return 0, errFromUnixErrno(err)
	}

	return n, nil
}

func (e *eventfd) WriteUint(val uint) error {
	for !e.lock.CompareAndSwap(false, true) {
		runtime.Gosched()
	}
	defer e.lock.Store(false)
	binary.LittleEndian.PutUint64(e.buf, uint64(val))
	_, err := e.Write(e.buf)

	return err
}

func (e *eventfd) Close() error {
	for !e.lock.CompareAndSwap(false, true) {
		runtime.Gosched()
	}

	err := unix.Close(e.fd)
	if err != nil {
		return errFromUnixErrno(err)
	}
	return nil
}
