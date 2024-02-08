// ©Hayabusa Cloud Co., Ltd. 2022. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

//go:build linux

package sox_test

import (
	"bytes"
	"hybscloud.com/sox"
	"io"
	"runtime"
	"testing"
)

func TestUDPSocket_ReadWrite(t *testing.T) {
	addr0, err := sox.ResolveUDPAddr("udp6", "[::1]:8088")
	if err != nil {
		t.Error(err)
		return
	}
	p := []byte("test0123456789")
	wait := make(chan struct{}, 1)
	go func() {
		conn, err := sox.ListenUDP6(addr0)
		if err != nil {
			t.Error(err)
			return
		}
		buf := make([]byte, len(p))
		wait <- struct{}{}
		for {
			_, addr, err := conn.RecvFrom(buf)
			if err == sox.ErrTemporarilyUnavailable {
				runtime.Gosched()
				continue
			}
			if err != nil {
				t.Error(err)
				return
			}
			_, err = conn.SendTo(buf, addr)
			if err != nil {
				t.Error(err)
				return
			}
		}
	}()

	addr1, err := sox.ResolveUDPAddr("udp6", "[::1]:8089")
	if err != nil {
		t.Error(err)
		return
	}

	<-wait
	conn, err := sox.DialUDP6(addr1, addr0)
	if err != nil {
		t.Error(err)
		return
	}

	for {
		n, err := conn.Write(p)
		if err != nil {
			t.Error(err)
			return
		}
		if n != len(p) {
			t.Error(io.ErrShortWrite)
			return
		}

		buf := make([]byte, len(p))
		n, err = conn.Read(buf)
		if err == sox.ErrTemporarilyUnavailable {
			runtime.Gosched()
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
