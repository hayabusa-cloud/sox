// ©Hayabusa Cloud Co., Ltd. 2024. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package sox

import (
	"sync/atomic"
)

// Spinlock is a synchronization type that provides a simple spinlock mechanism.
// The zero value for Spinlock is ready to use.
type Spinlock struct {
	_ noCopy

	sw SpinWait
	v  atomic.Uintptr
}

// TryLock attempts to acquire the Spinlock and returns true if successful,
// otherwise it returns false.
func (sl *Spinlock) TryLock() bool {
	v := sl.v.Add(1)

	return v == 1
}

// Lock attempts to acquire the spinlock. If it fails, it will spin wait until it succeeds.
func (sl *Spinlock) Lock() {
	for {
		if sl.TryLock() {
			break
		}
		sl.sw.Once()
	}
}

// Unlock releases the spinlock.
func (sl *Spinlock) Unlock() {
	sl.v.Swap(0)
}

// SpinOnce performs a single spin using the SpinWait synchronization type.
func (sl *Spinlock) SpinOnce() {
	sl.sw.Once()
}
