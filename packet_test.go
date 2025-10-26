// ©Hayabusa Cloud Co., Ltd. 2024. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package sox_test

import (
	"bytes"
	"code.hybscloud.com/sox"
	"encoding/binary"
	"io"
	"math"
	"testing"
	"unsafe"
)

func TestPacketReader(t *testing.T) {
	fn := func(t *testing.T, order binary.ByteOrder) {
		buf, offset := make([]byte, 0x1000), 0
		// push false
		buf[offset] = 0
		offset++
		// push true
		buf[offset] = 1
		offset++
		// push 0x40
		buf[offset] = 0x40
		offset++
		// push 0x1000
		order.PutUint16(buf[offset:], 0x1000)
		offset += 2
		// push 0x200000
		order.PutUint32(buf[offset:], 0x200000)
		offset += 4
		// push 0x300000000000
		order.PutUint64(buf[offset:], 0x300000000000)
		offset += 8
		// push 3.1416
		f32 := float32(3.1416)
		order.PutUint32(buf[offset:], math.Float32bits(f32))
		offset += 4
		// push 3.1415926535898
		f64 := 3.1415926535898
		order.PutUint64(buf[offset:], math.Float64bits(f64))
		offset += 8
		// push 1.234+5.678i
		c64 := complex(float32(1.234), float32(5.678))
		order.PutUint32(buf[offset:], math.Float32bits(real(c64)))
		offset += 4
		order.PutUint32(buf[offset:], math.Float32bits(imag(c64)))
		offset += 4
		// push 1.23456789+0.123456789i
		c128 := complex(1.23456789, 0.123456789)
		order.PutUint64(buf[offset:], math.Float64bits(real(c128)))
		offset += 8
		order.PutUint64(buf[offset:], math.Float64bits(imag(c128)))
		offset += 8
		// push item
		item1 := &testPacketItem{u64: 10, f64: 3.1415926, s: "sox"}
		n, err := item1.Read(buf[offset:])
		offset += n
		if err != nil {
			t.Errorf("packet item read: %v", err)
			return
		}
		// push non-buffer item
		item2 := &testPacketItemNonBuffer{u64: 10, f64: 3.1415926}
		b := bytes.NewBuffer(nil)
		nn, err := item2.WriteTo(b)
		if err != nil {
			t.Errorf("packet item write to: %v", err)
			return
		}
		n, err = b.Read(buf[offset : offset+int(nn)])
		offset += int(nn)
		if err != nil {
			t.Errorf("packek read buffer: %v", err)
			return
		}
		// push string
		s := "hello sox"
		copy(buf[offset:], s)
		offset += len(s)

		rd := sox.NewPacketReader(bytes.NewBuffer(buf), order)
		bVal, err := rd.ReadBool()
		if err != nil {
			t.Errorf("packet read bool: %v", err)
			return
		}
		if bVal != false {
			t.Errorf("packet read bool expected %v but got %v", false, bVal)
			return
		}
		bVal, err = rd.ReadBool()
		if err != nil {
			t.Errorf("packet read bool: %v", err)
			return
		}
		if bVal != true {
			t.Errorf("packet read bool expected %v but got %v", true, bVal)
			return
		}
		u8Val, err := rd.ReadUint8()
		if err != nil {
			t.Errorf("packet read uint8: %v", err)
			return
		}
		if u8Val != 0x40 {
			t.Errorf("packet read uint8 expected %v but got %v", 0x40, u8Val)
			return
		}
		u16Val, err := rd.ReadUint16()
		if err != nil {
			t.Errorf("packet read uint16: %v", err)
			return
		}
		if u16Val != 0x1000 {
			t.Errorf("packet read uint16 expected %v but got %v", 0x1000, u16Val)
			return
		}
		u32Val, err := rd.ReadUint32()
		if err != nil {
			t.Errorf("packet read uint32: %v", err)
			return
		}
		if u32Val != 0x200000 {
			t.Errorf("packet read uint32 expected %v but got %v", 0x200000, u32Val)
			return
		}
		u64Val, err := rd.ReadUint64()
		if err != nil {
			t.Errorf("packet read uint64: %v", err)
			return
		}
		if u64Val != 0x300000000000 {
			t.Errorf("packet read uint64 expected %v but got %v", 0x300000000000, u64Val)
			return
		}
		f32Val, err := rd.ReadFloat32()
		if err != nil {
			t.Errorf("packet read float32: %v", err)
			return
		}
		if f32Val != 3.1416 {
			t.Errorf("packet read float32 expected %v but got %v", 3.1416, f32Val)
			return
		}
		f64Val, err := rd.ReadFloat64()
		if err != nil {
			t.Errorf("packet read float64: %v", err)
			return
		}
		if f64Val != 3.1415926535898 {
			t.Errorf("packet read float64 expected %v but got %v", 3.1415926535898, f64Val)
			return
		}
		c64Val, err := rd.ReadComplex64()
		if err != nil {
			t.Errorf("packet read complex64: %v", err)
			return
		}
		if c64Val != complex(float32(1.234), float32(5.678)) {
			t.Errorf("packet read complex64 expected %v but got %v", complex(float32(1.234), float32(5.678)), c64Val)
			return
		}
		c128Val, err := rd.ReadComplex128()
		if err != nil {
			t.Errorf("packet read complex128: %v", err)
			return
		}
		if c128Val != complex(1.23456789, 0.123456789) {
			t.Errorf("packet read complex128 expected %v but got %v", complex(1.23456789, 0.123456789), c128Val)
			return
		}
		item1Val := &testPacketItem{u64: 0, f64: 0, s: ""}
		err = rd.ReadObject(item1Val)
		if err != nil {
			t.Errorf("packet read item: %v", err)
			return
		}
		if !item1Val.Equal(item1) {
			t.Errorf("packet read object expected %v but got %v", item1, item1Val)
			return
		}
		item2Val := &testPacketItemNonBuffer{u64: 0, f64: 0}
		err = rd.ReadObject(item2Val)
		if err != nil {
			t.Errorf("packet read item non buffer: %v", err)
			return
		}
		if !item2Val.Equal(item2) {
			t.Errorf("packet read object expected %v but got %v", item2, item2Val)
			return
		}
		sVal, err := rd.ReadFixedString(len(s))
		if err != nil {
			t.Errorf("packet read string: %v", err)
			return
		}
		if sVal != s {
			t.Errorf("packet read string expected %v but got %v", s, sVal)
			return
		}
	}

	t.Run("little endian", func(t *testing.T) {
		fn(t, binary.LittleEndian)
	})

	t.Run("big endian", func(t *testing.T) {
		fn(t, binary.BigEndian)
	})
}

