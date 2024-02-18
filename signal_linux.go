// ©Hayabusa Cloud Co., Ltd. 2024. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

//go:build linux

package sox

import (
	"encoding/binary"
	"golang.org/x/sys/unix"
	"unsafe"
)

const (
	_NSIG       = 64
	_NSIG_BPW   = __BITS_PER_LONG
	_NSIG_WORDS = (_NSIG / _NSIG_BPW)
)

type signalfd int

// NewSignalFile creates and returns a new signal fd
func NewSignalFile() (signalFile PollSignalfd, err error) {
	return newSignalfd()
}

func newSignalfd() (signalfd, error) {
	set := unix.Sigset_t{Val: [16]uint64{}}
	sigAddSet(&set, unix.SIGHUP)
	sigAddSet(&set, unix.SIGINT)
	sigAddSet(&set, unix.SIGQUIT)
	sigAddSet(&set, unix.SIGBUS)
	sigAddSet(&set, unix.SIGUSR1)
	sigAddSet(&set, unix.SIGUSR2)
	sigAddSet(&set, unix.SIGPIPE)
	sigAddSet(&set, unix.SIGTERM)
	sigAddSet(&set, unix.SIGCHLD)
	fd, err := unix.Signalfd(-1, &set, unix.SFD_CLOEXEC)
	if err != nil {
		return -1, errFromUnixErrno(err)
	}

	return signalfd(fd), nil
}

func (fd signalfd) Fd() int {
	return int(fd)
}

func (fd signalfd) Read(p []byte) (n int, err error) {
	n, err = unix.Read(int(fd), p)
	if err != nil {
		return n, errFromUnixErrno(err)
	}

	return n, nil
}

func (fd signalfd) ReadUint64() (val uint64, err error) {
	var buf [8]byte
	_, err = fd.Read(buf[:])
	if err != nil {
		return 0, errFromUnixErrno(err)
	}
	val = binary.LittleEndian.Uint64(buf[:])

	return val, nil
}

func (fd signalfd) ReadUint() (val uint, err error) {
	u64, err := fd.ReadUint64()
	return uint(u64), err
}

func (fd signalfd) ReadSiginfo() (sig unix.Signal, code int, err error) {
	var buf [128]byte
	_, err = fd.Read(buf[:])
	if err != nil {
		return -1, -1, err
	}
	s := (*unix.Siginfo)(unsafe.Pointer(&buf))

	return unix.Signal(s.Signo), int(s.Code), nil
}

func (fd signalfd) Close() error {
	err := unix.Close(int(fd))
	if err != nil {
		return errFromUnixErrno(err)
	}
	return nil
}

func sigAddSet(set *unix.Sigset_t, _sig unix.Signal) {
	sig := uint(_sig - 1)
	if _NSIG_WORDS == 1 {
		set.Val[0] |= 1 << sig
	} else {
		set.Val[sig/_NSIG_BPW] |= 1 << (sig % _NSIG_BPW)
	}
}
