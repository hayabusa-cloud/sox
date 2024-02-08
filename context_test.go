// ©Hayabusa Cloud Co., Ltd. 2022. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package sox_test

import (
	"context"
	"hybscloud.com/sox"
	"testing"
	"unsafe"
)

func TestContextWithUserdata(t *testing.T) {
	t.Run("numeric", func(t *testing.T) {
		val := sox.ContextUserdata[int](context.Background())
		if val != 0 {
			t.Errorf("expected val=%v but got %v", 0, val)
			return
		}
		ctx := sox.ContextWithUserdata[int](context.Background(), 5)
		val = sox.ContextUserdata[int](ctx)
		if val != 5 {
			t.Errorf("expected val=%v but got %v", 5, val)
			return
		}
		ctx = sox.ContextWithUserdata[int](ctx, 10)
		ctx = sox.ContextWithUserdata[int](ctx, 15)
		val = sox.ContextUserdata[int](ctx)
		if val != 15 {
			t.Errorf("expected val=%v but got %v", 15, val)
			return
		}
	})

	t.Run("string", func(t *testing.T) {
		val := sox.ContextUserdata[string](context.Background())
		if val != "" {
			t.Errorf("expected val=\"%v\" but got \"%v\"", "", val)
			return
		}
		ctx := sox.ContextWithUserdata[string](context.Background(), "sox")
		val = sox.ContextUserdata[string](ctx)
		if val != "sox" {
			t.Errorf("expected val=\"%v\" but got \"%v\"", "sox", val)
			return
		}
		ctx = sox.ContextWithUserdata[string](ctx, "git@hybscloud.com")
		ctx = sox.ContextWithUserdata[string](ctx, "io-library")
		val = sox.ContextUserdata[string](ctx)
		if val != "io-library" {
			t.Errorf("expected val=\"%v\" but got \"%v\"", "io-library", val)
			return
		}
	})

	t.Run("struct", func(t *testing.T) {
		type Struct struct {
			x int
		}

		val := sox.ContextUserdata[Struct](context.Background())
		if val.x != 0 {
			t.Errorf("expected val=%v but got %v", 0, val.x)
			return
		}
		ctx := sox.ContextWithUserdata[Struct](context.Background(), Struct{5})
		val = sox.ContextUserdata[Struct](ctx)
		if val.x != 5 {
			t.Errorf("expected val=%v but got %v", 5, val.x)
			return
		}
	})

	t.Run("pointer", func(t *testing.T) {
		type Struct struct {
			x int
		}

		val := sox.ContextUserdata[*Struct](context.Background())
		if val != nil {
			t.Errorf("expected val=%v but got %v", nil, val)
			return
		}
		p := &Struct{5}
		ctx := sox.ContextWithUserdata[*Struct](context.Background(), p)
		val = sox.ContextUserdata[*Struct](ctx)
		if val != p {
			t.Errorf("expected val=%v but got %v", p, val)
			return
		}
	})

	t.Run("container", func(t *testing.T) {
		val := sox.ContextUserdata[[]byte](context.Background())
		if val != nil {
			t.Errorf("expected val=%v but got %v", nil, val)
			return
		}
		s := []byte{5, 10, 15, 20, 25, 30, 35, 40, 45, 50}
		ctx := sox.ContextWithUserdata[[]byte](context.Background(), s)
		val = sox.ContextUserdata[[]byte](ctx)
		if len(val) != 10 || val[0] != 5 || val[9] != 50 {
			t.Errorf("expected val=%v but got %v", s, val)
			return
		}
		if unsafe.Pointer(&val[0]) != unsafe.Pointer(&s[0]) {
			t.Errorf("expected val=%v but got %v", s, val)
			return
		}
	})

	t.Run("channel", func(t *testing.T) {
		val := sox.ContextUserdata[chan struct{ int }](context.Background())
		if val != nil {
			t.Errorf("expected val=%v but got %v", nil, val)
			return
		}
		c := make(chan struct{ int }, 3)
		c <- struct{ int }{5}
		ctx := sox.ContextWithUserdata[chan struct{ int }](context.Background(), c)
		val = sox.ContextUserdata[chan struct{ int }](ctx)
		if len(val) != 1 || cap(val) != 3 {
			t.Errorf("expected val=%v but got %v", c, val)
			return
		}
		if e := <-c; e.int != 5 {
			t.Errorf("expected elem=%v but got %v", 5, e.int)
			return
		}
	})

	t.Run("interface", func(t *testing.T) {
		ctx := sox.ContextWithUserdata[any](context.Background(), 10)
		val := sox.ContextUserdata[any](ctx)
		if val, ok := val.(int); !ok || val != 10 {
			t.Errorf("expected val=%v but got %v", 10, val)
			return
		}
		ctx = sox.ContextWithUserdata[any](ctx, "sox")
		val = sox.ContextUserdata[any](ctx)
		if val, ok := val.(string); !ok || val != "sox" {
			t.Errorf("expected val=%v but got %v", "sox", val)
			return
		}
		i := sox.ContextUserdata[int](ctx)
		if i != 0 {
			t.Errorf("expected i=%d but got %d", 0, i)
			return
		}
	})

	t.Run("multiple values", func(t *testing.T) {
		ctx := sox.ContextWithUserdata[int](context.Background(), 5)
		ctx = sox.ContextWithUserdata[string](ctx, "git@hybscloud.com")
		ctx = sox.ContextWithUserdata[string](ctx, "sox")
		ctx = sox.ContextWithUserdata[float32](ctx, 10)
		a := sox.ContextUserdata[any](ctx)
		if a != nil {
			t.Errorf("expected a=%v but got %v", nil, a)
			return
		}
		f64 := sox.ContextUserdata[float64](ctx)
		if f64 != 0 {
			t.Errorf("expected f64=%f but got %f", 0.0, f64)
			return
		}
		i := sox.ContextUserdata[int](ctx)
		if i != 5 {
			t.Errorf("expected i=%d but got %d", 5, i)
			return
		}
		s := sox.ContextUserdata[string](ctx)
		if s != "sox" {
			t.Errorf("expected s=\"%v\" but got \"%v\"", "sox", s)
			return
		}
		f32 := sox.ContextUserdata[float32](ctx)
		if f32 != 10.0 {
			t.Errorf("expected f32=%f but got %f", 10.0, f32)
			return
		}
	})
}