func TestPacketWriter(t *testing.T) {
	fn := func(t *testing.T, order binary.ByteOrder) {
		wr := sox.NewPacketWriter(bytes.NewBuffer(nil), order)
		wr.WriteBool(false).WriteBool(true)
		wr.WriteInt8(-10).WriteUint8(10)
		wr.WriteInt16(-20).WriteUint16(20)
		wr.WriteInt32(-30).WriteUint32(30)
		wr.WriteInt64(-40).WriteUint64(40)
		wr.WriteFloat32(3.1416).WriteFloat64(3.1415926535898)
		wr.WriteComplex64(complex(float32(1.234), float32(5.678)))
		wr.WriteComplex128(complex(1.23456789, 0.123456789))
		item1Val := &testPacketItem{u64: 10, f64: 3.1415926, s: "sox"}
		err := wr.WriteObject(item1Val)
		if err != nil {
			t.Errorf("packet write object: %v", err)
			return
		}
		item2Val := &testPacketItemNonBuffer{u64: 10, f64: 3.1415926}
		err = wr.WriteObject(item2Val)
		if err != nil {
			t.Errorf("packet write object: %v", err)
			return
		}
		buffer := bytes.NewBuffer(nil)
		if _, ok := wr.(io.Reader); !ok {
			t.Errorf("packet writer not implemented io.Reader")
			return
		}
		_, err = io.Copy(buffer, wr.(io.Reader))
		if err != nil {
			t.Errorf("take packet contents: %v", err)
			return
		}
		got := buffer.Bytes()

		expected := make([]byte, 1+1+1+1+2+2+4+4+8+8+4+8+8+16+8+8+3+8+8)
		offset := 0
		expected[offset] = 0
		offset++
		expected[offset] = 1
		offset++
		expected[offset] = 246
		offset++
		expected[offset] = 10
		offset++
		order.PutUint16(expected[offset:offset+2], (1<<16)-20)
		offset += 2
		order.PutUint16(expected[offset:offset+2], 20)
		offset += 2
		order.PutUint32(expected[offset:offset+4], (1<<32)-30)
		offset += 4
		order.PutUint32(expected[offset:offset+4], 30)
		offset += 4
		order.PutUint64(expected[offset:offset+8], (1<<64)-40)
		offset += 8
		order.PutUint64(expected[offset:offset+8], 40)
		offset += 8
		order.PutUint32(expected[offset:offset+4], math.Float32bits(3.1416))
		offset += 4
		order.PutUint64(expected[offset:offset+8], math.Float64bits(3.1415926535898))
		offset += 8
		order.PutUint32(expected[offset:offset+4], math.Float32bits(1.234))
		offset += 4
		order.PutUint32(expected[offset:offset+4], math.Float32bits(5.678))
		offset += 4
		order.PutUint64(expected[offset:offset+8], math.Float64bits(1.23456789))
		offset += 8
		order.PutUint64(expected[offset:offset+8], math.Float64bits(0.123456789))
		offset += 8
		u64, f64 := 10, 3.1415926
		copy(expected[offset:offset+8], (*[8]byte)(unsafe.Pointer(&u64))[:])
		offset += 8
		copy(expected[offset:offset+8], (*[8]byte)(unsafe.Pointer(&f64))[:])
		offset += 8
		copy(expected[offset:offset+3], "sox")
		offset += 3
		copy(expected[offset:offset+8], (*[8]byte)(unsafe.Pointer(&u64))[:])
		offset += 8
		copy(expected[offset:offset+8], (*[8]byte)(unsafe.Pointer(&f64))[:])
		offset += 8

		if !bytes.Equal(got, expected) {
			t.Errorf("packet contents expected %v but got %v", expected, got)
			return
		}
	}

	t.Run("little endian", func(t *testing.T) {
		fn(t, binary.LittleEndian)
	})

	t.Run("big endian", func(t *testing.T) {
		fn(t, binary.BigEndian)
	})
}

