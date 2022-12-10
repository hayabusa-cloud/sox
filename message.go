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
	ReadByteOrder  binary.ByteOrder
	WriteByteOrder binary.ByteOrder
	ReadProto      UnderlyingProtocol
	WriteProto     UnderlyingProtocol
	Nonblock       bool
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
	messageHeaderLength     = 2
	messagePayloadMaxLength = 1 << (messageHeaderLength << 3)

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
	header []byte
	length int
	offset int
	count  atomic.Int32

	nonblock bool

	done bool
}

func (msg *message) close() error {
	if msg.done {
		return nil
	}
	for sw := NewSpinWaiter(); !sw.Closed(); {
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

	if msg.header == nil || len(msg.header) < messageHeaderLength {
		return 0, ErrMsgInvalidArguments
	}

	for rn := 0; msg.offset < messageHeaderLength; {
		rn, err = msg.readOnce(msg.header[msg.offset:messageHeaderLength])
		msg.offset += rn
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
	if msg.offset == messageHeaderLength {
		msg.length = msg.payloadLen(msg.header)
	}
	if msg.length > len(p) {
		return 0, io.ErrShortBuffer
	}
	for rn := 0; msg.offset < messageHeaderLength+msg.length; {
		rn, err = msg.readOnce(p[msg.offset-messageHeaderLength : msg.length])
		msg.offset += rn
		n += rn
		if err != nil && err != io.EOF && (err != ErrTemporarilyUnavailable || msg.nonblock) {
			return
		}
		if err == io.EOF {
			if msg.offset < messageHeaderLength+msg.length {
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
			return 0, err
		}
		if n > messagePayloadMaxLength {
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
	for sw := NewSpinWaiter(); !sw.Closed(); {
		oldStatus = msg.status.Load()
		if (oldStatus & ^messageStatusRead) == 0 {
			if msg.status.CompareAndSwap(oldStatus, oldStatus|messageStatusRead) {
				return oldStatus, true
			}
			sw.OnceWithLevel(spinWaitLevelAtomic)
			continue
		} else if !msg.nonblock {
			sw.Once()
			continue
		}
		return oldStatus, false
	}

	return oldStatus, false
}
func (msg *message) exitRead() (oldStatus uint32) {
	if msg.wr == nil {
		return 0
	}
	for sw := NewSpinWaiter().SetLevel(spinWaitLevelAtomic); !sw.Closed(); sw.Once() {
		oldStatus = msg.status.Load()
		if msg.status.CompareAndSwap(oldStatus, oldStatus&^messageStatusRead) {
			break
		}
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

	if msg.header == nil || len(msg.header) < messageHeaderLength {
		return 0, ErrMsgInvalidArguments
	}
	if len(p) > messagePayloadMaxLength {
		return 0, ErrMsgTooLong
	}

	if msg.offset == 0 {
		msg.length = len(p)
		msg.putPayloadLen(msg.header, msg.length)
	}
	for wn := 0; msg.offset < messageHeaderLength; {
		wn, err = msg.writeOnce(msg.header[msg.offset:messageHeaderLength])
		msg.offset += wn
		if err != nil && (err != ErrTemporarilyUnavailable || msg.nonblock) {
			return
		}
	}
	if msg.length != msg.offset-messageHeaderLength+len(p) {
		return 0, io.ErrShortWrite
	}
	for wn := 0; msg.offset < messageHeaderLength+msg.length; {
		wn, err = msg.writeOnce(p[:msg.length-(msg.offset-messageHeaderLength)])
		msg.offset += wn
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
	if len(p) > messagePayloadMaxLength {
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
	for sw := NewSpinWaiter(); !sw.Closed(); {
		oldStatus = msg.status.Load()
		if (oldStatus & ^messageStatusWrite) == 0 {
			if msg.status.CompareAndSwap(oldStatus, oldStatus|messageStatusWrite) {
				return oldStatus, true
			}
			sw.OnceWithLevel(spinWaitLevelAtomic)
			continue
		} else if !msg.nonblock {
			sw.Once()
			continue
		}
		return oldStatus, false
	}

	return oldStatus, false
}
func (msg *message) exitWrite() (oldStatus uint32) {
	if msg.rd == nil {
		return 0
	}
	for sw := NewSpinWaiter().SetLevel(spinWaitLevelAtomic); !sw.Closed(); sw.Once() {
		oldStatus = msg.status.Load()
		if msg.status.CompareAndSwap(oldStatus, oldStatus&^messageStatusWrite) {
			break
		}
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

func (msg *message) payloadLen(header []byte) int {
	return int(msg.rbo.Uint16(header))
}

func (msg *message) putPayloadLen(header []byte, length int) {
	msg.wbo.PutUint16(header, uint16(length))
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
		status:   atomic.Uint32{},
		header:   make([]byte, 2),
		length:   0,
		offset:   0,
		count:    atomic.Int32{},
		nonblock: opt.Nonblock,
		done:     false,
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
