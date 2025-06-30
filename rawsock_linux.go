// ©Hayabusa Cloud Co., Ltd. 2025. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

//go:build linux

package sox

import (
	"errors"
	"golang.org/x/sys/unix"
	"time"
)

type RawSocket struct {
	*socket
}

func newRawSocket(sa Sockaddr, protocol int) (*RawSocket, error) {
	network, fd, err := NetworkType(-1), 0, error(nil)
	if _, ok := sa.(*SockaddrInet4); ok {
		fd, err = newRaw4Socket(protocol)
		if err != nil {
			return nil, err
		}
		network = NetworkIPv4
	} else if _, ok = sa.(*SockaddrInet6); ok {
		fd, err = newRaw6Socket(protocol)
		if err != nil {
			return nil, err
		}
		network = NetworkIPv6
	} else {
		return nil, UnknownNetworkError("unexpected network family")
	}

	so := &RawSocket{socket: newSocket(network, fd, sa)}
	return so, nil
}

func (so *RawSocket) Protocol() UnderlyingProtocol {
	return UnderlyingProtocolRaw
}

func (so *RawSocket) RecvFrom(b []byte) (n int, addr Addr, err error) {
	n, sa, err := unix.Recvfrom(so.fd, b, 0)
	if err != nil {
		return n, nil, errFromUnixErrno(err)
	}

	return n, ipAddrFromSockaddr(sa), nil
}

func (so *RawSocket) SendTo(b []byte, raddr Addr) (n int, err error) {
	ra, ok := raddr.(*IPAddr)
	if !ok {
		return 0, InvalidAddrError(raddr.String())
	}
	err = unix.Sendto(so.fd, b, 0, inetAddrToSockaddr(ra))
	if err != nil {
		return 0, errFromUnixErrno(err)
	}

	return len(b), nil
}

type RawConn struct {
	*RawSocket
	laddr *IPAddr
	raddr *IPAddr
}

func (conn *RawConn) LocalAddr() Addr                    { return conn.laddr }
func (conn *RawConn) RemoteAddr() Addr                   { return conn.raddr }
func (conn *RawConn) SetDeadline(t time.Time) error      { return nil }
func (conn *RawConn) SetReadDeadline(t time.Time) error  { return nil }
func (conn *RawConn) SetWriteDeadline(t time.Time) error { return nil }
func (conn *RawConn) Read(p []byte) (n int, err error) {
	if conn.raddr == nil {
		n, _, err = conn.RecvFrom(p)
		return
	}
	n, err = unix.Read(conn.fd, p)
	if err != nil {
		return 0, errFromUnixErrno(err)
	}
	return n, nil
}
func (conn *RawConn) Write(p []byte) (n int, err error) {
	return conn.RawSocket.SendTo(p, conn.raddr)
}

func ListenRaw4(laddr *IPAddr, protocol int) (*RawConn, error) {
	if laddr == nil {
		return nil, InvalidAddrError("nil local address")
	}
	so, err := newRawSocket(ip4AddrToSockaddr(laddr), protocol)
	if err != nil {
		return nil, err
	}
	err = unix.Bind(so.fd, ip4AddrToSockaddr(laddr))
	if err != nil {
		return nil, errFromUnixErrno(err)
	}
	return &RawConn{RawSocket: so, laddr: laddr, raddr: nil}, nil
}

func ListenRaw6(laddr *IPAddr, protocol int) (*RawConn, error) {
	if laddr == nil {
		return nil, InvalidAddrError("nil local address")
	}
	so, err := newRawSocket(ip6AddrToSockaddr(laddr), protocol)
	if err != nil {
		return nil, err
	}
	err = unix.Bind(so.fd, ip6AddrToSockaddr(laddr))
	if err != nil {
		return nil, errFromUnixErrno(err)
	}
	return &RawConn{RawSocket: so, laddr: laddr, raddr: nil}, nil
}

func DialRaw4(laddr *IPAddr, raddr *IPAddr, protocol int) (*RawConn, error) {
	if laddr == nil {
		laddr = &IPAddr{IP: IPv4LoopBack}
	}
	if raddr == nil {
		return nil, &OpError{Op: "dial", Net: "raw4", Source: laddr, Addr: nil, Err: errors.New("missing address")}
	}
	lsa := ip4AddrToSockaddr(laddr)
	so, err := newRawSocket(lsa, protocol)
	if err != nil {
		return nil, err
	}
	err = connectWait(so.fd, ip4AddrToSockaddr(raddr))
	if err != nil {
		return nil, errFromUnixErrno(err)
	}
	return &RawConn{RawSocket: so, laddr: laddr, raddr: raddr}, nil
}

func DialRaw6(laddr *IPAddr, raddr *IPAddr, protocol int) (*RawConn, error) {
	if laddr == nil {
		laddr = &IPAddr{IP: IPv6LoopBack}
	}
	if raddr == nil {
		return nil, &OpError{Op: "dial", Net: "raw6", Source: laddr, Addr: nil, Err: errors.New("missing address")}
	}
	lsa := ip6AddrToSockaddr(laddr)
	so, err := newRawSocket(lsa, protocol)
	if err != nil {
		return nil, err
	}
	err = connectWait(so.fd, ip6AddrToSockaddr(raddr))
	if err != nil {
		return nil, errFromUnixErrno(err)
	}
	return &RawConn{RawSocket: so, laddr: laddr, raddr: raddr}, nil
}

func newRaw4Socket(protocol int) (fd int, err error) {
	fd, err = unix.Socket(unix.AF_INET, unix.SOCK_RAW|unix.SOCK_NONBLOCK|unix.SOCK_CLOEXEC, protocol)
	if err != nil {
		return 0, errFromUnixErrno(err)
	}
	return fd, nil
}

func newRaw6Socket(protocol int) (fd int, err error) {
	fd, err = unix.Socket(unix.AF_INET6, unix.SOCK_RAW|unix.SOCK_NONBLOCK|unix.SOCK_CLOEXEC, protocol)
	if err != nil {
		return 0, errFromUnixErrno(err)
	}
	return fd, nil
}
