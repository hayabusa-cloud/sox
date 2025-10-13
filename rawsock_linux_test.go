// ©Hayabusa Cloud Co., Ltd. 2025. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

//go:build linux

package sox_test

import (
	"sync"
	"testing"
	"time"

	"golang.org/x/sys/unix"
	"hybscloud.com/sox"
)

func TestRawSocket_ReadWrite(t *testing.T) {
	addr0, err := sox.ResolveIPAddr("ip4", "127.0.0.1")
	if err != nil {
		t.Error(err)
		return
	}

	p := []byte{8, 0, 0, 0, 0, 1, 0, 1}
	wg := sync.WaitGroup{}
	skip := false

	wg.Add(1)
	go func() {
		lis, err := sox.ListenRaw4(addr0, unix.IPPROTO_ICMP)
		if err != nil {
			skip = true
			wg.Done()
			return
		}
		defer func(lis *sox.RawConn) {
			_ = lis.Close()
		}(lis)
		wg.Done()

		time.Sleep(100 * time.Millisecond)
		buf := make([]byte, 64)
		n, _, err := lis.RecvFrom(buf)
		if err != nil {
			t.Errorf("recvfrom: %v", err)
			return
		}
		if n > 0 {
			_, err = lis.SendTo(buf[:n], addr0)
			if err != nil {
				t.Errorf("sendto: %v", err)
			}
		}
	}()
	wg.Wait()
	if skip {
		t.Skipf("listen raw4: %v", err)
		return
	}

	addr1, err := sox.ResolveIPAddr("ip4", "127.0.0.2")
	if err != nil {
		t.Error(err)
		return
	}

	conn, err := sox.DialRaw4(addr1, addr0, unix.IPPROTO_ICMP)
	if err != nil {
		t.Skipf("dial raw4: %v", err)
		return
	}
	defer func(conn *sox.RawConn) {
		_ = conn.Close()
	}(conn)

	n, err := conn.Write(p)
	if err != nil {
		t.Error(err)
		return
	}
	if n != len(p) {
		t.Errorf("expected %d bytes but wrote %d", len(p), n)
	}

	buf := make([]byte, 64)
	_, err = conn.Read(buf)
	if err != nil {
		t.Error(err)
		return
	}
}
