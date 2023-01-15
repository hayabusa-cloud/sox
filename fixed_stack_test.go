// ©Hayabusa Cloud Co., Ltd. 2023. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package sox_test

import (
	"hybscloud.com/sox"
	"io"
	"math"
	"sync"
	"testing"
)

func TestFixedStack_Series(t *testing.T) {
	t.Run("a little serial ops", func(t *testing.T) {
		s, err := sox.NewFixedStack[uintptr](func(options *sox.FixedStackOptions) {
			options.Capacity = 0x3
			options.Concurrent = false
			options.Nonblocking = true
		})
		if err != nil {
			t.Errorf("fixed stack new: %v", err)
			return
		}
		testFixedStackNonblocking(t, s)
	})

	t.Run("parallel produce and consume", func(t *testing.T) {
		s, err := sox.NewFixedStack[int](func(options *sox.FixedStackOptions) {
			options.Concurrent = false
			options.Nonblocking = false
		})
		if err != nil {
			t.Errorf("fixed stack new: %v", err)
			return
		}
		go func() {
			for i := 0; i < (1 << 20); i++ {
				err := s.Push(i)
				if err != nil {
					t.Errorf("fixed stack push: %v", err)
					break
				}
			}
			err := s.Close()
			if err != nil {
				t.Errorf("fixed stack close: %v", err)
				return
			}
		}()
		stack, index, next := make([]int, 0), make([]int, 1<<20), 0
		for i := 0; i < (1 << 20); i++ {
			index[i] = -1
		}
		for i := 0; i < (1 << 20); i++ {
			item, err := s.Pop()
			if err != nil {
				t.Errorf("fixed stack pop: %v", err)
				return
			}
			for j := next; j <= item; j++ {
				if index[j] >= 0 && stack[len(stack)-1] != j {
					t.Errorf("fixed stack bad result")
					return
				}
				index[j] = len(stack)
				stack = append(stack, j)
			}
			if next < item+1 {
				next = item + 1
			}
			stack = stack[:len(stack)-1]
		}
		_, err = s.Pop()
		if err != nil && err != io.EOF {
			t.Errorf("fixed stack pop: %v", err)
			return
		}
	})
}

func BenchmarkFixedStack_Parallel(b *testing.B) {
	s, err := sox.NewFixedStack[int](func(options *sox.FixedStackOptions) {
		options.Concurrent = false
		options.Nonblocking = false
	})
	if err != nil {
		b.Errorf("fixed stack new: %v", err)
		return
	}
	b.ResetTimer()
	go func() {
		for i := 0; i < b.N; i++ {
			err := s.Push(i)
			if err != nil {
				b.Errorf("fixed stack push: %v", err)
				break
			}
		}
		err := s.Close()
		if err != nil {
			b.Errorf("fixed stack close: %v", err)
			return
		}
	}()
	for i := 0; i < b.N; i++ {
		_, err = s.Pop()
		if err != nil {
			b.Errorf("fixed stack pop: %v", err)
			return
		}
	}
}

func TestFixedStack_Concurrent(t *testing.T) {
	t.Run("a little ops", func(t *testing.T) {
		s, err := sox.NewFixedStack[uintptr](func(options *sox.FixedStackOptions) {
			options.Capacity = 0x3
			options.Concurrent = true
			options.Nonblocking = true
		})
		if err != nil {
			t.Errorf("fixed stack new: %v", err)
			return
		}
		testFixedStackNonblocking(t, s)
	})

	t.Run("4 push goroutines 4 pop goroutines 32k buffer", func(t *testing.T) {
		s, err := sox.NewFixedStack[int64](func(options *sox.FixedStackOptions) {
			options.Concurrent = true
			options.Nonblocking = false
		})
		if err != nil {
			t.Errorf("fixed stack new: %v", err)
			return
		}
		testFixedStackConcurrent(t, s, 0x04, 0x2000)
	})

	t.Run("16 push goroutines 16 pop goroutines 32k buffer", func(t *testing.T) {
		s, err := sox.NewFixedStack[int64](func(options *sox.FixedStackOptions) {
			options.Concurrent = true
			options.Nonblocking = false
		})
		if err != nil {
			t.Errorf("fixed stack new: %v", err)
			return
		}
		testFixedStackConcurrent(t, s, 0x10, 0x2000)
	})

	t.Run("64 push goroutines 64 pop goroutines 32k buffer", func(t *testing.T) {
		s, err := sox.NewFixedStack[int64](func(options *sox.FixedStackOptions) {
			options.Concurrent = true
			options.Nonblocking = false
		})
		if err != nil {
			t.Errorf("fixed stack new: %v", err)
			return
		}
		testFixedStackConcurrent(t, s, 0x40, 0x2000)
	})

}

