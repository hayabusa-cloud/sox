// ©Hayabusa Cloud Co., Ltd. 2022. All rights reserved.
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

func TestNewRingQueue(t *testing.T) {
	t.Run("capacity 1", func(t *testing.T) {
		_, _, err := sox.NewRingQueue[uintptr](func(options *sox.RingQueueOptions) {
			options.Capacity = 1
			options.ConcurrentProduce = false
			options.ConcurrentConsume = false
		})
		if err != nil {
			t.Errorf("ring queue new: %v", err)
			return
		}
	})

	t.Run("capacity 0", func(t *testing.T) {
		c, p, err := sox.NewRingQueue[uintptr](func(options *sox.RingQueueOptions) {
			options.Capacity = 0
			options.ConcurrentProduce = false
			options.ConcurrentConsume = false
		})
		if c != nil || p != nil || err == nil {
			t.Errorf("ring queue new expected err but successed")
			return
		}
	})

	t.Run("capacity 65536", func(t *testing.T) {
		_, _, err := sox.NewRingQueue[uintptr](func(options *sox.RingQueueOptions) {
			options.Capacity = 1 << 16
			options.ConcurrentProduce = false
			options.ConcurrentConsume = false
		})
		if err != nil {
			t.Errorf("ring queue new: %v", err)
			return
		}
	})
}

func TestRingQueue_Series(t *testing.T) {
	t.Run("a little serial ops", func(t *testing.T) {
		c, p, err := sox.NewRingQueue[uintptr](func(options *sox.RingQueueOptions) {
			options.Capacity = 0x3
			options.ConcurrentProduce = false
			options.ConcurrentConsume = false
			options.Nonblocking = true
		})
		if err != nil {
			t.Errorf("ring queue new: %v", err)
			return
		}
		testRingQueueNonblocking(t, c, p)
	})

	t.Run("parallel produce and consume", func(t *testing.T) {
		c, p, err := sox.NewRingQueue[int](func(options *sox.RingQueueOptions) {
			options.ConcurrentProduce = false
			options.ConcurrentConsume = false
			options.Nonblocking = false
		})
		if err != nil {
			t.Errorf("ring queue new: %v", err)
			return
		}
		go func() {
			for i := range 1 << 20 {
				err := p.Produce(i)
				if err != nil {
					t.Errorf("ring producer produce: %v", err)
					break
				}
			}
			err := p.Close()
			if err != nil {
				t.Errorf("ring producer close: %v", err)
				return
			}
		}()
		for i := range (1 << 20) - 1 {
			item, err := c.Consume()
			if err != nil {
				t.Errorf("ring consumer consume: %v", err)
				return
			}
			if item != i {
				t.Errorf("ring consumed item expected %d but got %d", i, item)
				return
			}
		}
		item, err := c.Consume()
		if err != nil && err != io.EOF {
			t.Errorf("ring consumer consume: %v", err)
			return
		}
		if item != (1<<20)-1 {
			t.Errorf("ring consumed item expected %d but got %d", (1<<20)-1, item)
			return
		}
		item, err = c.Consume()
		if err != io.EOF {
			t.Errorf("ring consumer consume expected %v but got %v %v", io.EOF, item, err)
			return
		}
	})
}

func BenchmarkRingQueue_Parallel(b *testing.B) {
	c, p, err := sox.NewRingQueue[int](func(options *sox.RingQueueOptions) {
		options.ConcurrentProduce = false
		options.ConcurrentConsume = false
		options.Nonblocking = false
	})
	if err != nil {
		b.Errorf("ring queue new: %v", err)
		return
	}
	b.ResetTimer()
	go func() {
		for i := range b.N {
			err := p.Produce(i)
			if err != nil {
				b.Errorf("ring producer produce: %v", err)
				break
			}
		}
		err := p.Close()
		if err != nil {
			b.Errorf("ring producer close: %v", err)
			return
		}
	}()
	for range b.N {
		_, err = c.Consume()
		if err != nil {
			b.Errorf("ring consumer consume: %v", err)
			return
		}
	}
}

