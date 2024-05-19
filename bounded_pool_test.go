// ©Hayabusa Cloud Co., Ltd. 2024. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package sox_test

import (
	"bytes"
	"golang.org/x/sys/cpu"
	"hybscloud.com/sox"
	"runtime"
	"sync"
	"sync/atomic"
	"testing"
	"unsafe"
)

func TestBoundedPool_Block(t *testing.T) {
	t.Run("simple usage", func(t *testing.T) {
		pool := sox.NewBoundedPool[sox.Empty](sox.BoundedPoolPicoCapacity)
		pool.Fill(func() sox.Empty { return sox.Empty{} })
		c := pool.Cap()
		remapM := unsafe.Sizeof(cpu.CacheLinePad{}) / unsafe.Sizeof(atomic.Uint64{})
		remapN := max(1, uintptr(c)/remapM)
		remap := func(cursor uint32) int {
			p, q := uintptr(cursor)/remapN, uintptr(cursor)%remapN
			return int(q*remapM + p%remapM)
		}
		go func() {
			for i := range c {
				err := pool.Put(remap(uint32(c + i)))
				if err != nil {
					t.Errorf("put item back to pool: %v", err)
					return
				}
			}
		}()
		for i := range c * 2 {
			indirect, err := pool.Get()
			if err != nil {
				t.Errorf("get item from pool: %v", err)
				return
			}
			expected := remap(uint32(i))
			if expected != indirect {
				t.Errorf("expected indirect %v but got %v", expected, indirect)
				return
			}
		}
	})
	t.Run("concurrent pico", func(t *testing.T) {
		pool := sox.NewBoundedPool[sox.Empty](sox.BoundedPoolPicoCapacity)
		pool.Fill(func() sox.Empty { return sox.Empty{} })
		testBoundedPoolBlock(t, pool, 8, 4)
	})
	t.Run("concurrent nano", func(t *testing.T) {
		pool := sox.NewBoundedPool[sox.Empty](sox.BoundedPoolNanoCapacity)
		pool.Fill(func() sox.Empty { return sox.Empty{} })
		testBoundedPoolBlock(t, pool, 8, 4)
	})
	t.Run("concurrent micro", func(t *testing.T) {
		pool := sox.NewBoundedPool[sox.Empty](sox.BoundedPoolMicroCapacity)
		pool.Fill(func() sox.Empty { return sox.Empty{} })
		testBoundedPoolBlock(t, pool, 8, 4)
	})
}

func TestBoundedPool_Nonblock(t *testing.T) {
	pool := sox.NewBoundedPool[sox.Empty](sox.BoundedPoolPicoCapacity)
	pool.SetNonblock(true)
	pool.Fill(func() sox.Empty { return sox.Empty{} })
	indirect, err := pool.Get()
	if err != nil {
		t.Errorf("getting item from pool: %v", err)
		return
	}
	if indirect != 0 {
		t.Errorf("expected indirect %v but got %v", 0, indirect)
	}
	err = pool.Put(indirect)
	if err != nil {
		t.Errorf("put item back to pool: %v", err)
		return
	}
	err = pool.Put(0)
	if err != sox.ErrTemporarilyUnavailable {
		t.Errorf("expected temporarily unavailable error but got %v", err)
		return
	}
	vals := make([]int, 0)
	for range pool.Cap() {
		indirect, err = pool.Get()
		if err != nil {
			t.Errorf("getting item from pool: %v", err)
			return
		}
		vals = append(vals, indirect)
	}
	_, err = pool.Get()
	if err != sox.ErrTemporarilyUnavailable {
		t.Errorf("expected temporarily unavailable error but got %v", err)
		return
	}
	for _, val := range vals {
		err = pool.Put(val)
		if err != nil {
			t.Errorf("put item back to pool: %v", err)
			return
		}
	}
	err = pool.Put(0)
	if err != sox.ErrTemporarilyUnavailable {
		t.Errorf("expected temporarily unavailable error but got %v", err)
		return
	}
}

func TestBoundedPool_Value_SetValue(t *testing.T) {
	t.Run("int pool", func(t *testing.T) {
		intPool := sox.NewBoundedPool[int](1)
		intPool.Fill(func() int { return 0 })
		i, err := intPool.Get()
		if err != nil {
			t.Errorf("get item from pool: %v", err)
			return
		}
		intPool.SetValue(i, 10)
		err = intPool.Put(i)
		if err != nil {
			t.Errorf("put item into pool: %v", err)
			return
		}
		i, err = intPool.Get()
		if err != nil {
			t.Errorf("get item from pool: %v", err)
			return
		}
		val := intPool.Value(i)
		if val != 10 {
			t.Errorf("unexpected value: %v", val)
			return
		}
	})
	t.Run("float pool", func(t *testing.T) {
		floatPool := sox.NewBoundedPool[*float64](1)
		floatPool.Fill(func() *float64 {
			return new(float64)
		})
		f, err := floatPool.Get()
		if err != nil {
			t.Errorf("get item from pool: %v", err)
			return
		}
		*floatPool.Value(f) = 3.14
		err = floatPool.Put(f)
		if err != nil {
			t.Errorf("put item into pool: %v", err)
			return
		}
		f, err = floatPool.Get()
		if err != nil {
			t.Errorf("get item from pool: %v", err)
			return
		}
		val := floatPool.Value(f)
		if *val != 3.14 {
			t.Errorf("unexpected value: %v", *val)
			return
		}
	})
	t.Run("string pool", func(t *testing.T) {
		stringPool := sox.NewBoundedPool[string](1)
		stringPool.Fill(func() string { return "initial" })
		s, err := stringPool.Get()
		if err != nil {
			t.Errorf("get item from pool: %v", err)
			return
		}
		stringPool.SetValue(s, "updated")
		err = stringPool.Put(s)
		if err != nil {
			t.Errorf("put item into pool: %v", err)
			return
		}
		s, err = stringPool.Get()
		if err != nil {
			t.Errorf("get item from pool: %v", err)
			return
		}
		val := stringPool.Value(s)
		if val != "updated" {
			t.Errorf("unexpected value: %v", val)
			return
		}
	})
	t.Run("pico buffer pool", func(t *testing.T) {
		picoBufferPool := sox.NewBoundedPool[sox.PicoBuffer](1)
		picoBufferPool.Fill(func() sox.PicoBuffer { return sox.PicoBuffer{} })
		b, err := picoBufferPool.Get()
		if err != nil {
			t.Errorf("get buffer from pool: %v", err)
			return
		}
		picoBufferPool.SetValue(b, sox.PicoBuffer{1, 2, 3, 4, 5})
		err = picoBufferPool.Put(b)
		if err != nil {
			t.Errorf("put buffer into pool: %v", err)
			return
		}
		b, err = picoBufferPool.Get()
		if err != nil {
			t.Errorf("get buffer from pool: %v", err)
			return
		}
		val := picoBufferPool.Value(b)
		if !bytes.Equal(val[:5], []byte{1, 2, 3, 4, 5}) {
			t.Errorf("unexpected value: %v", val)
		}
	})
}

