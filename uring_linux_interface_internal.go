// ©Hayabusa Cloud Co., Ltd. 2025. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

//go:build linux

package sox

import (
	"context"
	"unsafe"
)

var defaultUringOptions = UringOptions{
	Entries:                 UringEntriesMedium,
	LockedBufferMem:         registerBufferDefaultMem,
	ReadBufferSize:          bufferSizeDefault,
	ReadBufferNum:           UringEntriesMedium * 2,
	WriteBufferSize:         bufferSizeDefault,
	WriteBufferNum:          UringEntriesMedium,
	MultiIssuers:            false,
	NotifySucceed:           false,
	IndirectSubmissionQueue: false,
}

type UringOptionFunc func(opt *UringOptions)

func (uo *UringOptions) Apply(opts ...UringOptionFunc) {
	for _, f := range opts {
		f(uo)
	}
}

var (
	UringLargeLockedBufferMemOptions UringOptionFunc = func(opt *UringOptions) {
		opt.LockedBufferMem = registerBufferSize * registerBufferNum
	}
	UringMultiSizeBufferOptions UringOptionFunc = func(opt *UringOptions) {
		opt.MultiSizeBuffer = 1
	}
)

func (ur *Uring) operationOptions(opts []UringOpOptionFunc) (flags uint8) {
	if len(opts) < 1 {
		return uringOpFlagsNone
	}
	opt := defaultUringOpOption
	opt.Apply(opts...)
	return opt.Flags
}

func (ur *Uring) readFixedOptions(bufIndex int, opts []UringOpOptionFunc) (flags uint8, offset int64, n int) {
	if len(opts) < 1 {
		bufSize := len(ur.ioUring.bufs[bufIndex])
		return uringOpFlagsNone, 0, bufSize
	}
	opt := defaultUringOpOption
	opt.Apply(opts...)

	bufSize := len(ur.ioUring.bufs[bufIndex])
	readSize := bufSize
	if opt.N != nil && *opt.N < bufSize {
		readSize = *opt.N
	}

	return opt.Flags, opt.Offset, readSize
}

func (ur *Uring) writeFixedOptions(opts []UringOpOptionFunc) (flags uint8, offset int64) {
	if len(opts) < 1 {
		return uringOpFlagsNone, 0
	}
	opt := defaultUringOpOption
	opt.Apply(opts...)

	return opt.Flags, opt.Offset
}

func (ur *Uring) pollAddOptions(opts []UringOpOptionFunc) (flags uint8) {
	if len(opts) < 1 {
		return uringOpFlagsNone
	}
	opt := defaultUringOpOption
	opt.Apply(opts...)
	return opt.Flags
}

func (ur *Uring) pollRemoveOptions(opts []UringOpOptionFunc) (flags uint8) {
	if len(opts) < 1 {
		return uringOpFlagsNone
	}
	opt := defaultUringOpOption
	opt.Apply(opts...)
	return opt.Flags
}

func (ur *Uring) syncFileRangeOptions(opts []UringOpOptionFunc) (flags uint8, offset int64, n int) {
	if len(opts) < 1 {
		return uringOpFlagsNone, 0, bufferSizeDefault
	}
	opt := defaultUringOpOption
	opt.Apply(opts...)

	size := bufferSizeDefault
	if opt.N != nil {
		size = *opt.N
	}
	return opt.Flags, opt.Offset, size
}
func (ur *Uring) sendmsgOptions(iovs [][]byte, opts []UringOpOptionFunc) (flags uint8, ioprio uint16) {
	if len(opts) < 1 {
		return uringOpFlagsNone, uringOpIoprioNone
	}
	opt := defaultUringOpOption
	opt.Apply(opts...)
	return opt.Flags, opt.Ioprio
}

func (ur *Uring) recvmsgOptions(iovs [][]byte, opts []UringOpOptionFunc) (flags uint8, ioprio uint16) {
	if len(opts) < 1 {
		return uringOpFlagsNone, uringOpIoprioNone
	}
	opt := defaultUringOpOption
	opt.Apply(opts...)
	return opt.Flags, opt.Ioprio
}

func (ur *Uring) timeoutOptions(opts []UringOpOptionFunc) (flags uint8, cnt int) {
	if len(opts) < 1 {
		return uringOpFlagsNone, 1
	}
	opt := defaultUringOpOption
	opt.Apply(opts...)
	return opt.Flags, opt.Count
}

func (ur *Uring) timeoutRemoveOptions(opts []UringOpOptionFunc) (flags uint8) {
	if len(opts) < 1 {
		return uringOpFlagsNone
	}
	opt := defaultUringOpOption
	opt.Apply(opts...)
	return opt.Flags
}

