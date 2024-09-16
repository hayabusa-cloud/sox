// ©Hayabusa Cloud Co., Ltd. 2024. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

//go:build linux

package sox

import (
	"context"
	"math"
	"unsafe"
)

type bufferGroupIndex uint16

const (
	bufferGroupIndexPico = iota
	bufferGroupIndexNano
	bufferGroupIndexMicro
	bufferGroupIndexSmall
	bufferGroupIndexMedium
	bufferGroupIndexLarge
	bufferGroupIndexHuge
	bufferGroupIndexGiant
	bufferGroupIndexEnd

	bufferGroupIndexMask = bufferGroupIndexEnd - 1
)

const (
	bufferNumPico = 1 << (15 - iota/2)
	bufferNumNano
	bufferNumMicro
	bufferNumSmall
	bufferNumMedium = 1 << (21 - iota*2)
	bufferNumLarge
	bufferNumHuge
	bufferNumGiant
)

const (
	registerBufferNum        = UringEntriesLarge
	registerBufferSize       = BufferSizeHuge
	registerBufferDefaultMem = 1 << 24
)

type (

	// RegisterBuffer represents a fixed-size buffer used for registering with the I/O ring.
	RegisterBuffer [registerBufferSize]byte

	// RegisterBufferPool represents a pool of fixed-size buffers used for registering with the I/O ring.
	RegisterBufferPool = BoundedPool[RegisterBuffer]
)

// NewRegisterBufferPool creates a new instance of RegisterBufferPool with the specified capacity.
func NewRegisterBufferPool(capacity int) *RegisterBufferPool {
	return NewBoundedPool[RegisterBuffer](capacity)
}

func newUringProvideBuffers(size, n int) *uringProvideBuffers {
	if size < 1 || size > (1<<24) {
		panic("size must be between 1 and 16777216")
	}
	if n < 1 || n > (1<<24) {
		panic("n must be between 1 and 16777216")
	}
	size--
	size |= size >> 1
	size |= size >> 2
	size |= size >> 4
	size |= size >> 8
	size |= size >> 16
	size++
	mask := n - 1
	mask |= mask >> 1
	mask |= mask >> 2
	mask |= mask >> 4
	mask |= mask >> 8
	mask |= mask >> 16
	n = mask + 1
	mask |= math.MaxInt16
	gn := 1
	if n > (1 << 15) {
		gn, n = n>>15, 1<<15
	}
	mem := AlignedMem(int(size) * n)
	ptr := unsafe.Pointer(unsafe.SliceData(mem))
	return &uringProvideBuffers{
		gn:    gn,
		n:     n,
		gMask: gn - 1,
		mask:  mask,
		mem:   mem,
		ptr:   ptr,
		size:  size,
	}
}

type uringProvideBuffers struct {
	gn        int
	n         int
	gMask     int
	mask      int
	mem       []byte
	ptr       unsafe.Pointer
	size      int
	gidOffset uint16
}

func (g *uringProvideBuffers) setGIDOffset(offset int) {
	g.gidOffset = uint16(offset)
}

func (g *uringProvideBuffers) register(ctx context.Context, ur *ioUring, flags uint8) error {
	for i := range g.gn {
		addr := unsafe.Add(g.ptr, g.size*g.n*i)
		gid := uint16(i) + g.gidOffset
		err := ur.provideBuffers(ctx, flags, g.n, 0, addr, g.size, gid)
		if err != nil {
			return err
		}
	}
	return nil
}

func (g *uringProvideBuffers) provide(ctx context.Context, ur *ioUring, flags uint8, group uint16, data []byte, index uint32) error {
	ptr, n := unsafe.Pointer(unsafe.SliceData(data)), len(data)
	err := ur.provideBuffers(ctx, flags, 1, int(index), ptr, n, group)
	if err != nil {
		return err
	}
	return nil
}

func (g *uringProvideBuffers) bufGroup(f PollFd) uint16 {
	return uint16(f.Fd()&g.gMask) + g.gidOffset
}

func (g *uringProvideBuffers) buf(bufGroup bufferGroupIndex, bufIndex uint32) []byte {
	i := (int(bufGroup)-int(g.gidOffset))*g.n + int(bufIndex)
	return unsafe.Slice((*byte)(unsafe.Add(g.ptr, g.size*i)), g.size)
}

