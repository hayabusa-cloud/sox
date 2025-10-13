// ©Hayabusa Cloud Co., Ltd. 2024. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package sox

import (
	"bytes"
	"encoding/binary"
	"io"
	"math"
)

// PacketReader is the interface that wraps common used methods
// that reads numerics, byte slices, strings etc.
type PacketReader interface {
	io.Reader
	io.WriterTo
	io.ByteReader
	io.RuneReader
	ReadBool() (bool, error)
	ReadInt8() (int8, error)
	ReadUint8() (uint8, error)
	ReadInt16() (int16, error)
	ReadUint16() (uint16, error)
	ReadInt32() (int32, error)
	ReadUint32() (uint32, error)
	ReadInt64() (int64, error)
	ReadUint64() (uint64, error)
	ReadFloat32() (float32, error)
	ReadFloat64() (float64, error)
	ReadComplex64() (complex64, error)
	ReadComplex128() (complex128, error)
	ReadObject(object io.Writer) error
	ReadBytes(delim byte) (line []byte, err error)
	ReadString(delim byte) (line string, err error)
	ReadFixedBytes(n int) ([]byte, error)
	ReadFixedString(n int) (string, error)
	Writer() PacketWriter
}

// PacketWriter is the interface that wraps common used methods
// that writes or puts numerics, byte slices, strings etc.
type PacketWriter interface {
	io.Writer
	io.ReaderFrom
	io.ByteWriter
	WriteRune(r rune) (n int, err error)
	WriteBool(val bool) PacketWriter
	WriteInt8(val int8) PacketWriter
	WriteUint8(val uint8) PacketWriter
	WriteInt16(val int16) PacketWriter
	WriteUint16(val uint16) PacketWriter
	WriteInt32(val int32) PacketWriter
	WriteUint32(val uint32) PacketWriter
	WriteInt64(val int64) PacketWriter
	WriteUint64(val uint64) PacketWriter
	WriteFloat32(val float32) PacketWriter
	WriteFloat64(val float64) PacketWriter
	WriteComplex64(val complex64) PacketWriter
	WriteComplex128(val complex128) PacketWriter
	WriteObject(object io.Reader) error
	WriteString(s string) (n int, err error)
	Reset()
	Reader() PacketReader
}

// NewPacketReader creates and returns a new PacketReader with the given buffer and byte order
func NewPacketReader(buffer *bytes.Buffer, order binary.ByteOrder) PacketReader {
	return &packetReader{
		Buffer: buffer,
		bo:     order,
	}
}

// NewPacketWriter creates and returns a new PacketWriter with the given buffer and byte order
func NewPacketWriter(buffer *bytes.Buffer, order binary.ByteOrder) PacketWriter {
	return &packetWriter{
		Buffer: buffer,
		bo:     order,
	}
}

type packetReader struct {
	*bytes.Buffer
	bo binary.ByteOrder
}

func (rd *packetReader) ReadBool() (bool, error) {
	val, err := rd.ReadByte()
	if val != 0 {
		return true, err
	}
	return false, err
}

func (rd *packetReader) ReadInt8() (int8, error) {
	val, err := rd.ReadByte()
	return int8(val), err
}

func (rd *packetReader) ReadUint8() (uint8, error) {
	return rd.ReadByte()
}

func (rd *packetReader) ReadInt16() (int16, error) {
	if rd.Buffer.Len() == 0 {
		return 0, io.EOF
	}
	if rd.Buffer.Len() < 2 {
		return 0, io.ErrShortBuffer
	}
	b := rd.Buffer.Next(2)
	return int16(rd.bo.Uint16(b)), nil
}

func (rd *packetReader) ReadUint16() (uint16, error) {
	if rd.Buffer.Len() == 0 {
		return 0, io.EOF
	}
	if rd.Buffer.Len() < 2 {
		return 0, io.ErrShortBuffer
	}
	b := rd.Buffer.Next(2)
	return rd.bo.Uint16(b), nil
}

