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
	procYieldCycles = 16
)

// SpinWait is a lightweight synchronization type that
// you can use in low-level scenarios with lower cost.
// The zero value for SpinWait is ready to use.
type SpinWait struct {
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
	SpinWaitLevelClient = iota
	SpinWaitLevelBlockingIO
	SpinWaitLevelConsume
	spinWaitLevelProduce
	spinWaitLevelAtomic
)

type ParamSpinWait struct {
	i     uint32
	level int8
	d     time.Duration
	limit uint32
	total int32
}

func NewParamSpinWait() *ParamSpinWait {
	return &ParamSpinWait{
		i:     0,
		level: SpinWaitLevelBlockingIO,
		d:     jiffies,
		limit: 0,
		total: 0,
	}
}

func (sw *ParamSpinWait) SetLevel(level int) *ParamSpinWait {
	if level < SpinWaitLevelClient {
		sw.level = SpinWaitLevelClient
		return sw
	}
	if level > spinWaitLevelAtomic {
		level = spinWaitLevelAtomic
	}
	sw.level = int8(level)

	return sw
}

func (sw *ParamSpinWait) SetLimit(limit int) *ParamSpinWait {
	if limit > math.MaxUint32-1 {
		limit = math.MaxUint32 - 1
	}
	sw.limit = uint32(limit)

	return sw
}

func (sw *ParamSpinWait) Once() {
	sw.once(sw.level)
}

func (sw *ParamSpinWait) OnceWithLevel(level int) {
	if level < SpinWaitLevelClient {
		level = SpinWaitLevelClient
	}
	if level > spinWaitLevelAtomic {
		level = spinWaitLevelAtomic
	}
	sw.once(int8(level))
}

func (sw *ParamSpinWait) WillYield() bool {
	return sw.willYield(sw.level)
}

func (sw *ParamSpinWait) WillYieldWithLevel(level int8) bool {
	return sw.willYield(level)
}

func (sw *ParamSpinWait) Reset() {
	sw.i = 0
	sw.total = 0
}

func (sw *ParamSpinWait) Closed() bool {
	return sw.limit > 0 && sw.i >= sw.limit
}

func (sw *ParamSpinWait) willYield(level int8) bool {
	x := int32(level << 1)
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
	if level <= SpinWaitLevelBlockingIO {
		time.Sleep(jiffies)
	} else {
		runtime.Gosched()
	}
}

//go:linkname procyield runtime.procyield
func procyield(cycles uint32)
