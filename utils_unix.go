//go:build unix

package sox

import (
	"golang.org/x/sys/unix"
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
