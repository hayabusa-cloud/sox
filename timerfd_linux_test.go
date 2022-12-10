//go:build linux

package sox

import (
	"testing"
	"time"
)

func TestTimer_Tick(t *testing.T) {
	fn := func(d time.Duration, t *testing.T) {
		tm, err := newTimerfd(d)
		if err != nil {
			t.Errorf("create timerfd fd: %v", err)
			return
		}
		defer tm.Close()

		tfd := tm.(*timerfd)
		time.Sleep(d + jiffies)
		_, err = tfd.Read(tfd.buf)
		if err != nil {
			t.Errorf("timerfd refresh: %v", err)
			return
		}
		if tm.Now().Sub(time.Now()).Abs() >= d/2+jiffies {
			t.Errorf("too large time difference")
			return
		}
	}

	t.Run("200msec", func(t *testing.T) {
		fn(200*time.Millisecond, t)
	})

	t.Run("1sec100msec", func(t *testing.T) {
		fn(1*time.Second+100*time.Millisecond, t)
	})
}