func (ur *Uring) acceptOptions(opts []UringOpOptionFunc) (flags uint8, ioprio uint16) {
	if len(opts) < 1 {
		return uringOpFlagsNone, uringOpIoprioNone
	}
	opt := defaultUringOpOption
	opt.Apply(opts...)
	return opt.Flags, opt.Ioprio
}
func (ur *Uring) asyncCancelOptions(opts []UringOpOptionFunc) (flags uint8) {
	if len(opts) < 1 {
		return uringOpFlagsNone
	}
	opt := defaultUringOpOption
	opt.Apply(opts...)
	return opt.Flags
}

func (ur *Uring) connectOptions(opts []UringOpOptionFunc) (flags uint8) {
	if len(opts) < 1 {
		return uringOpFlagsNone
	}
	opt := defaultUringOpOption
	opt.Apply(opts...)
	return opt.Flags
}

func (ur *Uring) openAtOptions(opts []UringOpOptionFunc) (flags uint8) {
	if len(opts) < 1 {
		return uringOpFlagsNone
	}
	opt := defaultUringOpOption
	opt.Apply(opts...)
	return opt.Flags
}

func (ur *Uring) closeOptions(opts []UringOpOptionFunc) (flags uint8) {
	if len(opts) < 1 {
		return uringOpFlagsNone
	}
	opt := defaultUringOpOption
	opt.Apply(opts...)
	return opt.Flags
}

func (ur *Uring) statxOptions(opts []UringOpOptionFunc) (flags uint8) {
	if len(opts) < 1 {
		return uringOpFlagsNone
	}
	opt := defaultUringOpOption
	opt.Apply(opts...)
	return opt.Flags
}

func (ur *Uring) readOptions(b []byte, opts []UringOpOptionFunc) (flags uint8, ioprio uint16, offset int64, n int) {
	if len(opts) < 1 {
		return uringOpFlagsNone, uringOpIoprioNone, 0, len(b)
	}
	opt := defaultUringOpOption
	opt.Apply(opts...)

	readSize := len(b)
	if opt.N != nil && *opt.N < len(b) {
		readSize = *opt.N
	}

	return opt.Flags, opt.Ioprio, opt.Offset, readSize
}

func (ur *Uring) writeOptions(b []byte, opts []UringOpOptionFunc) (flags uint8, ioprio uint16, offset int64, n int) {
	if len(opts) < 1 {
		return uringOpFlagsNone, uringOpIoprioNone, 0, len(b)
	}
	opt := defaultUringOpOption
	opt.Apply(opts...)

	writeSize := len(b)
	if opt.N != nil && *opt.N < len(b) {
		writeSize = *opt.N
	}

	return opt.Flags, opt.Ioprio, opt.Offset, writeSize
}

func (ur *Uring) fadviseOptions(b []byte, opts []UringOpOptionFunc) (flags uint8, offset int64, n int, advice int) {
	if len(opts) < 1 {
		return uringOpFlagsNone, 0, len(b), 0
	}
	opt := defaultUringOpOption
	opt.Apply(opts...)

	size := len(b)
	if opt.N != nil && *opt.N < len(b) {
		size = *opt.N
	}

	return opt.Flags, opt.Offset, size, opt.Advice
}

func (ur *Uring) madviseOptions(b []byte, opts []UringOpOptionFunc) (flags uint8, advice int) {
	if len(opts) < 1 {
		return uringOpFlagsNone, 0
	}
	opt := defaultUringOpOption
	opt.Apply(opts...)
	return opt.Flags, opt.Advice
}

func (ur *Uring) sendOptions(b []byte, opts []UringOpOptionFunc) (flags uint8, ioprio uint16, offset int64, n int) {
	if len(opts) < 1 {
		return uringOpFlagsNone, uringOpIoprioNone, 0, len(b)
	}
	opt := defaultUringOpOption
	opt.Apply(opts...)

	n = len(b)
	if opt.N != nil && *opt.N < len(b) {
		n = *opt.N
	}

	return opt.Flags, opt.Ioprio, opt.Offset, n
}

func (ur *Uring) receiveOptions(b []byte, opts []UringOpOptionFunc) (flags uint8, ioprio uint16, offset int64, n int) {
	if len(opts) < 1 {
		return uringOpFlagsNone, uringOpIoprioNone, 0, len(b)
	}
	opt := defaultUringOpOption
	opt.Apply(opts...)

	n = len(b)
	if opt.N != nil && *opt.N < len(b) {
		n = *opt.N
	}

	return opt.Flags, opt.Ioprio, opt.Offset, n
}

