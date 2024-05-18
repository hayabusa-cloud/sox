// ©Hayabusa Cloud Co., Ltd. 2022. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package sox

import (
	"math"
	"runtime"
	"time"
	_ "unsafe"
)

const (
	procYieldCycles          = 30
	spinWaitDurationBlocking = max(jiffies, 4*time.Millisecond)
	spinWaitDurationPending  = min(jiffies, time.Millisecond)
)

// SpinWait is a lightweight synchronization type that
// you can use in low-level scenarios with lower cost.
// The zero value for SpinWait is ready to use.
type SpinWait struct {
	_ noCopy

	i uint32
}

// Once performs a single spin
func (s *SpinWait) Once() {
	s.i++
	if s.WillYield() {
		runtime.Gosched()
		return
	}
	procyield(procYieldCycles)
}

// WillYield returns true if calling SpinOnce() will result
// in occurring a thread sleeping instead of a simply procyield()
func (s *SpinWait) WillYield() bool {
	return s.i >= 8
}

// Reset resets the counter in SpinWait
func (s *SpinWait) Reset() {
	s.i = 0
}

const (
	// SpinWaitLevelBlocking results in less CPU consumption but may cause a little delay
	// It is suggested to be used for io events polling etc.
	SpinWaitLevelBlocking = iota
	// SpinWaitLevelPending take balances of CPU consumption and delay
	// It is often to be used for producer-consumer ops on concurrent data structures
	SpinWaitLevelPending
	// SpinWaitLevelPreempting results in higher CPU consumption and lower delay,
	// which makes it suitable for time-sensitive operations on concurrent data structures.
	SpinWaitLevelPreempting
)

// ParamSpinWait is a type that represents a parameterized version of SpinWait.
// It is used for lightweight synchronization in low-level scenarios with lower cost.
// The zero value of ParamSpinWait is already initialized and ready to use.
type ParamSpinWait struct {
	_ noCopy

	i     uint32
	Level int8
	limit uint32
	total int32
}

// NewParamSpinWait returns a new instance of ParamSpinWait with default values.
func NewParamSpinWait() ParamSpinWait {
	return ParamSpinWait{}
}

// NewSpinWaitWithLevel returns a new instance of ParamSpinWait with the specified level.
// The level determines the preemptive behavior of the spin wait, where higher levels result in less delay but more CPU consumption.
func NewSpinWaitWithLevel(level int8) ParamSpinWait {
	return ParamSpinWait{Level: level}
}

// SetLevel sets the preemptive Level of the ParamSpinWait to the specified value.
// Higher preemptive level result in less delay, but more CPU consumption.
func (sw *ParamSpinWait) SetLevel(level int) *ParamSpinWait {
	if level < SpinWaitLevelBlocking {
		sw.Level = SpinWaitLevelBlocking
		return sw
	}
	if level > SpinWaitLevelPreempting {
		level = SpinWaitLevelPreempting
	}
	sw.Level = int8(level)

	return sw
}

// SetLimit sets the limit for the number of times the spin waite can spin before yielding.
func (sw *ParamSpinWait) SetLimit(limit int) *ParamSpinWait {
	if limit > math.MaxUint32-1 {
		limit = math.MaxUint32 - 1
	}
	sw.limit = uint32(limit)

	return sw
}

// Once performs a single spin operation based on the provided preemptive level.
func (sw *ParamSpinWait) Once() {
	sw.once(sw.Level)
}

// OnceWithLevel performs a single spin with the given preemptive level.
func (sw *ParamSpinWait) OnceWithLevel(level int) {
	if level < SpinWaitLevelBlocking {
		level = SpinWaitLevelBlocking
	}
	if level > SpinWaitLevelPreempting {
		level = SpinWaitLevelPreempting
	}
	sw.once(int8(level))
}

// WillYield returns true if the spin wait should yield the processor,
func (sw *ParamSpinWait) WillYield() bool {
	return sw.willYield(sw.Level)
}

// WillYieldWithLevel returns true if the spin wait should yield the processor based on the provided preemptive level.
func (sw *ParamSpinWait) WillYieldWithLevel(level int8) bool {
	return sw.willYield(level)
}

// Reset resets the internal counters of the ParamSpinWait instance to their initial values.
func (sw *ParamSpinWait) Reset() {
	sw.i = 0
	sw.total = 0
}

// Closed returns true if the limit is greater than 0 and the current count is equal or greater than the limit.
func (sw *ParamSpinWait) Closed() bool {
	return sw.limit > 0 && sw.i >= sw.limit
}

func (sw *ParamSpinWait) willYield(level int8) bool {
	x := int32(level << 2)
	if sw.i&(1<<(x-min(x, sw.total>>1))-1) != 0 {
		return false
	}

	return true
}

func (sw *ParamSpinWait) once(level int8) {
	sw.i++
	if !sw.willYield(level) {
		procyield(procYieldCycles)
		return
	}
	sw.total++
	if level <= SpinWaitLevelBlocking {
		time.Sleep(spinWaitDurationBlocking)
	} else if level <= SpinWaitLevelPending {
		time.Sleep(spinWaitDurationPending)
	}
	runtime.Gosched()
}

//go:linkname procyield runtime.procyield
func procyield(cycles uint32)
