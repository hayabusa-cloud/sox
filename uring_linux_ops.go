// ©Hayabusa Cloud Co., Ltd. 2022. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

//go:build linux

package sox

import (
	"context"
	"golang.org/x/sys/unix"
	"unsafe"
)

const (
	IORING_OP_NOP uint8 = iota
	IORING_OP_READV
	IORING_OP_WRITEV
	IORING_OP_FSYNC
	IORING_OP_READ_FIXED
	IORING_OP_WRITE_FIXED
	IORING_OP_POLL_ADD
	IORING_OP_POLL_REMOVE
	IORING_OP_SYNC_FILE_RANGE
	IORING_OP_SENDMSG
	IORING_OP_RECVMSG
	IORING_OP_TIMEOUT
	IORING_OP_TIMEOUT_REMOVE
	IORING_OP_ACCEPT
	IORING_OP_ASYNC_CANCEL
	IORING_OP_LINK_TIMEOUT
	IORING_OP_CONNECT
	IORING_OP_FALLOCATE
	IORING_OP_OPENAT
	IORING_OP_CLOSE
	IORING_OP_FILES_UPDATE
	IORING_OP_STATX
	IORING_OP_READ
	IORING_OP_WRITE
	IORING_OP_FADVISE
	IORING_OP_MADVISE
	IORING_OP_SEND
	IORING_OP_RECV
	IORING_OP_OPENAT2
	IORING_OP_EPOLL_CTL
	IORING_OP_SPLICE
	IORING_OP_PROVIDE_BUFFERS
	IORING_OP_REMOVE_BUFFERS
	IORING_OP_TEE
	IORING_OP_SHUTDOWN
	IORING_OP_RENAMEAT
	IORING_OP_UNLINKAT
	IORING_OP_MKDIRAT
	IORING_OP_SYMLINKAT
	IORING_OP_LINKAT
)

func (ur *ioUring) nop(ctx context.Context, fd int) error {
	return ur.submit(contextWithFD(ctx, fd), IORING_OP_NOP, fd, 0, 0, 0, 0)
}

func (ur *ioUring) readv(ctx context.Context, fd int, iov [][]byte) error {
	if iov == nil || len(iov) < 1 {
		return ErrInvalidParam
	}
	opcode := IORING_OP_READV
	addr, n := ioVecFromBytesSlice(iov)

	return ur.submit(contextWithFD(ctx, fd), opcode, fd, 0, uint64(addr), n, unix.MSG_WAITALL)
}

func (ur *ioUring) writev(ctx context.Context, fd int, iov [][]byte) error {
	if iov == nil || len(iov) < 1 {
		return ErrInvalidParam
	}

	opcode := IORING_OP_WRITEV
	addr, n := ioVecFromBytesSlice(iov)

	return ur.submit(contextWithFD(ctx, fd), opcode, fd, 0, uint64(addr), n, unix.MSG_ZEROCOPY)
}

func (ur *ioUring) fsync(ctx context.Context, fd int) error {
	return ur.submit(contextWithFD(ctx, fd), IORING_OP_FSYNC, fd, 0, 0, 0, 0)
}

func (ur *ioUring) sendmsg(ctx context.Context, fd int, buffers [][]byte, oob []byte, to unix.Sockaddr) error {
	saPtr, saN, err := unsafe.Pointer(uintptr(0)), 0, error(nil)
	if to != nil {
		saPtr, saN, err = sockaddr(to)
		if err != nil {
			return err
		}
	}
	opcode := IORING_OP_SENDMSG
	addr, n := ioVecFromBytesSlice(buffers)
	msg := unix.Msghdr{
		Name:       (*byte)(saPtr),
		Namelen:    uint32(saN),
		Iov:        (*unix.Iovec)(unsafe.Pointer(addr)),
		Iovlen:     uint64(n),
		Control:    nil,
		Controllen: 0,
	}
	if len(oob) > 0 {
		msg.Control = &oob[0]
		msg.Controllen = uint64(len(oob))
	}

	return ur.submit(contextWithFD(ctx, fd), opcode, fd, 0, uint64(addr), 1, unix.MSG_ZEROCOPY)
}

func (ur *ioUring) recvmsg(ctx context.Context, fd int, buffers [][]byte, oob []byte) error {
	from := unix.RawSockaddrAny{}
	opcode := IORING_OP_RECVMSG
	addr, n := ioVecFromBytesSlice(buffers)
	msg := unix.Msghdr{
		Name:       (*byte)(unsafe.Pointer(&from)),
		Namelen:    uint32(unix.SizeofSockaddrAny),
		Iov:        (*unix.Iovec)(unsafe.Pointer(addr)),
		Iovlen:     uint64(n),
		Control:    nil,
		Controllen: 0,
	}
	if len(oob) > 0 {
		msg.Control = &oob[0]
		msg.Controllen = uint64(len(oob))
	}

	return ur.submit(contextWithFD(ctx, fd), opcode, fd, 0, uint64(addr), 1, unix.MSG_WAITALL)
}

func (ur *ioUring) accept(ctx context.Context, fd int) error {
	return ur.submit(contextWithFD(ctx, fd), IORING_OP_ACCEPT, fd, 0, 0, 0, unix.SOCK_NONBLOCK|unix.SOCK_CLOEXEC)
}

func (ur *ioUring) close(ctx context.Context, fd int) error {
	return ur.submit(contextWithFD(ctx, fd), IORING_OP_CLOSE, fd, 0, 0, 0, 0)
}

func (ur *ioUring) read(ctx context.Context, fd int, p []byte) error {
	if p == nil || len(p) < 1 {
		return ErrInvalidParam
	}

	opcode := IORING_OP_READ
	addr := uint64(uintptr(unsafe.Pointer(&p[0])))

	return ur.submit(contextWithFD(ctx, fd), opcode, fd, 0, addr, len(p), 0)
}

func (ur *ioUring) write(ctx context.Context, fd int, p []byte, n int) error {
	if p == nil || len(p) < 1 {
		return ErrInvalidParam
	}

	opcode := IORING_OP_WRITE
	addr := uint64(uintptr(unsafe.Pointer(&p[0])))

	return ur.submit(contextWithFD(ctx, fd), opcode, fd, 0, addr, n, 0)
}

func (ur *ioUring) send(ctx context.Context, fd int, p []byte) error {
	if p == nil || len(p) < 1 {
		return ErrInvalidParam
	}
	opcode := IORING_OP_SEND
	addr := uint64(uintptr(unsafe.Pointer(&p[0])))

	return ur.submit(contextWithFD(ctx, fd), opcode, fd, 0, addr, len(p), unix.MSG_ZEROCOPY)
}

func (ur *ioUring) receive(ctx context.Context, fd int, p []byte) error {
	if p == nil || len(p) < 1 {
		return ErrInvalidParam
	}
	opcode := IORING_OP_RECV
	addr := uint64(uintptr(unsafe.Pointer(&p[0])))

	return ur.submit(contextWithFD(ctx, fd), opcode, fd, 0, addr, len(p), unix.MSG_WAITALL)
}

func (ur *ioUring) epollCtl(ctx context.Context, epfd int, op int, fd int, events uint32) error {
	e := unix.EpollEvent{Events: events, Fd: int32(fd)}
	opcode := IORING_OP_EPOLL_CTL
	addr := uint64(uintptr(unsafe.Pointer(&e)))

	return ur.submit(ctx, opcode, epfd, uint64(fd), addr, op, 0)
}