func TestRingQueue_ConcurrentProduce(t *testing.T) {
	t.Run("a little ops", func(t *testing.T) {
		c, p, err := sox.NewRingQueue[uintptr](func(options *sox.RingQueueOptions) {
			options.Capacity = 0x3
			options.ConcurrentProduce = true
			options.ConcurrentConsume = false
			options.Nonblocking = true
		})
		if err != nil {
			t.Errorf("ring queue new: %v", err)
			return
		}
		testRingQueueNonblocking(t, c, p)
	})

	t.Run("4 producer goroutines 32k buffer", func(t *testing.T) {
		c, p, err := sox.NewRingQueue[int64](func(options *sox.RingQueueOptions) {
			options.ConcurrentProduce = true
			options.ConcurrentConsume = false
			options.Nonblocking = false
		})
		if err != nil {
			t.Errorf("ring queue new: %v", err)
			return
		}
		testRingQueueConcurrentProduce(t, c, p, 0x04, 0x2000)
	})

	t.Run("16 producer goroutines 32k buffer", func(t *testing.T) {
		c, p, err := sox.NewRingQueue[int64](func(options *sox.RingQueueOptions) {
			options.ConcurrentProduce = true
			options.ConcurrentConsume = false
			options.Nonblocking = false
		})
		if err != nil {
			t.Errorf("ring queue new: %v", err)
			return
		}
		testRingQueueConcurrentProduce(t, c, p, 0x10, 0x2000)
	})

	t.Run("64 producer goroutines 32k buffer", func(t *testing.T) {
		c, p, err := sox.NewRingQueue[int64](func(options *sox.RingQueueOptions) {
			options.ConcurrentProduce = true
			options.ConcurrentConsume = false
			options.Nonblocking = false
		})
		if err != nil {
			t.Errorf("ring queue new: %v", err)
			return
		}
		testRingQueueConcurrentProduce(t, c, p, 0x40, 0x2000)
	})
}

func BenchmarkRingQueue_ConcurrentProduce(b *testing.B) {
	b.Run("1 producer", func(b *testing.B) {
		c, p, err := sox.NewRingQueue[int](func(options *sox.RingQueueOptions) {
			options.ConcurrentProduce = true
			options.ConcurrentConsume = false
			options.Nonblocking = false
		})
		if err != nil {
			b.Errorf("ring queue new: %v", err)
			return
		}
		b.ResetTimer()
		benchmarkRingQueueConcurrentProduce(b, c, p, 1)
	})

	b.Run("4 producers", func(b *testing.B) {
		c, p, err := sox.NewRingQueue[int](func(options *sox.RingQueueOptions) {
			options.ConcurrentProduce = true
			options.ConcurrentConsume = false
			options.Nonblocking = false
		})
		if err != nil {
			b.Errorf("ring queue new: %v", err)
			return
		}
		b.ResetTimer()
		benchmarkRingQueueConcurrentProduce(b, c, p, 4)
	})

	b.Run("16 producers", func(b *testing.B) {
		c, p, err := sox.NewRingQueue[int](func(options *sox.RingQueueOptions) {
			options.ConcurrentProduce = true
			options.ConcurrentConsume = false
			options.Nonblocking = false
		})
		if err != nil {
			b.Errorf("ring queue new: %v", err)
			return
		}
		b.ResetTimer()
		benchmarkRingQueueConcurrentProduce(b, c, p, 16)
	})
}

