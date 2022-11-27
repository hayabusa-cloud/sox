//go:build linux

package sox

import (
	"bytes"
	"io"
	"testing"
)

func TestTCPSocket_ReadWrite(t *testing.T) {
	addr0, err := ResolveTCPAddr("tcp6", "[::1]:8088")
	if err != nil {
		t.Error(err)
		return
	}
	p := []byte("test0123456789")
	wait := make(chan struct{}, 1)
	go func() {
		lis, err := ListenTCP6(addr0)
		if err != nil {
			t.Error(err)
			return
		}
		wait <- struct{}{}
		conn, err := lis.Accept()
		if err != nil {
			t.Error(err)
			return
		}
		buf := make([]byte, len(p))
		for {
			r := NewMessageReader(conn, MessageOptionsTCPSocket)
			rn, err := r.Read(buf)
			if err != nil {
				t.Errorf("read message: %v", err)
				return
			}
			if rn != len(p) || !bytes.Equal(p, buf[:rn]) {
				t.Errorf("read message expected %s but got %s", p, buf[:rn])
				return
			}
			w := NewMessageWriter(conn, MessageOptionsTCPSocket)
			wn, err := w.Write(buf[:rn])
			if err != nil {
				t.Errorf("write message: %v", err)
				return
			}
			if wn != rn {
				t.Errorf("short write")
				return
			}
			break
		}
	}()

	addr1, err := ResolveTCPAddr("tcp6", "[::1]:8089")
	if err != nil {
		t.Error(err)
		return
	}

	<-wait
	conn, err := DialTCP6(addr1, addr0)
	if err != nil {
		t.Error(err)
		return
	}

	for {
		w := NewMessageWriter(conn, MessageOptionsTCPSocket)
		n, err := w.Write(p)
		if err != nil {
			t.Error(err)
			return
		}
		if n != len(p) {
			t.Error(io.ErrShortWrite)
			return
		}

		buf := make([]byte, len(p))
		r := NewMessageReader(conn, MessageOptionsTCPSocket)
		n, err = r.Read(buf)
		if err == ErrTemporarilyUnavailable {
			Yield(jiffies)
			continue
		}
		if err != nil {
			t.Error(err)
			return
		}

		if !bytes.Equal(buf, p) {
			t.Errorf("udp read expected %s but %s", p, buf)
			return
		}
		break
	}
}
