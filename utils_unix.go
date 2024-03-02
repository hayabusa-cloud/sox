// ©Hayabusa Cloud Co., Ltd. 2022. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

//go:build unix

package sox

import (
	"golang.org/x/sys/unix"
	"unsafe"
)

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
		return ErrInterruptedSyscall
	case unix.EAGAIN:
		return ErrTemporarilyUnavailable
	case unix.EINPROGRESS:
		return ErrInProgress
	case unix.EFAULT:
		return ErrFaultParams
	case unix.EINVAL:
		return ErrInvalidParam
	case unix.EMFILE:
		return ErrProcessFileLimit
	case unix.ENFILE:
		return ErrSystemFileLimit
	case unix.ENODEV:
		return ErrNoDevice
	case unix.ENOMEM:
		return ErrNoAvailableMemory
	case unix.EPERM:
		return ErrNoPermission
	default:
		return errno
	}
}

func ioVecFromBytesSlice(iov [][]byte) (addr uintptr, n int) {
	vec := make([]unix.Iovec, len(iov))
	for i := range len(iov) {
		vec[i] = unix.Iovec{Base: &iov[i][0], Len: uint64(len(iov[i]))}
	}
	addr, n = uintptr(unsafe.Pointer(&iov)), len(iov)

	return
}
