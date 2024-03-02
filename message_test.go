// ©Hayabusa Cloud Co., Ltd. 2022. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package sox_test

import (
	"bytes"
	"encoding/binary"
	"hybscloud.com/sox"
	"io"
	"testing"
)

func TestMessage_ReadStream(t *testing.T) {
	t.Run("single message", func(t *testing.T) {
		rd, wr := io.Pipe()
		defer rd.Close()
		r := sox.NewMessageReader(rd, func(options *sox.MessageOptions) {
			options.ReadProto = sox.UnderlyingProtocolStream
			options.ReadByteOrder = binary.LittleEndian
			options.Nonblock = false
		})
		s := []byte{0x41, 0x42, 0x43, 0x44}
		h := [1]byte{byte(len(s))}
		go func() {
			defer wr.Close()
			n, err := wr.Write(append(h[:], s...))
			if err != nil {
				t.Errorf("write %d byte(s): %v\n", n, err)
				return
			}
		}()
		buf := make([]byte, len(s))
		n, err := r.Read(buf)
		if err != nil {
			t.Errorf("read %d byte(s): %v\n", n, err)
			return
		}
		if !bytes.Equal(buf[:n], s) {
			t.Errorf("excepted %x but %x\n", s, buf)
			return
		}
	})

	t.Run("multiple messages", func(t *testing.T) {
		rd, wr := io.Pipe()
		defer rd.Close()
		r := sox.NewMessageReader(rd, func(options *sox.MessageOptions) {
			options.ReadProto = sox.UnderlyingProtocolStream
			options.ReadByteOrder = binary.LittleEndian
			options.Nonblock = false
		})
		s := []byte{0x41, 0x42, 0x43, 0x44, 0x45}
		h := [1]byte{byte(len(s))}
		p := append(h[:], s...)
		p = append(p, p...)
		p = append(p, p...)
		go func() {
			defer wr.Close()
			n, err := wr.Write(p[0:3])
			if err != nil {
				t.Errorf("write %d byte(s): %v\n", n, err)
				return
			}
			n, err = wr.Write(p[3:6])
			if err != nil {
				t.Errorf("write %d byte(s): %v\n", n, err)
				return
			}
			n, err = wr.Write(p[6:9])
			if err != nil {
				t.Errorf("write %d byte(s): %v\n", n, err)
				return
			}
			n, err = wr.Write(p[9:12])
			if err != nil {
				t.Errorf("write %d byte(s): %v\n", n, err)
				return
			}
			n, err = wr.Write(p[12:15])
			if err != nil {
				t.Errorf("write %d byte(s): %v\n", n, err)
				return
			}
			n, err = wr.Write(p[15:])
			if err != nil {
				t.Errorf("write %d byte(s): %v\n", n, err)
				return
			}
		}()
		buf := make([]byte, len(s))
		for range 4 {
			n, err := r.Read(buf)
			if err != nil {
				t.Errorf("read %d byte(s): %v\n", n, err)
				return
			}
			if !bytes.Equal(buf[:n], s) {
				t.Errorf("excepted %x but %x\n", s, buf)
				return
			}
		}
	})

	t.Run("many messages", func(t *testing.T) {
		rd, wr := io.Pipe()
		defer rd.Close()
		r := sox.NewMessageReader(rd, func(options *sox.MessageOptions) {
			options.ReadProto = sox.UnderlyingProtocolStream
			options.ReadByteOrder = binary.LittleEndian
			options.Nonblock = false
		})
		s := []byte{0x41, 0x42, 0x43}
		h := [1]byte{byte(len(s))}
		p := append(h[:], s...)
		for range 10 {
			p = append(p, p...)
		}
		go func() {
			defer wr.Close()
			n, err := wr.Write(p[0:3])
			if err != nil {
				t.Errorf("write %d byte(s): %v\n", n, err)
				return
			}
			n, err = wr.Write(p[3:38])
			if err != nil {
				t.Errorf("write %d byte(s): %v\n", n, err)
				return
			}
			n, err = wr.Write(p[38:])
			if err != nil {
				t.Errorf("write %d byte(s): %v\n", n, err)
				return
			}
		}()
		buf := make([]byte, len(s))
		for range 1024 {
			n, err := r.Read(buf)
			if err != nil {
				t.Errorf("read %d byte(s): %v\n", n, err)
				return
			}
			if !bytes.Equal(buf[:n], s) {
				t.Errorf("excepted %x but %x\n", s, buf)
				return
			}
		}
	})

	t.Run("multiple large messages", func(t *testing.T) {
		rd, wr := io.Pipe()
		defer rd.Close()
		r := sox.NewMessageReader(rd, func(options *sox.MessageOptions) {
			options.ReadProto = sox.UnderlyingProtocolStream
			options.ReadByteOrder = binary.LittleEndian
			options.Nonblock = false
		})
		s := []byte{0x41, 0x42, 0x43, 0x44, 0x45, 0x46, 0x47, 0x48}
		s = append(s, s...)
		s = append(s, s...)
		s = append(s, s...)
		s = append(s, s...)
		s = append(s, s...)
		s = append(s, s...)
		s = append(s, s...)
		h := [3]byte{254, 0, 0}
		binary.LittleEndian.PutUint16(h[1:], uint16(len(s)))
		p := append(h[:], s...)
		p = append(p, p...)
		p = append(p, p...)
		go func() {
			defer wr.Close()
			n, err := wr.Write(p[0:15])
			if err != nil {
				t.Errorf("write %d byte(s): %v\n", n, err)
				return
			}
			n, err = wr.Write(p[15:512])
			if err != nil {
				t.Errorf("write %d byte(s): %v\n", n, err)
				return
			}
			n, err = wr.Write(p[512:])
			if err != nil {
				t.Errorf("write %d byte(s): %v\n", n, err)
				return
			}
		}()
		buf := make([]byte, len(s))
		for range 4 {
			n, err := r.Read(buf)
			if err != nil {
				t.Errorf("read %d byte(s): %v\n", n, err)
				return
			}
			if !bytes.Equal(buf[:n], s) {
				t.Errorf("excepted %x but %x\n", s, buf)
				return
			}
		}
	})

	t.Run("multiple huge messages", func(t *testing.T) {
		rd, wr := io.Pipe()
		defer rd.Close()
		r := sox.NewMessageReader(rd, func(options *sox.MessageOptions) {
			options.ReadProto = sox.UnderlyingProtocolStream
			options.ReadByteOrder = binary.LittleEndian
			options.Nonblock = false
		})
		s := []byte{0x41, 0x42, 0x43, 0x44, 0x45, 0x46, 0x47, 0x48}
		for range 16 {
			s = append(s, s...)
		}
		h := [8]byte{255, 0, 0, 0, 0, 0, 0, 0}
		binary.LittleEndian.PutUint64(h[:], uint64(len(s))<<8)
		h[0] = 255
		p := append(h[:], s...)
		p = append(p, p...)
		p = append(p, p...)
		go func() {
			defer wr.Close()
			n, err := wr.Write(p[:])
			if err != nil {
				t.Errorf("write %d byte(s): %v\n", n, err)
				return
			}
		}()
		buf := make([]byte, len(s))
		for range 4 {
			n, err := r.Read(buf)
			if err != nil {
				t.Errorf("read %d byte(s): %v\n", n, err)
				return
			}
			if !bytes.Equal(buf[:n], s) {
				t.Errorf("excepted %x but %x\n", s, buf)
				return
			}
		}
	})
}