func TestPacketReaderWriter(t *testing.T) {
	fn := func(t *testing.T, order binary.ByteOrder) {
		// write
		wr := sox.NewPacketWriter(bytes.NewBuffer(nil), order)
		wr.WriteBool(false).WriteBool(true)
		wr.WriteInt8(-10).WriteUint8(10)
		wr.WriteInt16(-20).WriteUint16(20)
		wr.WriteInt32(-30).WriteUint32(30)
		wr.WriteInt64(-40).WriteUint64(40)
		wr.WriteFloat32(3.1416).WriteFloat64(3.1415926535898)
		wr.WriteComplex64(complex(float32(1.234), float32(5.678)))
		wr.WriteComplex128(complex(1.23456789, 0.123456789))
		item1 := &testPacketItem{u64: 10, f64: 3.1415926, s: "sox"}
		err := wr.WriteObject(item1)
		if err != nil {
			t.Errorf("packet write object: %v", err)
			return
		}
		item2 := &testPacketItemNonBuffer{u64: 10, f64: 3.1415926}
		err = wr.WriteObject(item2)
		if err != nil {
			t.Errorf("packet write object: %v", err)
			return
		}
		_, err = wr.WriteString("hello sox")
		if err != nil {
			t.Errorf("packet write string: %v", err)
			return
		}

		// read
		rd := wr.Reader()
		bVal, err := rd.ReadBool()
		if err != nil {
			t.Errorf("packet read bool: %v", err)
			return
		}
		if bVal != false {
			t.Errorf("packet read bool expected %v but got %v", false, bVal)
			return
		}
		bVal, err = rd.ReadBool()
		if err != nil {
			t.Errorf("packet read bool: %v", err)
			return
		}
		if bVal != true {
			t.Errorf("packet read bool expected %v but got %v", true, bVal)
			return
		}
		i8Val, err := rd.ReadInt8()
		if err != nil {
			t.Errorf("packet read int8: %v", err)
			return
		}
		if i8Val != -10 {
			t.Errorf("packet read int8 expected %v but got %v", -10, i8Val)
			return
		}
		u8Val, err := rd.ReadUint8()
		if err != nil {
			t.Errorf("packet read uint8: %v", err)
			return
		}
		if u8Val != 10 {
			t.Errorf("packet read uint8 expected %v but got %v", 10, u8Val)
			return
		}
		i16Val, err := rd.ReadInt16()
		if err != nil {
			t.Errorf("packet read int16: %v", err)
			return
		}
		if i16Val != -20 {
			t.Errorf("packet read int16 expected %v but got %v", -20, i16Val)
			return
		}
		u16Val, err := rd.ReadUint16()
		if err != nil {
			t.Errorf("packet read uint16: %v", err)
			return
		}
		if u16Val != 20 {
			t.Errorf("packet read uint16 expected %v but got %v", 20, u16Val)
			return
		}
		i32Val, err := rd.ReadInt32()
		if err != nil {
			t.Errorf("packet read int32: %v", err)
			return
		}
		if i32Val != -30 {
			t.Errorf("packet read int32 expected %v but got %v", -30, i32Val)
			return
		}
		u32Val, err := rd.ReadUint32()
		if err != nil {
			t.Errorf("packet read uint32: %v", err)
			return
		}
		if u32Val != 30 {
			t.Errorf("packet read uint32 expected %v but got %v", 30, u32Val)
			return
		}
		i64Val, err := rd.ReadInt64()
		if err != nil {
			t.Errorf("packet read int64: %v", err)
			return
		}
		if i64Val != -40 {
			t.Errorf("packet read int64 expected %v but got %v", -40, i64Val)
			return
		}
		u64Val, err := rd.ReadUint64()
		if err != nil {
			t.Errorf("packet read uint64: %v", err)
			return
		}
		if u64Val != 40 {
			t.Errorf("packet read uint64 expected %v but got %v", 40, u64Val)
			return
		}
		f32Val, err := rd.ReadFloat32()
		if err != nil {
			t.Errorf("packet read float32: %v", err)
			return
		}
		if f32Val != 3.1416 {
			t.Errorf("packet read float32 expected %v but got %v", 3.1416, f32Val)
			return
		}
		f64Val, err := rd.ReadFloat64()
		if err != nil {
			t.Errorf("packet read float64: %v", err)
			return
		}
		if f64Val != 3.1415926535898 {
			t.Errorf("packet read float64 expected %v but got %v", 3.1415926535898, f64Val)
			return
		}
		c64Val, err := rd.ReadComplex64()
		if err != nil {
			t.Errorf("packet read complex64: %v", err)
			return
		}
		if c64Val != complex(float32(1.234), float32(5.678)) {
			t.Errorf("packet read complex64 expected %v but got %v", complex(float32(1.234), float32(5.678)), c64Val)
			return
		}
		c128Val, err := rd.ReadComplex128()
		if err != nil {
			t.Errorf("packet read complex128: %v", err)
			return
		}
		if c128Val != complex(1.23456789, 0.123456789) {
			t.Errorf("packet read complex128 expected %v but got %v", complex(1.23456789, 0.123456789), c128Val)
			return
		}
		item1Val := &testPacketItem{u64: 0, f64: 0, s: ""}
		err = rd.ReadObject(item1Val)
		if err != nil {
			t.Errorf("packet read item: %v", err)
			return
		}
		if !item1Val.Equal(item1) {
			t.Errorf("packet read object expected %v but got %v", item1, item1Val)
			return
		}
		item2Val := &testPacketItemNonBuffer{u64: 0, f64: 0}
		err = rd.ReadObject(item2Val)
		if err != nil {
			t.Errorf("packet read item non buffer: %v", err)
			return
		}
		if !item2Val.Equal(item2) {
			t.Errorf("packet read object expected %v but got %v", item2, item2Val)
			return
		}
		sVal, err := rd.ReadFixedString(len("hello sox"))
		if err != nil {
			t.Errorf("packet read string: %v", err)
			return
		}
		if sVal != "hello sox" {
			t.Errorf("packet read string expected %v but got %v", "hello sox", sVal)
			return
		}

		// write
		wr = rd.Writer()
		wr.WriteUint64(2000)
		wr.Reset()
		wr.WriteUint64(1000)

		// read
		rd = wr.Reader()
		u64Val, err = rd.ReadUint64()
		if err != nil {
			t.Errorf("packet read uint64: %v", err)
			return
		}
		if u64Val != 1000 {
			t.Errorf("packet read uint64 expected %v but got %v", 1000, u64Val)
			return
		}
	}

	t.Run("little endian", func(t *testing.T) {
		fn(t, binary.LittleEndian)
	})

	t.Run("big endian", func(t *testing.T) {
		fn(t, binary.BigEndian)
	})
}

