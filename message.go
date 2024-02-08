// ©Hayabusa Cloud Co., Ltd. 2022. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package sox

import (
	"bufio"
	"encoding/binary"
	"errors"
	"io"
	"sync/atomic"
)

// MessageOptions represents message feature options
type MessageOptions struct {
	// ReadByteOrder sets which byte order will be used when reading data
	ReadByteOrder binary.ByteOrder
	// WriteByteOrder sets which byte order will be used when writing data
	WriteByteOrder binary.ByteOrder
	// ReadProto sets which protocol type will be used when reading data
	ReadProto UnderlyingProtocol
	// WriteProto sets which protocol type will be used when writing data
	WriteProto UnderlyingProtocol
	// ReadLimit is the maximum message payload data size
	// A ReadLimit of zero indicates that there is no limit
	ReadLimit int
	// Nonblock if the nonblock flag is true, Message will not block on I/O
	Nonblock bool
}

var defaultMessageOptions = MessageOptions{
	ReadByteOrder:  binary.BigEndian,
	WriteByteOrder: binary.BigEndian,
	ReadProto:      UnderlyingProtocolStream,
	WriteProto:     UnderlyingProtocolStream,
	Nonblock:       false,
}

// MessageOptionsTCPSocket sets feature options for TCP sockets
var MessageOptionsTCPSocket = func(options *MessageOptions) {
	options.ReadByteOrder = binary.BigEndian
	options.WriteByteOrder = binary.BigEndian
	options.ReadProto = UnderlyingProtocolStream
	options.WriteProto = UnderlyingProtocolStream
}

// MessageOptionsSCTPSocket sets feature options for SCTP sockets
var MessageOptionsSCTPSocket = func(options *MessageOptions) {
	options.ReadByteOrder = binary.BigEndian
	options.WriteByteOrder = binary.BigEndian
	options.ReadProto = UnderlyingProtocolSeqPacket
	options.WriteProto = UnderlyingProtocolSeqPacket
}

// MessageOptionsNetworkOrder sets byte order to big endian
var MessageOptionsNetworkOrder = func(options *MessageOptions) {
	options.ReadByteOrder = binary.BigEndian
	options.WriteByteOrder = binary.BigEndian
}

// MessageOptionsNonblock sets message nonblock
var MessageOptionsNonblock = func(options *MessageOptions) {
	options.Nonblock = true
}

// NewMessageReader creates and returns a new io.Reader to read messages
func NewMessageReader(reader io.Reader, opts ...func(options *MessageOptions)) io.Reader {
	return &messageReader{message: newMessage(reader, nil, opts...)}
}

// NewMessageWriter creates and returns a new io.Writer to write messages
func NewMessageWriter(writer io.Writer, opts ...func(options *MessageOptions)) io.Writer {
	return &messageWriter{message: newMessage(nil, writer, opts...)}
}

// NewMessageReadWriter creates and returns a new io.ReadWriter to read and write messages
func NewMessageReadWriter(reader io.Reader, writer io.Writer, opts ...func(options *MessageOptions)) io.ReadWriter {
	return &messageReadWriter{
		messageReader: &messageReader{newMessage(reader, nil, opts...)},
		messageWriter: &messageWriter{newMessage(nil, writer, opts...)},
	}
}

// NewMessagePipe creates and returns a synchronous in-memory message pipe
func NewMessagePipe(opts ...func(options *MessageOptions)) (reader io.Reader, writer io.Writer) {
	r, w := io.Pipe()
	pipe := NewMessageReadWriter(r, w, opts...)
	reader, writer = pipe, pipe
	return
}

// UnderlyingProtocol represents transmission protocol features
type UnderlyingProtocol int

const (
	// UnderlyingProtocolStream means the underlying protocol works like stream
	UnderlyingProtocolStream UnderlyingProtocol = 1
	// UnderlyingProtocolDgram means the underlying protocol works like datagram
	UnderlyingProtocolDgram UnderlyingProtocol = 2
	// UnderlyingProtocolSeqPacket means the underlying protocol works like sequenced packet
	UnderlyingProtocolSeqPacket UnderlyingProtocol = 5
)

