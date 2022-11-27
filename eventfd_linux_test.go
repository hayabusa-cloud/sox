//go:build linux

package sox

import "testing"

func TestEventObject_ReadWrite(t *testing.T) {
	t.Run("single write read", func(t *testing.T) {
		evt, err := newEventfd()
		if err != nil {
			t.Errorf("new eventfd: %v", err)
			return
		}
		defer evt.Close()
		err = evt.WriteUint(10)
		if err != nil {
			t.Errorf("write event object: %v", err)
			return
		}
		val, err := evt.ReadUint()
		if err != nil {
			t.Errorf("read event object: %v", err)
			return
		}
		if val != 10 {
			t.Errorf("read expected 10 but got %d", val)
			return
		}
		val, err = evt.ReadUint()
		if err != ErrTemporarilyUnavailable {
			t.Errorf("read event object expected EGAIN but got: %d", val)
			return
		}
	})

	t.Run("multiple write and read", func(t *testing.T) {
		evt, err := newEventfd()
		if err != nil {
			t.Errorf("new eventfd: %v", err)
			return
		}
		defer evt.Close()
		err = evt.WriteUint(10)
		if err != nil {
			t.Errorf("write eventfd: %v", err)
			return
		}
		err = evt.WriteUint(20)
		if err != nil {
			t.Errorf("write eventfd: %v", err)
			return
		}
		err = evt.WriteUint(40)
		if err != nil {
			t.Errorf("write eventfd: %v", err)
			return
		}
		val, err := evt.ReadUint()
		if err != nil {
			t.Errorf("read eventfd: %v", err)
			return
		}
		if val != 70 {
			t.Errorf("read expected 10 but got %d", val)
			return
		}
		val, err = evt.ReadUint()
		if err != ErrTemporarilyUnavailable {
			t.Errorf("read event object expected EGAIN but got: %d", val)
			return
		}
		err = evt.WriteUint(10)
		if err != nil {
			t.Errorf("write eventfd: %v", err)
			return
		}
		val, err = evt.ReadUint()
		if err != nil {
			t.Errorf("read eventfd: %v", err)
			return
		}
		if val != 10 {
			t.Errorf("read expected 10 but got %d", val)
			return
		}
	})

	t.Run("write 0", func(t *testing.T) {
		evt, err := newEventfd()
		if err != nil {
			t.Errorf("new eventfd: %v", err)
			return
		}
		defer evt.Close()
		err = evt.WriteUint(0)
		if err != nil {
			t.Errorf("write event object: %v", err)
			return
		}
		val, err := evt.ReadUint()
		if err != ErrTemporarilyUnavailable {
			t.Errorf("read event object expected EGAIN but got: %d", val)
			return
		}
	})
}
