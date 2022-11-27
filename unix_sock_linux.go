//go:build linux

package sox

import (
	"errors"
	"golang.org/x/sys/unix"
	"time"
)

type UnixSocket struct {
	*socket
}

func newUnixSocket(sa unix.Sockaddr) (*UnixSocket, error) {
	fd, err := 0, error(nil)
	if _, ok := sa.(*unix.SockaddrUnix); ok {
		fd, err = unix.Socket(unix.AF_UNIX, unix.SOCK_SEQPACKET|unix.SOCK_NONBLOCK|unix.SOCK_CLOEXEC, 0)
		if err != nil {
			return nil, err
		}
	} else {
		return nil, UnknownNetworkError("unexpected family")
	}
	err = unix.SetsockoptInt(fd, unix.SOL_SOCKET, unix.SO_PASSSEC, 1)
	if err != nil {
		return nil, errFromUnixErrno(err)
	}

	so := &UnixSocket{socket: newSocket(NetworkUnix, fd, sa)}
	return so, nil
}

func newUnixSocketPair() (so [2]*UnixSocket, err error) {
	fd, err := unix.Socketpair(unix.AF_UNIX, unix.SOCK_SEQPACKET|unix.SOCK_NONBLOCK|unix.SOCK_CLOEXEC, 0)
	if err != nil {
		return [2]*UnixSocket{}, errFromUnixErrno(err)
	}

	so[0] = &UnixSocket{socket: newSocket(NetworkUnix, fd[0], &unix.SockaddrUnix{})}
	so[1] = &UnixSocket{socket: newSocket(NetworkUnix, fd[1], &unix.SockaddrUnix{})}
	return so, nil
}

func (so *UnixSocket) Protocol() UnderlyingProtocol {
	return UnderlyingProtocolSeqPacket
}

type UnixConn struct {
	*UnixSocket
	laddr *UnixAddr
	raddr *UnixAddr
}

func NewUnixConn(localAddr Addr, remoteSock *UnixSocket) (Conn, error) {
	if localAddr == nil {
		localAddr = &UnixAddr{}
	}
	unixAddr, ok := localAddr.(*UnixAddr)
	if !ok {
		return nil, &AddrError{Err: "unexpected address type", Addr: localAddr.String()}
	}

	remoteAddr := unixAddrFromSockaddr(remoteSock.sa, UnderlyingProtocolSeqPacket)
	return &UnixConn{UnixSocket: remoteSock, laddr: unixAddr, raddr: remoteAddr}, nil
}

func (conn *UnixConn) LocalAddr() Addr {
	return conn.laddr
}
func (conn *UnixConn) RemoteAddr() Addr {
	return conn.raddr
}
func (conn *UnixConn) SetDeadline(t time.Time) error {
	return nil
}
func (conn *UnixConn) SetReadDeadline(t time.Time) error {
	return nil
}
func (conn *UnixConn) SetWriteDeadline(t time.Time) error {
	return nil
}

type UnixListener struct {
	*UnixSocket
	laddr *UnixAddr
}

func (l *UnixListener) Accept() (Conn, error) {
	nfd, sa, err := acceptWait(l.fd)
	if err != nil {
		return nil, err
	}

	so := &UnixSocket{socket: newSocket(NetworkUnix, nfd, sa)}
	conn, err := NewUnixConn(l.Addr(), so)
	if err != nil {
		return nil, err
	}
	return conn, err
}

func (l *UnixListener) Close() error {
	sa := l.sa.(*unix.SockaddrUnix)
	if len(sa.Name) > 0 {
		_ = unix.Unlink(sa.Name)
	}
	return l.UnixSocket.Close()
}

func (l *UnixListener) Addr() Addr {
	if l.laddr != nil {
		return l.laddr
	}
	return unixAddrFromSockaddr(l.sa, UnderlyingProtocolSeqPacket)
}

func ListenUnix(laddr *UnixAddr) (*UnixListener, error) {
	if laddr == nil {
		return nil, InvalidAddrError("nil local address")
	}
	so, err := newUnixSocket(unixAddrToSockAddr(laddr))
	if err != nil {
		return nil, err
	}
	err = unix.Bind(so.fd, unixAddrToSockAddr(laddr))
	if err != nil {
		return nil, errFromUnixErrno(err)
	}
	err = unix.Listen(so.fd, defaultBacklog)
	if err != nil {
		return nil, errFromUnixErrno(err)
	}
	lis := &UnixListener{UnixSocket: so, laddr: laddr}
	return lis, nil
}

func DialUnix(laddr *UnixAddr, raddr *UnixAddr) (*UnixConn, error) {
	if raddr == nil {
		return nil, &OpError{Op: "dial", Net: "unix", Source: laddr, Addr: nil, Err: errors.New("missing address")}
	}
	so, err := newUnixSocket(unixAddrToSockAddr(laddr))
	if err != nil {
		return nil, err
	}
	err = connectWait(so.fd, unixAddrToSockAddr(raddr))
	if err != nil {
		return nil, err
	}

	conn := &UnixConn{
		UnixSocket: so,
		laddr:      laddr,
		raddr:      raddr,
	}
	return conn, nil
}
