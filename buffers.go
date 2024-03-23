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

const (
	_ = 1 << (iota * 3)
	BufferSizePico
	BufferSizeNano
	BufferSizeMicro
	BufferSizeSmall
	BufferSizeMedium
	BufferSizeLarge
	BufferSizeHuge
)

// PicoBuffer represents a byte array with size of BufferSizePico
type PicoBuffer [BufferSizePico]byte

func (b PicoBuffer) Reset() {}

// NanoBuffer represents a byte array with size of BufferSizeNano
type NanoBuffer [BufferSizeNano]byte

func (b NanoBuffer) Reset() {}

// MicroBuffer represents a byte array with size of BufferSizeMicro
type MicroBuffer [BufferSizeMicro]byte

func (b MicroBuffer) Reset() {}

// SmallBuffer represents a byte array with size of BufferSizeSmall
type SmallBuffer [BufferSizeSmall]byte

func (b SmallBuffer) Reset() {}

// MediumBuffer represents a byte array with size of BufferSizeMedium
type MediumBuffer [BufferSizeMedium]byte

func (b MediumBuffer) Reset() {}

// LargeBuffer represents a byte array with size of BufferSizeLarge
type LargeBuffer [BufferSizeLarge]byte

func (b LargeBuffer) Reset() {}

// HugeBuffer represents a byte array with size of BufferSizeHuge
type HugeBuffer [BufferSizeHuge]byte

func (b HugeBuffer) Reset() {}

func picoArrayFromSlice(s []byte, offset int64) PicoBuffer {
	ptr := unsafe.Add(unsafe.Pointer(unsafe.SliceData(s)), offset)
	return *(*[BufferSizePico]byte)(ptr)
}
func nanoArrayFromSlice(s []byte, offset int64) NanoBuffer {
	ptr := unsafe.Add(unsafe.Pointer(unsafe.SliceData(s)), offset)
	return *(*[BufferSizeNano]byte)(ptr)
}
func microArrayFromSlice(s []byte, offset int64) MicroBuffer {
	ptr := unsafe.Add(unsafe.Pointer(unsafe.SliceData(s)), offset)
	return *(*[BufferSizeMicro]byte)(ptr)
}
func smallArrayFromSlice(s []byte, offset int64) SmallBuffer {
	ptr := unsafe.Add(unsafe.Pointer(unsafe.SliceData(s)), offset)
	return *(*[BufferSizeSmall]byte)(ptr)
}
func mediumArrayFromSlice(s []byte, offset int64) MediumBuffer {
	ptr := unsafe.Add(unsafe.Pointer(unsafe.SliceData(s)), offset)
	return *(*[BufferSizeMedium]byte)(ptr)
}
func largeArrayFromSlice(s []byte, offset int64) LargeBuffer {
	ptr := unsafe.Add(unsafe.Pointer(unsafe.SliceData(s)), offset)
	return *(*[BufferSizeLarge]byte)(ptr)
}
func hugeArrayFromSlice(s []byte, offset int64) HugeBuffer {
	ptr := unsafe.Add(unsafe.Pointer(unsafe.SliceData(s)), offset)
	return *(*[BufferSizeHuge]byte)(ptr)
}

func sliceOfPicoArray(s []byte, offset int64, n int) []PicoBuffer {
	if n < 1 {
		panic("bad array num")
	}
	ret := make([]PicoBuffer, n)
	for i := range n {
		ret[i] = picoArrayFromSlice(s, offset)
		offset += BufferSizePico
	}

	return ret
}
func sliceOfNanoArray(s []byte, offset int64, n int) []NanoBuffer {
	if n < 1 {
		panic("bad array num")
	}
	ret := make([]NanoBuffer, n)
	for i := range n {
		ret[i] = nanoArrayFromSlice(s, offset)
		offset += BufferSizeNano
	}

	return ret
}
func sliceOfMicroArray(s []byte, offset int64, n int) []MicroBuffer {
	if n < 1 {
		panic("bad array num")
	}
	ret := make([]MicroBuffer, n)
	for i := range n {
		ret[i] = microArrayFromSlice(s, offset)
		offset += BufferSizeMicro
	}

	return ret
}
func sliceOfSmallArray(s []byte, offset int64, n int) []SmallBuffer {
	if n < 1 {
		panic("bad array num")
	}
	ret := make([]SmallBuffer, n)
	for i := range n {
		ret[i] = smallArrayFromSlice(s, offset)
		offset += BufferSizeSmall
	}

	return ret
}
func sliceOfMediumArray(s []byte, offset int64, n int) []MediumBuffer {
	if n < 1 {
		panic("bad array num")
	}
	ret := make([]MediumBuffer, n)
	for i := range n {
		ret[i] = mediumArrayFromSlice(s, offset)
		offset += BufferSizeMedium
	}

	return ret
}
func sliceOfLargeArray(s []byte, offset int64, n int) []LargeBuffer {
	if n < 1 {
		panic("bad array num")
	}
	ret := make([]LargeBuffer, n)
	for i := range n {
		ret[i] = largeArrayFromSlice(s, offset)
		offset += BufferSizeLarge
	}

	return ret
}
func sliceOfHugeArray(s []byte, offset int64, n int) []HugeBuffer {
	if n < 1 {
		panic("bad array num")
	}
	ret := make([]HugeBuffer, n)
	for i := range n {
		ret[i] = hugeArrayFromSlice(s, offset)
		offset += BufferSizeHuge
	}

	return ret
}
