// ©Hayabusa Cloud Co., Ltd. 2022. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package sox

import (
	"errors"
	"io"
	"math"
	"sync/atomic"
)

const (
	defaultRingQueueCapacity = math.MaxInt16
)

// ItemProducer is the interface that Produce items and can be Close
type ItemProducer[ItemType any] interface {
	// Produce produces items
	Produce(item ItemType) error
	// Close closed the ItemProducer
	Close() error
}

// ItemConsumer is the interface that Consume items
type ItemConsumer[ItemType any] interface {
	// Consume consumes items
	Consume() (item ItemType, err error)
}

// NewRingQueue creates a ring queue with given options
// and returns the consumer and the producer of it
func NewRingQueue[ItemType any](
	opts ...func(options *RingQueueOptions)) (
	consumer ItemConsumer[ItemType],
	producer ItemProducer[ItemType],
	err error) {
	o := &RingQueueOptions{
		Capacity:          defaultRingQueueCapacity,
		ConcurrentProduce: true,
		ConcurrentConsume: true,
		Nonblocking:       false,
	}
	for _, f := range opts {
		f(o)
	}
	if o.Capacity < 1 || o.Capacity >= (1<<30) {
		return nil, nil, errors.New("invalid ring queue capacity")
	}
	o.Capacity |= o.Capacity >> 1
	o.Capacity |= o.Capacity >> 2
	o.Capacity |= o.Capacity >> 4
	o.Capacity |= o.Capacity >> 8
	o.Capacity |= o.Capacity >> 16

	if !o.ConcurrentProduce && !o.ConcurrentConsume {
		ring := newRingQueue[ItemType](o)
		return ring, ring, nil
	} else if o.ConcurrentProduce && !o.ConcurrentConsume {
		ring := newRingQueueConcurrentProduce[ItemType](o)
		return ring, ring, nil
	} else if !o.ConcurrentProduce && o.ConcurrentConsume {
		ring := newRingQueueConcurrentConsume[ItemType](o)
		return ring, ring, nil
	}
	ring := newRingQueueConcurrent[ItemType](o)

	return ring, ring, nil
}

// RingQueueOptions holds optional parameters for RingQueue implementations
type RingQueueOptions struct {
	Capacity          int
	ConcurrentProduce bool
	ConcurrentConsume bool
	Nonblocking       bool
}

type ringQueue[T any] struct {
	*RingQueueOptions
	ring                 []T
	capacity, head, tail uint32
	closed               bool
}

func newRingQueue[T any](opt *RingQueueOptions) *ringQueue[T] {
	return &ringQueue[T]{
		RingQueueOptions: opt,
		ring:             make([]T, opt.Capacity+1),
		capacity:         uint32(opt.Capacity),
		head:             0,
		tail:             0,
		closed:           false,
	}
}

func (rq *ringQueue[T]) Produce(item T) error {
	for sw := NewSpinWait().SetLevel(spinWaitLevelProduce); !sw.Closed(); sw.Once() {
		if rq.closed {
			return io.ErrClosedPipe
		}
		if (rq.tail+1)&rq.capacity == rq.head {
			if rq.Nonblocking {
				return ErrTemporarilyUnavailable
			}
			continue
		}
		break
	}
	rq.ring[rq.tail] = item
	rq.tail = (rq.tail + 1) & rq.capacity

	return nil
}

func (rq *ringQueue[T]) Consume() (item T, err error) {
	for sw := NewSpinWait().SetLevel(spinWaitLevelProduce); !sw.Closed(); sw.Once() {
		if rq.head == rq.tail {
			if rq.closed {
				return item, io.EOF
			}
			if rq.Nonblocking {
				return item, ErrTemporarilyUnavailable
			}
			continue
		}
		break
	}
	item = rq.ring[rq.head]
	rq.head = (rq.head + 1) & rq.capacity

	return item, nil
}

func (rq *ringQueue[T]) Close() error {
	rq.closed = true

	return nil
}

const (
	ringQueueStatusWriting  = 1 << 31
	ringQueueStatusClosed   = 1 << 30
	ringQueueTailStatusMask = ringQueueStatusWriting | ringQueueStatusClosed
	ringQueueTailValueMask  = (1 << 30) - 1
)

