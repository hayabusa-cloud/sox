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
	// IPV4zero is an IPv4 address representing the zero value (0.0.0.0).
	IPV4zero = net.IPv4zero

	// IPV6unspecified is an IPv6 address representing the unspecified value (::).
	IPV6unspecified = net.IPv6unspecified

	// IPv4LoopBack is an IPv4 address representing the loopback address (127.0.0.1).
	IPv4LoopBack = net.IPv4(127, 0, 0, 1)

	// IPv6LoopBack is an IPv6 address representing the loopback address (::).
	IPv6LoopBack = net.IPv6loopback
)

// IP represents an IP address.
// It is a type alias for the net.IP type.
type IP = net.IP

// IPAddr represents a network address of type IP.
// It is a type alias for the net.IPAddr type.
type IPAddr = net.IPAddr

// TCPAddr represents the address of a TCP endpoint.
// It is a type alias for the net.TCPAddr type.
type TCPAddr = net.TCPAddr

// UDPAddr represents the address of a UDP endpoint.
// It is a type alias for the net.UDPAddr type.
type UDPAddr = net.UDPAddr

// SCTPAddr represents the address of a SCTP endpoint.
// It contains the IP address, port number, and IPv6 zone.
type SCTPAddr struct {
	IP   net.IP
	Port int
	Zone string
}

// Network returns the network name "sctp"
func (a *SCTPAddr) Network() string {
	return "sctp"
}

// String returns the string representation of the SCTPAddr.
// It returns "<nil>" if the SCTPAddr is nil.
// Otherwise, it combines the IP address, IPv6 zone (if present), and port number
// using the net.JoinHostPort function and returns the resulting string.
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
	// TCPAddrFromAddrPort refers to the net.TCPAddrFromAddrPort function
	// It returns addr as a [TCPAddr]. If addr.IsValid() is false,
	// then the returned TCPAddr will contain a nil IP field, indicating an
	// address-family-agnostic unspecified address.
	TCPAddrFromAddrPort = net.TCPAddrFromAddrPort

	// UDPAddrFromAddrPort refers to the net.UDPAddrFromAddrPort function
	// It returns addr as a UDPAddr. If addr.IsValid() is false,
	// then the returned UDPAddr will contain a nil IP field, indicating an
	// address-family-agnostic unspecified address.
	UDPAddrFromAddrPort = net.UDPAddrFromAddrPort
)

// SCTPAddrFromAddrPort returns a new SCTPAddr based on the given netip.AddrPort.
func SCTPAddrFromAddrPort(addr netip.AddrPort) *SCTPAddr {
	return &SCTPAddr{
		IP:   addr.Addr().AsSlice(),
		Zone: addr.Addr().Zone(),
		Port: int(addr.Port()),
	}
}

var (
	// ResolveIPAddr refers to the net.ResolveIPAddr function
	// It returns an address of the IP end point.
	ResolveIPAddr = net.ResolveIPAddr

	// ResolveTCPAddr refers to the net.ResolveTCPAddr function.
	// It returns a TCPAddr struct that contains IP and port information.
	ResolveTCPAddr = net.ResolveTCPAddr

	// ResolveUDPAddr refers to the net.ResolveUDPAddr function.
	// It takes a network type and a string representation of the address and returns a
	// UDPAddr struct that contains the IP and port information.
	ResolveUDPAddr = net.ResolveUDPAddr
)

// ResolveSCTPAddr resolves the SCTP network address of the given network and address string.
// It returns a new SCTPAddr based on the resolved address and network.
// Possible network values are "sctp", "sctp4", and "sctp6".
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

// IPAddrFromTCPAddr returns a new IPAddr based on the given TCPAddr.
func IPAddrFromTCPAddr(addr *TCPAddr) *IPAddr {
	return &IPAddr{IP: addr.IP, Zone: addr.Zone}
}

// IPAddrFromUDPAddr returns a new IPAddr based on the given UDPAddr.
// It sets the IP and Zone fields of the IPAddr with the values from the UDPAddr.
func IPAddrFromUDPAddr(addr *UDPAddr) *IPAddr {
	return &IPAddr{IP: addr.IP, Zone: addr.Zone}
}

// IPAddrFromSCTPAddr returns a new IPAddr based on the given SCTPAddr.
func IPAddrFromSCTPAddr(addr *SCTPAddr) *IPAddr {
	return &IPAddr{IP: addr.IP, Zone: addr.Zone}
}

// IP4AddressToBytes converts an IPv4 address to a byte array.
// If the given IP address is not an IPv4 address, it returns an empty byte array.
// The byte array contains the four octets of the IPv4 address in network byte order.
func IP4AddressToBytes(ip net.IP) [4]byte {
	ip4 := ip.To4()
	if ip4 == nil {
		return [4]byte{}
	}
	return [4]byte{ip4[0], ip4[1], ip4[2], ip4[3]}
}

// IP6AddressToBytes converts the given net.IPv6 address to a fixed-size byte array.
// The resulting byte array contains the individual bytes of the IPv6 address in the same order as the original address.
// Each byte of the byte array corresponds to a byte of the IPv6 address.
// For example, the first byte of the byte array corresponds to the first byte of the IPv6 address, and so on.
// The byte array has a length of 16 bytes.
//
// Note: This function assumes that the given net.IPv6 address is a valid IPv6 address.
func IP6AddressToBytes(ip net.IP) [16]byte {
	return [16]byte{
		ip[0], ip[1], ip[2], ip[3],
		ip[4], ip[5], ip[6], ip[7],
		ip[8], ip[9], ip[10], ip[11],
		ip[12], ip[13], ip[14], ip[15],
	}
}

func ipFamily(ip IP) NetworkType {
	if ip == nil || ip.IsUnspecified() {
		return NetworkIPv6
	}
	if len(ip) <= net.IPv4len {
		return NetworkIPv4
	}
	if ip.To4() != nil {
		return NetworkIPv4
	}
	return NetworkIPv6
}

func networkIPFamily(network string, ip IP) NetworkType {
	if strings.HasSuffix(network, "4") {
		return NetworkIPv4
	} else if strings.HasSuffix(network, "6") {
		return NetworkIPv6
	}
	return ipFamily(ip)
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

func ip6ZoneString(id int) string {
	if id == 0 {
		return ""
	}
	if i, err := net.InterfaceByIndex(id); err == nil {
		return i.Name
	}
	return ""
}
