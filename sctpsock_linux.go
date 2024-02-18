// ©Hayabusa Cloud Co., Ltd. 2023. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

//go:build linux

package sox

import (
	"errors"
	"golang.org/x/sys/unix"
	"time"
)

const (
	SOL_SCTP = 132
)

const (
	SCTP_RTOINFO   = 0
	SCTP_ASSOCINFO = 1
	SCTP_INITMSG   = 2
	SCTP_NODELAY   = 3

	SCTP_SOCKOPT_BINDX_ADD = 100
	SCTP_SOCKOPT_BINDX_REM = 101
	SCTP_SOCKOPT_CONNECTX  = 110
	SCTP_SOCKOPT_CONNECTX3 = 111
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
	nfd, sa, err := sctpAcceptWait(l)
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
	lsa := sctp4AddrToSockaddr(laddr)
	so, err := newSCTPSocket(lsa)
	if err != nil {
		return nil, err
	}
	err = sctpBindx(so, lsa)
	if err != nil {
		return nil, err
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
	lsa := sctp6AddrToSockaddr(laddr)
	so, err := newSCTPSocket(lsa)
	if err != nil {
		return nil, err
	}
	err = sctpBindx(so, lsa)
	if err != nil {
		return nil, err
	}
	err = unix.Listen(so.fd, defaultBacklog)
	if err != nil {
		return nil, errFromUnixErrno(err)
	}

	lis := &SCTPListener{SCTPSocket: so, laddr: laddr}
	return lis, nil
}

func DialSCTP4(laddr *SCTPAddr, raddr *SCTPAddr) (*SCTPConn, error) {
	if laddr == nil {
		laddr = &SCTPAddr{IP: IPv4LoopBack}
	}
	if raddr == nil {
		return nil, &OpError{Op: "dial", Net: "sctp4", Source: laddr, Addr: nil, Err: errors.New("missing address")}
	}
	lsa := sctp4AddrToSockaddr(laddr)
	so, err := newSCTPSocket(lsa)
	if err != nil {
		return nil, err
	}
	err = sctpBindx(so, lsa)
	if err != nil {
		return nil, err
	}
	conn := &SCTPConn{
		SCTPSocket: so,
		laddr:      laddr,
		raddr:      raddr,
	}
	err = sctpConnectx(so, sctp4AddrToSockaddr(raddr))
	if err != nil {
		return nil, err
	}

	return conn, nil
}

func DialSCTP6(laddr *SCTPAddr, raddr *SCTPAddr) (*SCTPConn, error) {
	if laddr == nil {
		laddr = &SCTPAddr{IP: IPv6LoopBack}
	}
	if raddr == nil {
		return nil, &OpError{Op: "dial", Net: "sctp6", Source: laddr, Addr: nil, Err: errors.New("missing address")}
	}
	lsa := sctp6AddrToSockaddr(laddr)
	so, err := newSCTPSocket(lsa)
	if err != nil {
		return nil, err
	}
	err = sctpBindx(so, lsa)
	if err != nil {
		return nil, err
	}
	conn := &SCTPConn{
		SCTPSocket: so,
		laddr:      laddr,
		raddr:      raddr,
	}
	err = sctpConnectx(so, sctp6AddrToSockaddr(raddr))
	if err != nil {
		return nil, err
	}

	return conn, nil
}

func newSCTP4Socket() (fd int, err error) {
	fd, err = unix.Socket(unix.AF_INET, unix.SOCK_STREAM|unix.SOCK_NONBLOCK|unix.SOCK_CLOEXEC, unix.IPPROTO_SCTP)
	if err != nil {
		return 0, errFromUnixErrno(err)
	}
	return fd, nil
}

func newSCTP6Socket() (fd int, err error) {
	fd, err = unix.Socket(unix.AF_INET6, unix.SOCK_STREAM|unix.SOCK_NONBLOCK|unix.SOCK_CLOEXEC, unix.IPPROTO_SCTP)
	if err != nil {
		return 0, errFromUnixErrno(err)
	}
	return fd, nil
}

func sctpBindx(so *SCTPSocket, sa unix.Sockaddr) error {
	ptr, n, err := sockaddr(sa)
	if err != nil {
		return err
	}
	_, _, errno := unix.Syscall6(
		unix.SYS_SETSOCKOPT,
		uintptr(so.fd),
		SOL_SCTP,
		SCTP_SOCKOPT_BINDX_ADD,
		uintptr(ptr),
		uintptr(n),
		0)
	if errno != 0 {
		return errFromUnixErrno(errno)
	}

	return nil
}

func sctpAcceptWait(lis *SCTPListener) (nfd int, sa unix.Sockaddr, err error) {
	for sw := NewParamSpinWait().SetLevel(SpinWaitLevelConsume); !sw.Closed(); sw.Once() {
		nfd, sa, err = unix.Accept4(lis.fd, unix.SOCK_NONBLOCK|unix.SOCK_CLOEXEC)
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

func sctpConnectx(so *SCTPSocket, sa unix.Sockaddr) error {
	ptr, n, err := sockaddr(sa)
	if err != nil {
		return err
	}
	_, _, errno := unix.Syscall6(
		unix.SYS_SETSOCKOPT,
		uintptr(so.fd),
		SOL_SCTP,
		SCTP_SOCKOPT_CONNECTX,
		uintptr(ptr),
		uintptr(n),
		0)
	if errno != unix.EINPROGRESS {
		return errFromUnixErrno(errno)
	}
	for sw := NewParamSpinWait(); !sw.Closed(); sw.Once() {
		val, err := unix.GetsockoptInt(so.fd, unix.SOL_SOCKET, unix.SO_ERROR)
		if err != nil {
			return errFromUnixErrno(err)
		}
		if val == 0 {
			break
		}
	}

	return nil
}
