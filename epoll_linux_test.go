//go:build linux

package sox

import (
	"testing"
	"time"
)

func TestEpoll(t *testing.T) {
	ep, err := newPoller(16)
	if err != nil {
		t.Errorf("new epoll: %v", err)
		return
	}
	defer ep.Close()

	efd1, err := newEventfd()
	if err != nil {
		t.Errorf("new event fd: %v", err)
		return
	}
	defer efd1.Close()
	err = efd1.WriteUint(5)
	if err != nil {
		t.Errorf("event fd write: %v", err)
		return
	}
	efd2, err := newEventfd()
	if err != nil {
		t.Errorf("new event fd: %v", err)
		return
	}
	defer efd2.Close()
	err = efd2.WriteUint(10)
	if err != nil {
		t.Errorf("event fd write: %v", err)
		return
	}

	d := time.Millisecond * 200
	events, err := ep.wait(d)
	if err != nil {
		t.Errorf("epoll wait: %v", err)
		return
	}
	if len(events) != 0 {
		t.Errorf("epoll wait expected event num=%d but got %v", 0, events)
		return
	}
	err = ep.add(efd1.fd, pollerEventIn)
	if err != nil {
		t.Errorf("epoll add fd=%d: %v", efd1.fd, err)
		return
	}
	events, err = ep.wait(d)
	if err != nil {
		t.Errorf("epoll wait: %v", err)
		return
	}
	if len(events) != 1 {
		t.Errorf("epoll wait expected event num=%d but got %v", 1, events)
		return
	}
	if int(events[0].Fd) != efd1.fd || events[0].Events != pollerEventIn {
		t.Errorf("epoll event expected fd=%d events=%d but got %v", efd1.fd, pollerEventIn, events[0])
		return
	}

	err = ep.add(efd2.fd, pollerEventIn)
	if err != nil {
		t.Errorf("epoll add fd=%d: %v", efd2.fd, err)
		return
	}
	events, err = ep.wait(d)
	if err != nil {
		t.Errorf("epoll wait: %v", err)
		return
	}
	if len(events) != 1 {
		t.Errorf("epoll wait expected event num=%d but got %v", 1, events)
		return
	}
	if int(events[0].Fd) != efd2.fd || events[0].Events != pollerEventIn {
		t.Errorf("epoll event expected fd=%d events=%d but got %v", efd2.fd, pollerEventIn, events[0])
		return
	}

	err = efd1.WriteUint(5)
	if err != nil {
		t.Errorf("event fd write: %v", err)
		return
	}
	err = efd2.WriteUint(10)
	if err != nil {
		t.Errorf("event fd write: %v", err)
		return
	}
	events, err = ep.wait(d)
	if err != nil {
		t.Errorf("epoll wait: %v", err)
		return
	}
	if len(events) != 2 {
		t.Errorf("epoll wait expected event num=%d but got %v", 2, events)
		return
	}

	err = ep.del(efd1.fd)
	if err != nil {
		t.Errorf("epoll del fd=%d: %v", efd1.fd, err)
		return
	}
	err = ep.del(efd2.fd)
	if err != nil {
		t.Errorf("epoll del fd=%d: %v", efd2.fd, err)
		return
	}
	err = efd1.WriteUint(5)
	if err != nil {
		t.Errorf("event fd write: %v", err)
		return
	}
	err = efd2.WriteUint(10)
	if err != nil {
		t.Errorf("event fd write: %v", err)
		return
	}
	events, err = ep.wait(d)
	if err != nil {
		t.Errorf("epoll wait: %v", err)
		return
	}
	if len(events) != 0 {
		t.Errorf("epoll wait expected event num=%d but got %v", 0, events)
		return
	}
}