func (g *uringProvideBuffers) data(f PollFd, bufIndex uint32, length int) []byte {
	i := (f.Fd()&g.gMask)*g.n + int(bufIndex)
	return unsafe.Slice((*byte)(unsafe.Add(g.ptr, g.size*i)), length)
}

func newUringBufferGroups(scale int) *uringProvideBufferGroups {
	if scale < 1 || scale > (1<<12) {
		panic("scale must be between 1 and 4096")
	}
	mask := scale - 1
	mask |= mask >> 1
	mask |= mask >> 2
	mask |= mask >> 4
	mask |= mask >> 8
	scale = mask + 1

	pico := AlignedMem(BufferSizePico * bufferNumPico * scale)
	nano := AlignedMem(BufferSizeNano * bufferNumNano * scale)
	micro := AlignedMem(BufferSizeMicro * bufferNumMicro * scale)
	small := AlignedMem(BufferSizeSmall * bufferNumSmall * scale)
	medium := AlignedMem(BufferSizeMedium * bufferNumMedium * scale)
	large := AlignedMem(BufferSizeLarge * bufferNumLarge * scale)
	huge := AlignedMem(BufferSizeHuge * bufferNumHuge * scale)
	giant := AlignedMem(BufferSizeGiant * bufferNumGiant * scale)

	return &uringProvideBufferGroups{
		scale:     scale,
		scaleMask: mask,

		picoMem:   pico,
		nanoMem:   nano,
		microMem:  micro,
		smallMem:  small,
		mediumMem: medium,
		largeMem:  large,
		hugeMem:   huge,
		giantMem:  giant,

		picoBuffers:   sliceOfPicoArray(pico, 0, len(pico)/BufferSizePico),
		nanoBuffers:   sliceOfNanoArray(nano, 0, len(nano)/BufferSizeNano),
		microBuffers:  sliceOfMicroArray(micro, 0, len(micro)/BufferSizeMicro),
		smallBuffers:  sliceOfSmallArray(small, 0, len(small)/BufferSizeSmall),
		mediumBuffers: sliceOfMediumArray(medium, 0, len(medium)/BufferSizeMedium),
		largeBuffers:  sliceOfLargeArray(large, 0, len(large)/BufferSizeLarge),
		hugeBuffers:   sliceOfHugeArray(huge, 0, len(huge)/BufferSizeHuge),
		giantBuffers:  sliceOfGiantArray(giant, 0, len(giant)/BufferSizeGiant),
	}
}

type uringProvideBufferGroups struct {
	scale     int
	scaleMask int

	picoMem   []byte
	nanoMem   []byte
	microMem  []byte
	smallMem  []byte
	mediumMem []byte
	largeMem  []byte
	hugeMem   []byte
	giantMem  []byte

	_             [0][1]byte
	picoBuffers   []PicoBuffer
	nanoBuffers   []NanoBuffer
	microBuffers  []MicroBuffer
	smallBuffers  []SmallBuffer
	mediumBuffers []MediumBuffer
	largeBuffers  []LargeBuffer
	hugeBuffers   []HugeBuffer
	giantBuffers  []GiantBuffer

	gidOffset uint16
}

func (g *uringProvideBufferGroups) setGIDOffset(offset int) {
	g.gidOffset = uint16(offset)
}