// PreserveBoundary returns true if the underlying protocol preserves message boundaries
func (t UnderlyingProtocol) PreserveBoundary() bool {
	switch t {
	case UnderlyingProtocolDgram, UnderlyingProtocolSeqPacket:
		return true
	default:
		return false
	}
}

//
// We defined an original message protocol format as follows:
//
// 0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1
// +--------------+-------------------------------+---------------+
// |Payload Length|    Extended Payload Length    | Ext. Length   |
// |     (8)      |            (16/56)            | continued ... |
// |              | (if payload length==254/255)  | if len == 255 |
// +--------------+ - - - - - - - - - - - - - - - - - - - - - - - +
// |   Extended payload length continued, if payload len == 255   |
// | - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - -+
// |                        Payload Data                          |
// + - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - -+
// :                 Payload Data continued ...                   :
// + - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - -+
// |                 Payload Data continued ...                   |
// +--------------------------------------------------------------+
//
// Payload length: 8 bits, 8+16 bits, or 8+56 bits
//   The length of the "Payload data", in bytes: if 0-253, that is the
//   payload length. If 254, the following 2 bytes interpreted as a
//   2-bytes unsigned integer are the payload length. If 255, the
//   following 7 bytes interpreted as a 56-bits unsigned integer are
//   the payload length. Multibyte length quantities are expressed in
//   network byte order. oad Length to encode the length.

var (
	// ErrMsgInvalidArguments will be returned when got invalid parameter
	ErrMsgInvalidArguments = errors.New("message invalid argument")
	// ErrMsgInvalidRead will be returned when read operation invalid
	ErrMsgInvalidRead = errors.New("message invalid read result")
	// ErrMsgInvalidWrite will be returned when write operation invalid
	ErrMsgInvalidWrite = errors.New("message invalid write result")
	// ErrMsgTooLong will be returned when try to read or write a message which is too long
	ErrMsgTooLong = errors.New("message too long")
	// ErrMsgClosed will be returned when try to read or write on a closed reader or writer
	ErrMsgClosed = errors.New("message closed")
)

const (
	messageHeaderLength           = 1
	messagePayloadMaxLength8Bits  = 1<<8 - 3
	messagePayloadMaxLength16Bits = 1<<16 - 1
	messagePayloadMaxLength56Bits = 1<<56 - 1

	messageStatusRead   uint32 = 4
	messageStatusWrite  uint32 = 2
	messageStatusClosed uint32 = 0x2000
)

type message struct {
	rd  io.Reader
	rbo binary.ByteOrder
	rpr UnderlyingProtocol
	wr  io.Writer
	wbo binary.ByteOrder
	wpr UnderlyingProtocol

	status atomic.Uint32
	header [8]byte
	length int64
	offset int64
	count  atomic.Int32

	readLimit int64
	nonblock  bool

	done bool
}

func (msg *message) close() error {
	if msg.done {
		return nil
	}
	for sw := NewParamSpinWait(); !sw.Closed(); {
		status := msg.status.Load()
		if (status & (messageStatusRead | messageStatusWrite)) == (messageStatusRead | messageStatusWrite) {
			if msg.nonblock {
				return ErrTemporarilyUnavailable
			}
			sw.Once()
			continue
		}

		if msg.status.CompareAndSwap(status, status|messageStatusClosed) {
			msg.done = true
			return nil
		}
		sw.OnceWithLevel(spinWaitLevelAtomic)
	}

	return nil
}

func (msg *message) setReadWriter(rw io.ReadWriter, order binary.ByteOrder, typ UnderlyingProtocol) {
	msg.setReader(rw, order, typ)
	msg.setWriter(rw, order, typ)
}
func (msg *message) setReader(r io.Reader, order binary.ByteOrder, typ UnderlyingProtocol) {
	msg.rd = r
	msg.rbo = order
	msg.rpr = typ
}
func (msg *message) setWriter(w io.Writer, order binary.ByteOrder, typ UnderlyingProtocol) {
	msg.wr = w
	msg.wbo = order
	msg.wpr = typ
}

