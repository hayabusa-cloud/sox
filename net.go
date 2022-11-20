package sox

import (
	"errors"
	"io"
	"net"
	"net/netip"
	"strconv"
)

var (
	ErrInterruptedSyscall     = errors.New("interrupted system call")
	ErrTemporarilyUnavailable = errors.New("resource temporarily unavailable")
	ErrInProgress             = errors.New("in progress")
	ErrFaultParams            = errors.New("fault parameters")
	ErrInvalidParam           = errors.New("invalid param")
	ErrProcessFileLimit       = errors.New("process open fd limit")
	ErrSystemFileLimit        = errors.New("system open fd limit")
	ErrNoDevice               = errors.New("no device")
	ErrNoAvailableMemory      = errors.New("no available kernel memory")
	ErrNoPermission           = errors.New("operation not permiited")
)

type Socket interface {
	FD() int
	io.Reader
	io.Writer
	io.Closer
}

type Conn = net.Conn

type Addr = net.Addr
type TCPAddr = net.TCPAddr
type UDPAddr = net.UDPAddr
type UnixAddr = net.UnixAddr

type SCTPAddr struct {
	IP   net.IP
	Port int
	Zone string
}

func (a *SCTPAddr) Network() string {
	return "sctp"
}

func (a *SCTPAddr) String() string {
	if a == nil {
		return "<nil>"
	}
	ip := ""
	if len(a.IP) > 0 {
		ip = a.IP.String()
	}
	if a.Zone != "" {
		return net.JoinHostPort(ip+"%"+a.Zone, strconv.Itoa(a.Port))
	}
	return net.JoinHostPort(ip, strconv.Itoa(a.Port))
}

var (
	TCPAddrFromAddrPort = net.TCPAddrFromAddrPort
	UDPAddrFromAddrPort = net.UDPAddrFromAddrPort
)

func SCTPAddrFromAddrPort(addr netip.AddrPort) *SCTPAddr {
	return &SCTPAddr{
		IP:   addr.Addr().AsSlice(),
		Zone: addr.Addr().Zone(),
		Port: int(addr.Port()),
	}
}

var (
	ResolveTCPAddr  = net.ResolveTCPAddr
	ResolveUDPAddr  = net.ResolveUDPAddr
	ResolveUnixAddr = net.ResolveUnixAddr
)

type OpError = net.OpError
type AddrError = net.AddrError
type InvalidAddrError = net.InvalidAddrError
type UnknownNetworkError = net.UnknownNetworkError

func IP4AddressToBytes(ip net.IP) [4]byte {
	ip4 := ip.To4()
	if ip4 == nil {
		return [4]byte{}
	}
	return [4]byte{ip4[0], ip4[1], ip4[2], ip4[3]}
}

func IP6AddressToBytes(ip net.IP) [16]byte {
	return [16]byte{
		ip[0], ip[1], ip[2], ip[3],
		ip[4], ip[5], ip[6], ip[7],
		ip[8], ip[9], ip[10], ip[11],
		ip[12], ip[13], ip[14], ip[15],
	}
}
