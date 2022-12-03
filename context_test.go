package sox

import (
	"context"
	"testing"
	"unsafe"
)

func TestContextWithUserdata(t *testing.T) {
	t.Run("numeric", func(t *testing.T) {
		val := ContextUserdata[int](context.Background())
		if val != 0 {
			t.Errorf("expected val=%v but got %v", 0, val)
			return
		}
		ctx := ContextWithUserdata[int](context.Background(), 5)
		val = ContextUserdata[int](ctx)
		if val != 5 {
			t.Errorf("expected val=%v but got %v", 5, val)
			return
		}
		ctx = ContextWithUserdata[int](ctx, 10)
		ctx = ContextWithUserdata[int](ctx, 15)
		val = ContextUserdata[int](ctx)
		if val != 15 {
			t.Errorf("expected val=%v but got %v", 15, val)
			return
		}
	})

	t.Run("string", func(t *testing.T) {
		val := ContextUserdata[string](context.Background())
		if val != "" {
			t.Errorf("expected val=\"%v\" but got \"%v\"", "", val)
			return
		}
		ctx := ContextWithUserdata[string](context.Background(), "sox")
		val = ContextUserdata[string](ctx)
		if val != "sox" {
			t.Errorf("expected val=\"%v\" but got \"%v\"", "sox", val)
			return
		}
		ctx = ContextWithUserdata[string](ctx, "git@hybscloud.com")
		ctx = ContextWithUserdata[string](ctx, "io-library")
		val = ContextUserdata[string](ctx)
		if val != "io-library" {
			t.Errorf("expected val=\"%v\" but got \"%v\"", "io-library", val)
			return
		}
	})

	t.Run("struct", func(t *testing.T) {
		type Struct struct {
			x int
		}

		val := ContextUserdata[Struct](context.Background())
		if val.x != 0 {
			t.Errorf("expected val=%v but got %v", 0, val.x)
			return
		}
		ctx := ContextWithUserdata[Struct](context.Background(), Struct{5})
		val = ContextUserdata[Struct](ctx)
		if val.x != 5 {
			t.Errorf("expected val=%v but got %v", 5, val.x)
			return
		}
	})

	t.Run("pointer", func(t *testing.T) {
		type Struct struct {
			x int
		}

		val := ContextUserdata[*Struct](context.Background())
		if val != nil {
			t.Errorf("expected val=%v but got %v", nil, val)
			return
		}
		p := &Struct{5}
		ctx := ContextWithUserdata[*Struct](context.Background(), p)
		val = ContextUserdata[*Struct](ctx)
		if val != p {
			t.Errorf("expected val=%v but got %v", p, val)
			return
		}
	})

	t.Run("container", func(t *testing.T) {
		val := ContextUserdata[[]byte](context.Background())
		if val != nil {
			t.Errorf("expected val=%v but got %v", nil, val)
			return
		}
		s := []byte{5, 10, 15, 20, 25, 30, 35, 40, 45, 50}
		ctx := ContextWithUserdata[[]byte](context.Background(), s)
		val = ContextUserdata[[]byte](ctx)
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
		val := ContextUserdata[chan struct{ int }](context.Background())
		if val != nil {
			t.Errorf("expected val=%v but got %v", nil, val)
			return
		}
		c := make(chan struct{ int }, 3)
		c <- struct{ int }{5}
		ctx := ContextWithUserdata[chan struct{ int }](context.Background(), c)
		val = ContextUserdata[chan struct{ int }](ctx)
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
		ctx := ContextWithUserdata[any](context.Background(), 10)
		val := ContextUserdata[any](ctx)
		if val, ok := val.(int); !ok || val != 10 {
			t.Errorf("expected val=%v but got %v", 10, val)
			return
		}
		ctx = ContextWithUserdata[any](ctx, "sox")
		val = ContextUserdata[any](ctx)
		if val, ok := val.(string); !ok || val != "sox" {
			t.Errorf("expected val=%v but got %v", "sox", val)
			return
		}
		i := ContextUserdata[int](ctx)
		if i != 0 {
			t.Errorf("expected i=%d but got %d", 0, i)
			return
		}
	})

	t.Run("multiple values", func(t *testing.T) {
		ctx := ContextWithUserdata[int](context.Background(), 5)
		ctx = ContextWithUserdata[string](ctx, "git@hybscloud.com")
		ctx = ContextWithUserdata[string](ctx, "sox")
		ctx = ContextWithUserdata[float32](ctx, 10)
		a := ContextUserdata[any](ctx)
		if a != nil {
			t.Errorf("expected a=%v but got %v", nil, a)
			return
		}
		f64 := ContextUserdata[float64](ctx)
		if f64 != 0 {
			t.Errorf("expected f64=%f but got %f", 0.0, f64)
			return
		}
		i := ContextUserdata[int](ctx)
		if i != 5 {
			t.Errorf("expected i=%d but got %d", 5, i)
			return
		}
		s := ContextUserdata[string](ctx)
		if s != "sox" {
			t.Errorf("expected s=\"%v\" but got \"%v\"", "sox", s)
			return
		}
		f32 := ContextUserdata[float32](ctx)
		if f32 != 10.0 {
			t.Errorf("expected f32=%f but got %f", 10.0, f32)
			return
		}
	})
}

func TestContextWithFD(t *testing.T) {
	t.Run("no value", func(t *testing.T) {
		val := contextFD(context.Background())
		if val != -1 {
			t.Errorf("expected fd=%d but got %d", -1, val)
		}
	})

	t.Run("set once", func(t *testing.T) {
		ctx := contextWithFD(context.Background(), 10)
		val := contextFD(ctx)
		if val != 10 {
			t.Errorf("expected fd=%d but got %d", 10, val)
		}
	})

	t.Run("set more than once", func(t *testing.T) {
		ctx := contextWithFD(context.Background(), 10)
		val := contextFD(ctx)
		if val != 10 {
			t.Errorf("expected fd=%d but got %d", 10, val)
			return
		}
		ctx = contextWithFD(ctx, 15)
		ctx = contextWithFD(ctx, 20)
		val = contextFD(ctx)
		if val != 20 {
			t.Errorf("expected fd=%d but got %d", 20, val)
			return
		}
	})
}
