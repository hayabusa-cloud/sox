// ©Hayabusa Cloud Co., Ltd. 2025. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

//go:build linux

package sox_test

import (
	"code.hybscloud.com/sox"
	"golang.org/x/sys/unix"
)

// test helper socket
type socket struct {
	network sox.NetworkType
	typ     sox.UnderlyingProtocol
	fd      int
	sa      sox.Sockaddr
}

func (s *socket) Fd() int {
	return s.fd
}

func (s *socket) NetworkType() sox.NetworkType {
	return s.network
}

func (s *socket) Protocol() sox.UnderlyingProtocol {
	if s.typ > 0 {
		return s.typ
	}
	v, err := unix.GetsockoptInt(s.fd, unix.SOL_SOCKET, unix.SO_TYPE)
	if err != nil {
		return 0
	}
	s.typ = sox.UnderlyingProtocol(v & 0xff)
	return s.typ
}

func (s *socket) Read(b []byte) (n int, err error) {
	n, err = unix.Read(s.fd, b)
	if err != nil {
		return n, errFromUnixErrno(err)
	}
	return n, nil
}

func (s *socket) Write(b []byte) (n int, err error) {
	err = unix.Send(s.fd, b, unix.MSG_ZEROCOPY)
	if err != nil {
		return 0, errFromUnixErrno(err)
	}
	return len(b), nil
}

func (s *socket) Close() error {
	return unix.Close(s.fd)
}

func errFromUnixErrno(err error) error {
	if err == nil {
		return nil
	}
	errno, ok := err.(unix.Errno)
	if !ok {
		return err
	}
	switch errno {
	case unix.EINTR:
		return sox.ErrInterruptedSyscall
	case unix.EAGAIN:
		return sox.ErrTemporarilyUnavailable
	case unix.EINPROGRESS:
		return sox.ErrInProgress
	case unix.EFAULT:
		return sox.ErrFaultParams
	case unix.EINVAL:
		return sox.ErrInvalidParam
	case unix.EMFILE:
		return sox.ErrProcessFileLimit
	case unix.ENFILE:
		return sox.ErrSystemFileLimit
	case unix.ENODEV:
		return sox.ErrNoDevice
	case unix.ENOMEM:
		return sox.ErrNoAvailableMemory
	case unix.EPERM:
		return sox.ErrNoPermission
	default:
		return errno
	}
}
