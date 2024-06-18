// ©Hayabusa Cloud Co., Ltd. 2024. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package sox

// Pool is a generic object pool
type Pool[T any] interface {
	// Put adds item to the pool
	// If the Pool is a nonblocking bounded pool, ErrTemporarilyUnavailable
	// will be returned when the Pool is full.
	// If the Pool is a bounded pool but not nonblocking,
	// Put would block until the Pool is no longer full
	// Note that if the Pool is not nonblocking, the return error can be ignored
	Put(item T) error
	// Get selects an arbitrary item from the Pool,
	// removes it from the Pool, and returns it to the caller.
	// If the Pool is empty and has been set to nonblocking,
	// ErrTemporarilyUnavailable will be returned.
	// If the Pool is empty and has been not set to nonblocking,
	// Get would block until any item put back to the Pool
	// Note that if the Pool is not nonblocking, the return error can be ignored
	Get() (item T, err error)
}

// IndirectPool is a generic object pool that Put and Get the int type indices of items instead of direct values.
// It extends the Pool[int] interface and adds Value and SetValue methods for working with indirect indices.
type IndirectPool[T any] interface {
	Pool[int]
	// Value returns the value with the given indirect index
	Value(indirect int) T
	// SetValue sets the value with the given indirect index
	SetValue(indirect int, item T)
}

type (

	// PicoBufferPool represents a generic object pool that uses indirect indices to
	// Put and Get items of type PicoBuffer.
	PicoBufferPool = IndirectPool[PicoBuffer]

	// NanoBufferPool represents a generic object pool that uses indirect indices to
	// Put and Get items of type NanoBuffer. It implements the IndirectPool interface.
	NanoBufferPool = IndirectPool[NanoBuffer]

	// MicroBufferPool represents a generic object pool that uses indirect indices to
	// Put and Get items of type MicroBuffer.
	MicroBufferPool = IndirectPool[MicroBuffer]

	// SmallBufferPool represents a generic object pool that uses indirect indices to
	// Put and Get items of type SmallBuffer.
	SmallBufferPool = IndirectPool[SmallBuffer]

	// MediumBufferPool represents a generic object pool that uses indirect indices to
	// Put and Get items of type MediumBuffer.
	MediumBufferPool = IndirectPool[MediumBuffer]

	// LargeBufferPool represents a generic object pool that uses indirect indices to
	// Put and Get items of type LargeBuffer.
	LargeBufferPool = IndirectPool[LargeBuffer]

	// HugeBufferPool represents a generic object pool that uses indirect indices to
	// Put and Get items of type HugeBuffer.
	HugeBufferPool = IndirectPool[HugeBuffer]

	// GiantBufferPool represents a generic object pool that uses indirect indices to
	// Put and Get items of type GiantBuffer
	GiantBufferPool = IndirectPool[GiantBuffer]
)
