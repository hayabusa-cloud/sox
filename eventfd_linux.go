// ©Hayabusa Cloud Co., Ltd. 2022. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

//go:build linux

package sox

import (
	"encoding/binary"
	"golang.org/x/sys/unix"
)

type eventfd int

// NewEventfd creates and returns a new nonblocking eventfd as a PollUintReadWriteCloser
func NewEventfd() (PollUintReadWriteCloser, error) {
	fd, err := unix.Eventfd(0, unix.EFD_NONBLOCK|unix.EFD_CLOEXEC)
	if err != nil {
		return nil, errFromUnixErrno(err)
	}

	return eventfd(fd), nil
}

func (fd eventfd) Fd() int {
	return int(fd)
}

func (fd eventfd) Read(p []byte) (n int, err error) {
	n, err = unix.Read(int(fd), p)
	if err != nil {
		return n, errFromUnixErrno(err)
	}

	return n, nil
}

func (fd eventfd) ReadUint64() (val uint64, err error) {
	var buf [8]byte
	_, err = fd.Read(buf[:])
	if err != nil {
		return 0, errFromUnixErrno(err)
	}
	val = binary.LittleEndian.Uint64(buf[:])

	return val, nil
}

func (fd eventfd) ReadUint() (val uint, err error) {
	u64, err := fd.ReadUint64()
	return uint(u64), err
}

func (fd eventfd) Write(p []byte) (n int, err error) {
	n, err = unix.Write(int(fd), p)
	if err != nil {
		return 0, errFromUnixErrno(err)
	}

	return n, nil
}

func (fd eventfd) WriteUint64(val uint64) error {
	var buf [8]byte
	binary.LittleEndian.PutUint64(buf[:], val)
	_, err := fd.Write(buf[:])

	return err
}

func (fd eventfd) WriteUint(val uint) error {
	return fd.WriteUint64(uint64(val))
}

func (fd eventfd) Close() error {
	err := unix.Close(int(fd))
	if err != nil {
		return errFromUnixErrno(err)
	}
	return nil
}