func (g *uringProvideBufferGroups) register(ctx context.Context, ur *ioUring, flags uint8) error {
	for i := range g.scale {
		addr := unsafe.Pointer(&g.picoBuffers[i*bufferNumPico])
		gid := uint16(i*bufferGroupIndexEnd+bufferGroupIndexPico) + g.gidOffset
		err := ur.provideBuffers(ctx, flags, bufferNumPico, 0, addr, BufferSizePico, gid)
		if err != nil {
			return err
		}
		addr = unsafe.Pointer(&g.nanoBuffers[i*bufferNumNano])
		gid = uint16(i*bufferGroupIndexEnd+bufferGroupIndexNano) + g.gidOffset
		err = ur.provideBuffers(ctx, flags, bufferNumNano, 0, addr, BufferSizeNano, gid)
		if err != nil {
			return err
		}
		addr = unsafe.Pointer(&g.microBuffers[i*bufferNumMicro])
		gid = uint16(i*bufferGroupIndexEnd+bufferGroupIndexMicro) + g.gidOffset
		err = ur.provideBuffers(ctx, flags, bufferNumMicro, 0, addr, BufferSizeMicro, gid)
		if err != nil {
			return err
		}
		addr = unsafe.Pointer(&g.smallBuffers[i*bufferNumSmall])
		gid = uint16(i*bufferGroupIndexEnd+bufferGroupIndexSmall) + g.gidOffset
		err = ur.provideBuffers(ctx, flags, bufferNumSmall, 0, addr, BufferSizeSmall, gid)
		if err != nil {
			return err
		}
		addr = unsafe.Pointer(&g.mediumBuffers[i*bufferNumMedium])
		gid = uint16(i*bufferGroupIndexEnd+bufferGroupIndexMedium) + g.gidOffset
		err = ur.provideBuffers(ctx, flags, bufferNumMedium, 0, addr, BufferSizeMedium, gid)
		if err != nil {
			return err
		}
		addr = unsafe.Pointer(&g.largeBuffers[i*bufferNumLarge])
		gid = uint16(i*bufferGroupIndexEnd+bufferGroupIndexLarge) + g.gidOffset
		err = ur.provideBuffers(ctx, flags, bufferNumLarge, 0, addr, BufferSizeLarge, gid)
		if err != nil {
			return err
		}
		addr = unsafe.Pointer(&g.hugeBuffers[i*bufferNumHuge])
		gid = uint16(i*bufferGroupIndexEnd+bufferGroupIndexHuge) + g.gidOffset
		err = ur.provideBuffers(ctx, flags, bufferNumHuge, 0, addr, BufferSizeHuge, gid)
		if err != nil {
			return err
		}
		addr = unsafe.Pointer(&g.giantBuffers[i*bufferNumGiant])
		gid = uint16(i*bufferGroupIndexEnd+bufferGroupIndexGiant) + g.gidOffset
		err = ur.provideBuffers(ctx, flags, bufferNumGiant, 0, addr, BufferSizeGiant, gid)
		if err != nil {
			return err
		}
	}
	return nil
}

func (g *uringProvideBufferGroups) provide(ctx context.Context, ur *ioUring, flags uint8, group uint16, data []byte, index uint32) error {
	ptr, n := unsafe.Pointer(unsafe.SliceData(data)), len(data)
	err := ur.provideBuffers(ctx, flags, 1, int(index), ptr, n, group)
	if err != nil {
		return err
	}
	return nil
}

func (g *uringProvideBufferGroups) bufGroup(f PollFd) uint16 {
	offset := uint16(f.Fd() & g.scaleMask)
	return bufferGroupIndexEnd*offset + bufferGroupIndexMedium + g.gidOffset
}

func (g *uringProvideBufferGroups) bufGroupBySize(f PollFd, size int) uint16 {
	offset := uint16(f.Fd() & g.scaleMask)
	if size <= BufferSizePico {
		return bufferGroupIndexEnd*offset + bufferGroupIndexPico + g.gidOffset
	} else if size <= BufferSizeNano {
		return bufferGroupIndexEnd*offset + bufferGroupIndexNano + g.gidOffset
	} else if size <= BufferSizeMicro {
		return bufferGroupIndexEnd*offset + bufferGroupIndexMicro + g.gidOffset
	} else if size <= BufferSizeSmall {
		return bufferGroupIndexEnd*offset + bufferGroupIndexSmall + g.gidOffset
	} else if size <= BufferSizeMedium {
		return bufferGroupIndexEnd*offset + bufferGroupIndexMedium + g.gidOffset
	} else if size <= BufferSizeLarge {
		return bufferGroupIndexEnd*offset + bufferGroupIndexLarge + g.gidOffset
	} else if size <= BufferSizeHuge {
		return bufferGroupIndexEnd*offset + bufferGroupIndexHuge + g.gidOffset
	} else if size <= BufferSizeGiant {
		return bufferGroupIndexEnd*offset + bufferGroupIndexGiant + g.gidOffset
	}
	return bufferGroupIndexEnd*(offset+1) + g.gidOffset
}

