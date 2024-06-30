// ©Hayabusa Cloud Co., Ltd. 2024. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package sox

func Listen(network string, address string) (listener Listener, err error) {
	switch network {
	case "tcp", "tcp4", "tcp6":
		a, err := ResolveTCPAddr(network, address)
		if err != nil {
			return nil, err
		}
		return ListenTCP(network, a)
	case "sctp", "sctp4", "sctp6":
		a, err := ResolveSCTPAddr(network, address)
		if err != nil {
			return nil, err
		}
		return ListenSCTP(network, a)
	case "unix", "unixpacket":
		addr, err := ResolveUnixAddr(network, address)
		if err != nil {
			return nil, err
		}
		return ListenUnix(addr)
	}

	return nil, &OpError{Op: "listen", Net: network, Source: nil, Addr: nil, Err: UnknownNetworkError(network)}
}

func ListenTCP(network string, laddr *TCPAddr) (listener Listener, err error) {
	if laddr == nil {
		laddr = &TCPAddr{}
	}
	switch network {
	case "tcp":
		if ipFamily(laddr.IP) == NetworkIPv4 {
			return ListenTCP4(laddr)
		}
		return ListenTCP6(laddr)
	case "tcp4":
		return ListenTCP4(laddr)
	case "tcp6":
		return ListenTCP6(laddr)
	}
	return nil, &OpError{Op: "listen", Net: network, Source: nil, Addr: laddr, Err: UnknownNetworkError(network)}
}

func ListenUDP(network string, laddr *UDPAddr) (conn *UDPConn, err error) {
	if laddr == nil {
		laddr = &UDPAddr{}
	}
	switch network {
	case "udp":
		if ipFamily(laddr.IP) == NetworkIPv4 {
			return ListenUDP4(laddr)
		}
		return ListenUDP6(laddr)
	case "udp4":
		return ListenUDP4(laddr)
	case "udp6":
		return ListenUDP6(laddr)
	}
	return nil, &OpError{Op: "listen", Net: network, Source: nil, Addr: laddr, Err: UnknownNetworkError(network)}
}

func ListenSCTP(network string, laddr *SCTPAddr) (listener Listener, err error) {
	if laddr == nil {
		laddr = &SCTPAddr{}
	}
	switch network {
	case "sctp":
		if ipFamily(laddr.IP) == NetworkIPv4 {
			return ListenSCTP4(laddr)
		}
		return ListenSCTP6(laddr)
	case "sctp4":
		return ListenSCTP4(laddr)
	case "sctp6":
		return ListenSCTP6(laddr)
	}
	return nil, &OpError{Op: "listen", Net: network, Source: nil, Addr: laddr, Err: UnknownNetworkError(network)}
}
