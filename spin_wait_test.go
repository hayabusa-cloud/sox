// ©Hayabusa Cloud Co., Ltd. 2022. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package sox

import (
	"math"
	"runtime"
	"sync/atomic"
	"testing"
	"time"
)

func TestSpinWaiter(t *testing.T) {
	fn := func(x *atomic.Int32) {
		for {
			val := x.Load()
			if val < 0 {
				break
			}
			if val > math.MaxInt32/2 {
				val = math.MaxInt32 / 2
			}
			if !x.CompareAndSwap(val, val+1) {
				runtime.Gosched()
				continue
			}
			if val > math.MaxInt16 {
				time.Sleep(jiffies)
			}
		}
	}
	t.Run("common usage", func(t *testing.T) {
		x := atomic.Int32{}
		go fn(&x)
		for sw := NewSpinWait(); !sw.Closed(); sw.Once() {
			val := x.Load()
			//  some actions
			runtime.Gosched()

			if !x.CompareAndSwap(val, -1) {
				continue
			}
			break
		}
		res := x.Load()
		if res != -1 {
			t.Errorf("wronf value of x expected %d but got %d", -1, res)
			return
		}
	})

	t.Run("level 0", func(t *testing.T) {
		sw := NewSpinWait().SetLevel(SpinWaitLevelClient)
		for i := 0; i < 1<<4; i++ {
			sw.Once()
		}
		if sw.total != 1<<5 {
			t.Errorf("expected total wait %d but got %d", 1<<5, sw.total)
		}
	})

	t.Run("level 1", func(t *testing.T) {
		sw := NewSpinWait().SetLevel(SpinWaitLevelBlockingIO)
		for i := 0; i < 1<<5; i++ {
			sw.Once()
		}
		if sw.total != 1<<5 {
			t.Errorf("expected total wait %d but got %d", 1<<5, sw.total)
		}
	})

	t.Run("level 2", func(t *testing.T) {
		sw := NewSpinWait().SetLevel(SpinWaitLevelConsume)
		for i := 0; i < 1<<7; i++ {
			sw.Once()
		}
		if sw.total != 1<<6 {
			t.Errorf("expected total wait %d but got %d", 1<<6, sw.total)
		}
	})

	t.Run("level 3", func(t *testing.T) {
		sw := NewSpinWait().SetLevel(spinWaitLevelProduce)
		for i := 0; i < 1<<10; i++ {
			sw.Once()
		}
		if sw.total <= 1<<5 || sw.total >= 1<<6 {
			t.Errorf("expected total wait between %d and %d but got %d", 1<<5, 1<<6, sw.total)
		}
	})

	t.Run("level 4", func(t *testing.T) {
		sw := NewSpinWait().SetLevel(spinWaitLevelAtomic)
		for i := 0; i < 1<<16; i++ {
			sw.Once()
		}
		if sw.total != 1 {
			t.Errorf("expected total wait %d but got %d", 1, sw.total)
		}
	})

	t.Run("timeout", func(t *testing.T) {
		sw := NewSpinWait().SetLimit(128)
		for ; !sw.Closed(); sw.Once() {
		}
		if sw.i != 128 {
			t.Errorf("expected spin waiter wait %d times but executed %d times", 64, sw.i)
		}
	})
}
