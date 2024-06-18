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

var PageSize = uintptr(os.Getpagesize())

// AlignedMem returns a byte slice with the specified size
// and starting address aligned to the memory page size.
func AlignedMem(size int) []byte {
	p := make([]byte, uintptr(size)+PageSize-1)
	ptr := uintptr(unsafe.Pointer(unsafe.SliceData(p)))
	ptr = ((ptr + PageSize - 1) / PageSize) * PageSize
	return unsafe.Slice((*byte)(unsafe.Pointer(ptr)), size)
}

// AlignedMemBlocks returns n bytes slices which
// has length with memory page size and address
// starts from multiple of memory page size
func AlignedMemBlocks(n int) (blocks [][]byte) {
	if n < 1 {
		panic("bad block num")
	}
	blocks = make([][]byte, n)
	p := make([]byte, int(PageSize)*(n+1))
	ptr := uintptr(unsafe.Pointer(&p[0]))
	off := ptr - (ptr & ^(PageSize - 1))
	for i := range n {
		blocks[i] = unsafe.Slice(&p[uintptr(i)*PageSize-off], PageSize)
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
	_ = 1 << (iota * 2)
	_

	// BufferSizePico represents the size of a pico buffer
	BufferSizePico

	// BufferSizeNano represents the size of a nano buffer.
	BufferSizeNano

	// BufferSizeMicro represents the size of a micro buffer.
	BufferSizeMicro

	// BufferSizeSmall represents the size of a small buffer.
	BufferSizeSmall

	// BufferSizeMedium represents the size of a medium buffer.
	BufferSizeMedium

	// BufferSizeLarge represents the size of a large buffer.
	BufferSizeLarge

	// BufferSizeHuge represents the size of a huge buffer.
	BufferSizeHuge

	// BufferSizeGiant represents the size of a giant buffer.
	BufferSizeGiant

	bufferSizeDefault = BufferSizeMedium
)

// NewPicoBuffer returns a new PicoBuffer.
func NewPicoBuffer() PicoBuffer { return PicoBuffer{} }

// NewNanoBuffer returns a new instance of NanoBuffer.
func NewNanoBuffer() NanoBuffer { return NanoBuffer{} }

// NewMicroBuffer returns a new instance of MicroBuffer.
func NewMicroBuffer() MicroBuffer { return MicroBuffer{} }

// NewSmallBuffer returns a new instance of SmallBuffer.
func NewSmallBuffer() SmallBuffer { return SmallBuffer{} }

// NewMediumBuffer returns a new instance of MediumBuffer.
func NewMediumBuffer() MediumBuffer { return MediumBuffer{} }

// NewLargeBuffer returns a new instance of LargeBuffer.
func NewLargeBuffer() LargeBuffer { return LargeBuffer{} }

// NewHugeBuffer returns a new instance of HugeBuffer.
func NewHugeBuffer() HugeBuffer { return HugeBuffer{} }

// NewGiantBuffer returns a new instance of GiantBuffer.
func NewGiantBuffer() GiantBuffer { return GiantBuffer{} }

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

// GiantBuffer represents a byte array with size of BufferSizeGiant
type GiantBuffer [BufferSizeGiant]byte

func (b GiantBuffer) Reset() {}

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

func giantArrayFromSlice(s []byte, offset int64) GiantBuffer {
	ptr := unsafe.Add(unsafe.Pointer(unsafe.SliceData(s)), offset)
	return *(*[BufferSizeGiant]byte)(ptr)
}

func sliceOfPicoArray(s []byte, offset int64, n int) []PicoBuffer {
	if n < 1 {
		panic("bad array num")
	}
	base := unsafe.Pointer(unsafe.SliceData(s))
	return unsafe.Slice((*PicoBuffer)(unsafe.Add(base, offset)), n)
}
func sliceOfNanoArray(s []byte, offset int64, n int) []NanoBuffer {
	if n < 1 {
		panic("bad array num")
	}
	base := unsafe.Pointer(unsafe.SliceData(s))
	return unsafe.Slice((*NanoBuffer)(unsafe.Add(base, offset)), n)
}
func sliceOfMicroArray(s []byte, offset int64, n int) []MicroBuffer {
	if n < 1 {
		panic("bad array num")
	}
	base := unsafe.Pointer(unsafe.SliceData(s))
	return unsafe.Slice((*MicroBuffer)(unsafe.Add(base, offset)), n)
}
func sliceOfSmallArray(s []byte, offset int64, n int) []SmallBuffer {
	if n < 1 {
		panic("bad array num")
	}
	base := unsafe.Pointer(unsafe.SliceData(s))
	return unsafe.Slice((*SmallBuffer)(unsafe.Add(base, offset)), n)
}
func sliceOfMediumArray(s []byte, offset int64, n int) []MediumBuffer {
	if n < 1 {
		panic("bad array num")
	}
	base := unsafe.Pointer(unsafe.SliceData(s))
	return unsafe.Slice((*MediumBuffer)(unsafe.Add(base, offset)), n)
}
func sliceOfLargeArray(s []byte, offset int64, n int) []LargeBuffer {
	if n < 1 {
		panic("bad array num")
	}
	base := unsafe.Pointer(unsafe.SliceData(s))
	return unsafe.Slice((*LargeBuffer)(unsafe.Add(base, offset)), n)
}
func sliceOfHugeArray(s []byte, offset int64, n int) []HugeBuffer {
	if n < 1 {
		panic("bad array num")
	}
	base := unsafe.Pointer(unsafe.SliceData(s))
	return unsafe.Slice((*HugeBuffer)(unsafe.Add(base, offset)), n)
}
func sliceOfGiantArray(s []byte, offset int64, n int) []GiantBuffer {
	if n < 1 {
		panic("bad array num")
	}
	base := unsafe.Pointer(unsafe.SliceData(s))
	return unsafe.Slice((*GiantBuffer)(unsafe.Add(base, offset)), n)
}