type ringQueueConcurrentProduce[T any] struct {
	*RingQueueOptions
	ring           []T
	capacity, head uint32
	*ringQueueConcurrentClose
}

func newRingQueueConcurrentProduce[T any](opt *RingQueueOptions) *ringQueueConcurrentProduce[T] {
	return &ringQueueConcurrentProduce[T]{
		RingQueueOptions:         opt,
		ring:                     make([]T, opt.Capacity+1),
		capacity:                 uint32(opt.Capacity),
		head:                     0,
		ringQueueConcurrentClose: newRingQueueConcurrentClose(),
	}
}

func (rq *ringQueueConcurrentProduce[T]) Produce(item T) error {
	for sw := NewSpinWait().SetLevel(spinWaitLevelProduce); !sw.Closed(); {
		tail := rq.tail.Load()
		if tail&ringQueueStatusWriting == ringQueueStatusWriting {
			sw.Once()
			continue
		}
		if tail&ringQueueStatusClosed == ringQueueStatusClosed {
			return io.ErrClosedPipe
		}
		if ((tail&ringQueueTailValueMask)+1)&rq.capacity == rq.head {
			if rq.Nonblocking {
				break
			}
			sw.Once()
			continue
		}
		newTailStatus, newTailVal := (tail|ringQueueStatusWriting)&ringQueueTailStatusMask, (tail+1)&rq.capacity
		if swapped := rq.tail.CompareAndSwap(tail, newTailStatus|newTailVal); !swapped {
			sw.OnceWithLevel(spinWaitLevelAtomic)
			continue
		}
		rq.ring[tail&ringQueueTailValueMask] = item
		newTailStatus &= ringQueueTailStatusMask ^ ringQueueStatusWriting
		rq.tail.Store(newTailStatus | newTailVal)

		return nil
	}

	return ErrTemporarilyUnavailable
}

func (rq *ringQueueConcurrentProduce[T]) Consume() (item T, err error) {
	for sw := NewSpinWait().SetLevel(SpinWaitLevelConsume); !sw.Closed(); sw.Once() {
		tail := rq.tail.Load()
		if tail&ringQueueStatusWriting == ringQueueStatusWriting {
			continue
		}
		tailStatus, tailVal := tail&ringQueueTailStatusMask, tail&ringQueueTailValueMask
		if rq.head == tailVal {
			if tailStatus&ringQueueStatusClosed == ringQueueStatusClosed {
				return item, io.EOF
			}
			if rq.Nonblocking {
				return item, ErrTemporarilyUnavailable
			}
			continue
		}
		item = rq.ring[rq.head]
		rq.head = (rq.head + 1) & rq.capacity

		return item, nil
	}

	return
}

type ringQueueConcurrentConsume[T any] struct {
	*RingQueueOptions
	ring     []T
	capacity uint32
	head     atomic.Uint32
	tail     uint32
	closed   bool
}

func newRingQueueConcurrentConsume[T any](opt *RingQueueOptions) *ringQueueConcurrentConsume[T] {
	return &ringQueueConcurrentConsume[T]{
		RingQueueOptions: opt,
		ring:             make([]T, opt.Capacity+1),
		capacity:         uint32(opt.Capacity),
		head:             atomic.Uint32{},
		tail:             0,
		closed:           false,
	}
}

func (rq *ringQueueConcurrentConsume[T]) Produce(item T) error {
	if rq.closed {
		return io.ErrClosedPipe
	}
	for sw := NewSpinWait().SetLevel(spinWaitLevelProduce); !sw.Closed(); sw.Once() {
		if (rq.tail+1)&rq.capacity == rq.head.Load()&rq.capacity {
			if rq.Nonblocking {
				break
			}
			continue
		}
		rq.ring[rq.tail] = item
		rq.tail = (rq.tail + 1) & rq.capacity

		return nil
	}

	return ErrTemporarilyUnavailable
}