func (rd *packetReader) ReadInt32() (int32, error) {
	if rd.Buffer.Len() == 0 {
		return 0, io.EOF
	}
	if rd.Buffer.Len() < 4 {
		return 0, io.ErrShortBuffer
	}
	b := rd.Buffer.Next(4)
	return int32(rd.bo.Uint32(b)), nil
}

func (rd *packetReader) ReadUint32() (uint32, error) {
	if rd.Buffer.Len() == 0 {
		return 0, io.EOF
	}
	if rd.Buffer.Len() < 4 {
		return 0, io.ErrShortBuffer
	}
	b := rd.Buffer.Next(4)
	return rd.bo.Uint32(b), nil
}

func (rd *packetReader) ReadInt64() (int64, error) {
	if rd.Buffer.Len() == 0 {
		return 0, io.EOF
	}
	if rd.Buffer.Len() < 8 {
		return 0, io.ErrShortBuffer
	}
	b := rd.Buffer.Next(8)
	return int64(rd.bo.Uint64(b)), nil
}

func (rd *packetReader) ReadUint64() (uint64, error) {
	if rd.Buffer.Len() == 0 {
		return 0, io.EOF
	}
	if rd.Buffer.Len() < 8 {
		return 0, io.ErrShortBuffer
	}
	b := rd.Buffer.Next(8)
	return rd.bo.Uint64(b), nil
}

func (rd *packetReader) ReadFloat32() (float32, error) {
	u32, err := rd.ReadUint32()
	if err != nil {
		return 0, err
	}
	f32 := math.Float32frombits(u32)
	return f32, nil
}

func (rd *packetReader) ReadFloat64() (float64, error) {
	u64, err := rd.ReadUint64()
	if err != nil {
		return 0, err
	}
	f64 := math.Float64frombits(u64)
	return f64, nil
}

func (rd *packetReader) ReadComplex64() (complex64, error) {
	if rd.Buffer.Len() == 0 {
		return 0, io.EOF
	}
	if rd.Buffer.Len() < 8 {
		return 0, io.ErrShortBuffer
	}
	r := rd.bo.Uint32(rd.Buffer.Next(4))
	i := rd.bo.Uint32(rd.Buffer.Next(4))
	c := complex(math.Float32frombits(r), math.Float32frombits(i))
	return c, nil
}

func (rd *packetReader) ReadComplex128() (complex128, error) {
	if rd.Buffer.Len() == 0 {
		return 0, io.EOF
	}
	if rd.Buffer.Len() < 16 {
		return 0, io.ErrShortBuffer
	}
	r := rd.bo.Uint64(rd.Buffer.Next(8))
	i := rd.bo.Uint64(rd.Buffer.Next(8))
	c := complex(math.Float64frombits(r), math.Float64frombits(i))
	return c, nil
}

func (rd *packetReader) ReadObject(object io.Writer) error {
	if object, ok := object.(io.ReaderFrom); ok {
		_, err := object.ReadFrom(rd.Buffer)
		return err
	}
	n, err := object.Write(rd.Buffer.Bytes())
	rd.Buffer.Next(n)
	return err
}

func (rd *packetReader) ReadFixedBytes(n int) ([]byte, error) {
	if rd.Buffer.Len() == 0 {
		return []byte{}, io.EOF
	}
	if rd.Buffer.Len() < n {
		return []byte{}, io.ErrShortBuffer
	}
	return rd.Buffer.Next(n), nil
}

