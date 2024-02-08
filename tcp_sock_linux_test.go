// ©Hayabusa Cloud Co., Ltd. 2022. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

//go:build linux

package sox_test

import (
	"bytes"
	"hybscloud.com/sox"
	"io"
	"testing"
)

func TestTCPSocket_ReadWrite(t *testing.T) {
	addr0, err := sox.ResolveTCPAddr("tcp6", "[::1]:8088")
	if err != nil {
		t.Error(err)
		return
	}
	p := []byte("test0123456789")
	wait := make(chan struct{}, 1)
	go func() {
		lis, err := sox.ListenTCP6(addr0)
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
			r := sox.NewMessageReader(conn, sox.MessageOptionsTCPSocket)
			rn, err := r.Read(buf)
			if err != nil {
				t.Errorf("read message: %v", err)
				return
			}
			if rn != len(p) || !bytes.Equal(p, buf[:rn]) {
				t.Errorf("read message expected %s but got %s", p, buf[:rn])
				return
			}
			w := sox.NewMessageWriter(conn, sox.MessageOptionsTCPSocket)
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

	addr1, err := sox.ResolveTCPAddr("tcp6", "[::1]:8089")
	if err != nil {
		t.Error(err)
		return
	}

	<-wait
	conn, err := sox.DialTCP6(addr1, addr0)
	if err != nil {
		t.Error(err)
		return
	}

	for sw := sox.NewParamSpinWait(); !sw.Closed(); sw.Once() {
		w := sox.NewMessageWriter(conn, sox.MessageOptionsTCPSocket)
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
		r := sox.NewMessageReader(conn, sox.MessageOptionsTCPSocket)
		n, err = r.Read(buf)
		if err == sox.ErrTemporarilyUnavailable {
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
