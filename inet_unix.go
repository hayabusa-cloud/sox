// ©Hayabusa Cloud Co., Ltd. 2022. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

//go:build unix

package sox

import (
	"encoding/binary"
	"golang.org/x/sys/unix"
	"net/netip"
	"unsafe"
)

func inetAddrFromAddrPort(addrPort netip.AddrPort) Sockaddr {
	if addrPort.Addr().Is4() {
		return &SockaddrInet4{
			Port: int(addrPort.Port()),
			Addr: addrPort.Addr().As4(),
		}
	} else if addrPort.Addr().Is6() {
		return &SockaddrInet6{
			Port: int(addrPort.Port()),
			Addr: addrPort.Addr().As16(),
		}
	}
	return &SockaddrInet6{}
}

func ipAddrFromSockaddr(addr Sockaddr) *IPAddr {
	if addr == nil {
		return nil
	}
	switch addr.(type) {
	case *SockaddrInet4:
		a := addr.(*SockaddrInet4)
		return &IPAddr{IP: a.Addr[:]}
	case *SockaddrInet6:
		a := addr.(*SockaddrInet6)
		return &IPAddr{
			IP:   a.Addr[:],
			Zone: ip6ZoneString(int(a.ZoneId)),
		}
	}
	return nil
}
func unixAddrFromSockaddr(sa Sockaddr, proto UnderlyingProtocol) *UnixAddr {
	switch proto {
	case UnderlyingProtocolStream:
		return &UnixAddr{Name: sa.(*SockaddrUnix).Name, Net: "unix"}
	case UnderlyingProtocolDgram:
		return &UnixAddr{Name: sa.(*SockaddrUnix).Name, Net: "unixgram"}
	case UnderlyingProtocolSeqPacket:
		return &UnixAddr{Name: sa.(*SockaddrUnix).Name, Net: "unixpacket"}
	}
	return nil
}

func addrPortFromSockaddr(addr Sockaddr) netip.AddrPort {
	if addr == nil {
		return netip.AddrPort{}
	}
	switch addr.(type) {
	case *SockaddrInet4:
		a := addr.(*SockaddrInet4)
		return netip.AddrPortFrom(netip.AddrFrom4(a.Addr), uint16(a.Port))
	case *SockaddrInet6:
		a := addr.(*SockaddrInet6)
		return netip.AddrPortFrom(netip.AddrFrom16(a.Addr), uint16(a.Port))
	}
	return netip.AddrPort{}
}

