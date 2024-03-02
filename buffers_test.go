// ©Hayabusa Cloud Co., Ltd. 2022. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package sox_test

import (
	"hybscloud.com/sox"
	"os"
	"testing"
	"unsafe"
)

func TestAlignedMemBlocks(t *testing.T) {
	t.Run("single", func(t *testing.T) {
		b := sox.AlignedMemBlocks(1)[0]
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
		blocks := sox.AlignedMemBlocks(n)
		for i := range n {
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
