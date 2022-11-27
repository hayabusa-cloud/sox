//go:build linux

package sox

import (
	"errors"
	"golang.org/x/sys/unix"
	"time"
)

type TCPSocket struct {
	*socket
}

func newTCPSocket(sa unix.Sockaddr) (*TCPSocket, error) {
	network, fd, err := NetworkType(-1), 0, error(nil)
	if _, ok := sa.(*unix.SockaddrInet4); ok {
		fd, err = newTCP4Socket()
		if err != nil {
			return nil, err
		}
		network = NetworkIPv4
	} else if _, ok = sa.(*unix.SockaddrInet6); ok {
		fd, err = newTCP6Socket()
		if err != nil {
			return nil, err
		}
		network = NetworkIPv6
	} else {
		return nil, UnknownNetworkError("unexpected family")
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

	so := &TCPSocket{socket: newSocket(network, fd, sa)}
	return so, nil
}

func (so *TCPSocket) Protocol() UnderlyingProtocol {
	return UnderlyingProtocolStream
}

type TCPConn struct {
	*TCPSocket
	laddr *TCPAddr
	raddr *TCPAddr
}

func NewTCPConn(localAddr Addr, remoteSock *TCPSocket) (Conn, error) {
	if localAddr == nil {
		localAddr = &TCPAddr{}
	}
	tcpAddr, ok := localAddr.(*TCPAddr)
	if !ok {
		return nil, &AddrError{Err: "unexpected address type", Addr: localAddr.String()}
	}
	err := unix.SetsockoptInt(remoteSock.fd, unix.SOL_SOCKET, unix.SO_ZEROCOPY, 1)
	if err != nil {
		return nil, errFromUnixErrno(err)
	}
	remoteAddr := TCPAddrFromAddrPort(addrPortFromSockaddr(remoteSock.sa))
	return &TCPConn{TCPSocket: remoteSock, laddr: tcpAddr, raddr: remoteAddr}, nil
}

func (conn *TCPConn) LocalAddr() Addr {
	return conn.laddr
}
func (conn *TCPConn) RemoteAddr() Addr {
	return conn.raddr
}
func (conn *TCPConn) SetDeadline(t time.Time) error {
	return nil
}
func (conn *TCPConn) SetReadDeadline(t time.Time) error {
	return nil
}
func (conn *TCPConn) SetWriteDeadline(t time.Time) error {
	return nil
}

type TCPListener struct {
	*TCPSocket
	laddr *TCPAddr
}

func (l *TCPListener) Accept() (Conn, error) {
	nfd, sa, err := acceptWait(l.fd)
	if err != nil {
		return nil, err
	}

	so := &TCPSocket{socket: newSocket(l.network, nfd, sa)}
	conn, err := NewTCPConn(l.Addr(), so)
	if err != nil {
		return nil, err
	}
	return conn, err
}

func (l *TCPListener) Addr() Addr {
	if l.laddr != nil {
		return l.laddr
	}
	return TCPAddrFromAddrPort(addrPortFromSockaddr(l.sa))
}

func ListenTCP4(laddr *TCPAddr) (*TCPListener, error) {
	if laddr == nil {
		return nil, InvalidAddrError("nil local address")
	}
	so, err := newTCPSocket(tcp4AddrToSockaddr(laddr))
	if err != nil {
		return nil, err
	}
	err = unix.Bind(so.fd, tcp4AddrToSockaddr(laddr))
	if err != nil {
		return nil, errFromUnixErrno(err)
	}
	err = unix.Listen(so.fd, defaultBacklog)
	if err != nil {
		return nil, errFromUnixErrno(err)
	}

	lis := &TCPListener{TCPSocket: so, laddr: laddr}
	return lis, nil
}

func ListenTCP6(laddr *TCPAddr) (*TCPListener, error) {
	if laddr == nil {
		return nil, InvalidAddrError("nil local address")
	}
	so, err := newTCPSocket(tcp6AddrToSockaddr(laddr))
	if err != nil {
		return nil, err
	}
	err = unix.Bind(so.fd, tcp6AddrToSockaddr(laddr))
	if err != nil {
		return nil, errFromUnixErrno(err)
	}
	err = unix.Listen(so.fd, defaultBacklog)
	if err != nil {
		return nil, errFromUnixErrno(err)
	}

	lis := &TCPListener{TCPSocket: so, laddr: laddr}
	return lis, nil
}

func DialTCP4(laddr *TCPAddr, raddr *TCPAddr) (*TCPConn, error) {
	if raddr == nil {
		return nil, &OpError{Op: "dial", Net: "tcp4", Source: laddr, Addr: nil, Err: errors.New("missing address")}
	}
	so, err := newTCPSocket(tcp4AddrToSockaddr(laddr))
	if err != nil {
		return nil, err
	}
	err = connectWait(so.fd, tcp4AddrToSockaddr(raddr))
	if err != nil {
		return nil, err
	}

	conn := &TCPConn{
		TCPSocket: so,
		laddr:     laddr,
		raddr:     raddr,
	}
	return conn, nil
}

func DialTCP6(laddr *TCPAddr, raddr *TCPAddr) (*TCPConn, error) {
	if raddr == nil {
		return nil, &OpError{Op: "dial", Net: "udp6", Source: laddr, Addr: nil, Err: errors.New("missing address")}
	}
	so, err := newTCPSocket(tcp6AddrToSockaddr(laddr))
	if err != nil {
		return nil, err
	}
	err = connectWait(so.fd, tcp6AddrToSockaddr(raddr))
	if err != nil {
		return nil, err
	}

	conn := &TCPConn{
		TCPSocket: so,
		laddr:     laddr,
		raddr:     raddr,
	}
	return conn, nil
}

func newTCP4Socket() (fd int, err error) {
	fd, err = unix.Socket(unix.AF_INET, unix.SOCK_STREAM|unix.SOCK_NONBLOCK|unix.SOCK_CLOEXEC, unix.IPPROTO_TCP)
	if err != nil {
		return 0, errFromUnixErrno(err)
	}
	return fd, nil
}

func newTCP6Socket() (fd int, err error) {
	fd, err = unix.Socket(unix.AF_INET6, unix.SOCK_STREAM|unix.SOCK_NONBLOCK|unix.SOCK_CLOEXEC, unix.IPPROTO_TCP)
	if err != nil {
		return 0, errFromUnixErrno(err)
	}
	return fd, nil
}