func (msg *message) read(p []byte) (n int, err error) {
	if msg.done {
		return 0, io.EOF
	}
	if _, ok := msg.enterRead(); !ok {
		return 0, ErrTemporarilyUnavailable
	}
	if msg.rpr.PreserveBoundary() {
		return msg.readPacket(p)
	}
	return msg.readStream(p)
}

func (msg *message) readStream(p []byte) (n int, err error) {
	defer func() {
		if err != ErrTemporarilyUnavailable {
			msg.exitRead()
		}
	}()

	for rn := 0; msg.offset < messageHeaderLength; {
		rn, err = msg.readOnce(msg.header[msg.offset:messageHeaderLength])
		msg.offset += int64(rn)
		if err != nil && err != io.EOF && (err != ErrTemporarilyUnavailable || msg.nonblock) {
			return
		}
		if err == io.EOF {
			if msg.offset < messageHeaderLength {
				return 0, io.ErrUnexpectedEOF
			}
			break
		}
	}
	exLengthBytes := int64(0)
	if msg.offset >= messageHeaderLength {
		if msg.readLimit > 0 && msg.length > msg.readLimit {
			return 0, ErrMsgTooLong
		}
		if msg.header[0] == messagePayloadMaxLength8Bits+1 {
			exLengthBytes = 2
		} else if msg.header[0] == messagePayloadMaxLength8Bits+2 {
			exLengthBytes = 7
		}
	}
	if len(msg.header) < int(messageHeaderLength+exLengthBytes) {
		return 0, io.ErrShortBuffer
	}
	if msg.offset < messageHeaderLength+exLengthBytes {
		for rn := 0; msg.offset < messageHeaderLength+exLengthBytes; {
			rn, err = msg.readOnce(msg.header[messageHeaderLength : messageHeaderLength+exLengthBytes])
			msg.offset += int64(rn)
			if err != nil && err != io.EOF && (err != ErrTemporarilyUnavailable || msg.nonblock) {
				return
			}
			if err == io.EOF {
				if msg.offset < messageHeaderLength+exLengthBytes {
					return 0, io.ErrUnexpectedEOF
				}
				break
			}
		}
	}
	if msg.offset == messageHeaderLength+exLengthBytes {
		if exLengthBytes == 2 {
			msg.length = int64(msg.rbo.Uint16(msg.header[messageHeaderLength : messageHeaderLength+exLengthBytes]))
		} else if exLengthBytes == 7 {
			u64 := msg.rbo.Uint64(msg.header[:])
			if msg.rbo == binary.LittleEndian {
				msg.length = int64(u64 >> 8)
			} else if msg.rbo == binary.BigEndian {
				msg.length = int64(u64 & messagePayloadMaxLength56Bits)
			}
		} else {
			msg.length = int64(msg.header[0])
		}
	}
	if msg.readLimit > 0 && msg.length > msg.readLimit {
		return 0, ErrMsgTooLong
	}
	// we assume that generally a 4K buffer p []byte will be given
	if msg.length > int64(len(p)) {
		if msg.length < (1 << 16) {
			// TODO: acquire a 64k buffer from pool
		} else if msg.length < (1 << 20) {
			// TODO: acquire a 1m buffer from pool
		} else if msg.length < (1 << 24) {
			// TODO: acquire a 16m buffer from pool
		} else {
			// TODO: non-buffered work mode
		}
		return 0, io.ErrShortBuffer
	}
	for rn := 0; msg.offset < messageHeaderLength+exLengthBytes+msg.length; {
		rn, err = msg.readOnce(p[msg.offset-messageHeaderLength-exLengthBytes : msg.length])
		msg.offset += int64(rn)
		n += rn
		if err != nil && err != io.EOF && (err != ErrTemporarilyUnavailable || msg.nonblock) {
			return
		}
		if err == io.EOF {
			if msg.offset < messageHeaderLength+exLengthBytes+msg.length {
				return n, io.ErrUnexpectedEOF
			}
			break
		}
	}

	msg.count.Add(-1)
	msg.reset()
	return
}
func (msg *message) readPacket(p []byte) (n int, err error) {
	defer msg.exitRead()
	for {
		n, err = msg.readOnce(p)
		if err == ErrTemporarilyUnavailable {
			if msg.nonblock {
				return
			}
			continue
		}
		if err != nil && err != io.EOF {
			return
		}
		if n > messagePayloadMaxLength56Bits {
			return n, ErrMsgTooLong
		} else if n == len(p) {
			break
		}
	}

	msg.count.Add(-1)
	msg.reset()
	return
}
func (msg *message) readOnce(p []byte) (n int, err error) {
	if msg.rd == nil {
		return 0, ErrMsgInvalidArguments
	}
	for {
		n, err = msg.rd.Read(p)
		if err != ErrTemporarilyUnavailable {
			break
		}
		if msg.nonblock {
			break
		}
	}
	return
}
func (msg *message) enterRead() (oldStatus uint32, ok bool) {
	if msg.wr == nil {
		return 0, true
	}
	sw := SpinWait{}
	for {
		oldStatus = msg.status.Load()
		if (oldStatus & ^messageStatusRead) == 0 {
			if msg.status.CompareAndSwap(oldStatus, oldStatus|messageStatusRead) {
				return oldStatus, true
			}
			sw.Once()
			continue
		} else if !msg.nonblock {
			sw.Once()
			continue
		}
		return oldStatus, false
	}
}
func (msg *message) exitRead() (oldStatus uint32) {
	if msg.wr == nil {
		return 0
	}
	sw := SpinWait{}
	for {
		oldStatus = msg.status.Load()
		if msg.status.CompareAndSwap(oldStatus, oldStatus&^messageStatusRead) {
			break
		}
		sw.Once()
	}
	return
}

