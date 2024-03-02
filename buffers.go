// ©Hayabusa Cloud Co., Ltd. 2022. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package sox

import (
	"net"
	"os"
	"unsafe"
)

// Buffers is alias of net.Buffers
type Buffers = net.Buffers

// AlignedMemBlocks returns n bytes slices which
// has length with memory page size and address
// starts from multiple of memory page size
func AlignedMemBlocks(n int) (blocks [][]byte) {
	if n < 1 {
		panic("bad block num")
	}
	blocks = make([][]byte, n)
	size := os.Getpagesize()
	p := make([]byte, size*(n+1))
	ptr := uintptr(unsafe.Pointer(&p[0]))
	off := ptr - (ptr & ^(uintptr(size) - 1))
	for i := range n {
		blocks[i] = unsafe.Slice(&p[i*size-int(off)], size)
	}
	return
}

// AlignedMemBlock returns one aligned block
func AlignedMemBlock() []byte {
	return AlignedMemBlocks(1)[0]
}

// NewBuffers creates and initializes a new Buffers with given n and size
func NewBuffers(n int, size int) Buffers {
	if n < 1 {
		return Buffers{}
	}
	ret := make(Buffers, n)
	for i := range n {
		if size > 0 {
			ret[i] = make([]byte, size)
		} else {
			ret[i] = []byte{}
		}
	}

	return ret
}
