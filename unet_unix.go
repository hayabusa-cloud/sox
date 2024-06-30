// ©Hayabusa Cloud Co., Ltd. 2022. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

//go:build unix

package sox

import (
	"golang.org/x/sys/unix"
	"net"
	"slices"
	"unsafe"
)

// UnixAddr represents a Unix domain socket address.
// It is an alias for the net.UnixAddr type.
type UnixAddr = net.UnixAddr

var (
	// ResolveUnixAddr is a function that resolves a Unix network address.
	ResolveUnixAddr = net.ResolveUnixAddr
)

func unixAddrToSockaddr(addr *UnixAddr) *SockaddrUnix {
	return &SockaddrUnix{
		Name: addr.Name,
	}
}

func unixSockaddr(sa *SockaddrUnix) (ptr unsafe.Pointer, n int, err error) {
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

func unixSockaddrData(sa *SockaddrUnix) (data []byte, err error) {
	ptr, n, _ := unixSockaddr(sa)
	return unsafe.Slice((*byte)(ptr), n), nil
}

func unixSockaddrFromData(data []byte) *SockaddrUnix {
	raw := (*unix.RawSockaddrUnix)(unsafe.Pointer(unsafe.SliceData(data)))
	sa := SockaddrUnix{}
	if raw.Path[0] == 0 {
		sa.Name = "@"
	} else {
		l := slices.Index(raw.Path[:], 0)
		sa.Name = unsafe.String((*byte)(unsafe.Pointer(&raw.Path[0])), l)
	}
	fRawPtr := unsafe.Add(unsafe.Pointer(&sa), unsafe.Sizeof(sa.Name))
	*(*unix.RawSockaddrUnix)(fRawPtr) = *raw
	return &sa
}