func (ur *Uring) openAt2Options(opts []UringOpOptionFunc) (flags uint8) {
	if len(opts) < 1 {
		return uringOpFlagsNone
	}
	opt := defaultUringOpOption
	opt.Apply(opts...)
	return opt.Flags
}

func (ur *Uring) epollCtlOptions(opts []UringOpOptionFunc) (flags uint8) {
	if len(opts) < 1 {
		return uringOpFlagsNone
	}
	opt := defaultUringOpOption
	opt.Apply(opts...)
	return opt.Flags
}

func (ur *Uring) spliceOptions(opts []UringOpOptionFunc) (flags uint8, n int) {
	if len(opts) < 1 {
		return uringOpFlagsNone, bufferSizeDefault
	}
	opt := defaultUringOpOption
	opt.Apply(opts...)

	size := bufferSizeDefault
	if opt.N != nil {
		size = *opt.N
	}
	return opt.Flags, size
}

func (ur *Uring) teeOptions(opts []UringOpOptionFunc) (flags uint8, n int) {
	if len(opts) < 1 {
		return uringOpFlagsNone, bufferSizeDefault
	}
	opt := defaultUringOpOption
	opt.Apply(opts...)

	size := bufferSizeDefault
	if opt.N != nil {
		size = *opt.N
	}
	return opt.Flags, size
}

func (ur *Uring) shutdownOptions(opts []UringOpOptionFunc) (flags uint8) {
	if len(opts) < 1 {
		return uringOpFlagsNone
	}
	opt := defaultUringOpOption
	opt.Apply(opts...)
	return opt.Flags
}

func (ur *Uring) renameAtOptions(opts []UringOpOptionFunc) (flags uint8) {
	if len(opts) < 1 {
		return uringOpFlagsNone
	}
	opt := defaultUringOpOption
	opt.Apply(opts...)
	return opt.Flags
}

func (ur *Uring) unlinkAtOptions(opts []UringOpOptionFunc) (flags uint8) {
	if len(opts) < 1 {
		return uringOpFlagsNone
	}
	opt := defaultUringOpOption
	opt.Apply(opts...)
	return opt.Flags
}

func (ur *Uring) mkdirAtOptions(opts []UringOpOptionFunc) (flags uint8) {
	if len(opts) < 1 {
		return uringOpFlagsNone
	}
	opt := defaultUringOpOption
	opt.Apply(opts...)
	return opt.Flags
}

func (ur *Uring) symlinkAtOptions(opts []UringOpOptionFunc) (flags uint8) {
	if len(opts) < 1 {
		return uringOpFlagsNone
	}
	opt := defaultUringOpOption
	opt.Apply(opts...)
	return opt.Flags
}

func (ur *Uring) linkAtOptions(opts []UringOpOptionFunc) (flags uint8) {
	if len(opts) < 1 {
		return uringOpFlagsNone
	}
	opt := defaultUringOpOption
	opt.Apply(opts...)
	return opt.Flags
}

func (ur *Uring) receiveWithBufferSelectOptions(f PollFd, opts []UringOpOptionFunc) (flags uint8, ioprio uint16, bufferSize int, bufferGroup uint16) {
	if len(opts) < 1 {
		return uringOpFlagsNone, uringOpIoprioNone, bufferSizeDefault, ur.getBufferGid(f, bufferSizeDefault)
	}
	opt := defaultUringOpOption
	opt.Apply(opts...)
	return opt.Flags, opt.Ioprio, opt.ReadBufferSize, ur.getBufferGid(f, opt.ReadBufferSize)
}

func (ur *Uring) socketOptions(opts []UringOpOptionFunc) (flags uint8, callback uint8, i uint32) {
	if len(opts) < 1 {
		return uringOpFlagsNone, uringOpCallbackNone, 0
	}
	opt := defaultUringOpOption
	opt.Apply(opts...)
	return opt.Flags, opt.Callback, opt.FileIndex
}

func (ur *Uring) bindOptions(opts []UringOpOptionFunc) (flags uint8, callback uint8) {
	if len(opts) < 1 {
		return uringOpFlagsNone, uringOpCallbackNone
	}
	opt := defaultUringOpOption
	opt.Apply(opts...)
	return opt.Flags, opt.Callback
}

func (ur *Uring) listenOptions(opts []UringOpOptionFunc) (flags uint8, backlog int) {
	if len(opts) < 1 {
		return uringOpFlagsNone, defaultBacklog
	}
	opt := defaultUringOpOption
	opt.Apply(opts...)
	return opt.Flags, opt.Backlog
}

