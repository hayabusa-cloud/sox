// ©Hayabusa Cloud Co., Ltd. 2024. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

//go:build linux

package sox_test

import (
	"golang.org/x/sys/unix"
	"hybscloud.com/sox"
	"os"
	"testing"
)

func TestSignalFile(t *testing.T) {
	s, err := sox.NewSignalFile()
	if err != nil {
		t.Errorf("new signalfd: %v", err)
		return
	}
	if s.Fd() < 0 {
		t.Errorf("new signalfd: %v", s.Fd())
		return
	}
	go func() {
		_ = unix.Kill(os.Getpid(), unix.SIGINT)
	}()
	sig, _, err := s.ReadSiginfo()
	if err != nil {
		t.Errorf("signal fd read: %v", err)
		return
	}
	if sig != unix.SIGINT {
		t.Errorf("signal fd read expected SIGINT but got %v", sig)
		return
	}
}
