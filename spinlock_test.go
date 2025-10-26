// ©Hayabusa Cloud Co., Ltd. 2024. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package sox_test

import (
	"code.hybscloud.com/sox"
	"context"
	"sync"
	"testing"
)

func TestSpinlock_TryLock(t *testing.T) {
	tests := []struct {
		name     string
		setup    func(*sox.Spinlock)
		expected bool
	}{
		{
			name:     "Unlocked lock should be lockable",
			setup:    func(sl *sox.Spinlock) {},
			expected: true,
		},
		{
			name: "Locked lock should not be lockable",
			setup: func(sl *sox.Spinlock) {
				sl.TryLock()
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sl := sox.Spinlock{}
			tt.setup(&sl)
			res := sl.TryLock()
			if res != tt.expected {
				t.Errorf("TryLock() = %v; want %v", res, tt.expected)
			}
		})
	}
}

func TestSpinlock_Lock(t *testing.T) {
	wg := sync.WaitGroup{}
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			sl := sox.Spinlock{}
			sl.Lock()
			locked := sl.TryLock()
			if locked {
				t.Errorf("Lock() failed to lock; TryLock() = true; want false")
			}
			wg.Done()
		}()
	}
	wg.Wait()
}

func TestSpinlock_Unlock(t *testing.T) {
	sl := sox.Spinlock{}
	sl.Lock()
	sl.Unlock()
	if !sl.TryLock() {
		t.Errorf("Unlock() failed; TryLock() = false; want true")
	}
}

func TestSpinlock_SpinOnce(t *testing.T) {
	sl := sox.Spinlock{}
	done := make(chan bool)

	if !sl.TryLock() {
		t.Error("Expected lock to succeed")
	}

	go func() {
		sl.SpinOnce()
		done <- true
	}()

	sl.Unlock()

	_, ok := <-done
	if !ok {
		t.Error("Expected SpinOnce to unblock")
	}
}

func BenchmarkSpinlock(b *testing.B) {
	b.Run("parallel 1", func(b *testing.B) {
		benchmarkSpinlock(b, 1)
	})
	b.Run("parallel 4", func(b *testing.B) {
		benchmarkSpinlock(b, 4)
	})
	b.Run("parallel 16", func(b *testing.B) {
		benchmarkSpinlock(b, 16)
	})
	b.Run("parallel 64", func(b *testing.B) {
		benchmarkSpinlock(b, 64)
	})
}

func benchmarkSpinlock(b *testing.B, parallel int) {
	sl := sox.Spinlock{}
	sl.Lock()
	wg := sync.WaitGroup{}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go func(ctx context.Context) {
		for {
			select {
			case <-ctx.Done():
				return
			default:
				sl.Unlock()
			}
		}
	}(ctx)
	b.ResetTimer()
	for range parallel {
		wg.Add(1)
		go func() {
			for range b.N/parallel + 1 {
				sl.Lock()
			}
			wg.Done()
		}()
	}
	wg.Wait()
	b.StopTimer()
}