func (rq *ringQueueConcurrentConsume[T]) Consume() (item T, err error) {
	for sw := NewSpinWait().SetLevel(SpinWaitLevelConsume); !sw.Closed(); sw.Once() {
		head := rq.head.Load()
		if head == rq.tail {
			if rq.closed {
				return item, io.EOF
			}
			if rq.Nonblocking {
				return item, ErrTemporarilyUnavailable
			}
			continue
		}
		item = rq.ring[head]
		if swapped := rq.head.CompareAndSwap(head, (head+1)&rq.capacity); !swapped {
			continue
		}

		return item, nil
	}

	return
}

func (rq *ringQueueConcurrentConsume[T]) Close() error {
	rq.closed = true

	return nil
}

type ringQueueConcurrent[T any] struct {
	*RingQueueOptions
	ring     []T
	capacity uint32
	head     atomic.Uint32
	*ringQueueConcurrentClose
}

func newRingQueueConcurrent[T any](opt *RingQueueOptions) *ringQueueConcurrent[T] {
	return &ringQueueConcurrent[T]{
		RingQueueOptions:         opt,
		ring:                     make([]T, opt.Capacity+1),
		capacity:                 uint32(opt.Capacity),
		head:                     atomic.Uint32{},
		ringQueueConcurrentClose: newRingQueueConcurrentClose(),
	}
}

func (rq *ringQueueConcurrent[T]) Produce(item T) error {
	for sw := NewSpinWait().SetLevel(spinWaitLevelProduce); !sw.Closed(); {
		tail := rq.tail.Load()
		if tail&ringQueueStatusWriting == ringQueueStatusWriting {
			sw.Once()
			continue
		}
		if tail&ringQueueStatusClosed == ringQueueStatusClosed {
			return io.ErrClosedPipe
		}
		if (tail+1)&rq.capacity == rq.head.Load() {
			if rq.Nonblocking {
				break
			}
			sw.Once()
			continue
		}
		newTailStatus, newTailVal := (tail|ringQueueStatusWriting)&ringQueueTailStatusMask, (tail+1)&rq.capacity
		if swapped := rq.tail.CompareAndSwap(tail, newTailStatus|newTailVal); !swapped {
			sw.OnceWithLevel(spinWaitLevelAtomic)
			continue
		}
		rq.ring[tail&ringQueueTailValueMask] = item
		newTailStatus &= ringQueueTailStatusMask ^ ringQueueStatusWriting
		rq.tail.Store(newTailStatus | newTailVal)

		return nil
	}

	return ErrTemporarilyUnavailable
}

func (rq *ringQueueConcurrent[T]) Consume() (item T, err error) {
	for sw := NewSpinWait().SetLevel(SpinWaitLevelConsume); !sw.Closed(); {
		head, tail := rq.head.Load(), rq.tail.Load()
		if head == tail&ringQueueTailValueMask {
			if tail&ringQueueStatusClosed == ringQueueStatusClosed {
				return item, io.EOF
			}
			if rq.Nonblocking {
				return item, ErrTemporarilyUnavailable
			}
			sw.OnceWithLevel(SpinWaitLevelConsume)
			continue
		}
		if tail&ringQueueStatusWriting == ringQueueStatusWriting {
			sw.OnceWithLevel(SpinWaitLevelConsume)
			continue
		}
		item = rq.ring[head]
		if swapped := rq.head.CompareAndSwap(head, (head+1)&rq.capacity); !swapped {
			sw.OnceWithLevel(spinWaitLevelAtomic)
			continue
		}

		return item, nil
	}

	return
}

type ringQueueConcurrentClose struct {
	tail atomic.Uint32
}

func newRingQueueConcurrentClose() *ringQueueConcurrentClose {
	return &ringQueueConcurrentClose{tail: atomic.Uint32{}}
}

func (rq *ringQueueConcurrentClose) Close() error {
	for sw := NewSpinWait().SetLevel(spinWaitLevelProduce); !sw.Closed(); {
		tail := rq.tail.Load()
		if tail&ringQueueStatusClosed == ringQueueStatusClosed {
			return nil
		}
		if tail&ringQueueStatusWriting == ringQueueStatusWriting {
			sw.Once()
			continue
		}
		if swapped := rq.tail.CompareAndSwap(tail, tail|ringQueueStatusClosed); !swapped {
			sw.OnceWithLevel(spinWaitLevelAtomic)
			continue
		}

		break
	}

	return nil
}