func TestRingQueue_ConcurrentConsume(t *testing.T) {
	t.Run("a little ops", func(t *testing.T) {
		c, p, err := sox.NewRingQueue[uintptr](func(options *sox.RingQueueOptions) {
			options.Capacity = 0x3
			options.ConcurrentProduce = false
			options.ConcurrentConsume = true
			options.Nonblocking = true
		})
		if err != nil {
			t.Errorf("ring queue new: %v", err)
			return
		}
		testRingQueueNonblocking(t, c, p)
	})

	t.Run("4 consume goroutines 32k buffer", func(t *testing.T) {
		c, p, err := sox.NewRingQueue[int64](func(options *sox.RingQueueOptions) {
			options.ConcurrentProduce = false
			options.ConcurrentConsume = true
			options.Nonblocking = false
		})
		if err != nil {
			t.Errorf("ring queue new: %v", err)
			return
		}
		testRingQueueConcurrentConsume(t, c, p, 0x04, 0x2000)
	})

	t.Run("16 consume goroutines 32k buffer", func(t *testing.T) {
		c, p, err := sox.NewRingQueue[int64](func(options *sox.RingQueueOptions) {
			options.ConcurrentProduce = false
			options.ConcurrentConsume = true
			options.Nonblocking = false
		})
		if err != nil {
			t.Errorf("ring queue new: %v", err)
			return
		}
		testRingQueueConcurrentConsume(t, c, p, 0x10, 0x2000)
	})

	t.Run("64 consume goroutines 32k buffer", func(t *testing.T) {
		c, p, err := sox.NewRingQueue[int64](func(options *sox.RingQueueOptions) {
			options.ConcurrentProduce = false
			options.ConcurrentConsume = true
			options.Nonblocking = false
		})
		if err != nil {
			t.Errorf("ring queue new: %v", err)
			return
		}
		testRingQueueConcurrentConsume(t, c, p, 0x40, 0x2000)
	})
}

func BenchmarkRingQueue_ConcurrentConsume(b *testing.B) {
	b.Run("1 consumer", func(b *testing.B) {
		c, p, err := sox.NewRingQueue[int](func(options *sox.RingQueueOptions) {
			options.ConcurrentProduce = false
			options.ConcurrentConsume = true
			options.Nonblocking = false
		})
		if err != nil {
			b.Errorf("ring queue new: %v", err)
			return
		}
		b.ResetTimer()
		benchmarkRingQueueConcurrentConsume(b, c, p, 1)
	})

	b.Run("4 consumers", func(b *testing.B) {
		c, p, err := sox.NewRingQueue[int](func(options *sox.RingQueueOptions) {
			options.ConcurrentProduce = false
			options.ConcurrentConsume = true
			options.Nonblocking = false
		})
		if err != nil {
			b.Errorf("ring queue new: %v", err)
			return
		}
		b.ResetTimer()
		benchmarkRingQueueConcurrentConsume(b, c, p, 4)
	})

	b.Run("16 consumers", func(b *testing.B) {
		c, p, err := sox.NewRingQueue[int](func(options *sox.RingQueueOptions) {
			options.ConcurrentProduce = false
			options.ConcurrentConsume = true
			options.Nonblocking = false
		})
		if err != nil {
			b.Errorf("ring queue new: %v", err)
			return
		}
		b.ResetTimer()
		benchmarkRingQueueConcurrentConsume(b, c, p, 16)
	})
}

func TestRingQueue_Concurrent(t *testing.T) {
	t.Run("a little ops", func(t *testing.T) {
		c, p, err := sox.NewRingQueue[uintptr](func(options *sox.RingQueueOptions) {
			options.Capacity = 0x3
			options.ConcurrentProduce = true
			options.ConcurrentConsume = true
			options.Nonblocking = true
		})
		if err != nil {
			t.Errorf("ring queue new: %v", err)
			return
		}
		testRingQueueNonblocking(t, c, p)
	})

	t.Run("16 producer goroutines 32k buffer", func(t *testing.T) {
		c, p, err := sox.NewRingQueue[int64](func(options *sox.RingQueueOptions) {
			options.ConcurrentProduce = true
			options.ConcurrentConsume = true
			options.Nonblocking = false
		})
		if err != nil {
			t.Errorf("ring queue new: %v", err)
			return
		}
		testRingQueueConcurrentProduce(t, c, p, 0x10, 0x2000)
	})

	t.Run("16 consume goroutines 32k buffer", func(t *testing.T) {
		c, p, err := sox.NewRingQueue[int64](func(options *sox.RingQueueOptions) {
			options.ConcurrentProduce = true
			options.ConcurrentConsume = true
			options.Nonblocking = false
		})
		if err != nil {
			t.Errorf("ring queue new: %v", err)
			return
		}
		testRingQueueConcurrentConsume(t, c, p, 0x10, 0x2000)
	})

	t.Run("16 produce and 16 consume goroutines 32k buffer", func(t *testing.T) {
		c, p, err := sox.NewRingQueue[int64](func(options *sox.RingQueueOptions) {
			options.ConcurrentProduce = true
			options.ConcurrentConsume = true
			options.Nonblocking = false
		})
		if err != nil {
			t.Errorf("ring queue new: %v", err)
			return
		}
		testRingQueueConcurrent(t, c, p, 0x10, 0x10, 0x2000)
	})

	t.Run("4 produce and 64 consume goroutines 32k buffer", func(t *testing.T) {
		c, p, err := sox.NewRingQueue[int64](func(options *sox.RingQueueOptions) {
			options.ConcurrentProduce = true
			options.ConcurrentConsume = true
			options.Nonblocking = false
		})
		if err != nil {
			t.Errorf("ring queue new: %v", err)
			return
		}
		testRingQueueConcurrent(t, c, p, 0x40, 0x04, 0x2000)
	})

	t.Run("64 produce and 4 consume goroutines 32k buffer", func(t *testing.T) {
		c, p, err := sox.NewRingQueue[int64](func(options *sox.RingQueueOptions) {
			options.Capacity = (1 << 22) - 1
			options.ConcurrentProduce = true
			options.ConcurrentConsume = true
			options.Nonblocking = false
		})
		if err != nil {
			t.Errorf("ring queue new: %v", err)
			return
		}
		testRingQueueConcurrent(t, c, p, 0x04, 0x40, 0x2000)
	})
}

