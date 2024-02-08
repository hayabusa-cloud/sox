// ©Hayabusa Cloud Co., Ltd. 2022. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

//go:build unix

package sox

import (
	"encoding/binary"
	"golang.org/x/sys/unix"
	"net"
	"net/netip"
	"unsafe"
)

const (
	defaultBacklog = 511
)

func acceptWait(fd int) (nfd int, sa unix.Sockaddr, err error) {
	for sw := NewParamSpinWait().SetLevel(SpinWaitLevelConsume); !sw.Closed(); sw.Once() {
		nfd, sa, err = unix.Accept4(fd, unix.SOCK_NONBLOCK|unix.SOCK_CLOEXEC)
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

func connectWait(fd int, sa unix.Sockaddr) error {
	if err := unix.Connect(fd, sa); err == nil {
		return nil
	} else if err != unix.EINPROGRESS {
		return errFromUnixErrno(err)
	}
	for sw := NewParamSpinWait(); !sw.Closed(); sw.Once() {
		val, err := unix.GetsockoptInt(fd, unix.SOL_SOCKET, unix.SO_ERROR)
		if err != nil {
			return errFromUnixErrno(err)
		}
		if val == 0 {
			break
		}
	}
	return nil
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

func unixAddrFromSockaddr(sa unix.Sockaddr, proto UnderlyingProtocol) *net.UnixAddr {
	switch proto {
	case UnderlyingProtocolStream:
		return &net.UnixAddr{Name: sa.(*unix.SockaddrUnix).Name, Net: "unix"}
	case UnderlyingProtocolDgram:
		return &net.UnixAddr{Name: sa.(*unix.SockaddrUnix).Name, Net: "unixgram"}
	case UnderlyingProtocolSeqPacket:
		return &net.UnixAddr{Name: sa.(*unix.SockaddrUnix).Name, Net: "unixpacket"}
	}
	return nil
}

func addrPortFromSockaddr(addr unix.Sockaddr) netip.AddrPort {
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

func inet4AddrToSockaddr(addr Addr) unix.Sockaddr {
	switch a := addr.(type) {
	case *IPAddr:
		return ip4AddrToSockaddr(a)
	case *TCPAddr:
		return tcp4AddrToSockaddr(a)
	case *UDPAddr:
		return udp4AddrToSockaddr(a)
	case *SCTPAddr:
		return sctp4AddrToSockaddr(a)
	default:
		return nil
	}
}

func inet6AddrToSockaddr(addr Addr) unix.Sockaddr {
	switch a := addr.(type) {
	case *IPAddr:
		return ip6AddrToSockaddr(a)
	case *TCPAddr:
		return tcp6AddrToSockaddr(a)
	case *UDPAddr:
		return udp6AddrToSockaddr(a)
	case *SCTPAddr:
		return sctp6AddrToSockaddr(a)
	default:
		return nil
	}
}

func ip4AddrToSockaddr(addr *IPAddr) unix.Sockaddr {
	return &unix.SockaddrInet4{
		Addr: IP4AddressToBytes(addr.IP),
	}
}

func ip6AddrToSockaddr(addr *IPAddr) unix.Sockaddr {
	return &unix.SockaddrInet6{
		ZoneId: uint32(ip6ZoneID(addr.Zone)),
		Addr:   IP6AddressToBytes(addr.IP),
	}
}

func tcp4AddrToSockaddr(addr *TCPAddr) unix.Sockaddr {
	return &unix.SockaddrInet4{
		Port: addr.Port,
		Addr: IP4AddressToBytes(addr.IP),
	}
}

func tcp6AddrToSockaddr(addr *TCPAddr) unix.Sockaddr {
	return &unix.SockaddrInet6{
		Port:   addr.Port,
		ZoneId: uint32(ip6ZoneID(addr.Zone)),
		Addr:   IP6AddressToBytes(addr.IP),
	}
}

func udp4AddrToSockaddr(addr *UDPAddr) unix.Sockaddr {
	return &unix.SockaddrInet4{
		Port: addr.Port,
		Addr: IP4AddressToBytes(addr.IP),
	}
}

func udp6AddrToSockaddr(addr *UDPAddr) unix.Sockaddr {

	return &unix.SockaddrInet6{
		Port:   addr.Port,
		ZoneId: uint32(ip6ZoneID(addr.Zone)),
		Addr:   IP6AddressToBytes(addr.IP),
	}
}

func sctp4AddrToSockaddr(addr *SCTPAddr) unix.Sockaddr {
	return &unix.SockaddrInet4{
		Port: addr.Port,
		Addr: IP4AddressToBytes(addr.IP),
	}
}

func sctp6AddrToSockaddr(addr *SCTPAddr) unix.Sockaddr {
	return &unix.SockaddrInet6{
		Port:   addr.Port,
		ZoneId: uint32(ip6ZoneID(addr.Zone)),
		Addr:   IP6AddressToBytes(addr.IP),
	}
}

func sockaddr(sa unix.Sockaddr) (ptr unsafe.Pointer, n int, err error) {
	switch sa.(type) {
	case *unix.SockaddrInet4:
		return inet4Sockaddr(sa.(*unix.SockaddrInet4))
	case *unix.SockaddrInet6:
		return inet6Sockaddr(sa.(*unix.SockaddrInet6))
	case *unix.SockaddrUnix:
		if len(sa.(*unix.SockaddrUnix).Name) >= unix.SizeofSockaddrUnix-unix.SizeofShort {
			return unsafe.Pointer(uintptr(0)), 0, ErrInvalidParam
		}
		return unixSockaddr(sa.(*unix.SockaddrUnix))
	default:
		panic("unimplemented sockaddr type")
	}
}

func inet4Sockaddr(sa *unix.SockaddrInet4) (ptr unsafe.Pointer, n int, err error) {
	rawSa := &unix.RawSockaddrInet4{
		Family: unix.AF_INET,
		Addr:   sa.Addr,
	}
	rawP := (*[]byte)(unsafe.Pointer(&rawSa.Port))
	binary.BigEndian.PutUint16(*rawP, uint16(sa.Port))
	return unsafe.Pointer(rawSa), unix.SizeofSockaddrInet4, nil
}

func inet6Sockaddr(sa *unix.SockaddrInet6) (ptr unsafe.Pointer, n int, err error) {
	rawSa := &unix.RawSockaddrInet6{
		Family:   unix.AF_INET6,
		Addr:     sa.Addr,
		Scope_id: sa.ZoneId,
	}
	rawP := (*[]byte)(unsafe.Pointer(&rawSa.Port))
	binary.BigEndian.PutUint16(*rawP, uint16(sa.Port))
	return unsafe.Pointer(rawSa), unix.SizeofSockaddrInet6, nil
}
