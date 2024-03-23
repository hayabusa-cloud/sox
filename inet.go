// ©Hayabusa Cloud Co., Ltd. 2022. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package sox

import (
	"context"
	"net"
	"net/netip"
	"strconv"
	"strings"
)

var (
	IPV4zero        = net.IPv4zero
	IPV6unspecified = net.IPv6unspecified
	IPv4LoopBack    = net.IPv4(127, 0, 0, 1)
	IPv6LoopBack    = net.IPv6loopback
)

type IP = net.IP
type IPAddr = net.IPAddr
type TCPAddr = net.TCPAddr
type UDPAddr = net.UDPAddr

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
	ResolveIPAddr  = net.ResolveIPAddr
	ResolveTCPAddr = net.ResolveTCPAddr
	ResolveUDPAddr = net.ResolveUDPAddr
)

func ResolveSCTPAddr(network, address string) (*SCTPAddr, error) {
	switch network {
	case "sctp", "sctp4", "sctp6":
	case "": // a hint wildcard for Go 1.0 undocumented behavior
		network = "sctp"
	default:
		return nil, UnknownNetworkError(network)
	}
	var want6 bool
	if network == "sctp" || network == "sctp6" {
		want6 = true
	}
	if network == "sctp" && strings.ContainsAny(address, ":[") {
		want6 = true
	}
	if addrPort, err := netip.ParseAddrPort(address); err == nil {
		if addrPort.Addr().Is6() && want6 {
			return SCTPAddrFromAddrPort(addrPort), nil
		}
		if addrPort.Addr().Is4() && network != "sctp6" {
			return SCTPAddrFromAddrPort(addrPort), nil
		}
	}
	addrList, err := net.DefaultResolver.LookupAddr(context.Background(), address)
	if err != nil {
		return nil, err
	}
	var addr4 *SCTPAddr = nil
	for _, addr := range addrList {
		addrPort, err := netip.ParseAddrPort(addr)
		if err != nil {
			continue
		}
		if want6 && addrPort.Addr().Is6() {
			return SCTPAddrFromAddrPort(addrPort), nil
		}
		if addr4 == nil && addrPort.Addr().Is4() {
			addr4 = SCTPAddrFromAddrPort(addrPort)
			if !want6 {
				return addr4, nil
			}
		}
	}

	return addr4, nil
}

func IPAddrFromTCPAddr(addr *TCPAddr) *IPAddr {
	return &IPAddr{IP: addr.IP, Zone: addr.Zone}
}

func IPAddrFromUDPAddr(addr *UDPAddr) *IPAddr {
	return &IPAddr{IP: addr.IP, Zone: addr.Zone}
}

func IPAddrFromSCTPAddr(addr *SCTPAddr) *IPAddr {
	return &IPAddr{IP: addr.IP, Zone: addr.Zone}
}

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

func ip6ZoneID(zone string) int {
	if zone == "" {
		return 0
	}
	i, err := net.InterfaceByName(zone)
	if err != nil {
		panic(err)
	}
	return i.Index
}