func BenchmarkFixedStackConcurrent(b *testing.B) {
	b.Run("1 push goroutine 1 pop goroutine", func(b *testing.B) {
		s, err := sox.NewFixedStack[int](func(options *sox.FixedStackOptions) {
			options.Concurrent = true
			options.Nonblocking = false
		})
		if err != nil {
			b.Errorf("fixed stack new: %v", err)
			return
		}
		b.ResetTimer()
		benchmarkFixedStackConcurrent(b, s, 1)
	})

	b.Run("4 push goroutines 4 pop goroutines", func(b *testing.B) {
		s, err := sox.NewFixedStack[int](func(options *sox.FixedStackOptions) {
			options.Concurrent = true
			options.Nonblocking = false
		})
		if err != nil {
			b.Errorf("fixed stack new: %v", err)
			return
		}
		b.ResetTimer()
		benchmarkFixedStackConcurrent(b, s, 4)
	})

	b.Run("16 push goroutines 16 pop goroutines", func(b *testing.B) {
		s, err := sox.NewFixedStack[int](func(options *sox.FixedStackOptions) {
			options.Concurrent = true
			options.Nonblocking = false
		})
		if err != nil {
			b.Errorf("fixed stack new: %v", err)
			return
		}
		b.ResetTimer()
		benchmarkFixedStackConcurrent(b, s, 16)
	})

	b.Run("64 push goroutines 64 pop goroutines", func(b *testing.B) {
		s, err := sox.NewFixedStack[int](func(options *sox.FixedStackOptions) {
			options.Concurrent = true
			options.Nonblocking = false
		})
		if err != nil {
			b.Errorf("fixed stack new: %v", err)
			return
		}
		b.ResetTimer()
		benchmarkFixedStackConcurrent(b, s, 64)
	})
}