func TestMessage_WriteStream(t *testing.T) {
	t.Run("single message", func(t *testing.T) {
		rd, wr := io.Pipe()
		defer rd.Close()
		r := sox.NewMessageReader(rd, func(options *sox.MessageOptions) {
			options.ReadProto = sox.UnderlyingProtocolStream
			options.ReadByteOrder = binary.LittleEndian
			options.Nonblock = false
		})
		w := sox.NewMessageWriter(wr, func(options *sox.MessageOptions) {
			options.WriteProto = sox.UnderlyingProtocolStream
			options.WriteByteOrder = binary.LittleEndian
			options.Nonblock = false
		})
		s := []byte{0x41, 0x42, 0x43, 0x44}
		go func() {
			defer wr.Close()
			n, err := w.Write(s)
			if err != nil {
				t.Errorf("write %d byte(s): %v\n", n, err)
				return
			}
		}()
		buf := make([]byte, len(s))
		n, err := r.Read(buf)
		if err != nil {
			t.Errorf("read %d byte(s): %v\n", n, err)
			return
		}
		if n != len(s) {
			t.Error("read short bytes")
			return
		}
		if !bytes.Equal(buf[:n], s) {
			t.Errorf("excepted %x but %x\n", s, buf)
			return
		}
	})

	t.Run("multiple messages", func(t *testing.T) {
		rd, wr := io.Pipe()
		defer rd.Close()
		r := sox.NewMessageReader(rd, func(options *sox.MessageOptions) {
			options.ReadProto = sox.UnderlyingProtocolStream
			options.ReadByteOrder = binary.LittleEndian
			options.Nonblock = false
		})
		w := sox.NewMessageWriter(wr, func(options *sox.MessageOptions) {
			options.WriteProto = sox.UnderlyingProtocolStream
			options.WriteByteOrder = binary.LittleEndian
			options.Nonblock = false
		})
		s := []byte{0x41, 0x42, 0x43, 0x44}
		go func() {
			defer wr.Close()
			for range 8 {
				n, err := w.Write(s)
				if err != nil {
					t.Errorf("write %d byte(s): %v\n", n, err)
					return
				}
			}
		}()
		buf := make([]byte, len(s))
		for range 8 {
			n, err := r.Read(buf)
			if err != nil {
				t.Errorf("read %d byte(s): %v\n", n, err)
				return
			}
			if n != len(s) {
				t.Error("read short bytes")
				return
			}
			if !bytes.Equal(buf[:n], s) {
				t.Errorf("excepted %x but %x\n", s, buf)
				return
			}
		}
	})

	t.Run("multiple large messages", func(t *testing.T) {
		rd, wr := io.Pipe()
		defer rd.Close()
		r := sox.NewMessageReader(rd, func(options *sox.MessageOptions) {
			options.ReadProto = sox.UnderlyingProtocolStream
			options.ReadByteOrder = binary.LittleEndian
			options.Nonblock = false
		})
		w := sox.NewMessageWriter(wr, func(options *sox.MessageOptions) {
			options.WriteProto = sox.UnderlyingProtocolStream
			options.WriteByteOrder = binary.LittleEndian
			options.Nonblock = false
		})
		s := []byte{0x41, 0x42, 0x43, 0x44, 0x45, 0x46, 0x47, 0x48}
		for range 8 {
			s = append(s, s...)
		}
		go func() {
			defer wr.Close()
			for range 8 {
				n, err := w.Write(s)
				if err != nil {
					t.Errorf("write %d byte(s): %v\n", n, err)
					return
				}
			}
		}()
		buf := make([]byte, len(s))
		for range 8 {
			n, err := r.Read(buf)
			if err != nil {
				t.Errorf("read %d byte(s): %v\n", n, err)
				return
			}
			if n != len(s) {
				t.Error("read short bytes")
				return
			}
			if !bytes.Equal(buf[:n], s) {
				t.Errorf("excepted %x but %x\n", s, buf)
				return
			}
		}
	})

	t.Run("multiple huge messages", func(t *testing.T) {
		rd, wr := io.Pipe()
		defer rd.Close()
		r := sox.NewMessageReader(rd, func(options *sox.MessageOptions) {
			options.ReadProto = sox.UnderlyingProtocolStream
			options.ReadByteOrder = binary.LittleEndian
			options.Nonblock = false
		})
		w := sox.NewMessageWriter(wr, func(options *sox.MessageOptions) {
			options.WriteProto = sox.UnderlyingProtocolStream
			options.WriteByteOrder = binary.LittleEndian
			options.Nonblock = false
		})
		s := []byte{0x41, 0x42, 0x43, 0x44, 0x45, 0x46, 0x47, 0x48}
		for range 16 {
			s = append(s, s...)
		}
		go func() {
			defer wr.Close()
			for range 8 {
				n, err := w.Write(s)
				if err != nil {
					t.Errorf("write %d byte(s): %v\n", n, err)
					return
				}
			}
		}()
		buf := make([]byte, len(s))
		for range 8 {
			n, err := r.Read(buf)
			if err != nil {
				t.Errorf("read %d byte(s): %v\n", n, err)
				return
			}
			if n != len(s) {
				t.Error("read short bytes")
				return
			}
			if !bytes.Equal(buf[:n], s) {
				t.Errorf("excepted %x but %x\n", s, buf)
				return
			}
		}
	})
}

