// ©Hayabusa Cloud Co., Ltd. 2022. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

//go:build linux

package sox

import (
	"encoding/binary"
	"golang.org/x/sys/unix"
	"time"
)

type timerfd struct {
	tickedAt  time.Time
	tickCount uint64
	buf       []byte

	startedAt time.Time
	d         time.Duration
	fd        int
}

func newTimerfd(d time.Duration) (Timer, error) {
	if d <= 0 {
		return nil, ErrInvalidParam
	}

	fd, err := unix.TimerfdCreate(unix.CLOCK_MONOTONIC, unix.TFD_NONBLOCK|unix.TFD_CLOEXEC)
	if err != nil {
		return nil, errFromUnixErrno(err)
	}

	sec, nano := int64(d.Truncate(time.Second)/time.Second), d.Nanoseconds()%int64(time.Second/time.Nanosecond)
	err = unix.TimerfdSettime(fd, 0, &unix.ItimerSpec{
		Interval: unix.Timespec{Sec: sec, Nsec: nano},
		Value:    unix.NsecToTimespec(d.Nanoseconds()),
	}, nil)
	if err != nil {
		return nil, errFromUnixErrno(err)
	}

	return &timerfd{fd: fd, buf: make([]byte, 8), tickCount: 0, startedAt: time.Now().Local(), d: d}, nil
}

func (tm *timerfd) Fd() int {
	return tm.fd
}

func (tm *timerfd) Now() time.Time {
	return tm.tickedAt
}

func (tm *timerfd) Read(p []byte) (n int, err error) {
	n, err = unix.Read(tm.fd, p)
	if err != nil {
		return n, errFromUnixErrno(err)
	}
	tm.tickCount = binary.LittleEndian.Uint64(tm.buf)
	tm.tickedAt = tm.startedAt.Add(tm.d * time.Duration(tm.tickCount))

	return n, nil
}

func (tm *timerfd) Close() error {
	err := unix.Close(tm.fd)
	if err != nil {
		return errFromUnixErrno(err)
	}
	return nil
}
