package sox

import (
	"net"
	"net/netip"
	"strconv"
)

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