func BenchmarkPacketReader(b *testing.B) {
	fn := func(b *testing.B, size int) {
		contents := make([]byte, size*b.N)
		for i := 0; i < size*b.N; i += 16 {
			binary.BigEndian.PutUint64(contents[i:i+8], 10)
			binary.BigEndian.PutUint64(contents[i+8:i+16], math.Float64bits(3.1415926))
		}
		b.ResetTimer()
		wr := sox.NewPacketReader(bytes.NewBuffer(contents[:]), binary.BigEndian)
		for range b.N {
			for i := 0; i < size; i += 16 {
				uVal, err := wr.ReadUint64()
				if err != nil {
					b.Error(err)
					return
				}
				if uVal != 10 {
					b.Error("bad value")
					return
				}
				fVal, err := wr.ReadFloat64()
				if err != nil {
					b.Error(err)
					return
				}
				if fVal != 3.1415926 {
					b.Error("bad value")
					return
				}
			}
		}
	}

	b.Run("16 bytes packet", func(b *testing.B) {
		fn(b, 16)
	})

	b.Run("64 bytes packet", func(b *testing.B) {
		fn(b, 64)
	})

	b.Run("256 bytes packet", func(b *testing.B) {
		fn(b, 256)
	})

	b.Run("1k bytes packet", func(b *testing.B) {
		fn(b, 1<<10)
	})

	b.Run("4k bytes packet", func(b *testing.B) {
		fn(b, 1<<12)
	})
}