func testFixedStackNonblocking(t *testing.T, s sox.Stack[uintptr]) {
	item, err := s.Pop()
	if err != sox.ErrTemporarilyUnavailable {
		t.Errorf("fixed stack pop expected ErrTemporarilyUnavailable but got %v %v", item, err)
		return
	}
	err = s.Push(1)
	if err != nil {
		t.Errorf("fixed stack push: %v", err)
		return
	}
	item, err = s.Pop()
	if err != nil {
		t.Errorf("fixed stack pop: %v", err)
		return
	}
	if item != 1 {
		t.Errorf("fixed stack pop item expected %d but got %d", 1, item)
		return
	}
	item, err = s.Pop()
	if err != sox.ErrTemporarilyUnavailable {
		t.Errorf("fixed stack expected ErrTemporarilyUnavailable but got %v %v", item, err)
		return
	}
	err = s.Push(2)
	if err != nil {
		t.Errorf("fixed stack push: %v", err)
		return
	}
	err = s.Push(3)
	if err != nil {
		t.Errorf("fixed stack push: %v", err)
		return
	}
	err = s.Push(4)
	if err != nil {
		t.Errorf("fixed stack push: %v", err)
		return
	}
	item, err = s.Pop()
	if err != nil {
		t.Errorf("fixed stack pop: %v", err)
		return
	}
	if item != 4 {
		t.Errorf("fixed stack pop item expected %d but got %d", 4, item)
		return
	}
	err = s.Push(5)
	if err != nil {
		t.Errorf("fixed stack push: %v", err)
		return
	}
	err = s.Push(6)
	if err != sox.ErrTemporarilyUnavailable {
		t.Errorf("fixed stack expected ErrTemporarilyUnavailable but got %v", err)
		return
	}
	item, err = s.Pop()
	if err != nil {
		t.Errorf("fixed stack pop: %v", err)
		return
	}
	if item != 5 {
		t.Errorf("fixed stack pop item expected %d but got %d", 5, item)
		return
	}
	item, err = s.Pop()
	if err != nil {
		t.Errorf("fixed stack pop: %v", err)
		return
	}
	if item != 3 {
		t.Errorf("fixed stack pop item expected %d but got %d", 3, item)
		return
	}
	err = s.Push(7)
	if err != nil {
		t.Errorf("fixed stack push: %v", err)
		return
	}
	item, err = s.Pop()
	if err != nil {
		t.Errorf("fixed stack pop: %v", err)
		return
	}
	if item != 7 {
		t.Errorf("fixed stack pop item expected %d but got %d", 7, item)
		return
	}
	item, err = s.Pop()
	if err != nil {
		t.Errorf("fixed stack pop: %v", err)
		return
	}
	if item != 2 {
		t.Errorf("fixed stack pop item expected %d but got %d", 2, item)
		return
	}
	item, err = s.Pop()
	if err != sox.ErrTemporarilyUnavailable {
		t.Errorf("fixed stack pop expected %v but got %v %v", sox.ErrTemporarilyUnavailable, item, err)
		return
	}
	err = s.Close()
	if err != nil {
		t.Errorf("fixed stack pop: %v", err)
		return
	}
	item, err = s.Pop()
	if err != io.EOF {
		t.Errorf("fixed stack pop expected %v but got %v %v", io.EOF, item, err)
		return
	}
}

func testFixedStackConcurrent(t *testing.T, s sox.Stack[int64], m int, n int) {
	stack, index, next := make([][]int64, m), make([][]int, m), make([]int64, m)
	locks := make([]sync.Mutex, m)
	for i := 0; i < m; i++ {
		stack[i] = make([]int64, 0, n+1)
		index[i] = make([]int, n)
		for j := 0; j < n; j++ {
			index[i][j] = -1
		}
		next[i] = 0
	}
	wg := sync.WaitGroup{}
	for h := 0; h < m; h++ {
		wg.Add(1)
		go func() {
			for i := 0; i < n; i++ {
				item, err := s.Pop()
				if err != nil {
					t.Errorf("fixed stack pop: %v", err)
					return
				}
				high, low := item>>32, item&math.MaxUint32
				locks[high].Lock()
				st, idx, val := stack[high], index[high], low
				for j := next[high]; j <= val; j++ {
					if idx[j] >= 0 && st[len(st)-1] != j {
						t.Errorf("fixed stack bad result")
						return
					}
					idx[j] = len(st)
					st = append(st, j)
				}
				if next[high] < val+1 {
					next[high] = val + 1
				}
				st = st[:len(st)-1]
				stack[high] = st
				locks[high].Unlock()
			}
			wg.Done()
		}()
	}
	for i := 0; i < m; i++ {
		go func(i int) {
			for j := 0; j < n; j++ {
				err := s.Push(int64(i<<32) | int64(j))
				if err != nil {
					t.Errorf("fixed stack push: %v", err)
					return
				}
			}
		}(i)
	}
	wg.Wait()
}

func benchmarkFixedStackConcurrent(b *testing.B, s sox.Stack[int], m int) {
	for i := 0; i < m; i++ {
		go func(i int) {
			for j := 0; j < b.N/m+1; j++ {
				err := s.Push(j)
				if err != nil {
					b.Errorf("fixed stack push: %v", err)
					return
				}
			}
		}(i)
	}
	wg := sync.WaitGroup{}
	for i := 0; i < m; i++ {
		wg.Add(1)
		go func(i int) {
			for j := 0; j < b.N/m; j++ {
				_, err := s.Pop()
				if err != nil {
					b.Errorf("fixed stack pop: %v", err)
					return
				}
			}
			wg.Done()
		}(i)
	}
	wg.Wait()
}