func (ur *Uring) getBufferGid(f PollFd, size int) uint16 {
	if ur.buffers != nil {
		return ur.buffers.bufGroup(f)
	} else if ur.bufferGroups != nil {
		return ur.bufferGroups.bufGroupBySize(f, size)
	}
	panic("Uring has neither buffers nor bufferGroups set")
}

func (ur *Uring) buf(bufGroup bufferGroupIndex, bufIndex uint32) []byte {
	if ur.buffers != nil {
		return ur.buffers.buf(bufGroup, bufIndex)
	} else if ur.bufferGroups != nil {
		return ur.bufferGroups.buf(bufGroup, bufIndex)
	}
	panic("Uring has neither buffers nor bufferGroups set")
}

func (ur *Uring) socketOpUserdata(callback uint8, domain NetworkType, sa []byte) ioUringUserdata {
	ret := make(ioUringUserdata, unsafe.Sizeof(callback)+unsafe.Sizeof(uint16(0))+unsafe.Sizeof(uint16(0))+uintptr(len(sa)))
	i := uintptr(0)
	ret[i] = callback
	i += unsafe.Sizeof(callback)
	ur.Features.UserdataByteOrder.PutUint16(ret[i:], uint16(domain&0xffff))
	i += unsafe.Sizeof(uint16(domain))
	ur.Features.UserdataByteOrder.PutUint16(ret[i:], uint16(len(sa)&0xffff))
	i += unsafe.Sizeof(uint16(0))
	if len(sa) > 0 {
		copy(ret[i:], sa)
	}
	return ret
}

func (ur *Uring) bindOpUserdata(callback uint8, domain NetworkType, sa []byte) ioUringUserdata {
	return ur.socketOpUserdata(callback, domain, sa)
}

func (ur *Uring) listenOpUserdata(callback uint8, domain NetworkType, sa []byte) ioUringUserdata {
	return ur.bindOpUserdata(callback, domain, sa)
}

func (ur *Uring) acceptOpUserData(domain NetworkType) ioUringUserdata {
	ret := make(ioUringUserdata, unsafe.Sizeof(uint16(0)))
	ur.Features.UserdataByteOrder.PutUint16(ret[:], uint16(domain&0xffff))
	return ret
}

func (ur *Uring) socketFromUringEvent(ev *UringEvent) (so *socket, err error) {
	i := uintptr(0)
	callback := uint8(ev.UserData[i])
	i += unsafe.Sizeof(callback)
	domain := NetworkType(ur.Features.UserdataByteOrder.Uint16(ev.UserData[i:]))
	i += unsafe.Sizeof(uint16(0))
	fd := ev.FD
	if fd == 0 && ev.Result > 0 {
		fd = ioUringFd(int(ev.Result))
	}
	so = newSocket(domain, int(fd), nil)
	l := ur.Features.UserdataByteOrder.Uint16(ev.UserData[i:])
	i += unsafe.Sizeof(uint16(0))
	saRaw := ev.UserData[i : i+uintptr(l)]
	switch domain {
	case NetworkIPv4:
		so.sa = inet4SockaddrFromData(saRaw)
	case NetworkIPv6:
		so.sa = inet6SockaddrFromData(saRaw)
	case NetworkUnix:
		so.sa = unixSockaddrFromData(saRaw)
	}
	if ev.Operation == IORING_OP_SOCKET && callback >= uringOpCallbackBind {
		ctx := ContextWithUserdata(context.Background(), ur.bindOpUserdata(callback, so.network, saRaw))
		err = ur.bind(ctx, uringOpFlagsNone, so.fd, so.sa)
		if err != nil {
			return so, err
		}
	} else if ev.Operation == IORING_OP_BIND && callback >= uringOpCallbackListen {
		ctx := ContextWithUserdata(context.Background(), ur.listenOpUserdata(callback, so.network, saRaw))
		err = ur.listen(ctx, uringOpFlagsNone, so.fd, defaultBacklog)
		if err != nil {
			return so, err
		}
	}
	return so, nil
}

func (ur *Uring) listenerSocketFromUringEvent(ev *UringEvent) *socket {
	domain := NetworkType(ur.Features.UserdataByteOrder.Uint16(ev.UserData[:]))
	fd := int(ev.UringEventData.FD)
	return newSocket(domain, fd, nil)
}

func (ur *Uring) acceptedConnFromUringEvent(ev *UringEvent) *socket {
	domain := NetworkType(ur.Features.UserdataByteOrder.Uint16(ev.UserData[:]))
	fd := int(ev.Result)
	return newSocket(domain, fd, nil)
}
