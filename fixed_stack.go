// ©Hayabusa Cloud Co., Ltd. 2023. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package sox

import (
	"errors"
	"io"
	"math"
	"sync/atomic"
)

// Stack is the interface that wraps Push and Pop operations on a Stack
type Stack[ItemType any] interface {
	// Push inserts an element at the of the Stack
	// It returns io.ErrClosesPipe if the Stack is closed.
	// If the Stack is full and the Nonblocking option is set as true,
	// an ErrTemporarilyUnavailable will be returned.
	// If the Stack is full and the Nonblocking option is set as false,
	// the Push operation blocks until the Stack became not full
	Push(item ItemType) error
	// Pop removes and returns the element at the top of the Stack
	// When the Stack is empty,
	//   It returns the zero-value and io.EOF if the Stack is already closed.
	//   It returns the zero-value and ErrTemporarilyUnavailable if the Stack is set as Nonblocking.
	//   If the stack is not set as Nonblocking, it blocks until any element has been Pushed.
	Pop() (item ItemType, err error)
	// Close closes the Stack. The return value is always nil.
	Close() error
}

const (
	defaultFixedStackCapacity = math.MaxInt16
)

// NewFixedStack creates and returns a fixed capacity Stack with the given options
func NewFixedStack[ItemType any](opts ...func(options *FixedStackOptions)) (Stack[ItemType], error) {
	o := &FixedStackOptions{
		Capacity:    defaultFixedStackCapacity,
		Concurrent:  true,
		Nonblocking: false,
	}
	for _, f := range opts {
		f(o)
	}
	if o.Capacity < 1 || o.Capacity >= (1<<30) {
		return nil, errors.New("invalid fixed stack capacity")
	}
	o.Capacity |= o.Capacity >> 1
	o.Capacity |= o.Capacity >> 2
	o.Capacity |= o.Capacity >> 4
	o.Capacity |= o.Capacity >> 8
	o.Capacity |= o.Capacity >> 16

	s := make([]ItemType, o.Capacity)
	if o.Concurrent {
		stack := newFixedStackConcurrent[ItemType](s, o)
		return stack, nil
	}
	stack := newFixedStack[ItemType](s, o)

	return stack, nil
}

// FixedStackOptions holds optional parameters for Stack implementations
type FixedStackOptions struct {
	// Capacity specifies the capacity of Stack. The default Capacity is 32K
	Capacity uint32
	// Concurrent specifies whether the Stack works in Concurrent mode or not
	// Concurrent should be set as true,
	//   if there are multiple goroutines doing Push or multiple goroutines doing Pop
	Concurrent bool
	// Nonblocking specifies whether the Push or Pop operations will NOT block
	// even if it is temporarily unavailable or not
	Nonblocking bool
}

type fixedStack[T any] struct {
	*FixedStackOptions
	stack []T
	top   atomic.Uint32
}

func newFixedStack[T any](stack []T, opt *FixedStackOptions) *fixedStack[T] {
	return &fixedStack[T]{
		FixedStackOptions: opt,
		stack:             stack,
		top:               atomic.Uint32{},
	}
}

func (s *fixedStack[T]) Push(item T) error {
	for sw := NewSpinWait().SetLevel(spinWaitLevelProduce); !sw.Closed(); sw.Once() {
		top := s.top.Load()
		if top&fixedStackStatusClosed == fixedStackStatusClosed {
			return io.ErrClosedPipe
		}
		if top&fixedStackTopValueMask >= s.Capacity {
			if s.Nonblocking {
				return ErrTemporarilyUnavailable
			}
			continue
		}
		s.stack[top] = item
		if !s.top.CompareAndSwap(top, top+1) {
			continue
		}
		break
	}

	return nil
}

func (s *fixedStack[T]) Pop() (item T, err error) {
	for sw := NewSpinWait().SetLevel(SpinWaitLevelConsume); !sw.Closed(); sw.Once() {
		top := s.top.Load()
		if top&fixedStackTopValueMask <= 0 {
			if top&fixedStackStatusClosed == fixedStackStatusClosed {
				return item, io.EOF
			}
			if s.Nonblocking {
				return item, ErrTemporarilyUnavailable
			}
			continue
		}
		item = s.stack[s.top.Add(math.MaxUint32)&fixedStackTopValueMask]
		break
	}

	return item, nil
}