func TestMessage_PipeStream(t *testing.T) {
	t.Run("16 bytes messages", func(t *testing.T) {
		buf := make([]byte, 1024)
		r, w := sox.NewMessagePipe(func(options *sox.MessageOptions) {
			options.ReadProto = sox.UnderlyingProtocolStream
			options.ReadByteOrder = binary.BigEndian
			options.WriteProto = sox.UnderlyingProtocolStream
			options.WriteByteOrder = binary.BigEndian
			options.Nonblock = false
		})
		s := []byte{0x41, 0x42, 0x43, 0x44, 0x45, 0x46, 0x47, 0x48}
		s = append(s, s...)
		go func() {
			for range 64 {
				n, err := w.Write(s)
				if err != nil {
					t.Errorf("write %d byte(s): %v\n", n, err)
					return
				}
			}
		}()

		for range 64 {
			n, err := r.Read(buf)
			if err != nil {
				t.Errorf("read %d byte(s): %v\n", n, err)
				return
			}
			if n != len(s) {
				t.Error("read short bytes")
				return
			}
			if !bytes.Equal(buf[:n], s) {
				t.Errorf("excepted %x but %x\n", s, buf)
				return
			}
		}
	})

	t.Run("253 bytes messages", func(t *testing.T) {
		buf := make([]byte, 1024)
		r, w := sox.NewMessagePipe(func(options *sox.MessageOptions) {
			options.ReadProto = sox.UnderlyingProtocolStream
			options.ReadByteOrder = binary.BigEndian
			options.WriteProto = sox.UnderlyingProtocolStream
			options.WriteByteOrder = binary.BigEndian
			options.Nonblock = false
		})
		s := [253]byte{}
		for i := range len(s) {
			s[i] = 0x40 + byte(i&0xf)
		}
		go func() {
			for range 64 {
				n, err := w.Write(s[:])
				if err != nil {
					t.Errorf("write %d byte(s): %v\n", n, err)
					return
				}
			}
		}()

		for range 64 {
			n, err := r.Read(buf)
			if err != nil {
				t.Errorf("read %d byte(s): %v\n", n, err)
				return
			}
			if n != len(s) {
				t.Error("read short bytes")
				return
			}
			if !bytes.Equal(buf[:n], s[:]) {
				t.Errorf("excepted %x but %x\n", s, buf)
				return
			}
		}
	})

	t.Run("254 bytes messages", func(t *testing.T) {
		buf := make([]byte, 1024)
		r, w := sox.NewMessagePipe(func(options *sox.MessageOptions) {
			options.ReadProto = sox.UnderlyingProtocolStream
			options.ReadByteOrder = binary.BigEndian
			options.WriteProto = sox.UnderlyingProtocolStream
			options.WriteByteOrder = binary.BigEndian
			options.Nonblock = false
		})
		s := [254]byte{}
		for i := range len(s) {
			s[i] = 0x40 + byte(i&0xf)
		}
		go func() {
			for range 64 {
				n, err := w.Write(s[:])
				if err != nil {
					t.Errorf("write %d byte(s): %v\n", n, err)
					return
				}
			}
		}()

		for range 64 {
			n, err := r.Read(buf)
			if err != nil {
				t.Errorf("read %d byte(s): %v\n", n, err)
				return
			}
			if n != len(s) {
				t.Error("read short bytes")
				return
			}
			if !bytes.Equal(buf[:n], s[:]) {
				t.Errorf("excepted %x but %x\n", s, buf)
				return
			}
		}
	})

	t.Run("255 bytes messages", func(t *testing.T) {
		buf := make([]byte, 1024)
		r, w := sox.NewMessagePipe(func(options *sox.MessageOptions) {
			options.ReadProto = sox.UnderlyingProtocolStream
			options.ReadByteOrder = binary.BigEndian
			options.WriteProto = sox.UnderlyingProtocolStream
			options.WriteByteOrder = binary.BigEndian
			options.Nonblock = false
		})
		s := [255]byte{}
		for i := range len(s) {
			s[i] = 0x40 + byte(i&0xf)
		}
		go func() {
			for range 64 {
				n, err := w.Write(s[:])
				if err != nil {
					t.Errorf("write %d byte(s): %v\n", n, err)
					return
				}
			}
		}()

		for range 64 {
			n, err := r.Read(buf)
			if err != nil {
				t.Errorf("read %d byte(s): %v\n", n, err)
				return
			}
			if n != len(s) {
				t.Error("read short bytes")
				return
			}
			if !bytes.Equal(buf[:n], s[:]) {
				t.Errorf("excepted %x but %x\n", s, buf)
				return
			}
		}
	})

	t.Run("256 bytes messages", func(t *testing.T) {
		buf := make([]byte, 1024)
		r, w := sox.NewMessagePipe(func(options *sox.MessageOptions) {
			options.ReadProto = sox.UnderlyingProtocolStream
			options.ReadByteOrder = binary.BigEndian
			options.WriteProto = sox.UnderlyingProtocolStream
			options.WriteByteOrder = binary.BigEndian
			options.Nonblock = false
		})
		s := [256]byte{}
		for i := range len(s) {
			s[i] = 0x40 + byte(i&0xf)
		}
		go func() {
			for range 64 {
				n, err := w.Write(s[:])
				if err != nil {
					t.Errorf("write %d byte(s): %v\n", n, err)
					return
				}
			}
		}()

		for range 64 {
			n, err := r.Read(buf)
			if err != nil {
				t.Errorf("read %d byte(s): %v\n", n, err)
				return
			}
			if n != len(s) {
				t.Error("read short bytes")
				return
			}
			if !bytes.Equal(buf[:n], s[:]) {
				t.Errorf("excepted %x but %x\n", s, buf)
				return
			}
		}
	})

	t.Run("500 bytes messages", func(t *testing.T) {
		buf := make([]byte, 1024)
		r, w := sox.NewMessagePipe(func(options *sox.MessageOptions) {
			options.ReadProto = sox.UnderlyingProtocolStream
			options.ReadByteOrder = binary.BigEndian
			options.WriteProto = sox.UnderlyingProtocolStream
			options.WriteByteOrder = binary.BigEndian
			options.Nonblock = false
		})
		s := [500]byte{}
		for i := range len(s) {
			s[i] = 0x40 + byte(i&0xf)
		}
		go func() {
			for range 64 {
				n, err := w.Write(s[:])
				if err != nil {
					t.Errorf("write %d byte(s): %v\n", n, err)
					return
				}
			}
		}()

		for range 64 {
			n, err := r.Read(buf)
			if err != nil {
				t.Errorf("read %d byte(s): %v\n", n, err)
				return
			}
			if n != len(s) {
				t.Error("read short bytes")
				return
			}
			if !bytes.Equal(buf[:n], s[:]) {
				t.Errorf("excepted %x but %x\n", s, buf)
				return
			}
		}
	})

	t.Run("20000 bytes messages", func(t *testing.T) {
		buf := make([]byte, 20000)
		r, w := sox.NewMessagePipe(func(options *sox.MessageOptions) {
			options.ReadProto = sox.UnderlyingProtocolStream
			options.ReadByteOrder = binary.BigEndian
			options.WriteProto = sox.UnderlyingProtocolStream
			options.WriteByteOrder = binary.BigEndian
			options.Nonblock = false
		})
		s := [20000]byte{}
		for i := range len(s) {
			s[i] = 0x40 + byte(i&0xf)
		}
		go func() {
			for range 64 {
				n, err := w.Write(s[:])
				if err != nil {
					t.Errorf("write %d byte(s): %v\n", n, err)
					return
				}
			}
		}()

		for range 64 {
			n, err := r.Read(buf)
			if err != nil {
				t.Errorf("read %d byte(s): %v\n", n, err)
				return
			}
			if n != len(s) {
				t.Error("read short bytes")
				return
			}
			if !bytes.Equal(buf[:n], s[:]) {
				t.Errorf("excepted %x but %x\n", s, buf)
				return
			}
		}
	})

	t.Run("100000 bytes messages", func(t *testing.T) {
		buf := make([]byte, 100000)
		r, w := sox.NewMessagePipe(func(options *sox.MessageOptions) {
			options.ReadProto = sox.UnderlyingProtocolStream
			options.ReadByteOrder = binary.BigEndian
			options.WriteProto = sox.UnderlyingProtocolStream
			options.WriteByteOrder = binary.BigEndian
			options.Nonblock = false
		})
		s := [100000]byte{}
		for i := range len(s) {
			s[i] = 0x40 + byte(i&0xf)
		}
		go func() {
			for range 64 {
				n, err := w.Write(s[:])
				if err != nil {
					t.Errorf("write %d byte(s): %v\n", n, err)
					return
				}
			}
		}()

		for range 64 {
			n, err := r.Read(buf)
			if err != nil {
				t.Errorf("read %d byte(s): %v\n", n, err)
				return
			}
			if n != len(s) {
				t.Error("read short bytes")
				return
			}
			if !bytes.Equal(buf[:n], s[:]) {
				t.Errorf("excepted %x but %x\n", s, buf)
				return
			}
		}
	})

	t.Run("read limit", func(t *testing.T) {
		buf := make([]byte, 64)
		r, w := sox.NewMessagePipe(func(options *sox.MessageOptions) {
			options.ReadProto = sox.UnderlyingProtocolStream
			options.ReadByteOrder = binary.BigEndian
			options.WriteProto = sox.UnderlyingProtocolStream
			options.WriteByteOrder = binary.BigEndian
			options.ReadLimit = 16
			options.Nonblock = false
		})
		s := [64]byte{}
		for i := range len(s) {
			s[i] = 0x40 + byte(i&0xf)
		}
		go func() {
			n, err := w.Write(s[:])
			if err != nil {
				t.Errorf("write %d byte(s): %v\n", n, err)
				return
			}
		}()

		_, err := r.Read(buf)
		if err != sox.ErrMsgTooLong {
			t.Errorf("read expected %v but got %v\n", sox.ErrMsgTooLong, err)
			return
		}
	})
}

