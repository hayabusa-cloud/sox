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

func Dial(network string, address string) (conn Conn, err error) {
	switch network {
	case "tcp":
		a, err := ResolveTCPAddr(network, address)
		if err != nil {
			return nil, err
		}
		if ipFamily(a.IP) == NetworkIPv4 {
			return DialTCP4(nil, a)
		}
		return DialTCP6(nil, a)
	case "tcp4":
		a, err := ResolveTCPAddr(network, address)
		if err != nil {
			return nil, err
		}
		return DialTCP4(nil, a)
	case "tcp6":
		a, err := ResolveTCPAddr(network, address)
		if err != nil {
			return nil, err
		}
		return DialTCP6(nil, a)
	case "udp":
		a, err := ResolveUDPAddr(network, address)
		if err != nil {
			return nil, err
		}
		if ipFamily(a.IP) == NetworkIPv4 {
			return DialUDP4(nil, a)
		}
		return DialUDP6(nil, a)
	case "udp4":
		a, err := ResolveUDPAddr(network, address)
		if err != nil {
			return nil, err
		}
		return DialUDP4(nil, a)
	case "udp6":
		a, err := ResolveUDPAddr(network, address)
		if err != nil {
			return nil, err
		}
		return DialUDP6(nil, a)
	case "sctp":
		a, err := ResolveSCTPAddr(network, address)
		if err != nil {
			return nil, err
		}
		if ipFamily(a.IP) == NetworkIPv4 {
			return DialSCTP4(nil, a)
		}
		return DialSCTP6(nil, a)
	case "sctp4":
		a, err := ResolveSCTPAddr(network, address)
		if err != nil {
			return nil, err
		}
		return DialSCTP4(nil, a)
	case "sctp6":
		a, err := ResolveSCTPAddr(network, address)
		if err != nil {
			return nil, err
		}
		return DialSCTP6(nil, a)
	case "unix", "unixpacket":
		addr, err := ResolveUnixAddr(network, address)
		if err != nil {
			return nil, err
		}
		return DialUnix(nil, addr)
	}

	return nil, &OpError{Op: "dial", Net: network, Source: nil, Addr: nil, Err: UnknownNetworkError(network)}
}

func SocketPair(network string) (so [2]Socket, err error) {
	switch network {
	case "unix", "unixpacket":
		uSockets, err := newUnixSocketPair()
		if err != nil {
			return [2]Socket{}, err
		}
		so[0] = uSockets[0]
		so[1] = uSockets[1]
		return so, nil
	}

	return [2]Socket{}, &OpError{Op: "socketpair", Net: network, Source: nil, Addr: nil, Err: UnknownNetworkError(network)}
}