func BenchmarkRingQueue_Concurrent(b *testing.B) {
	b.Run("1 producer 1 consumer", func(b *testing.B) {
		c, p, err := sox.NewRingQueue[int](func(options *sox.RingQueueOptions) {
			options.ConcurrentProduce = true
			options.ConcurrentConsume = true
			options.Nonblocking = false
		})
		if err != nil {
			b.Errorf("ring queue new: %v", err)
			return
		}
		b.ResetTimer()
		benchmarkRingQueueConcurrent(b, c, p, 1, 1)
	})

	b.Run("4 producers 4 consumers", func(b *testing.B) {
		c, p, err := sox.NewRingQueue[int](func(options *sox.RingQueueOptions) {
			options.ConcurrentProduce = true
			options.ConcurrentConsume = true
			options.Nonblocking = false
		})
		if err != nil {
			b.Errorf("ring queue new: %v", err)
			return
		}
		b.ResetTimer()
		benchmarkRingQueueConcurrent(b, c, p, 4, 4)
	})

	b.Run("16 producers 16 consumers", func(b *testing.B) {
		c, p, err := sox.NewRingQueue[int](func(options *sox.RingQueueOptions) {
			options.ConcurrentProduce = true
			options.ConcurrentConsume = true
			options.Nonblocking = false
		})
		if err != nil {
			b.Errorf("ring queue new: %v", err)
			return
		}
		b.ResetTimer()
		benchmarkRingQueueConcurrent(b, c, p, 16, 16)
	})

	b.Run("4 producers 16 consumers", func(b *testing.B) {
		c, p, err := sox.NewRingQueue[int](func(options *sox.RingQueueOptions) {
			options.ConcurrentProduce = true
			options.ConcurrentConsume = true
			options.Nonblocking = false
		})
		if err != nil {
			b.Errorf("ring queue new: %v", err)
			return
		}
		b.ResetTimer()
		benchmarkRingQueueConcurrent(b, c, p, 16, 4)
	})

	b.Run("16 producers 4 consumers", func(b *testing.B) {
		c, p, err := sox.NewRingQueue[int](func(options *sox.RingQueueOptions) {
			options.ConcurrentProduce = true
			options.ConcurrentConsume = true
			options.Nonblocking = false
		})
		if err != nil {
			b.Errorf("ring queue new: %v", err)
			return
		}
		b.ResetTimer()
		benchmarkRingQueueConcurrent(b, c, p, 4, 16)
	})
}

