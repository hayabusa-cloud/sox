// ©Hayabusa Cloud Co., Ltd. 2024. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

//go:build unix

package sox

import (
	"golang.org/x/sys/unix"
	"net"
)

type (
	// Sockaddr is an alias for the unix.Sockaddr type.
	Sockaddr = unix.Sockaddr
	// SockaddrInet4 is an alias for the unix.SockaddrInet4 type.
	SockaddrInet4 = unix.SockaddrInet4
	// SockaddrInet6 is an alias for the unix.SockaddrInet6 type.
	SockaddrInet6 = unix.SockaddrInet6
	// SockaddrUnix is an alias for the unix.SockaddrUnix type.
	SockaddrUnix = unix.SockaddrUnix
)

// NetworkAddressToSockaddr returns the corresponding Sockaddr type socket address.
// It resolves the network address and creates a Sockaddr based on the network type.
func NetworkAddressToSockaddr(network, address string) (Sockaddr, error) {
	switch network {
	case "ip", "ip4", "ip6":
		a, err := ResolveIPAddr(network, address)
		if err != nil {
			return nil, err
		}
		return ipAddrToSockaddr(network, a), nil
	case "tcp", "tcp4", "tcp6":
		a, err := ResolveTCPAddr(network, address)
		if err != nil {
			return nil, err
		}
		return tcpAddrToSockaddr(network, a), nil
	case "udp", "udp4", "udp6":
		a, err := ResolveUDPAddr(network, address)
		if err != nil {
			return nil, err
		}
		return udpAddrToSockaddr(network, a), nil
	case "sctp", "sctp4", "sctp6":
		a, err := ResolveSCTPAddr(network, address)
		if err != nil {
			return nil, err
		}
		return sctpAddrToSockaddr(network, a), nil
	case "unix", "unixgram", "unixpacket":
		a, err := ResolveUnixAddr(network, address)
		if err != nil {
			return nil, err
		}
		return unixAddrToSockaddr(a), nil
	}
	return nil, UnknownNetworkError(network)
}

// AddrToSockaddr takes a network address and returns the Sockaddr type socket address.
func AddrToSockaddr(addr Addr) Sockaddr {
	switch addr := addr.(type) {
	case *IPAddr:
		return ipAddrToSockaddr("ip", addr)
	case *TCPAddr:
		return tcpAddrToSockaddr("tcp", addr)
	case *UDPAddr:
		return udpAddrToSockaddr("udp", addr)
	case *SCTPAddr:
		return sctpAddrToSockaddr("sctp", addr)
	case *UnixAddr:
		return unixAddrToSockaddr(addr)
	default:
		panic(net.InvalidAddrError(addr.String()))
	}
}
