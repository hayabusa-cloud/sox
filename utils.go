// ©Hayabusa Cloud Co., Ltd. 2024. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package sox

import (
	"runtime"
	"time"
)

type noCopy struct{}

type Empty = struct{}

// Yield pauses the current goroutine, allowing other goroutines to run,
// or optionally sleeps for a specific duration.
func Yield(ticks ...int) {
	runtime.Gosched()
	d := jiffy
	if len(ticks) > 0 {
		d = time.Duration(ticks[0]) * jiffy
	}
	if d > 0 {
		time.Sleep(d)
	}
}

// MemoryBarrier provides a memory barrier operation to ensure ordering of memory accesses.
func MemoryBarrier() {
	memoryBarrier()
}