func (g *uringProvideBufferGroups) bufGroupByData(f PollFd, b []byte) uint16 {
	if b == nil {
		return g.bufGroup(f)
	}
	return g.bufGroupBySize(f, len(b))
}

func (g *uringProvideBufferGroups) buf(bufGroup bufferGroupIndex, bufIndex uint32) []byte {
	switch (uint16(bufGroup) - g.gidOffset) & bufferGroupIndexMask {
	case bufferGroupIndexPico:
		i := (uint32(bufGroup)-uint32(g.gidOffset))/bufferGroupIndexEnd*bufferNumPico + bufIndex
		return g.picoBuffers[i][:]
	case bufferGroupIndexNano:
		i := (uint32(bufGroup)-uint32(g.gidOffset))/bufferGroupIndexEnd*bufferNumNano + bufIndex
		return g.nanoBuffers[i][:]
	case bufferGroupIndexMicro:
		i := (uint32(bufGroup)-uint32(g.gidOffset))/bufferGroupIndexEnd*bufferNumMicro + bufIndex
		return g.microBuffers[i][:]
	case bufferGroupIndexSmall:
		i := (uint32(bufGroup)-uint32(g.gidOffset))/bufferGroupIndexEnd*bufferNumSmall + bufIndex
		return g.smallBuffers[i][:]
	case bufferGroupIndexMedium:
		i := (uint32(bufGroup)-uint32(g.gidOffset))/bufferGroupIndexEnd*bufferNumMedium + bufIndex
		return g.mediumBuffers[i][:]
	case bufferGroupIndexLarge:
		i := (uint32(bufGroup)-uint32(g.gidOffset))/bufferGroupIndexEnd*bufferNumLarge + bufIndex
		return g.largeBuffers[i][:]
	case bufferGroupIndexHuge:
		i := (uint32(bufGroup)-uint32(g.gidOffset))/bufferGroupIndexEnd*bufferNumHuge + bufIndex
		return g.hugeBuffers[i][:]
	case bufferGroupIndexGiant:
		i := (uint32(bufGroup)-uint32(g.gidOffset))/bufferGroupIndexEnd*bufferNumGiant + bufIndex
		return g.giantBuffers[i][:]
	default:
		return []byte{}
	}
}

func (g *uringProvideBufferGroups) picoData(f PollFd, bufIndex uint32, length int) []byte {
	return g.picoBuffers[(f.Fd()&g.scaleMask)*bufferNumPico+int(bufIndex)][:length]
}

func (g *uringProvideBufferGroups) nanoData(f PollFd, bufIndex uint32, length int) []byte {
	return g.nanoBuffers[(f.Fd()&g.scaleMask)*bufferNumNano+int(bufIndex)][:length]
}

func (g *uringProvideBufferGroups) microData(f PollFd, bufIndex uint32, length int) []byte {
	return g.microBuffers[(f.Fd()&g.scaleMask)*bufferNumMicro+int(bufIndex)][:length]
}

func (g *uringProvideBufferGroups) smallData(f PollFd, bufIndex uint32, length int) []byte {
	return g.smallBuffers[(f.Fd()&g.scaleMask)*bufferNumSmall+int(bufIndex)][:length]
}

func (g *uringProvideBufferGroups) mediumData(f PollFd, bufIndex uint32, length int) []byte {
	return g.mediumBuffers[(f.Fd()&g.scaleMask)*bufferNumMedium+int(bufIndex)][:length]
}

func (g *uringProvideBufferGroups) largeData(f PollFd, bufIndex uint32, length int) []byte {
	return g.largeBuffers[(f.Fd()&g.scaleMask)*bufferNumLarge+int(bufIndex)][:length]
}

func (g *uringProvideBufferGroups) hugeData(f PollFd, bufIndex uint32, length int) []byte {
	return g.hugeBuffers[(f.Fd()&g.scaleMask)*bufferNumHuge+int(bufIndex)][:length]
}

func (g *uringProvideBufferGroups) giantData(f PollFd, bufIndex uint32, length int) []byte {
	return g.giantBuffers[(f.Fd()&g.scaleMask)*bufferNumGiant+int(bufIndex)][:length]
}
