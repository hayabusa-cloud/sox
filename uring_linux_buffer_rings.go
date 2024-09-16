// ©Hayabusa Cloud Co., Ltd. 2024. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

//go:build linux

package sox

import (
	"sync/atomic"
	"unsafe"
)

func newUringBufferRings() *uringBufferRings {
	ret := uringBufferRings{
		rings:  make([]*ioUringBufRing, 0),
		counts: make([]uintptr, 0),
		masks:  make([]uintptr, 0),
	}
	return &ret
}

type uringBufferRings struct {
	rings  []*ioUringBufRing
	locks  []Spinlock
	counts []uintptr
	masks  []uintptr
}

func (rings *uringBufferRings) registerBuffers(ur *ioUring, b *uringProvideBuffers) error {
	for i := range b.gn {
		gid := uint16(i) + b.gidOffset
		ptr := unsafe.Add(b.ptr, b.size*b.n*i)
		br, err := rings.registerGroup(ur, b.n, gid, ptr, b.size)
		if err != nil {
			return err
		}
		rings.rings = append(rings.rings, br)
		rings.locks = append(rings.locks, Spinlock{})
		rings.counts = append(rings.counts, uintptr(b.n))
		rings.masks = append(rings.masks, uintptr(b.n-1))
	}
	return nil
}

func (rings *uringBufferRings) registerGroups(ur *ioUring, g *uringProvideBufferGroups) error {
	for i := range g.scale {
		gid := uint16(i)*bufferGroupIndexEnd + bufferGroupIndexPico + g.gidOffset
		ptr := unsafe.Pointer(&g.picoBuffers[i*bufferNumPico])
		br, err := rings.registerGroup(ur, bufferNumPico, gid, ptr, BufferSizePico)
		if err != nil {
			return err
		}
		rings.rings = append(rings.rings, br)
		rings.locks = append(rings.locks, Spinlock{})
		rings.counts = append(rings.counts, bufferNumPico)
		rings.masks = append(rings.masks, bufferNumPico-1)

		gid = uint16(i)*bufferGroupIndexEnd + bufferGroupIndexNano + g.gidOffset
		ptr = unsafe.Pointer(&g.nanoBuffers[i*bufferNumNano])
		br, err = rings.registerGroup(ur, bufferNumNano, gid, ptr, BufferSizeNano)
		if err != nil {
			return err
		}
		rings.rings = append(rings.rings, br)
		rings.locks = append(rings.locks, Spinlock{})
		rings.counts = append(rings.counts, bufferNumNano)
		rings.masks = append(rings.masks, bufferNumNano-1)

		gid = uint16(i)*bufferGroupIndexEnd + bufferGroupIndexMicro + g.gidOffset
		ptr = unsafe.Pointer(&g.microBuffers[i*bufferNumMicro])
		br, err = rings.registerGroup(ur, bufferNumMicro, gid, ptr, BufferSizeMicro)
		if err != nil {
			return err
		}
		rings.rings = append(rings.rings, br)
		rings.locks = append(rings.locks, Spinlock{})
		rings.counts = append(rings.counts, bufferNumMicro)
		rings.masks = append(rings.masks, bufferNumMicro-1)

		gid = uint16(i)*bufferGroupIndexEnd + bufferGroupIndexSmall + g.gidOffset
		ptr = unsafe.Pointer(&g.smallBuffers[i*bufferNumSmall])
		br, err = rings.registerGroup(ur, bufferNumSmall, gid, ptr, BufferSizeSmall)
		if err != nil {
			return err
		}
		rings.rings = append(rings.rings, br)
		rings.locks = append(rings.locks, Spinlock{})
		rings.counts = append(rings.counts, bufferNumSmall)
		rings.masks = append(rings.masks, bufferNumSmall-1)

		gid = uint16(i)*bufferGroupIndexEnd + bufferGroupIndexMedium + g.gidOffset
		ptr = unsafe.Pointer(&g.mediumBuffers[i*bufferNumMedium])
		br, err = rings.registerGroup(ur, bufferNumMedium, gid, ptr, BufferSizeMedium)
		if err != nil {
			return err
		}
		rings.rings = append(rings.rings, br)
		rings.locks = append(rings.locks, Spinlock{})
		rings.counts = append(rings.counts, bufferNumMedium)
		rings.masks = append(rings.masks, bufferNumMedium-1)

		gid = uint16(i)*bufferGroupIndexEnd + bufferGroupIndexLarge + g.gidOffset
		ptr = unsafe.Pointer(&g.largeBuffers[i*bufferNumLarge])
		br, err = rings.registerGroup(ur, bufferNumLarge, gid, ptr, BufferSizeLarge)
		if err != nil {
			return err
		}
		rings.rings = append(rings.rings, br)
		rings.locks = append(rings.locks, Spinlock{})
		rings.counts = append(rings.counts, bufferNumLarge)
		rings.masks = append(rings.masks, bufferNumLarge-1)

		gid = uint16(i)*bufferGroupIndexEnd + bufferGroupIndexHuge + g.gidOffset
		ptr = unsafe.Pointer(&g.hugeBuffers[i*bufferNumHuge])
		br, err = rings.registerGroup(ur, bufferNumHuge, gid, ptr, BufferSizeHuge)
		if err != nil {
			return err
		}
		rings.rings = append(rings.rings, br)
		rings.locks = append(rings.locks, Spinlock{})
		rings.counts = append(rings.counts, bufferNumHuge)
		rings.masks = append(rings.masks, bufferNumHuge-1)

		gid = uint16(i)*bufferGroupIndexEnd + bufferGroupIndexGiant + g.gidOffset
		ptr = unsafe.Pointer(&g.giantBuffers[i*bufferNumGiant])
		br, err = rings.registerGroup(ur, bufferNumGiant, gid, ptr, BufferSizeGiant)
		if err != nil {
			return err
		}
		rings.rings = append(rings.rings, br)
		rings.locks = append(rings.locks, Spinlock{})
		rings.counts = append(rings.counts, bufferNumGiant)
		rings.masks = append(rings.masks, bufferNumGiant-1)
	}
	return nil
}

func (rings *uringBufferRings) registerGroup(ur *ioUring, entries int, gid uint16, ptr unsafe.Pointer, size int) (*ioUringBufRing, error) {
	r, err := ur.registerBufRing(entries, gid)
	if err != nil {
		return nil, err
	}
	ur.bufRingInit(r)
	mask := entries - 1
	base := uintptr(ptr)
	for i := range entries {
		ur.bufRingAdd(r, base+uintptr(i*size), size, uint16(i), uintptr(mask), uintptr(i))
	}
	return r, nil
}

func (rings *uringBufferRings) provide(ur *ioUring, gidOffset, group uint16, addr unsafe.Pointer, size int, index uint32) {
	br := rings.rings[group-gidOffset]
	offset := atomic.AddUintptr(&rings.counts[group-gidOffset], 1)
	ur.bufRingAdd(br, uintptr(addr), size, uint16(index), rings.masks[group-gidOffset], offset-1)
}

func (rings *uringBufferRings) advance(ur *ioUring) {
	for i, r := range rings.rings {
		if r == nil {
			break
		}
		rings.locks[i].Lock()
		ur.bufRingAdvance(r, int(rings.counts[i]))
		rings.counts[i] = 0
		rings.locks[i].Unlock()
	}
}