func (msg *message) write(p []byte) (n int, err error) {
	if msg.done {
		return 0, ErrMsgClosed
	}
	if _, ok := msg.enterWrite(); !ok {
		return 0, ErrTemporarilyUnavailable
	}
	if msg.wpr.PreserveBoundary() {
		return msg.writePacket(p)
	}
	return msg.writeStream(p)
}

func (msg *message) writeStream(p []byte) (n int, err error) {
	defer func() {
		if err != ErrTemporarilyUnavailable {
			msg.exitWrite()
		}
	}()

	if msg.length > messagePayloadMaxLength56Bits {
		return 0, ErrMsgTooLong
	}

	if msg.offset == 0 {
		msg.length = int64(len(p))
	}
	exLengthBytes := int64(0)
	if msg.length <= messagePayloadMaxLength8Bits {
		exLengthBytes = 0
	} else if msg.length <= messagePayloadMaxLength16Bits {
		exLengthBytes = 2
	} else {
		exLengthBytes = 7
	}
	if msg.offset == 0 {
		if msg.length <= messagePayloadMaxLength8Bits {
			msg.header[0] = byte(msg.length)
		} else if msg.length <= messagePayloadMaxLength16Bits {
			msg.header[0] = messagePayloadMaxLength8Bits + 1
			msg.wbo.PutUint16(msg.header[messageHeaderLength:messageHeaderLength+exLengthBytes], uint16(msg.length))
		} else {
			if msg.wbo == binary.LittleEndian {
				msg.wbo.PutUint64(msg.header[:], uint64(msg.length)<<8)
			} else {
				msg.wbo.PutUint64(msg.header[:], uint64(msg.length&messagePayloadMaxLength56Bits))
			}
			msg.header[0] = messagePayloadMaxLength8Bits + 2
		}
	}
	for wn := 0; msg.offset < messageHeaderLength+exLengthBytes; {
		wn, err = msg.writeOnce(msg.header[msg.offset : messageHeaderLength+exLengthBytes])
		msg.offset += int64(wn)
		if err != nil && (err != ErrTemporarilyUnavailable || msg.nonblock) {
			return
		}
	}
	if msg.length != msg.offset-messageHeaderLength-exLengthBytes+int64(len(p)) {
		return 0, io.ErrShortWrite
	}
	for wn := 0; msg.offset < messageHeaderLength+exLengthBytes+msg.length; {
		wn, err = msg.writeOnce(p[:msg.length-(msg.offset-messageHeaderLength-exLengthBytes)])
		msg.offset += int64(wn)
		n += wn
		if err != nil && (err != ErrTemporarilyUnavailable || msg.nonblock) {
			break
		}
	}

	if msg.offset < messageHeaderLength+msg.length {
		return n, io.ErrShortWrite
	}

	msg.count.Add(1)
	msg.reset()
	return
}
func (msg *message) writePacket(p []byte) (n int, err error) {
	defer msg.exitWrite()
	if len(p) > messagePayloadMaxLength56Bits {
		return 0, bufio.ErrTooLong
	}
	for {
		n, err = msg.writeOnce(p)
		if err == ErrTemporarilyUnavailable {
			if msg.nonblock {
				return
			}
			continue
		}
		if err != nil {
			return
		}
		if n < len(p) {
			return n, io.ErrShortWrite
		} else if n == len(p) {
			break
		}
	}

	msg.count.Add(1)
	msg.reset()
	return
}
func (msg *message) writeOnce(p []byte) (n int, err error) {
	if msg.wr == nil {
		return 0, ErrMsgInvalidArguments
	}
	for {
		n, err = msg.wr.Write(p)
		if err != ErrTemporarilyUnavailable {
			break
		}
		if msg.nonblock {
			break
		}
	}
	return
}

