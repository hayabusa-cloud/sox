// ©Hayabusa Cloud Co., Ltd. 2024. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

//go:build unix

package sox

import (
	"golang.org/x/sys/unix"
	"net"
)

// Sockaddr is an alias for the unix.Sockaddr type.
type Sockaddr = unix.Sockaddr

// AddrToSockaddr takes a network address and returns the Sockaddr type socket address.
func AddrToSockaddr(addr Addr) Sockaddr {
	switch addr := addr.(type) {
	case *IPAddr:
		return ipAddrToSockaddr(addr)
	case *TCPAddr:
		return tcpAddrToSockaddr(addr)
	case *UDPAddr:
		return udpAddrToSockaddr(addr)
	case *SCTPAddr:
		return sctpAddrToSockaddr(addr)
	case *UnixAddr:
		return unixAddrToSockaddr(addr)
	default:
		panic(net.InvalidAddrError(addr.String()))
	}
}