func BenchmarkBoundedPool(b *testing.B) {
	b.Run("pico capacity", func(b *testing.B) {
		pool := sox.NewBoundedPool[sox.Empty](sox.BoundedPoolPicoCapacity)
		pool.Fill(func() sox.Empty { return sox.Empty{} })
		benchmarkBoundedPool(b, pool)
	})

	b.Run("nano capacity", func(b *testing.B) {
		pool := sox.NewBoundedPool[sox.Empty](sox.BoundedPoolNanoCapacity)
		pool.Fill(func() sox.Empty { return sox.Empty{} })
		benchmarkBoundedPool(b, pool)
	})

	b.Run("micro capacity", func(b *testing.B) {
		pool := sox.NewBoundedPool[sox.Empty](sox.BoundedPoolMicroCapacity)
		pool.Fill(func() sox.Empty { return sox.Empty{} })
		benchmarkBoundedPool(b, pool)
	})

	b.Run("small capacity", func(b *testing.B) {
		pool := sox.NewBoundedPool[sox.Empty](sox.BoundedPoolSmallCapacity)
		pool.Fill(func() sox.Empty { return sox.Empty{} })
		benchmarkBoundedPool(b, pool)
	})

	b.Run("medium capacity", func(b *testing.B) {
		pool := sox.NewBoundedPool[sox.Empty](sox.BoundedPoolMediumCapacity)
		pool.Fill(func() sox.Empty { return sox.Empty{} })
		benchmarkBoundedPool(b, pool)
	})

	b.Run("large capacity", func(b *testing.B) {
		pool := sox.NewBoundedPool[sox.Empty](sox.BoundedPoolLargeCapacity)
		pool.Fill(func() sox.Empty { return sox.Empty{} })
		benchmarkBoundedPool(b, pool)
	})

	b.Run("huge capacity", func(b *testing.B) {
		pool := sox.NewBoundedPool[sox.Empty](sox.BoundedPoolHugeCapacity)
		pool.Fill(func() sox.Empty { return sox.Empty{} })
		benchmarkBoundedPool(b, pool)
	})

	b.Run("giant capacity", func(b *testing.B) {
		pool := sox.NewBoundedPool[sox.Empty](sox.BoundedPoolGiantCapacity)
		pool.Fill(func() sox.Empty { return sox.Empty{} })
		benchmarkBoundedPool(b, pool)
	})
}

func testBoundedPoolBlock[T sox.BoundedPoolItem](t *testing.T, pool *sox.BoundedPool[T], parallel, times int) {
	n := pool.Cap()
	indirect, err := pool.Get()
	if err != nil {
		t.Errorf("get item from pool: %v", err)
		return
	}
	if indirect != 0 {
		t.Errorf("expected indirect %v but got %v", 0, indirect)
		return
	}
	cnt := make([]atomic.Int32, n)
	wg := sync.WaitGroup{}
	for p := range parallel {
		wg.Add(1)
		go func(p int) {
			defer wg.Done()
			for range n * times {
				indirect, err := pool.Get()
				if err != nil {
					t.Errorf("get item from pool: %v", err)
					return
				}
				if cnt[indirect].Add(1) > 1 {
					t.Errorf("get same indirect %v simultaneously", indirect)
				}
				runtime.Gosched()
				err = pool.Put(indirect)
				cnt[indirect].Add(-1)
				if err != nil {
					t.Errorf("put item back to pool: %v", err)
					return
				}
			}
		}(p)
	}
	wg.Wait()
}

func benchmarkBoundedPool[T sox.BoundedPoolItem](b *testing.B, pool *sox.BoundedPool[T]) {
	idxCh := make(chan int)
	b.ResetTimer()
	go func() {
		for range b.N/2 + 1 {
			indirect, err := pool.Get()
			if err != nil {
				b.Fail()
				break
			}
			idxCh <- indirect
		}
	}()
	for range b.N/2 + 1 {
		indirect := <-idxCh
		err := pool.Put(indirect)
		if err != nil {
			b.Fail()
			break
		}
	}
}
