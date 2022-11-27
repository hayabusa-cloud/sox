//go:build unix

package sox

import (
	"golang.org/x/sys/unix"
	"net"
	"unsafe"
)

type UnixAddr = net.UnixAddr

var (
	ResolveUnixAddr = net.ResolveUnixAddr
)

func unixAddrToSockAddr(addr *net.UnixAddr) unix.Sockaddr {
	return &unix.SockaddrUnix{
		Name: addr.Name,
	}
}

func unixSockaddr(sa *unix.SockaddrUnix) (ptr unsafe.Pointer, n int, err error) {
	rawSa := &unix.RawSockaddrUnix{
		Family: unix.AF_UNIX,
	}
	pp := (*[]byte)(unsafe.Pointer(&rawSa.Path[0]))
	copy(*pp, sa.Name)
	(*pp)[len(sa.Name)] = 0
	n = int(unsafe.Sizeof(rawSa.Family)) + len(sa.Name)
	if sa.Name == "@" {
		*pp = []byte("")
	} else {
		n++
	}
	return ptr, n, nil
}
