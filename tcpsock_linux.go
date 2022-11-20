//go:build linux

package sox

import (
	"errors"
	"golang.org/x/sys/unix"
	"time"
)

type TCPSocket struct {
	fd int
	sa unix.Sockaddr
}

func newTCPSocket(sa unix.Sockaddr) (*TCPSocket, error) {
	fd, err := 0, error(nil)
	if _, ok := sa.(*unix.SockaddrInet4); ok {
		fd, err = newTCP4Socket()
		if err != nil {
			return nil, err
		}
	} else if _, ok = sa.(*unix.SockaddrInet6); ok {
		fd, err = newTCP6Socket()
		if err != nil {
			return nil, err
		}
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

	so := &TCPSocket{
		fd: fd,
		sa: sa,
	}
	return so, nil
}

func (so *TCPSocket) FD() int {
	return so.fd
}

func (so *TCPSocket) Read(b []byte) (n int, err error) {
	offset := 0
	for offset < len(b) {
		n, err = unix.Read(so.fd, b[offset:])
		if n > 0 {
			offset += n
		}
		if err != nil && err != unix.EINTR {
			break
		}
	}
	return offset, errFromUnixErrno(err)
}

// todo: poll the completed notification from uring to know
// it is safe to reuse a previously passed buffer or not
func (so *TCPSocket) Write(b []byte) (n int, err error) {
	for {
		err = unix.Send(so.fd, b, unix.MSG_ZEROCOPY)
		if err == unix.EINTR {
			continue
		}
		break
	}
	return len(b), errFromUnixErrno(err)
}

func (so *TCPSocket) Close() error {
	return nil
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
	remoteAddr := TCPAddrFromAddrPort(addrPortFromSockAddr(remoteSock.sa))
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

	so := &TCPSocket{fd: nfd, sa: sa}
	conn, err := NewTCPConn(l.Addr(), so)
	if err != nil {
		return nil, err
	}
	return conn, err
}

func (l *TCPListener) Close() error {
	return l.TCPSocket.Close()
}

func (l *TCPListener) Addr() Addr {
	if l.laddr != nil {
		return l.laddr
	}
	return TCPAddrFromAddrPort(addrPortFromSockAddr(l.sa))
}

func ListenTCP4(laddr *TCPAddr) (*TCPListener, error) {
	if laddr == nil {
		return nil, InvalidAddrError("nil local address")
	}
	so, err := newTCPSocket(convertTCP4Addr(laddr))
	if err != nil {
		return nil, err
	}
	err = unix.Bind(so.fd, convertTCP4Addr(laddr))
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

func ListenTCP6(laddr *TCPAddr, zoneID int) (*TCPListener, error) {
	if laddr == nil {
		return nil, InvalidAddrError("nil local address")
	}
	so, err := newTCPSocket(convertTCP6Addr(laddr, zoneID))
	if err != nil {
		return nil, err
	}
	err = unix.Bind(so.fd, convertTCP6Addr(laddr, zoneID))
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
	so, err := newTCPSocket(convertTCP4Addr(laddr))
	if err != nil {
		return nil, err
	}
	err = connectWait(so.fd, convertTCP6Addr(raddr, 64))
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

func DialTCP6(laddr *TCPAddr, zoneID int, raddr *TCPAddr) (*TCPConn, error) {
	if raddr == nil {
		return nil, &OpError{Op: "dial", Net: "udp6", Source: laddr, Addr: nil, Err: errors.New("missing address")}
	}
	so, err := newTCPSocket(convertTCP6Addr(laddr, zoneID))
	if err != nil {
		return nil, err
	}
	err = connectWait(so.fd, convertTCP6Addr(raddr, 64))
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
