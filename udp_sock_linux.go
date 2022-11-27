//go:build linux

package sox

import (
	"errors"
	"golang.org/x/sys/unix"
	"time"
)

type UDPSocket struct {
	*socket
}

func newUDPSocket(sa unix.Sockaddr) (*UDPSocket, error) {
	network, fd, err := NetworkType(-1), 0, error(nil)
	if _, ok := sa.(*unix.SockaddrInet4); ok {
		fd, err = newUDP4Socket()
		if err != nil {
			return nil, err
		}
		network = NetworkIPv4
	} else if _, ok = sa.(*unix.SockaddrInet6); ok {
		fd, err = newUDP6Socket()
		if err != nil {
			return nil, err
		}
		network = NetworkIPv6
	} else {
		return nil, UnknownNetworkError("unexpected family")
	}
	err = unix.SetsockoptInt(fd, unix.SOL_SOCKET, unix.SO_BROADCAST, 1)
	if err != nil {
		return nil, errFromUnixErrno(err)
	}
	err = unix.SetsockoptInt(fd, unix.SOL_SOCKET, unix.SO_REUSEADDR, 1)
	if err != nil {
		return nil, errFromUnixErrno(err)
	}
	err = unix.SetsockoptInt(fd, unix.SOL_SOCKET, unix.SO_REUSEPORT, 1)
	if err != nil {
		return nil, errFromUnixErrno(err)
	}
	err = unix.SetsockoptInt(fd, unix.SOL_SOCKET, unix.SO_ZEROCOPY, 1)
	if err != nil {
		return nil, errFromUnixErrno(err)
	}

	so := &UDPSocket{socket: newSocket(network, fd, sa)}
	return so, nil
}

func (so *UDPSocket) Protocol() UnderlyingProtocol {
	return UnderlyingProtocolDgram
}

func (so *UDPSocket) Dial4(raddr *UDPAddr) (conn *UDPConn, err error) {
	err = unix.Connect(so.fd, udp4AddrToSockaddr(raddr))
	if err != nil {
		return nil, errFromUnixErrno(err)
	}
	return &UDPConn{UDPSocket: so, laddr: UDPAddrFromAddrPort(addrPortFromSockaddr(so.sa)), raddr: raddr}, nil
}

func (so *UDPSocket) Dial6(raddr *UDPAddr) (conn *UDPConn, err error) {
	err = unix.Connect(so.fd, udp6AddrToSockaddr(raddr))
	if err != nil {
		return nil, errFromUnixErrno(err)
	}
	return &UDPConn{UDPSocket: so, laddr: UDPAddrFromAddrPort(addrPortFromSockaddr(so.sa)), raddr: raddr}, nil
}

func (so *UDPSocket) RecvFrom(b []byte) (n int, addr Addr, err error) {
	n, sa, err := unix.Recvfrom(so.fd, b, 0)
	if err != nil {
		return n, nil, errFromUnixErrno(err)
	}

	return n, UDPAddrFromAddrPort(addrPortFromSockaddr(sa)), nil
}

func (so *UDPSocket) SendTo(b []byte, raddr Addr) (n int, err error) {
	ra, ok := raddr.(*UDPAddr)
	if !ok {
		return 0, InvalidAddrError(raddr.String())
	}
	err = unix.Sendto(so.fd, b, 0, inetAddrFromAddrPort(ra.AddrPort()))
	if err != nil {
		return 0, errFromUnixErrno(err)
	}

	return len(b), nil
}

type UDPConn struct {
	*UDPSocket
	laddr *UDPAddr
	raddr *UDPAddr
}

func NewUDPConn(localAddr Addr, remoteSock *UDPSocket) (Conn, error) {
	if localAddr == nil {
		localAddr = &UDPAddr{}
	}
	udpAddr, ok := localAddr.(*UDPAddr)
	if !ok {
		return nil, &AddrError{Err: "unexpected address type", Addr: localAddr.String()}
	}
	err := unix.SetsockoptInt(remoteSock.fd, unix.SOL_SOCKET, unix.SO_ZEROCOPY, 1)
	if err != nil {
		return nil, errFromUnixErrno(err)
	}
	remoteAddr := UDPAddrFromAddrPort(addrPortFromSockaddr(remoteSock.sa))
	return &UDPConn{UDPSocket: remoteSock, laddr: udpAddr, raddr: remoteAddr}, nil
}

func (conn *UDPConn) LocalAddr() Addr {
	return conn.laddr
}
func (conn *UDPConn) RemoteAddr() Addr {
	return conn.raddr
}
func (conn *UDPConn) SetDeadline(t time.Time) error {
	return nil
}
func (conn *UDPConn) SetReadDeadline(t time.Time) error {
	return nil
}
func (conn *UDPConn) SetWriteDeadline(t time.Time) error {
	return nil
}
func (conn *UDPConn) Write(p []byte) (n int, err error) {
	return conn.UDPSocket.SendTo(p, conn.raddr)
}

func ListenUDP4(laddr *UDPAddr) (*UDPConn, error) {
	if laddr == nil {
		return nil, InvalidAddrError("nil local address")
	}
	so, err := newUDPSocket(udp4AddrToSockaddr(laddr))
	if err != nil {
		return nil, err
	}
	err = unix.Bind(so.fd, udp4AddrToSockaddr(laddr))
	if err != nil {
		return nil, errFromUnixErrno(err)
	}
	return &UDPConn{UDPSocket: so, laddr: laddr, raddr: nil}, nil
}

func ListenUDP6(laddr *UDPAddr) (*UDPConn, error) {
	if laddr == nil {
		return nil, InvalidAddrError("nil local address")
	}
	so, err := newUDPSocket(udp6AddrToSockaddr(laddr))
	if err != nil {
		return nil, err
	}
	err = unix.Bind(so.fd, udp6AddrToSockaddr(laddr))
	if err != nil {
		return nil, errFromUnixErrno(err)
	}
	return &UDPConn{UDPSocket: so, laddr: laddr, raddr: nil}, nil
}

func DialUDP4(laddr *UDPAddr, raddr *UDPAddr) (*UDPConn, error) {
	if raddr == nil {
		return nil, &OpError{Op: "dial", Net: "udp4", Source: laddr, Addr: nil, Err: errors.New("missing address")}
	}
	so, err := newUDPSocket(udp4AddrToSockaddr(laddr))
	if err != nil {
		return nil, err
	}
	return so.Dial4(raddr)
}

func DialUDP6(laddr *UDPAddr, raddr *UDPAddr) (*UDPConn, error) {
	if raddr == nil {
		return nil, &OpError{Op: "dial", Net: "udp6", Source: laddr, Addr: nil, Err: errors.New("missing address")}
	}
	so, err := newUDPSocket(udp6AddrToSockaddr(laddr))
	if err != nil {
		return nil, err
	}
	return so.Dial6(raddr)
}

func newUDP4Socket() (fd int, err error) {
	fd, err = unix.Socket(unix.AF_INET, unix.SOCK_DGRAM|unix.SOCK_NONBLOCK|unix.SOCK_CLOEXEC, unix.IPPROTO_UDP)
	if err != nil {
		return 0, errFromUnixErrno(err)
	}
	return
}

func newUDP6Socket() (fd int, err error) {
	fd, err = unix.Socket(unix.AF_INET6, unix.SOCK_DGRAM|unix.SOCK_NONBLOCK|unix.SOCK_CLOEXEC, unix.IPPROTO_UDP)
	if err != nil {
		return 0, errFromUnixErrno(err)
	}
	return
}
