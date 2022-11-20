package sox

import (
	"os"
	"runtime"
	"time"
	"unsafe"
)

// AlignedMemBlocks returns n bytes slices which
// has length with memory page size and address
// starts from multiple of memory page size
func AlignedMemBlocks(n int) (blocks [][]byte) {
	if n < 1 {
		panic("bad block num")
	}
	blocks = make([][]byte, n)
	size := os.Getpagesize()
	p := make([]byte, size*(n+1))
	ptr := uintptr(unsafe.Pointer(&p[0]))
	off := ptr - (ptr & ^(uintptr(size) - 1))
	for i := 0; i < n; i++ {
		blocks[i] = unsafe.Slice(&p[i*size-int(off)], size)
	}
	return
}

// AlignedMemBlock returns one aligned block
func AlignedMemBlock() []byte {
	return AlignedMemBlocks(1)[0]
}

// Yield yields the current goroutine
// If d > 0, Yield() also sleeps d duration
func Yield(d time.Duration) {
	if d > 0 {
		time.Sleep(d)
		return
	}
	runtime.Gosched()
}
