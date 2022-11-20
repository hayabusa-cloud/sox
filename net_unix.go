//go:build unix

package sox

import (
	"golang.org/x/sys/unix"
	"net"
	"net/netip"
	"unsafe"
)

const (
	defaultBacklog = 511
)

func acceptWait(fd int) (nfd int, sa unix.Sockaddr, err error) {
	for {
		nfd, sa, err = unix.Accept4(fd, unix.SOCK_NONBLOCK|unix.SOCK_CLOEXEC)
		if err == unix.EAGAIN || err == unix.EWOULDBLOCK {
			Yield(jiffies)
			continue
		}
		if err != nil {
			return 0, nil, errFromUnixErrno(err)
		}
		break
	}
	return
}

func connectWait(fd int, sa unix.Sockaddr) error {
	if err := unix.Connect(fd, sa); err == nil {
		return nil
	} else if err != unix.EINPROGRESS {
		return errFromUnixErrno(err)
	}
	for {
		val, err := unix.GetsockoptInt(fd, unix.SOL_SOCKET, unix.SO_ERROR)
		if err != nil {
			return errFromUnixErrno(err)
		}
		if val == 0 {
			break
		}
		Yield(jiffies)
	}
	return nil
}

func ioVecFromSliceOfBytes(iov [][]byte) (addr uint64, n int) {
	vec := make([]unix.Iovec, len(iov))
	for i := 0; i < len(iov); i++ {
		vec[i] = unix.Iovec{Base: &iov[i][0], Len: uint64(len(iov[i]))}
	}
	addr, n = uint64(uintptr(unsafe.Pointer(&iov))), len(iov)

	return
}

func inetAddrFromAddrPort(addrPort netip.AddrPort) unix.Sockaddr {
	if addrPort.Addr().Is4() {
		return &unix.SockaddrInet4{
			Port: int(addrPort.Port()),
			Addr: addrPort.Addr().As4(),
		}
	} else if addrPort.Addr().Is6() {
		return &unix.SockaddrInet6{
			Port: int(addrPort.Port()),
			Addr: addrPort.Addr().As16(),
		}
	}
	return &unix.SockaddrInet6{}
}

func unixAddrFromSockAddr(sa unix.Sockaddr, proto UnderlyingProtocol) *net.UnixAddr {
	switch proto {
	case UnderlyingProtocolStream:
		return &net.UnixAddr{Name: sa.(*unix.SockaddrUnix).Name, Net: "unix"}
	case UnderlyingProtocolSeqPacket:
		return &net.UnixAddr{Name: sa.(*unix.SockaddrUnix).Name, Net: "unixpacket"}
	}
	return nil
}

func addrPortFromSockAddr(addr unix.Sockaddr) netip.AddrPort {
	if addr == nil {
		return netip.AddrPort{}
	}
	switch addr.(type) {
	case *unix.SockaddrInet4:
		a := addr.(*unix.SockaddrInet4)
		return netip.AddrPortFrom(netip.AddrFrom4(a.Addr), uint16(a.Port))
	case *unix.SockaddrInet6:
		a := addr.(*unix.SockaddrInet6)
		return netip.AddrPortFrom(netip.AddrFrom16(a.Addr), uint16(a.Port))
	}
	return netip.AddrPort{}
}

func convertTCP4Addr(addr *net.TCPAddr) unix.Sockaddr {
	return &unix.SockaddrInet4{
		Port: addr.Port,
		Addr: IP4AddressToBytes(addr.IP),
	}
}

func convertTCP6Addr(addr *net.TCPAddr, zoneID int) unix.Sockaddr {
	return &unix.SockaddrInet6{
		Port:   addr.Port,
		ZoneId: uint32(zoneID),
		Addr:   IP6AddressToBytes(addr.IP),
	}
}

func convertUDP4Addr(addr *net.UDPAddr) unix.Sockaddr {
	return &unix.SockaddrInet4{
		Port: addr.Port,
		Addr: IP4AddressToBytes(addr.IP),
	}
}

func convertUDP6Addr(addr *net.UDPAddr, zoneID int) unix.Sockaddr {
	return &unix.SockaddrInet6{
		Port:   addr.Port,
		ZoneId: uint32(zoneID),
		Addr:   IP6AddressToBytes(addr.IP),
	}
}

func convertSCTP4Addr(addr *SCTPAddr) unix.Sockaddr {
	return &unix.SockaddrInet4{
		Port: addr.Port,
		Addr: IP4AddressToBytes(addr.IP),
	}
}

func convertSCTP6Addr(addr *SCTPAddr, zoneID int) unix.Sockaddr {
	return &unix.SockaddrInet6{
		Port:   addr.Port,
		ZoneId: uint32(zoneID),
		Addr:   IP6AddressToBytes(addr.IP),
	}
}

func convertUnixAddr(addr *net.UnixAddr) unix.Sockaddr {
	return &unix.SockaddrUnix{
		Name: addr.Name,
	}
}