func BenchmarkPacketWriter(b *testing.B) {
	fn := func(b *testing.B, size int) {
		b.ResetTimer()
		wr := sox.NewPacketWriter(bytes.NewBuffer(nil), binary.BigEndian)
		for range b.N {
			for i := 0; i < size; i += 16 {
				wr.WriteUint64(10).WriteFloat64(3.1415926)
			}
		}
		wr.Reset()
	}

	b.Run("16 bytes packet", func(b *testing.B) {
		fn(b, 16)
	})

	b.Run("64 bytes packet", func(b *testing.B) {
		fn(b, 64)
	})

	b.Run("256 bytes packet", func(b *testing.B) {
		fn(b, 256)
	})

	b.Run("1k bytes packet", func(b *testing.B) {
		fn(b, 1<<10)
	})

	b.Run("4k bytes packet", func(b *testing.B) {
		fn(b, 1<<12)
	})
}

type testPacketItem struct {
	u64    uint64
	f64    float64
	s      string
	offset int
}

func (i *testPacketItem) Read(buf []byte) (int, error) {
	if i.offset >= 8+8+len(i.s) {
		return 0, io.EOF
	}
	if len(buf) < 8+8+len(i.s)-i.offset {
		return 0, io.ErrShortBuffer
	}
	copy(buf[0:], (*[8]byte)(unsafe.Pointer(&i.u64))[:])
	i.offset += 8
	copy(buf[8:], (*[8]byte)(unsafe.Pointer(&i.f64))[:])
	i.offset += 8
	copy(buf[16:], i.s)
	i.offset += len(i.s)
	return 8 + 8 + len(i.s), nil
}

func (i *testPacketItem) Write(buf []byte) (int, error) {
	if i.offset >= 8+8+len("sox") {
		return 0, nil
	}
	if len(buf) < 8+8+len("sox")-i.offset {
		return 0, io.ErrShortWrite
	}
	i.u64 = *(*uint64)(unsafe.Pointer(&buf[0]))
	i.f64 = *(*float64)(unsafe.Pointer(&buf[8]))
	i.s = string(buf[16 : 16+len("sox")])
	i.offset += 8 + 8 + len(i.s)
	return 8 + 8 + len(i.s), nil
}

func (i *testPacketItem) Equal(other *testPacketItem) bool {
	if i.u64 != other.u64 {
		return false
	}
	if i.f64 != other.f64 {
		return false
	}
	return i.s == other.s
}

type testPacketItemNonBuffer struct {
	u64    uint64
	f64    float64
	offset int
}

func (i *testPacketItemNonBuffer) Read(buf []byte) (int, error) {
	panic("should enter WriteTo but entered Read branch")
}

func (i *testPacketItemNonBuffer) Write(buf []byte) (int, error) {
	panic("should enter ReadFrom but entered Write branch")
}

func (i *testPacketItemNonBuffer) Equal(other *testPacketItemNonBuffer) bool {
	if i.u64 != other.u64 {
		return false
	}
	return i.f64 == other.f64
}

func (i *testPacketItemNonBuffer) WriteTo(w io.Writer) (n int64, err error) {
	if i.offset >= 16 {
		return 0, nil
	}
	nn := 0
	if i.offset == 0 {
		nn, err = w.Write((*[8]byte)(unsafe.Pointer(&i.u64))[:])
		n += int64(nn)
		i.offset += nn
		if err != nil {
			return
		}
	}
	if i.offset == 8 {
		nn, err = w.Write((*[8]byte)(unsafe.Pointer(&i.f64))[:])
		n += int64(nn)
		i.offset += nn
		if err != nil {
			return
		}
	}
	return
}

func (i *testPacketItemNonBuffer) ReadFrom(r io.Reader) (n int64, err error) {
	if i.offset >= 16 {
		return 0, nil
	}
	nn := 0
	if i.offset == 0 {
		nn, err = r.Read((*[8]byte)(unsafe.Pointer(&i.u64))[:])
		n += int64(nn)
		i.offset += nn
		if err != nil {
			return
		}
	}
	if i.offset == 8 {
		nn, err = r.Read((*[8]byte)(unsafe.Pointer(&i.f64))[:])
		n += int64(nn)
		i.offset += nn
	}
	return
}
