// ©Hayabusa Cloud Co., Ltd. 2023. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package sox_test

import (
	"code.hybscloud.com/sox"
	"math"
	"runtime"
	"sync/atomic"
	"testing"
	"time"
)

func TestParamSpinWait(t *testing.T) {
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
				time.Sleep(time.Millisecond)
			}
		}
	}
	t.Run("common usage", func(t *testing.T) {
		x := atomic.Int32{}
		go fn(&x)
		for sw := sox.NewParamSpinWait(); !sw.Closed(); sw.Once() {
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
		sw := sox.NewSpinWaitWithLevel(sox.SpinWaitLevelBlocking)
		total := 0
		for range 1 << 4 {
			if sw.WillYield() {
				total++
			}
			sw.Once()
		}
		if total != 1<<4 {
			t.Errorf("expected total wait %d but got %d", 1<<4, total)
		}
	})

	t.Run("level 1", func(t *testing.T) {
		sw := sox.NewSpinWaitWithLevel(sox.SpinWaitLevelPending)
		total := 0
		for range 1 << 6 {
			if sw.WillYield() {
				total++
			}
			sw.Once()
		}
		if 1<<3 > total && total <= 1<<4 {
			t.Errorf("expected total wait between %d and %d but got %d", 1<<3, 1<<4, total)
		}
	})

	t.Run("level 2", func(t *testing.T) {
		sw := sox.NewSpinWaitWithLevel(sox.SpinWaitLevelPreempting)
		total := 0
		for range 1 << 8 {
			if sw.WillYield() {
				total++
			}
			sw.Once()
		}
		if total > 1 {
			t.Errorf("expected total wait<=%d but got %d", 1, total)
		}
	})

	t.Run("timeout", func(t *testing.T) {
		sw := sox.NewParamSpinWait()
		sw.SetLimit(128)
		cnt := 0
		for ; !sw.Closed(); sw.Once() {
			cnt++
		}
		if cnt != 128 {
			t.Errorf("expected spin waiter wait %d times but executed %d times", 128, cnt)
		}
	})
}
