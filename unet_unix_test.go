// ©Hayabusa Cloud Co., Ltd. 2022. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

//go:build linux

package sox

import (
	"bytes"
	"encoding/binary"
	"golang.org/x/sys/unix"
	"testing"
	"unsafe"
)

func TestUnixSockaddr(t *testing.T) {
	t.Run("abstract", func(t *testing.T) {
		unixAddr, _ := ResolveUnixAddr("unixpacket", "@")
		sa := unixAddrToSockaddr(unixAddr).(*unix.SockaddrUnix)
		if unixAddr.Name != sa.Name {
			t.Errorf("unix sockaddr expected name=%s but got %s", unixAddr.Name, sa.Name)
			return
		}
		ptr, n, err := unixSockaddr(sa)
		if err != nil {
			t.Errorf("unix sockaddr: %v", err)
			return
		}
		if n != unix.SizeofShort+1 {
			t.Errorf("unix sockaddr bad length expected 3 but got %d", n)
			return
		}
		b := unsafe.Slice((*byte)(ptr), unix.SizeofShort+1)
		family := binary.LittleEndian.Uint16(b[:unix.SizeofShort])
		if unix.AF_UNIX != family {
			t.Errorf("unix sockaddr bad family expected %s but got %x", "AF_UNIX", family)
			return
		}
		if b[unix.SizeofShort] != 0 {
			t.Errorf("unix sockaddr expected \\0 but got %s", b[unix.SizeofShort:])
			return
		}
	})

	t.Run("pathname", func(t *testing.T) {
		name := "uds_test"
		unixAddr, _ := ResolveUnixAddr("unixpacket", name)
		sa := unixAddrToSockaddr(unixAddr).(*unix.SockaddrUnix)
		if unixAddr.Name != sa.Name {
			t.Errorf("unix sockaddr expected name=%s but got %s", unixAddr.Name, sa.Name)
			return
		}
		ptr, n, err := unixSockaddr(sa)
		if err != nil {
			t.Errorf("unix sockaddr: %v", err)
			return
		}
		if n != unix.SizeofShort+len(name)+1 {
			t.Errorf("unix sockaddr bad length expected %d but got %d", len(name)+1, n)
			return
		}
		b := unsafe.Slice((*byte)(ptr), n)
		family := binary.LittleEndian.Uint16(b[:unix.SizeofShort])
		if unix.AF_UNIX != family {
			t.Errorf("unix sockaddr bad family expected %s but got %x", "AF_UNIX", family)
			return
		}
		if bytes.Equal(b[unix.SizeofShort:], []byte(name)) {
			t.Errorf("unix sockaddr expected %s but got %s", name, b[unix.SizeofShort:])
			return
		}
	})
}
