package sox

import (
	"math"
	"time"
)

const (
	SpinWaitLevelClient = iota
	SpinWaitLevelBlockingIO
	SpinWaitLevelConsume
	spinWaitLevelProduce
	spinWaitLevelAtomic
	spinWaitLevelNoSleep
)

type SpinWait struct {
	i     uint32
	level int8
	d     time.Duration
	limit uint32
	total int32
}

func NewSpinWait() *SpinWait {
	return &SpinWait{
		i:     0,
		level: SpinWaitLevelBlockingIO,
		d:     jiffies,
		limit: 0,
		total: 0,
	}
}

func (sw *SpinWait) SetLevel(level int) *SpinWait {
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

func (sw *SpinWait) SetLimit(limit int) *SpinWait {
	if limit > math.MaxUint32-1 {
		limit = math.MaxUint32 - 1
	}
	sw.limit = uint32(limit)

	return sw
}

func (sw *SpinWait) SetDuration(d time.Duration) *SpinWait {
	if d < jiffies {
		d = jiffies
	}
	sw.d = d

	return sw
}

func (sw *SpinWait) Once() {
	sw.once(sw.level)
}

func (sw *SpinWait) OnceWithLevel(level int) {
	if level < SpinWaitLevelClient {
		level = SpinWaitLevelClient
	}
	if level > spinWaitLevelNoSleep {
		level = spinWaitLevelNoSleep
	}
	sw.once(int8(level))
}

func (sw *SpinWait) Reset() {
	sw.i = 0
	sw.total = 0
}

func (sw *SpinWait) Closed() bool {
	return sw.limit > 0 && sw.i >= sw.limit
}

func (sw *SpinWait) once(level int8) {
	d := int32(0)
	for i := sw.i; i > 0; i &= i - 1 {
		d++
	}
	d = d >> level
	if d > 0 {
		sw.total += d
		time.Sleep(time.Duration(d) * sw.d)
	}
	sw.i++
}
