package sox

import (
	"bytes"
	"encoding/binary"
	"io"
	"testing"
)

func TestMessage_ReadStream(t *testing.T) {
	t.Run("single message", func(t *testing.T) {
		rd, wr := io.Pipe()
		defer rd.Close()
		r := NewMessageReader(rd, func(options *MessageOptions) {
			options.ReadProto = UnderlyingProtocolStream
			options.ReadByteOrder = binary.LittleEndian
			options.Nonblock = false
		})
		s := []byte{0x41, 0x42, 0x43, 0x44}
		h := []byte{0, 0}
		binary.LittleEndian.PutUint16(h, uint16(len(s)))
		go func() {
			defer wr.Close()
			n, err := wr.Write(append(h, s...))
			if err != nil {
				t.Errorf("write %d byte(s): %v\n", n, err)
				return
			}
		}()
		buf := make([]byte, 64)
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
	t.Run("multiple messages on small packets", func(t *testing.T) {
		rd, wr := io.Pipe()
		defer rd.Close()
		r := NewMessageReader(rd, func(options *MessageOptions) {
			options.ReadProto = UnderlyingProtocolStream
			options.ReadByteOrder = binary.LittleEndian
			options.Nonblock = false
		})
		s := []byte{0x41, 0x42, 0x43, 0x44, 0x45}
		h := []byte{0, 0}
		binary.LittleEndian.PutUint16(h, uint16(len(s)))
		p := append(h, s...)
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
		buf := make([]byte, 64)
		for i := 0; i < 4; i++ {
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
	t.Run("multiple messages on large packets", func(t *testing.T) {
		rd, wr := io.Pipe()
		defer rd.Close()
		r := NewMessageReader(rd, func(options *MessageOptions) {
			options.ReadProto = UnderlyingProtocolStream
			options.ReadByteOrder = binary.LittleEndian
			options.Nonblock = false
		})
		s := []byte{0x41, 0x42, 0x43}
		h := []byte{0, 0}
		binary.LittleEndian.PutUint16(h, uint16(len(s)))
		p := append(h, s...)
		p = append(p, p...)
		p = append(p, p...)
		p = append(p, p...)
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
		buf := make([]byte, 64)
		for i := 0; i < 8; i++ {
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
		r := NewMessageReader(rd, func(options *MessageOptions) {
			options.ReadProto = UnderlyingProtocolStream
			options.ReadByteOrder = binary.LittleEndian
			options.Nonblock = false
		})
		w := NewMessageWriter(wr, func(options *MessageOptions) {
			options.WriteProto = UnderlyingProtocolStream
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
		buf := make([]byte, 64)
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
		r := NewMessageReader(rd, func(options *MessageOptions) {
			options.ReadProto = UnderlyingProtocolStream
			options.ReadByteOrder = binary.LittleEndian
			options.Nonblock = false
		})
		w := NewMessageWriter(wr, func(options *MessageOptions) {
			options.WriteProto = UnderlyingProtocolStream
			options.WriteByteOrder = binary.LittleEndian
			options.Nonblock = false
		})
		s := []byte{0x41, 0x42, 0x43, 0x44}
		go func() {
			defer wr.Close()
			for i := 0; i < 8; i++ {
				n, err := w.Write(s)
				if err != nil {
					t.Errorf("write %d byte(s): %v\n", n, err)
					return
				}
			}
		}()
		buf := make([]byte, 64)
		for i := 0; i < 8; i++ {
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
	t.Run("multiple messages", func(t *testing.T) {
		buf := make([]byte, 1024)
		r, w := NewMessagePipe(func(options *MessageOptions) {
			options.ReadProto = UnderlyingProtocolStream
			options.ReadByteOrder = binary.BigEndian
			options.WriteProto = UnderlyingProtocolStream
			options.WriteByteOrder = binary.BigEndian
			options.Nonblock = false
		})
		s := []byte{0x41, 0x42, 0x43, 0x44}
		go func() {
			for i := 0; i < 64; i++ {
				n, err := w.Write(s)
				if err != nil {
					t.Errorf("write %d byte(s): %v\n", n, err)
					return
				}
			}
		}()

		for i := 0; i < 64; i++ {
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
