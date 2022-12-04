package sox

import (
	"math"
	"runtime"
	"time"
)

const (
	SpinWaitLevelClient = iota
	SpinWaitLevelBlockingIO
	SpinWaitLevelConsume
	SpinWaitLevelProduce
	spinWaitLevelAtomic
	spinWaitLevelNoSleep
)

type SpinWaiter struct {
	i     uint32
	level int8
	d     time.Duration
	limit uint32
	total int32
}

func NewSpinWaiter() *SpinWaiter {
	return &SpinWaiter{
		i:     0,
		level: SpinWaitLevelBlockingIO,
		d:     jiffies,
		limit: 0,
		total: 0,
	}
}

func (sw *SpinWaiter) SetLevel(level int) *SpinWaiter {
	if level < SpinWaitLevelClient {
		sw.level = SpinWaitLevelClient
		return sw
	}
	if level > spinWaitLevelNoSleep {
		level = spinWaitLevelNoSleep
	}
	sw.level = int8(level)

	return sw
}

func (sw *SpinWaiter) SetLimit(limit int) *SpinWaiter {
	if limit > math.MaxUint32-1 {
		limit = math.MaxUint32 - 1
	}
	sw.limit = uint32(limit)

	return sw
}

func (sw *SpinWaiter) SetDuration(d time.Duration) *SpinWaiter {
	if d < jiffies {
		d = jiffies
	}
	sw.d = d

	return sw
}

func (sw *SpinWaiter) Reset() {
	sw.i = 0
	sw.total = 0
}

func (sw *SpinWaiter) Closed() bool {
	return sw.limit > 0 && sw.i >= sw.limit
}

func (sw *SpinWaiter) Once() {
	d := int32(0)
	for i := sw.i; i > 0; i &= i - 1 {
		d++
	}
	d = d >> sw.level
	if d > 0 {
		sw.total += d
		time.Sleep(time.Duration(d) * sw.d)
	} else {
		runtime.Gosched()
	}
	sw.i++
}