func (msg *message) enterWrite() (oldStatus uint32, ok bool) {
	if msg.rd == nil {
		return 0, true
	}
	sw := SpinWait{}
	for {
		oldStatus = msg.status.Load()
		if (oldStatus & ^messageStatusWrite) == 0 {
			if msg.status.CompareAndSwap(oldStatus, oldStatus|messageStatusWrite) {
				return oldStatus, true
			}
			sw.Once()
			continue
		} else if !msg.nonblock {
			sw.Once()
			continue
		}
		return oldStatus, false
	}
}
func (msg *message) exitWrite() (oldStatus uint32) {
	if msg.rd == nil {
		return 0
	}
	sw := SpinWait{}
	for {
		oldStatus = msg.status.Load()
		if msg.status.CompareAndSwap(oldStatus, oldStatus&^messageStatusWrite) {
			break
		}
		sw.Once()
	}
	return
}

func (msg *message) readFrom(reader io.Reader) (n int64, err error) {
	if msg.done {
		return 0, ErrMsgClosed
	}
	if msg.wr == nil {
		return 0, ErrMsgInvalidArguments
	}
	return io.Copy(msg.wr, reader)
}

func (msg *message) writeTo(writer io.Writer) (n int64, err error) {
	if msg.done {
		return 0, io.EOF
	}
	if msg.rd == nil {
		return 0, ErrMsgInvalidArguments
	}
	return io.Copy(writer, msg.rd)
}

func (msg *message) reset() {
	msg.offset = 0
}

func newMessage(reader io.Reader, writer io.Writer, opts ...func(options *MessageOptions)) *message {
	opt := defaultMessageOptions
	for _, fn := range opts {
		fn(&opt)
	}

	m := &message{
		status:    atomic.Uint32{},
		header:    [8]byte{},
		length:    0,
		offset:    0,
		count:     atomic.Int32{},
		readLimit: int64(opt.ReadLimit),
		nonblock:  opt.Nonblock,
		done:      false,
	}
	if reader != nil {
		m.setReader(reader, opt.ReadByteOrder, opt.ReadProto)
	}
	if writer != nil {
		m.setWriter(writer, opt.WriteByteOrder, opt.WriteProto)
	}
	return m
}

type messageReader struct {
	*message
}

func (msg *messageReader) Read(b []byte) (n int, err error) {
	return msg.read(b)
}

func (msg *messageReader) WriteTo(writer io.Writer) (n int64, err error) {
	return msg.writeTo(writer)
}

type messageWriter struct {
	*message
}

func (msg *messageWriter) Write(b []byte) (n int, err error) {
	return msg.write(b)
}

func (msg *messageWriter) ReadFrom(reader io.Reader) (n int64, err error) {
	return msg.readFrom(reader)
}

type messageReadWriter struct {
	*messageReader
	*messageWriter
}
