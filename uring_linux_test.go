//go:build linux

package sox

import (
	"bytes"
	"context"
	"golang.org/x/sys/unix"
	"io"
	"os"
	"testing"
	"time"
)

func TestIOUring_DefaultMode(t *testing.T) {
	t.Run("read local file", func(t *testing.T) {
		ur, err := newIoUring(16)
		if err != nil {
			t.Errorf("new io-uring: %v", err)
			return
		}

		defer os.Remove("test_f_direct.txt")
		f, err := os.OpenFile("test_f_direct.txt", os.O_RDWR|os.O_CREATE|unix.O_DIRECT, 0660)
		if err != nil {
			t.Errorf("open file: %v", err)
			return
		}
		defer f.Close()

		s := AlignedMemBlock()
		copy(s, "test0123456789")
		n, err := f.Write(s)
		if err != nil {
			t.Errorf("write file: %v", err)
			return
		}
		if n != len(s) {
			t.Errorf("short write file expected %d but written %d bytes", len(s), n)
			return
		}

		_, err = f.Seek(0, 0)
		if err != nil {
			t.Errorf("seek file: %v", err)
			return
		}

		payload := AlignedMemBlock()
		err = ur.read(context.TODO(), int(f.Fd()), payload)
		if err != nil {
			t.Errorf("submission readv: %v", err)
			return
		}

		err = ur.enter()
		if err != nil {
			t.Errorf("io_ring_enter: %v", err)
			return
		}

		dl := time.Now().Add(time.Second)
		for sw := NewSpinWaiter(); !sw.Closed(); sw.Once() {
			cqe, err := ur.wait()
			if err == ErrTemporarilyUnavailable {
				if time.Now().After(dl) {
					t.Error("write file timeout")
					return
				}
				continue
			}
			if err != nil {
				t.Errorf("wait completion: %v", err)
				return
			}
			if cqe.res < 0 {
				t.Errorf("read file: %v", errFromUnixErrno(unix.Errno(-cqe.res)))
				return
			}
			if cqe.res != int32(len(payload)) {
				t.Errorf("read file: %v", io.ErrShortWrite)
				return
			}
			break
		}

		if !bytes.Equal(payload, s) {
			t.Error("file read write wrong")
			return
		}
	})

	t.Run("write local file", func(t *testing.T) {
		ur, err := newIoUring(16)
		if err != nil {
			t.Errorf("new io-uring: %v", err)
			return
		}

		defer os.Remove("test_f_direct.txt")
		f, err := os.OpenFile("test_f_direct.txt", os.O_RDWR|os.O_CREATE|unix.O_DIRECT, 0660)
		if err != nil {
			t.Errorf("open file: %v", err)
			return
		}
		defer f.Close()

		s := "test0123456789"
		payload := AlignedMemBlock()
		copy(payload, s)
		err = ur.write(context.TODO(), int(f.Fd()), payload, len(payload))
		if err != nil {
			t.Errorf("submission write: %v", err)
			return
		}

		err = ur.enter()
		if err != nil {
			t.Errorf("io_ring_enter: %v", err)
			return
		}

		dl := time.Now().Add(time.Second)
		for sw := NewSpinWaiter(); !sw.Closed(); sw.Once() {
			cqe, err := ur.wait()
			if err == ErrTemporarilyUnavailable {
				if time.Now().After(dl) {
					t.Error("write file timeout")
					return
				}
				continue
			}
			if err != nil {
				t.Errorf("wait completion: %v", err)
				return
			}
			if cqe.res < 0 {
				t.Errorf("write file: %v", errFromUnixErrno(unix.Errno(-cqe.res)))
				return
			}
			if cqe.res != int32(len(payload)) {
				t.Errorf("write file: %v", io.ErrShortWrite)
				return
			}
			break
		}

		_, err = f.Seek(0, 0)
		if err != nil {
			t.Errorf("seek file: %v", err)
			return
		}

		b := AlignedMemBlock()
		_, err = f.Read(b)
		if err != nil && err != io.EOF {
			t.Errorf("read file: %v", err)
			return
		}

		if !bytes.Equal(payload[:len(s)], b[:len(s)]) {
			t.Error("file read write wrong")
			return
		}
	})

	t.Run("read unix socket", func(t *testing.T) {
		ur, err := newIoUring(16)
		if err != nil {
			t.Errorf("new io-uring: %v", err)
			return
		}

		so, err := newUnixSocketPair()
		if err != nil {
			t.Errorf("unix socket pair: %v", err)
			return
		}

		wb := []byte("test0123456789")
		err = unix.Send(so[1].fd, wb, 0)
		if err != nil {
			t.Errorf("socked send: %v", errFromUnixErrno(err))
			return
		}

		rb := make([]byte, len(wb))
		err = ur.receive(context.TODO(), so[0].fd, rb)
		if err != nil {
			t.Errorf("submit recv: %v", err)
			return
		}

		err = ur.enter()
		if err != nil {
			t.Errorf("io_uring enter: %v", err)
			return
		}

		dl := time.Now().Add(time.Second)
		for sw := NewSpinWaiter(); !sw.Closed(); sw.Once() {
			cqe, err := ur.wait()
			if err == ErrTemporarilyUnavailable {
				if time.Now().After(dl) {
					t.Error("read socket timeout")
					return
				}
				continue
			}
			if err != nil {
				t.Errorf("wait completion: %v", err)
				return
			}
			if cqe.res < 0 {
				t.Errorf("socket recv: %v", errFromUnixErrno(unix.Errno(-cqe.res)))
				return
			}
			if cqe.res != int32(len(wb)) {
				t.Errorf("socket recv: %v", io.ErrShortWrite)
				return
			}
			break
		}

		if !bytes.Equal(wb, rb) {
			t.Errorf("socked recv expected %s but got %s", wb, rb)
			return
		}
		return
	})

	t.Run("write unix socket", func(t *testing.T) {
		ur, err := newIoUring(16)
		if err != nil {
			t.Errorf("new io-uring: %v", err)
			return
		}

		so, err := newUnixSocketPair()
		if err != nil {
			t.Errorf("unix socket pair: %v", err)
			return
		}

		wb := []byte("test0123456789")
		err = ur.send(context.TODO(), so[1].fd, wb)
		if err != nil {
			t.Errorf("submit send: %v", err)
			return
		}

		err = ur.enter()
		if err != nil {
			t.Errorf("io_uring enter: %v", err)
			return
		}

		dl := time.Now().Add(time.Second)
		for sw := NewSpinWaiter(); !sw.Closed(); sw.Once() {
			cqe, err := ur.wait()
			if err == ErrTemporarilyUnavailable {
				if time.Now().After(dl) {
					t.Error("write socket timeout")
					return
				}
				continue
			}
			if err != nil {
				t.Errorf("wait completion: %v", err)
				return
			}
			if cqe.res < 0 {
				t.Errorf("socket send: %v", errFromUnixErrno(unix.Errno(-cqe.res)))
				return
			}
			if cqe.res != int32(len(wb)) {
				t.Errorf("socket send: %v", io.ErrShortWrite)
				return
			}
			break
		}

		rb := make([]byte, len(wb))
		n, _, err := unix.Recvfrom(so[0].fd, rb, unix.MSG_WAITALL)
		if err != nil {
			t.Errorf("socked recv: %v", errFromUnixErrno(err))
			return
		}
		if n != len(wb) {
			t.Errorf("socked recv: %v", io.ErrUnexpectedEOF)
			return
		}
		if !bytes.Equal(wb, rb) {
			t.Errorf("socked recv expected %s but got %s", wb, rb)
			return
		}

		return
	})
}