func (rd *packetReader) ReadFixedString(n int) (string, error) {
	b, err := rd.ReadFixedBytes(n)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

// Writer exports a new PacketWriter with the same buffer and byte order
func (rd *packetReader) Writer() PacketWriter {
	return &packetWriter{
		Buffer: rd.Buffer,
		bo:     rd.bo,
		buf:    [16]byte{},
	}
}

type packetWriter struct {
	*bytes.Buffer
	bo  binary.ByteOrder
	buf [16]byte
}

func (wr *packetWriter) WriteBool(val bool) PacketWriter {
	var b byte = 0
	if val {
		b = 1
	}
	_ = wr.WriteByte(b)
	return wr
}

func (wr *packetWriter) WriteInt8(val int8) PacketWriter {
	_ = wr.WriteByte(byte(val))
	return wr
}

func (wr *packetWriter) WriteUint8(val uint8) PacketWriter {
	_ = wr.WriteByte(val)
	return wr
}

func (wr *packetWriter) WriteInt16(val int16) PacketWriter {
	wr.bo.PutUint16(wr.buf[:2], uint16(val))
	_, _ = wr.Buffer.Write(wr.buf[:2])
	return wr
}

func (wr *packetWriter) WriteUint16(val uint16) PacketWriter {
	wr.bo.PutUint16(wr.buf[:2], val)
	_, _ = wr.Buffer.Write(wr.buf[:2])
	return wr
}

func (wr *packetWriter) WriteInt32(val int32) PacketWriter {
	wr.bo.PutUint32(wr.buf[:4], uint32(val))
	_, _ = wr.Buffer.Write(wr.buf[:4])
	return wr
}

func (wr *packetWriter) WriteUint32(val uint32) PacketWriter {
	wr.bo.PutUint32(wr.buf[:4], val)
	_, _ = wr.Buffer.Write(wr.buf[:4])
	return wr
}

func (wr *packetWriter) WriteInt64(val int64) PacketWriter {
	wr.bo.PutUint64(wr.buf[:8], uint64(val))
	_, _ = wr.Buffer.Write(wr.buf[:8])
	return wr
}

func (wr *packetWriter) WriteUint64(val uint64) PacketWriter {
	wr.bo.PutUint64(wr.buf[:8], val)
	_, _ = wr.Buffer.Write(wr.buf[:8])
	return wr
}

func (wr *packetWriter) WriteFloat32(val float32) PacketWriter {
	wr.bo.PutUint32(wr.buf[:4], math.Float32bits(val))
	_, _ = wr.Buffer.Write(wr.buf[:4])
	return wr
}

func (wr *packetWriter) WriteFloat64(val float64) PacketWriter {
	wr.bo.PutUint64(wr.buf[:8], math.Float64bits(val))
	_, _ = wr.Buffer.Write(wr.buf[:8])
	return wr
}

func (wr *packetWriter) WriteComplex64(val complex64) PacketWriter {
	wr.bo.PutUint32(wr.buf[:4], math.Float32bits(real(val)))
	wr.bo.PutUint32(wr.buf[4:8], math.Float32bits(imag(val)))
	_, _ = wr.Buffer.Write(wr.buf[:8])
	return wr
}

func (wr *packetWriter) WriteComplex128(val complex128) PacketWriter {
	wr.bo.PutUint64(wr.buf[:8], math.Float64bits(real(val)))
	wr.bo.PutUint64(wr.buf[8:16], math.Float64bits(imag(val)))
	_, _ = wr.Buffer.Write(wr.buf[:16])
	return wr
}

func (wr *packetWriter) WriteOpcodeCommand(opcode int, cmd int) PacketWriter {
	_ = wr.Buffer.WriteByte(byte(opcode) | byte(cmd))
	return wr
}

func (wr *packetWriter) WriteObject(object io.Reader) error {
	if object, ok := object.(io.WriterTo); ok {
		_, err := object.WriteTo(wr.Buffer)
		return err
	}
	n, err := wr.Buffer.ReadFrom(object)
	if n == 0 && err == io.EOF {
		return nil
	}
	return err
}

// Reader exports a new PacketReader with the same buffer and byte order
func (wr *packetWriter) Reader() PacketReader {
	return &packetReader{
		Buffer: wr.Buffer,
		bo:     wr.bo,
	}
}