func BenchmarkMessage_Stream(b *testing.B) {
	b.Run("16 bytes message", func(b *testing.B) {
		r, w := sox.NewMessagePipe(func(options *sox.MessageOptions) {
			options.ReadProto = sox.UnderlyingProtocolStream
			options.ReadByteOrder = binary.BigEndian
			options.WriteProto = sox.UnderlyingProtocolStream
			options.WriteByteOrder = binary.BigEndian
			options.Nonblock = false
		})
		benchmarkMessageStream(b, r, w, 16)
	})

	b.Run("64 bytes message", func(b *testing.B) {
		r, w := sox.NewMessagePipe(func(options *sox.MessageOptions) {
			options.ReadProto = sox.UnderlyingProtocolStream
			options.ReadByteOrder = binary.BigEndian
			options.WriteProto = sox.UnderlyingProtocolStream
			options.WriteByteOrder = binary.BigEndian
			options.Nonblock = false
		})
		benchmarkMessageStream(b, r, w, 64)
	})

	b.Run("256 bytes message", func(b *testing.B) {
		r, w := sox.NewMessagePipe(func(options *sox.MessageOptions) {
			options.ReadProto = sox.UnderlyingProtocolStream
			options.ReadByteOrder = binary.BigEndian
			options.WriteProto = sox.UnderlyingProtocolStream
			options.WriteByteOrder = binary.BigEndian
			options.Nonblock = false
		})
		benchmarkMessageStream(b, r, w, 256)
	})

	b.Run("1k bytes message", func(b *testing.B) {
		r, w := sox.NewMessagePipe(func(options *sox.MessageOptions) {
			options.ReadProto = sox.UnderlyingProtocolStream
			options.ReadByteOrder = binary.BigEndian
			options.WriteProto = sox.UnderlyingProtocolStream
			options.WriteByteOrder = binary.BigEndian
			options.Nonblock = false
		})
		benchmarkMessageStream(b, r, w, 1<<10)
	})

	b.Run("4k bytes message", func(b *testing.B) {
		r, w := sox.NewMessagePipe(func(options *sox.MessageOptions) {
			options.ReadProto = sox.UnderlyingProtocolStream
			options.ReadByteOrder = binary.BigEndian
			options.WriteProto = sox.UnderlyingProtocolStream
			options.WriteByteOrder = binary.BigEndian
			options.Nonblock = false
		})
		benchmarkMessageStream(b, r, w, 1<<12)
	})

	b.Run("16k bytes message", func(b *testing.B) {
		r, w := sox.NewMessagePipe(func(options *sox.MessageOptions) {
			options.ReadProto = sox.UnderlyingProtocolStream
			options.ReadByteOrder = binary.BigEndian
			options.WriteProto = sox.UnderlyingProtocolStream
			options.WriteByteOrder = binary.BigEndian
			options.Nonblock = false
		})
		benchmarkMessageStream(b, r, w, 1<<14)
	})

	b.Run("64k bytes message", func(b *testing.B) {
		r, w := sox.NewMessagePipe(func(options *sox.MessageOptions) {
			options.ReadProto = sox.UnderlyingProtocolStream
			options.ReadByteOrder = binary.BigEndian
			options.WriteProto = sox.UnderlyingProtocolStream
			options.WriteByteOrder = binary.BigEndian
			options.Nonblock = false
		})
		benchmarkMessageStream(b, r, w, 1<<16)
	})

	b.Run("256k bytes message", func(b *testing.B) {
		r, w := sox.NewMessagePipe(func(options *sox.MessageOptions) {
			options.ReadProto = sox.UnderlyingProtocolStream
			options.ReadByteOrder = binary.BigEndian
			options.WriteProto = sox.UnderlyingProtocolStream
			options.WriteByteOrder = binary.BigEndian
			options.Nonblock = false
		})
		benchmarkMessageStream(b, r, w, 1<<18)
	})

	b.Run("1m bytes message", func(b *testing.B) {
		r, w := sox.NewMessagePipe(func(options *sox.MessageOptions) {
			options.ReadProto = sox.UnderlyingProtocolStream
			options.ReadByteOrder = binary.BigEndian
			options.WriteProto = sox.UnderlyingProtocolStream
			options.WriteByteOrder = binary.BigEndian
			options.Nonblock = false
		})
		benchmarkMessageStream(b, r, w, 1<<20)
	})

	b.Run("4m bytes message", func(b *testing.B) {
		r, w := sox.NewMessagePipe(func(options *sox.MessageOptions) {
			options.ReadProto = sox.UnderlyingProtocolStream
			options.ReadByteOrder = binary.BigEndian
			options.WriteProto = sox.UnderlyingProtocolStream
			options.WriteByteOrder = binary.BigEndian
			options.Nonblock = false
		})
		benchmarkMessageStream(b, r, w, 1<<22)
	})

	b.Run("16m bytes message", func(b *testing.B) {
		r, w := sox.NewMessagePipe(func(options *sox.MessageOptions) {
			options.ReadProto = sox.UnderlyingProtocolStream
			options.ReadByteOrder = binary.BigEndian
			options.WriteProto = sox.UnderlyingProtocolStream
			options.WriteByteOrder = binary.BigEndian
			options.Nonblock = false
		})
		benchmarkMessageStream(b, r, w, 1<<24)
	})

	b.Run("64m bytes message", func(b *testing.B) {
		r, w := sox.NewMessagePipe(func(options *sox.MessageOptions) {
			options.ReadProto = sox.UnderlyingProtocolStream
			options.ReadByteOrder = binary.BigEndian
			options.WriteProto = sox.UnderlyingProtocolStream
			options.WriteByteOrder = binary.BigEndian
			options.Nonblock = false
		})
		benchmarkMessageStream(b, r, w, 1<<26)
	})

	b.Run("256m bytes message", func(b *testing.B) {
		r, w := sox.NewMessagePipe(func(options *sox.MessageOptions) {
			options.ReadProto = sox.UnderlyingProtocolStream
			options.ReadByteOrder = binary.BigEndian
			options.WriteProto = sox.UnderlyingProtocolStream
			options.WriteByteOrder = binary.BigEndian
			options.Nonblock = false
		})
		benchmarkMessageStream(b, r, w, 1<<28)
	})
}

func benchmarkMessageStream(b *testing.B, r io.Reader, w io.Writer, l int) {
	wBuf, rBuf := make([]byte, l), make([]byte, l)
	go func() {
		for range b.N {
			n, err := w.Write(wBuf)
			if err != nil {
				b.Errorf("write %d byte(s): %v\n", n, err)
				return
			}
		}
	}()

	b.ResetTimer()
	for range b.N {
		n, err := r.Read(rBuf)
		if err != nil {
			b.Errorf("read %d byte(s): %v\n", n, err)
			return
		}
		if n != len(wBuf) {
			b.Error("read short bytes")
			return
		}
	}
}