func testRingQueueNonblocking(t *testing.T, c sox.ItemConsumer[uintptr], p sox.ItemProducer[uintptr]) {
	item, err := c.Consume()
	if err != sox.ErrTemporarilyUnavailable {
		t.Errorf("ring consumer expected ErrTemporarilyUnavailable but got %v %v", item, err)
		return
	}
	err = p.Produce(1)
	if err != nil {
		t.Errorf("ring producer produce: %v", err)
		return
	}
	item, err = c.Consume()
	if err != nil {
		t.Errorf("ring consumer consume: %v", err)
		return
	}
	if item != 1 {
		t.Errorf("ring consumed item expected %d but got %d", 1, item)
		return
	}
	item, err = c.Consume()
	if err != sox.ErrTemporarilyUnavailable {
		t.Errorf("ring consumer expected ErrTemporarilyUnavailable but got %v %v", item, err)
		return
	}
	err = p.Produce(2)
	if err != nil {
		t.Errorf("ring producer produce: %v", err)
		return
	}
	err = p.Produce(3)
	if err != nil {
		t.Errorf("ring producer produce: %v", err)
		return
	}
	err = p.Produce(4)
	if err != nil {
		t.Errorf("ring producer produce: %v", err)
		return
	}
	item, err = c.Consume()
	if err != nil {
		t.Errorf("ring consumer consume: %v", err)
		return
	}
	if item != 2 {
		t.Errorf("ring consumed item expected %d but got %d", 2, item)
		return
	}
	err = p.Produce(5)
	if err != nil {
		t.Errorf("ring producer produce: %v", err)
		return
	}
	err = p.Produce(6)
	if err != sox.ErrTemporarilyUnavailable {
		t.Errorf("ring producer expected ErrTemporarilyUnavailable but got %v", err)
		return
	}
	item, err = c.Consume()
	if err != nil {
		t.Errorf("ring consumer consume: %v", err)
		return
	}
	if item != 3 {
		t.Errorf("ring consumed item expected %d but got %d", 3, item)
		return
	}
	item, err = c.Consume()
	if err != nil {
		t.Errorf("ring consumer consume: %v", err)
		return
	}
	if item != 4 {
		t.Errorf("ring consumed item expected %d but got %d", 4, item)
		return
	}
	err = p.Produce(7)
	if err != nil {
		t.Errorf("ring producer produce: %v", err)
		return
	}
	err = p.Close()
	if err != nil {
		t.Errorf("ring producer close: %v", err)
		return
	}
	err = p.Produce(8)
	if err != io.ErrClosedPipe {
		t.Errorf("ring closed producer expected %v but got %v", io.ErrClosedPipe, err)
		return
	}
	item, err = c.Consume()
	if err != nil {
		t.Errorf("ring consumer consume: %v", err)
		return
	}
	if item != 5 {
		t.Errorf("ring consumed item expected %d but got %d", 5, item)
		return
	}
	item, err = c.Consume()
	if err != nil {
		t.Errorf("ring consumer consume: %v", err)
		return
	}
	if item != 7 {
		t.Errorf("ring consumed item expected %d but got %d", 6, item)
		return
	}
	item, err = c.Consume()
	if err != io.EOF {
		t.Errorf("ring consumer consume expected %v but got %v %v", io.EOF, item, err)
		return
	}
}

func testRingQueueConcurrentProduce(t *testing.T, c sox.ItemConsumer[int64], p sox.ItemProducer[int64], m int, n int) {
	last := make([]int64, m)
	for i := range m {
		last[i] = -1
	}
	for i := range m {
		go func(i int) {
			for j := range n {
				err := p.Produce(int64(i<<32) | int64(j))
				if err != nil {
					t.Errorf("ring producer produce: %v", err)
					return
				}
			}
		}(i)
	}
	for range m * n {
		item, err := c.Consume()
		if err != nil {
			t.Errorf("ring consumer consume: %v", err)
			return
		}
		high, low := item>>32, item&math.MaxUint32
		if low <= last[high] {
			t.Errorf("ring producer out of order")
			return
		}
		last[high] = low
	}
	err := p.Close()
	if err != nil {
		t.Errorf("ring producer close: %v", err)
		return
	}
	item, err := c.Consume()
	if err != io.EOF {
		t.Errorf("ring consumer consume expected %v but got %v %v", io.EOF, item, err)
		return
	}
}

