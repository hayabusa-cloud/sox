//go:build linux

package sox

import (
	"errors"
	"golang.org/x/sys/unix"
	"time"
)

type SCTPSocket struct {
	*socket
}

func newSCTPSocket(sa unix.Sockaddr) (*SCTPSocket, error) {
	network, fd, err := NetworkType(-1), 0, error(nil)
	if _, ok := sa.(*unix.SockaddrInet4); ok {
		fd, err = newSCTP4Socket()
		if err != nil {
			return nil, err
		}
		network = NetworkIPv4
	} else if _, ok = sa.(*unix.SockaddrInet6); ok {
		fd, err = newSCTP6Socket()
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

	so := &SCTPSocket{socket: newSocket(network, fd, sa)}
	return so, nil
}

func (so *SCTPSocket) Protocol() UnderlyingProtocol {
	return UnderlyingProtocolSeqPacket
}

type SCTPConn struct {
	*SCTPSocket
	laddr *SCTPAddr
	raddr *SCTPAddr
}

func NewSCTPConn(localAddr Addr, remoteSock *SCTPSocket) (Conn, error) {
	if localAddr == nil {
		localAddr = &SCTPAddr{}
	}
	sctpAddr, ok := localAddr.(*SCTPAddr)
	if !ok {
		return nil, &AddrError{Err: "unexpected address type", Addr: localAddr.String()}
	}
	err := unix.SetsockoptInt(remoteSock.fd, unix.SOL_SOCKET, unix.SO_ZEROCOPY, 1)
	if err != nil {
		return nil, errFromUnixErrno(err)
	}
	remoteAddr := SCTPAddrFromAddrPort(addrPortFromSockaddr(remoteSock.sa))
	return &SCTPConn{SCTPSocket: remoteSock, laddr: sctpAddr, raddr: remoteAddr}, nil
}

func (conn *SCTPConn) LocalAddr() Addr {
	return conn.laddr
}
func (conn *SCTPConn) RemoteAddr() Addr {
	return conn.raddr
}
func (conn *SCTPConn) SetDeadline(t time.Time) error {
	return nil
}
func (conn *SCTPConn) SetReadDeadline(t time.Time) error {
	return nil
}
func (conn *SCTPConn) SetWriteDeadline(t time.Time) error {
	return nil
}

type SCTPListener struct {
	*SCTPSocket
	laddr *SCTPAddr
}

func (l *SCTPListener) Accept() (Conn, error) {
	nfd, sa, err := acceptWait(l.fd)
	if err != nil {
		return nil, err
	}

	so := &SCTPSocket{socket: newSocket(l.network, nfd, sa)}
	conn, err := NewSCTPConn(l.laddr, so)
	if err != nil {
		return nil, err
	}
	return conn, err
}
func (l *SCTPListener) Close() error {
	return l.SCTPSocket.Close()
}
func (l *SCTPListener) Addr() Addr {
	if l.laddr != nil {
		return l.laddr
	}
	return SCTPAddrFromAddrPort(addrPortFromSockaddr(l.sa))
}

func ListenSCTP4(laddr *SCTPAddr) (*SCTPListener, error) {
	if laddr == nil {
		return nil, InvalidAddrError("nil local address")
	}
	so, err := newSCTPSocket(sctp4AddrToSockaddr(laddr))
	if err != nil {
		return nil, err
	}
	err = unix.Bind(so.fd, sctp4AddrToSockaddr(laddr))
	if err != nil {
		return nil, errFromUnixErrno(err)
	}
	err = unix.Listen(so.fd, defaultBacklog)
	if err != nil {
		return nil, errFromUnixErrno(err)
	}

	lis := &SCTPListener{SCTPSocket: so, laddr: laddr}
	return lis, nil
}

func ListenSCTP6(laddr *SCTPAddr) (*SCTPListener, error) {
	if laddr == nil {
		return nil, InvalidAddrError("nil local address")
	}
	so, err := newSCTPSocket(sctp6AddrToSockaddr(laddr))
	if err != nil {
		return nil, err
	}
	err = unix.Bind(so.fd, sctp6AddrToSockaddr(laddr))
	if err != nil {
		return nil, errFromUnixErrno(err)
	}
	err = unix.Listen(so.fd, defaultBacklog)
	if err != nil {
		return nil, errFromUnixErrno(err)
	}

	lis := &SCTPListener{SCTPSocket: so, laddr: laddr}
	return lis, nil
}

func DialSCTP4(laddr *SCTPAddr, raddr *SCTPAddr) (*SCTPConn, error) {
	if raddr == nil {
		return nil, &OpError{Op: "dial", Net: "sctp4", Source: laddr, Addr: nil, Err: errors.New("missing address")}
	}
	so, err := newSCTPSocket(sctp4AddrToSockaddr(laddr))
	if err != nil {
		return nil, err
	}
	err = unix.Connect(so.fd, sctp4AddrToSockaddr(raddr))
	if err != nil {
		return nil, err
	}
	conn := &SCTPConn{
		SCTPSocket: so,
		laddr:      laddr,
		raddr:      raddr,
	}
	return conn, nil
}

func DialSCTP6(laddr *SCTPAddr, raddr *SCTPAddr) (*SCTPConn, error) {
	if raddr == nil {
		return nil, &OpError{Op: "dial", Net: "sctp4", Source: laddr, Addr: nil, Err: errors.New("missing address")}
	}
	so, err := newSCTPSocket(sctp6AddrToSockaddr(laddr))
	if err != nil {
		return nil, err
	}
	err = unix.Connect(so.fd, sctp6AddrToSockaddr(raddr))
	if err != nil {
		return nil, err
	}
	conn := &SCTPConn{
		SCTPSocket: so,
		laddr:      laddr,
		raddr:      raddr,
	}
	return conn, nil
}

func newSCTP4Socket() (fd int, err error) {
	fd, err = unix.Socket(unix.AF_INET, unix.SOCK_SEQPACKET|unix.SOCK_NONBLOCK|unix.SOCK_CLOEXEC, unix.IPPROTO_SCTP)
	if err != nil {
		return 0, errFromUnixErrno(err)
	}
	return fd, nil
}

func newSCTP6Socket() (fd int, err error) {
	fd, err = unix.Socket(unix.AF_INET6, unix.SOCK_SEQPACKET|unix.SOCK_NONBLOCK|unix.SOCK_CLOEXEC, unix.IPPROTO_SCTP)
	if err != nil {
		return 0, errFromUnixErrno(err)
	}
	return fd, nil
}
