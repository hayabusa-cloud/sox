// ©Hayabusa Cloud Co., Ltd. 2022. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

//go:build unix

package sox

import (
	"golang.org/x/sys/unix"
)

type socket struct {
	network NetworkType
	fd      int
	sa      unix.Sockaddr
}

func newSocket(network NetworkType, fd int, sa unix.Sockaddr) *socket {
	return &socket{fd: fd, sa: sa, network: network}
}

func (so *socket) Fd() int {
	return so.fd
}

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

func (so *socket) Recvmsg(buffers [][]byte, oob []byte) (n, oobn int, recvflags int, from unix.Sockaddr, err error) {
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
