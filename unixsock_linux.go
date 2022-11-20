//go:build linux

package sox

import (
	"errors"
	"golang.org/x/sys/unix"
	"time"
)

type UnixSocket struct {
	fd int
	sa unix.Sockaddr
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

	so := &UnixSocket{
		fd: fd,
		sa: sa,
	}
	return so, nil
}

func newUnixSocketPair() (so [2]*UnixSocket, err error) {
	fd, err := unix.Socketpair(unix.AF_UNIX, unix.SOCK_SEQPACKET|unix.SOCK_NONBLOCK|unix.SOCK_CLOEXEC, 0)
	if err != nil {
		return [2]*UnixSocket{}, errFromUnixErrno(err)
	}

	so[0] = &UnixSocket{fd: fd[0], sa: &unix.SockaddrUnix{}}
	so[1] = &UnixSocket{fd: fd[1], sa: &unix.SockaddrUnix{}}
	return so, nil
}

func (so *UnixSocket) FD() int {
	return so.fd
}

func (so *UnixSocket) Read(b []byte) (n int, err error) {
	n, err = unix.Read(so.fd, b)
	if err != nil {
		return n, errFromUnixErrno(err)
	}
	return n, nil
}

func (so *UnixSocket) Write(b []byte) (n int, err error) {
	n, err = unix.Write(so.fd, b)
	if err != nil {
		return n, errFromUnixErrno(err)
	}
	return n, nil
}

func (so *UnixSocket) Close() error {
	return nil
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

	remoteAddr := unixAddrFromSockAddr(remoteSock.sa, UnderlyingProtocolSeqPacket)
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

	so := &UnixSocket{fd: nfd, sa: sa}
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
	return unixAddrFromSockAddr(l.sa, UnderlyingProtocolSeqPacket)
}

func ListenUnix(laddr *UnixAddr) (*UnixListener, error) {
	if laddr == nil {
		return nil, InvalidAddrError("nil local address")
	}
	so, err := newUnixSocket(convertUnixAddr(laddr))
	if err != nil {
		return nil, err
	}
	err = unix.Bind(so.fd, convertUnixAddr(laddr))
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
	so, err := newUnixSocket(convertUnixAddr(laddr))
	if err != nil {
		return nil, err
	}
	err = connectWait(so.fd, convertUnixAddr(raddr))
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