func benchmarkRingQueueConcurrentProduce(b *testing.B, c sox.ItemConsumer[int], p sox.ItemProducer[int], num int) {
	for range num {
		go func() {
			for j := range b.N/num + 1 {
				err := p.Produce(j)
				if err != nil {
					b.Errorf("ring producer produce: %v", err)
					return
				}
			}
		}()
	}
	for range b.N {
		_, err := c.Consume()
		if err != nil {
			b.Errorf("ring consumer consume: %v", err)
			return
		}
	}
}

func testRingQueueConcurrentConsume(t *testing.T, c sox.ItemConsumer[int64], p sox.ItemProducer[int64], m int, n int) {
	wg := sync.WaitGroup{}
	for range m {
		wg.Add(1)
		go func() {
			last := int64(-1)
			for range n {
				item, err := c.Consume()
				if err != nil {
					t.Errorf("ring consumer consume: %v", err)
					return
				}
				if item <= last {
					t.Errorf("ring consumer out of order")
					return
				}
				last = item
			}
			wg.Done()
		}()
	}
	for i := range m * n {
		err := p.Produce(int64(i))
		if err != nil {
			t.Errorf("ring producer produce: %v", err)
			return
		}
	}
	wg.Wait()
	err := p.Close()
	if err != nil {
		t.Errorf("ring producer close: %v", err)
		return
	}
	item, err := c.Consume()
	if err != io.EOF {
		t.Errorf("ring consumer consume expected %v but got %v %v", io.EOF, item, err)
		return
	}
}

func benchmarkRingQueueConcurrentConsume(b *testing.B, c sox.ItemConsumer[int], p sox.ItemProducer[int], num int) {
	wg := sync.WaitGroup{}
	for range num {
		wg.Add(1)
		go func() {
			for range b.N / num {
				_, err := c.Consume()
				if err != nil {
					b.Errorf("ring consumer consume: %v", err)
					return
				}
			}
			wg.Done()
		}()
	}
	for i := range b.N {
		err := p.Produce(i)
		if err != nil {
			b.Errorf("ring producer produce: %v", err)
			return
		}
	}
	wg.Wait()
}

func testRingQueueConcurrent(t *testing.T, c sox.ItemConsumer[int64], p sox.ItemProducer[int64], cNum, pNum int, n int) {
	for i := range pNum {
		go func(i int) {
			for j := range n {
				err := p.Produce(int64(i<<32) | int64(j))
				if err != nil {
					t.Errorf("ring producer produce: %v", err)
					return
				}
			}
		}(i)
	}
	wg := sync.WaitGroup{}
	for i := range cNum {
		wg.Add(1)
		go func(i int) {
			last := make([]int64, pNum)
			for j := range pNum {
				last[j] = -1
			}
			for range n * pNum / cNum {
				item, err := c.Consume()
				if err != nil {
					t.Errorf("ring consumer consume: %v", err)
					return
				}
				high, low := item>>32, item&math.MaxUint32
				if low <= last[high] {
					t.Errorf("ring produce comsume out of order")
					return
				}
				last[high] = low
			}
			wg.Done()
		}(i)
	}
	wg.Wait()
	err := p.Close()
	if err != nil {
		t.Errorf("ring producer close: %v", err)
		return
	}
	item, err := c.Consume()
	if err != io.EOF {
		t.Errorf("ring consumer consume expected %v but got %v %v", io.EOF, item, err)
		return
	}
}

func benchmarkRingQueueConcurrent(b *testing.B, c sox.ItemConsumer[int], p sox.ItemProducer[int], cNum, pNum int) {
	for i := range pNum {
		go func(i int) {
			for j := range b.N/pNum + 1 {
				err := p.Produce(j)
				if err != nil {
					b.Errorf("ring producer produce: %v", err)
					return
				}
			}
		}(i)
	}
	wg := sync.WaitGroup{}
	for i := range cNum {
		wg.Add(1)
		go func(i int) {
			for range b.N / cNum {
				_, err := c.Consume()
				if err != nil {
					b.Errorf("ring consumer consume: %v", err)
					return
				}
			}
			wg.Done()
		}(i)
	}
	wg.Wait()
}
