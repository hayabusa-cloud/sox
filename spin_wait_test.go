// ©Hayabusa Cloud Co., Ltd. 2023. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package sox_test

import (
	"hybscloud.com/sox"
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
		sw := sox.NewParamSpinWait().SetLevel(sox.SpinWaitLevelClient)
		total := 0
		for i := 0; i < 1<<4; i++ {
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
		sw := sox.NewParamSpinWait().SetLevel(sox.SpinWaitLevelBlockingIO)
		total := 0
		for i := 0; i < 1<<5; i++ {
			if sw.WillYield() {
				total++
			}
			sw.Once()
		}
		if total <= 1<<4 {
			t.Errorf("expected total wait>%d but got %d", 1<<4, total)
		}
	})

	t.Run("level 2", func(t *testing.T) {
		sw := sox.NewParamSpinWait().SetLevel(sox.SpinWaitLevelConsume)
		total := 0
		for i := 0; i < 1<<7; i++ {
			if sw.WillYield() {
				total++
			}
			sw.Once()
		}
		if total <= 1<<6 {
			t.Errorf("expected total wait>%d but got %d", 1<<6, total)
		}
	})

	t.Run("level 3", func(t *testing.T) {
		sw := sox.NewParamSpinWait().SetLevel(sox.SpinWaitLevelConsume + 1)
		total := 0
		for i := 0; i < 1<<10; i++ {
			if sw.WillYield() {
				total++
			}
			sw.Once()
		}
		if total <= 1<<9 || total >= 1<<10 {
			t.Errorf("expected total wait between %d and %d but got %d", 1<<9, 1<<10, total)
		}
	})

	t.Run("level 4", func(t *testing.T) {
		sw := sox.NewParamSpinWait().SetLevel(sox.SpinWaitLevelConsume + 2)
		total := 0
		for i := 0; i < 1<<14; i++ {
			if sw.WillYield() {
				total++
			}
			sw.Once()
		}
		if total <= 1<<13 || total >= 1<<14 {
			t.Errorf("expected total wait between %d and %d but got %d", 1<<13, 1<<14, total)
		}
	})

	t.Run("timeout", func(t *testing.T) {
		sw := sox.NewParamSpinWait().SetLimit(128)
		cnt := 0
		for ; !sw.Closed(); sw.Once() {
			cnt++
		}
		if cnt != 128 {
			t.Errorf("expected spin waiter wait %d times but executed %d times", 128, cnt)
		}
	})
}
