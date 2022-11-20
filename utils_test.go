package sox

import (
	"os"
	"testing"
	"unsafe"
)

func TestAlignedBlocks(t *testing.T) {
	t.Run("single", func(t *testing.T) {
		b := AlignedMemBlocks(1)[0]
		if len(b) != os.Getpagesize() {
			t.Errorf("expected slice len %d but got %d", os.Getpagesize(), len(b))
			return
		}
		if cap(b) != os.Getpagesize() {
			t.Errorf("expected slice cap %d but got %d", os.Getpagesize(), cap(b))
			return
		}
		ptr := uintptr(unsafe.Pointer(&b[0]))
		if ptr%uintptr(os.Getpagesize()) != 0 {
			t.Errorf("memory not aligned")
			return
		}
	})

	t.Run("multiple", func(t *testing.T) {
		const n = 1024
		blocks := AlignedMemBlocks(n)
		for i := 0; i < n; i++ {
			b := blocks[i]
			if len(b) != os.Getpagesize() {
				t.Errorf("expected slice len %d but got %d", os.Getpagesize(), len(b))
				return
			}
			if cap(b) != os.Getpagesize() {
				t.Errorf("expected slice cap %d but got %d", os.Getpagesize(), cap(b))
				return
			}
			ptr := uintptr(unsafe.Pointer(&b[0]))
			if ptr%uintptr(os.Getpagesize()) != 0 {
				t.Errorf("memory not aligned")
				return
			}
		}
	})
}