func inet4AddrToSockaddr(addr Addr) *SockaddrInet4 {
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

func inet6AddrToSockaddr(addr Addr) *SockaddrInet6 {
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

func ip4AddrToSockaddr(addr *IPAddr) *SockaddrInet4 {
	ip := addr.IP
	if ip == nil {
		ip = IPV4zero
	}
	return &SockaddrInet4{
		Addr: IP4AddressToBytes(addr.IP),
	}
}

func ip4AddrPortToSockaddr(addr *IPAddr, port int) *SockaddrInet4 {
	ip := addr.IP
	if ip == nil {
		ip = IPV4zero
	}
	return &SockaddrInet4{
		Port: port,
		Addr: IP4AddressToBytes(ip),
	}
}

func ip6AddrToSockaddr(addr *IPAddr) *SockaddrInet6 {
	ip := addr.IP
	if ip == nil {
		ip = IPV6unspecified
	}
	return &SockaddrInet6{
		ZoneId: uint32(ip6ZoneID(addr.Zone)),
		Addr:   IP6AddressToBytes(ip),
	}
}

func ip6AddrPortToSockaddr(addr *IPAddr, port int) *SockaddrInet6 {
	ip := addr.IP
	if ip == nil {
		ip = IPV6unspecified
	}
	return &SockaddrInet6{
		Port:   port,
		ZoneId: uint32(ip6ZoneID(addr.Zone)),
		Addr:   IP6AddressToBytes(ip),
	}
}

func ipAddrToSockaddr(network string, addr *IPAddr) Sockaddr {
	if networkIPFamily(network, addr.IP) == NetworkIPv4 {
		return ip4AddrToSockaddr(addr)
	}
	return ip6AddrToSockaddr(addr)
}

func ipAddrPortToSockaddr(network string, addr *IPAddr, port int) Sockaddr {
	if networkIPFamily(network, addr.IP) == NetworkIPv4 {
		return ip4AddrPortToSockaddr(addr, port)
	}
	return ip6AddrPortToSockaddr(addr, port)
}

func tcp4AddrToSockaddr(addr *TCPAddr) *SockaddrInet4 {
	ip := addr.IP
	if ip == nil {
		ip = IPV4zero
	}
	return &SockaddrInet4{
		Port: addr.Port,
		Addr: IP4AddressToBytes(ip),
	}
}

func tcp6AddrToSockaddr(addr *TCPAddr) *SockaddrInet6 {
	ip := addr.IP
	if ip == nil {
		ip = IPV6unspecified
	}
	return &SockaddrInet6{
		Port:   addr.Port,
		ZoneId: uint32(ip6ZoneID(addr.Zone)),
		Addr:   IP6AddressToBytes(ip),
	}
}

func tcpAddrToSockaddr(network string, addr *TCPAddr) Sockaddr {
	return ipAddrPortToSockaddr(network, IPAddrFromTCPAddr(addr), addr.Port)
}

func udp4AddrToSockaddr(addr *UDPAddr) *SockaddrInet4 {
	ip := addr.IP
	if ip == nil {
		ip = IPV4zero
	}
	return &SockaddrInet4{
		Port: addr.Port,
		Addr: IP4AddressToBytes(ip),
	}
}

func udp6AddrToSockaddr(addr *UDPAddr) *SockaddrInet6 {
	ip := addr.IP
	if ip == nil {
		ip = IPV6unspecified
	}
	return &SockaddrInet6{
		Port:   addr.Port,
		ZoneId: uint32(ip6ZoneID(addr.Zone)),
		Addr:   IP6AddressToBytes(ip),
	}
}

func udpAddrToSockaddr(network string, addr *UDPAddr) Sockaddr {
	return ipAddrPortToSockaddr(network, IPAddrFromUDPAddr(addr), addr.Port)
}

func sctp4AddrToSockaddr(addr *SCTPAddr) *SockaddrInet4 {
	ip := addr.IP
	if ip == nil {
		ip = IPV4zero
	}
	return &SockaddrInet4{
		Port: addr.Port,
		Addr: IP4AddressToBytes(ip),
	}
}

func sctp6AddrToSockaddr(addr *SCTPAddr) *SockaddrInet6 {
	ip := addr.IP
	if ip == nil {
		ip = IPV6unspecified
	}
	return &SockaddrInet6{
		Port:   addr.Port,
		ZoneId: uint32(ip6ZoneID(addr.Zone)),
		Addr:   IP6AddressToBytes(ip),
	}
}

func sctpAddrToSockaddr(network string, addr *SCTPAddr) Sockaddr {
	return ipAddrPortToSockaddr(network, IPAddrFromSCTPAddr(addr), addr.Port)
}

func inetAddrToSockaddr(addr Addr) Sockaddr {
	switch addr := addr.(type) {
	case *IPAddr:
		return ipAddrToSockaddr("ip", addr)
	case *TCPAddr:
		return tcpAddrToSockaddr("tcp", addr)
	case *UDPAddr:
		return udpAddrToSockaddr("udp", addr)
	case *SCTPAddr:
		return sctpAddrToSockaddr("sctp", addr)
	default:
		panic("addr must be an internet address")
	}
}

func sockaddr(sa Sockaddr) (ptr unsafe.Pointer, n int, err error) {
	switch sa.(type) {
	case *SockaddrInet4:
		return inet4Sockaddr(sa.(*SockaddrInet4))
	case *SockaddrInet6:
		return inet6Sockaddr(sa.(*SockaddrInet6))
	case *SockaddrUnix:
		if len(sa.(*SockaddrUnix).Name) >= unix.SizeofSockaddrUnix-unix.SizeofShort {
			return unsafe.Pointer(uintptr(0)), 0, ErrInvalidParam
		}
		return unixSockaddr(sa.(*SockaddrUnix))
	default:
		panic("unimplemented sockaddr type")
	}
}

func inet4Sockaddr(sa *SockaddrInet4) (ptr unsafe.Pointer, n int, err error) {
	rawSa := &unix.RawSockaddrInet4{
		Family: unix.AF_INET,
		Addr:   sa.Addr,
	}
	b := unsafe.Slice((*byte)(unsafe.Pointer(rawSa)), unix.SizeofSockaddrInet4)
	portOffset, portSize := unsafe.Sizeof(uint16(0)), unsafe.Sizeof(uint16(0))
	binary.BigEndian.PutUint16(b[portOffset:portOffset+portSize], uint16(sa.Port))
	return unsafe.Pointer(unsafe.SliceData(b)), unix.SizeofSockaddrInet4, nil
}

func inet6Sockaddr(sa *SockaddrInet6) (ptr unsafe.Pointer, n int, err error) {
	rawSa := &unix.RawSockaddrInet6{
		Family:   unix.AF_INET6,
		Addr:     sa.Addr,
		Scope_id: sa.ZoneId,
	}
	b := unsafe.Slice((*byte)(unsafe.Pointer(rawSa)), unix.SizeofSockaddrInet6)
	portOffset, portSize := unsafe.Sizeof(uint16(0)), unsafe.Sizeof(uint16(0))
	binary.BigEndian.PutUint16(b[portOffset:portOffset+portSize], uint16(sa.Port))
	return unsafe.Pointer(unsafe.SliceData(b)), unix.SizeofSockaddrInet6, nil
}

func sockaddrData(sa Sockaddr) (data []byte, err error) {
	switch sa.(type) {
	case *SockaddrInet4:
		return inet4SockaddrData(sa.(*SockaddrInet4))
	case *SockaddrInet6:
		return inet6SockaddrData(sa.(*SockaddrInet6))
	case *SockaddrUnix:
		if len(sa.(*SockaddrUnix).Name) >= unix.SizeofSockaddrUnix-unix.SizeofShort {
			return nil, ErrInvalidParam
		}
		return unixSockaddrData(sa.(*SockaddrUnix))
	default:
		panic("unimplemented sockaddr type")
	}
}

func inet4SockaddrData(sa *SockaddrInet4) (data []byte, err error) {
	rawSa := &unix.RawSockaddrInet4{
		Family: unix.AF_INET,
		Port:   uint16(sa.Port),
		Addr:   sa.Addr,
	}
	b := unsafe.Slice((*byte)(unsafe.Pointer(rawSa)), unix.SizeofSockaddrInet4)
	return b[:unix.SizeofSockaddrInet4], nil
}

func inet6SockaddrData(sa *SockaddrInet6) (data []byte, err error) {
	rawSa := &unix.RawSockaddrInet6{
		Family:   unix.AF_INET6,
		Port:     uint16(sa.Port),
		Addr:     sa.Addr,
		Scope_id: sa.ZoneId,
	}
	b := unsafe.Slice((*byte)(unsafe.Pointer(rawSa)), unix.SizeofSockaddrInet6)
	return b[:unix.SizeofSockaddrInet6], nil
}

func inet4SockaddrFromData(data []byte) *SockaddrInet4 {
	raw := (*unix.RawSockaddrInet4)(unsafe.Pointer(unsafe.SliceData(data)))
	sa := SockaddrInet4{
		Port: int(raw.Port),
		Addr: raw.Addr,
	}
	fRawPtr := unsafe.Add(unsafe.Pointer(&sa), unsafe.Sizeof(sa.Port)+unsafe.Sizeof(sa.Addr))
	*(*unix.RawSockaddrInet4)(fRawPtr) = *raw
	return &sa
}

func inet6SockaddrFromData(data []byte) *SockaddrInet6 {
	raw := (*unix.RawSockaddrInet6)(unsafe.Pointer(unsafe.SliceData(data)))
	sa := SockaddrInet6{
		Port:   int(raw.Port),
		ZoneId: raw.Scope_id,
		Addr:   raw.Addr,
	}
	fRawPtr := unsafe.Add(unsafe.Pointer(&sa), unsafe.Sizeof(sa.Port)+unsafe.Sizeof(sa.ZoneId)+unsafe.Sizeof(sa.Addr))
	*(*unix.RawSockaddrInet6)(fRawPtr) = *raw
	return &sa
}