func (s *fixedStack[T]) Close() error {
	for sw := NewSpinWait().SetLevel(spinWaitLevelProduce); !sw.Closed(); {
		top := s.top.Load()
		if top&fixedStackStatusClosed == fixedStackStatusClosed {
			return nil
		}
		if swapped := s.top.CompareAndSwap(top, top|fixedStackStatusClosed); !swapped {
			sw.OnceWithLevel(spinWaitLevelAtomic)
			continue
		}

		break
	}

	return nil
}

const (
	fixedStackStatusWriting = 1 << 31
	fixedStackStatusClosed  = 1 << 30
	fixedStackStatusMask    = fixedStackStatusWriting | fixedStackStatusClosed
	fixedStackTopValueMask  = (1 << 30) - 1
)

type fixedStackConcurrent[T any] struct {
	*FixedStackOptions
	stack []T
	top   atomic.Uint32
}

func newFixedStackConcurrent[T any](stack []T, opt *FixedStackOptions) *fixedStackConcurrent[T] {
	return &fixedStackConcurrent[T]{
		FixedStackOptions: opt,
		stack:             stack,
		top:               atomic.Uint32{},
	}
}

func (s *fixedStackConcurrent[T]) Push(item T) error {
	for sw := NewSpinWait().SetLevel(spinWaitLevelProduce); !sw.Closed(); {
		top := s.top.Load()
		if top&fixedStackStatusWriting == fixedStackStatusWriting {
			sw.Once()
			continue
		}
		if top&fixedStackStatusClosed == fixedStackStatusClosed {
			return io.ErrClosedPipe
		}
		if top&fixedStackTopValueMask >= s.Capacity {
			if s.Nonblocking {
				return ErrTemporarilyUnavailable
			}
			sw.Once()
			continue
		}
		newTop := fixedStackStatusWriting | (top&fixedStackTopValueMask + 1)
		if !s.top.CompareAndSwap(top, newTop) {
			sw.OnceWithLevel(spinWaitLevelAtomic)
			continue
		}
		s.stack[top&fixedStackTopValueMask] = item
		s.top.Store(newTop&(fixedStackStatusMask^fixedStackStatusWriting) | newTop&fixedStackTopValueMask)
		break
	}

	return nil
}

func (s *fixedStackConcurrent[T]) Pop() (item T, err error) {
	for sw := NewSpinWait().SetLevel(SpinWaitLevelConsume); !sw.Closed(); {
		top := s.top.Load()
		if top&fixedStackTopValueMask <= 0 {
			if top&fixedStackStatusClosed == fixedStackStatusClosed {
				return item, io.EOF
			}
			if s.Nonblocking {
				return item, ErrTemporarilyUnavailable
			}
			sw.Once()
			continue
		}
		if top&fixedStackStatusWriting == fixedStackStatusWriting {
			sw.Once()
			continue
		}
		newTop := fixedStackStatusWriting | (top&fixedStackTopValueMask - 1)
		if !s.top.CompareAndSwap(top, newTop) {
			sw.OnceWithLevel(spinWaitLevelAtomic)
			continue
		}
		item = s.stack[newTop&fixedStackTopValueMask]
		s.top.Store(newTop&(fixedStackStatusMask^fixedStackStatusWriting) | newTop&fixedStackTopValueMask)
		break
	}

	return item, nil
}

func (s *fixedStackConcurrent[T]) Close() error {
	for sw := NewSpinWait().SetLevel(spinWaitLevelProduce); !sw.Closed(); {
		top := s.top.Load()
		if top&fixedStackStatusClosed == fixedStackStatusClosed {
			return nil
		}
		if top&fixedStackStatusWriting == fixedStackStatusWriting {
			sw.Once()
			continue
		}
		if swapped := s.top.CompareAndSwap(top, top|fixedStackStatusClosed); !swapped {
			sw.OnceWithLevel(spinWaitLevelAtomic)
			continue
		}

		break
	}

	return nil
}
