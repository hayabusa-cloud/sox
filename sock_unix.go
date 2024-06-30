// ©Hayabusa Cloud Co., Ltd. 2022. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

//go:build unix

package sox

import (
	"golang.org/x/sys/unix"
	"unsafe"
)

const (

	// SockShutdownRead is a constant that represents the action of shutting down
	// the reading side of a socket. It is used in conjunction with the unix.SHUT_RD
	// constant to specify the shutdown action.
	SockShutdownRead = unix.SHUT_RD

	// SockShutdownWrite represents a constant that is used to shutdown the write
	// function of a socket. It is typically used with the SockShutdown function
	// in the unix package.
	SockShutdownWrite = unix.SHUT_WR

	// SockShutdownReadWrite is a constant that represents the value `unix.SHUT_RDWR`.
	// It can be used to specify shutting down both the reading and writing operations
	// of a socket.
	SockShutdownReadWrite = unix.SHUT_RDWR
)

const (
	defaultBacklog = 511
)

type socket struct {
	network NetworkType
	fd      int
	sa      Sockaddr
}

func newSocket(network NetworkType, fd int, sa Sockaddr) *socket {
	return &socket{fd: fd, sa: sa, network: network}
}

func (so *socket) Fd() int {
	return so.fd
}

func (so *socket) NetworkType() NetworkType { return so.network }

func (so *socket) Readv(iovs [][]byte) (n int, err error) {
	n, err = unix.Readv(so.fd, iovs)
	if err != nil {
		return n, errFromUnixErrno(err)
	}
	return n, nil
}

func (so *socket) Writev(iovs [][]byte) (n int, err error) {
	n, err = unix.Writev(so.fd, iovs)
	if err != nil {
		return n, errFromUnixErrno(err)
	}
	return n, nil
}

func (so *socket) Recvmsg(buffers [][]byte, oob []byte) (n, oobn int, recvflags int, from Sockaddr, err error) {
	n, oobn, recvflags, from, err = unix.RecvmsgBuffers(so.fd, buffers, oob, unix.MSG_WAITALL)
	if err != nil {
		return 0, 0, recvflags, nil, errFromUnixErrno(err)
	}
	return
}

func (so *socket) Sendmsg(buffers [][]byte, oob []byte, to Addr) (n int, err error) {
	sa := unix.Sockaddr(nil)
	if to != nil {
		switch so.network {
		case NetworkUnix:
			sa = unixAddrToSockaddr(to.(*UnixAddr))
		case NetworkIPv4:
			sa = inet4AddrToSockaddr(to)
		case NetworkIPv6:
			sa = inet6AddrToSockaddr(to)
		}
	}
	n, err = unix.SendmsgBuffers(so.fd, buffers, oob, sa, unix.MSG_ZEROCOPY)
	if err != nil {
		return 0, errFromUnixErrno(err)
	}
	return n, nil
}

func (so *socket) Read(b []byte) (n int, err error) {
	n, err = unix.Read(so.fd, b)
	if err != nil {
		return n, errFromUnixErrno(err)
	}
	return n, nil
}

// todo: poll the completed notification from uring to know
// it is safe to reuse a previously passed buffer or not
func (so *socket) Write(b []byte) (n int, err error) {
	err = unix.Send(so.fd, b, unix.MSG_ZEROCOPY)
	if err != nil {
		return 0, errFromUnixErrno(err)
	}
	return len(b), nil
}

func (so *socket) Close() error {
	return unix.Close(so.fd)
}

func acceptWait(fd int) (nfd int, sa Sockaddr, err error) {
	for sw := NewParamSpinWait(); !sw.Closed(); sw.Once() {
		nfd, sa, err = unix.Accept4(fd, unix.SOCK_NONBLOCK|unix.SOCK_CLOEXEC)
		if err == unix.EAGAIN || err == unix.EWOULDBLOCK {
			continue
		}
		if err != nil {
			return 0, nil, errFromUnixErrno(err)
		}
		break
	}
	return
}

func connectWait(fd int, sa Sockaddr) error {
	if err := unix.Connect(fd, sa); err == nil {
		return nil
	} else if err != unix.EINPROGRESS {
		return errFromUnixErrno(err)
	}
	for sw := NewParamSpinWait(); !sw.Closed(); sw.Once() {
		val, err := unix.GetsockoptInt(fd, unix.SOL_SOCKET, unix.SO_ERROR)
		if err != nil {
			return errFromUnixErrno(err)
		}
		if val == 0 {
			break
		}
	}
	return nil
}

func getSockOptWait(fd int, level int, opt int, optVal unsafe.Pointer, optLen unsafe.Pointer) error {
	for sw := NewParamSpinWait(); !sw.Closed(); sw.Once() {
		_, _, errno := unix.Syscall6(unix.SYS_GETSOCKOPT, uintptr(fd), uintptr(level), uintptr(opt), uintptr(optVal), uintptr(optLen), 0)
		if errno == 0 {
			break
		} else if errno == unix.EINPROGRESS {
			continue
		}
		return errFromUnixErrno(errno)
	}
	return nil
}
