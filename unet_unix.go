// ©Hayabusa Cloud Co., Ltd. 2022. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

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

func unixAddrToSockaddr(addr *net.UnixAddr) unix.Sockaddr {
	return &unix.SockaddrUnix{
		Name: addr.Name,
	}
}

func unixSockaddr(sa *unix.SockaddrUnix) (ptr unsafe.Pointer, n int, err error) {
	rawSa := &unix.RawSockaddrUnix{
		Family: unix.AF_UNIX,
	}
	pp := (*byte)(unsafe.Pointer(&rawSa.Path[0]))
	copy(unsafe.Slice(pp, len(rawSa.Path)), sa.Name)
	n = int(unsafe.Sizeof(rawSa.Family)) + len(sa.Name)
	if sa.Name == "@" {
		*pp = 0
	} else {
		n++
	}
	ptr = unsafe.Pointer(rawSa)

	return ptr, n, nil
}
