// ©Hayabusa Cloud Co., Ltd. 2022. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

//go:build linux

package sox

import (
	"context"
	"golang.org/x/sys/unix"
	"math"
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
	IORING_OP_MSG_RING
	IORING_OP_FSETXATTR
	IORING_OP_SETXATTR
	IORING_OP_FGETXATTR
	IORING_OP_GETXATTR
	IORING_OP_SOCKET
	IORING_OP_URING_CMD
	IORING_OP_SEND_ZC
	IORING_OP_SENDMSG_ZC
	IORING_OP_READ_MULTISHOT
	IORING_OP_WAITID
	IORING_OP_FUTEX_WAIT
	IORING_OP_FUTEX_WAKE
	IORING_OP_FUTEX_WAITV
	IORING_OP_FIXED_FD_INSTALL
)

func (ur *ioUring) nop(ctx context.Context, flags uint8, fd int) error {
	return ur.submit3(ContextWithUserdata(ctx, ioUringFd(fd)), IORING_OP_NOP, flags, 0, fd, 0, 0)
}

func (ur *ioUring) readv(ctx context.Context, flags uint8, fd int, iov [][]byte) error {
	if iov == nil || len(iov) < 1 {
		return ErrInvalidParam
	}
	opcode := IORING_OP_READV
	addr, n := ioVecFromBytesSlice(iov)

	return ur.submit6(ContextWithUserdata(ctx, ioUringFd(fd)), opcode, flags, 0, fd, 0, uint64(addr), n, unix.MSG_WAITALL)
}

func (ur *ioUring) writev(ctx context.Context, flags uint8, fd int, iov [][]byte) error {
	if iov == nil || len(iov) < 1 {
		return ErrInvalidParam
	}

	opcode := IORING_OP_WRITEV
	addr, n := ioVecFromBytesSlice(iov)

	return ur.submit3(ContextWithUserdata(ctx, ioUringFd(fd)), opcode, flags, 0, fd, uint64(addr), n)
}

func (ur *ioUring) fsync(ctx context.Context, flags uint8, fd int) error {
	return ur.submit3(ContextWithUserdata(ctx, ioUringFd(fd)), IORING_OP_FSYNC, flags, 0, fd, 0, 0)
}

func (ur *ioUring) readFixed(ctx context.Context, flags uint8, fd int, i int) (buf []byte, err error) {
	if i < 0 || i >= len(ur.bufs) {
		return nil, ErrInvalidParam
	}
	opcode := IORING_OP_READ_FIXED
	addr := uint64(uintptr(unsafe.Pointer(unsafe.SliceData(ur.bufs[i]))))

	return ur.bufs[i], ur.submit9(ContextWithUserdata(ctx, ioUringFd(fd)), opcode, flags, 0, fd, 0, addr, len(ur.bufs[i]), unix.MSG_WAITALL, uint16(i), 0, 0)
}

func (ur *ioUring) writeFixed(ctx context.Context, flags uint8, fd int, i int, n int) error {
	if i < 0 || i >= len(ur.bufs) || n < 0 || n > len(ur.bufs[i]) {
		return ErrInvalidParam
	}
	opcode := IORING_OP_WRITE_FIXED
	addr := uint64(uintptr(unsafe.Pointer(unsafe.SliceData(ur.bufs[i]))))

	return ur.submit9(ContextWithUserdata(ctx, ioUringFd(fd)), opcode, flags, 0, fd, 0, addr, n, 0, uint16(i), 0, 0)
}

func (ur *ioUring) pollAdd(ctx context.Context, flags uint8, fd int, how int, events int) error {
	return ur.submit6(ctx, IORING_OP_POLL_ADD, flags, 0, fd, 0, 0, how, uint32(events))
}

func (ur *ioUring) pollRemove(ctx context.Context, flags uint8) error {
	return ur.submit3(ctx, IORING_OP_POLL_REMOVE, flags, 0, 0, 0, 0)
}

func (ur *ioUring) syncFileRange(ctx context.Context, flags uint8, fd int, off int64, n int, uflags int) error {
	return ur.submit6(ctx, IORING_OP_SYNC_FILE_RANGE, flags, 0, fd, uint64(off), 0, n, uint32(uflags))
}

func (ur *ioUring) sendmsg(ctx context.Context, flags uint8, ioprio uint16, fd int, buffers [][]byte, oob []byte, to unix.Sockaddr) error {
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

	return ur.submit6(ContextWithUserdata(ctx, ioUringFd(fd)), opcode, flags, ioprio, fd, 0, uint64(addr), 1, 0)
}

func (ur *ioUring) recvmsg(ctx context.Context, flags uint8, ioprio uint16, fd int, buffers [][]byte, oob []byte) error {
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

	return ur.submit6(ContextWithUserdata(ctx, ioUringFd(fd)), opcode, flags, ioprio, fd, 0, uint64(addr), 1, unix.MSG_WAITALL)
}

const (
	IORING_TIMEOUT_ABS = 1 << iota
	IORING_TIMEOUT_UPDATE
	IORING_TIMEOUT_BOOTTIME
	IORING_TIMEOUT_REALTIME
	IORING_LINK_TIMEOUT_UPDATE
	IORING_TIMEOUT_ETIME_SUCCESS
	IORING_TIMEOUT_MULTISHOT
	IORING_TIMEOUT_CLOCK_MASK  = IORING_TIMEOUT_BOOTTIME | IORING_TIMEOUT_REALTIME
	IORING_TIMEOUT_UPDATE_MASK = IORING_TIMEOUT_UPDATE | IORING_LINK_TIMEOUT_UPDATE
)

func (ur *ioUring) timeout(ctx context.Context, flags uint8, cnt int, ts *unix.Timespec, uflags int) error {
	addr := uint64(uintptr(unsafe.Pointer(ts)))
	return ur.submit6(ctx, IORING_OP_TIMEOUT, flags, 0, 0, uint64(cnt), addr, 1, uint32(uflags))
}

func (ur *ioUring) timeoutUpdate(ctx context.Context, flags uint8, userData unsafe.Pointer, ts *unix.Timespec, uflags int) error {
	addr := uint64(uintptr(userData))
	addr2 := uint64(uintptr(unsafe.Pointer(ts)))
	return ur.submit6(ctx, IORING_OP_TIMEOUT_REMOVE, flags, 0, 0, addr2, addr, 0, uint32(uflags|IORING_TIMEOUT_UPDATE))
}

func (ur *ioUring) timeoutRemove(ctx context.Context, flags uint8, userData unsafe.Pointer, uflags int) error {
	addr := uint64(uintptr(userData))
	return ur.submit6(ctx, IORING_OP_TIMEOUT_REMOVE, flags, 0, 0, 0, addr, 0, uint32(uflags))
}

const (
	IORING_ACCEPT_MULTISHOT  = 1 << 0
	IORING_ACCEPT_DONTWAIT   = 1 << 1
	IORING_ACCEPT_POLL_FIRST = 1 << 2
)

func (ur *ioUring) accept(ctx context.Context, flags uint8, ioprio uint16, fd int) error {
	return ur.submit6(ContextWithUserdata(ctx, ioUringFd(fd)), IORING_OP_ACCEPT, flags, ioprio, fd, 0, 0, 0, unix.SOCK_NONBLOCK|unix.SOCK_CLOEXEC)
}

func (ur *ioUring) asyncCancel(ctx context.Context, flags uint8, userdata *ioUringCtx) error {
	addr := uint64(uintptr(unsafe.Pointer(userdata)))
	return ur.submit3(ctx, IORING_OP_ASYNC_CANCEL, flags, 0, 0, addr, 0)
}

func (ur *ioUring) linkTimeout(ctx context.Context, flags uint8, ts *unix.Timespec, uflags int) error {
	addr := uint64(uintptr(unsafe.Pointer(ts)))
	return ur.submit6(ctx, IORING_OP_LINK_TIMEOUT, flags, 0, 0, 0, addr, 1, uint32(uflags))
}

func (ur *ioUring) connect(ctx context.Context, flags uint8, fd int, sa unix.Sockaddr) error {
	ptr, n, err := sockaddr(sa)
	if err != nil {
		return err
	}

	return ur.submit6(ContextWithUserdata(ctx, ioUringFd(fd)), IORING_OP_CONNECT, flags, 0, fd, uint64(n), uint64(uintptr(ptr)), 0, 0)
}

func (ur *ioUring) fAllocate(ctx context.Context, flags uint8, fd int, mode uint32, off int64, len int64) error {
	return ur.submit6(ContextWithUserdata(ctx, ioUringFd(fd)), IORING_OP_FALLOCATE, flags, 0, fd, uint64(off), uint64(len), int(mode), 0)
}

func (ur *ioUring) openAt(ctx context.Context, flags uint8, dirfd int, pathname string, uflags int, mode uint32) error {
	ptr, err := unix.BytePtrFromString(pathname)
	if err != nil {
		return err
	}
	addr := uint64(uintptr(unsafe.Pointer(ptr)))

	return ur.submit6(ContextWithUserdata(ctx, ioUringFd(dirfd)), IORING_OP_OPENAT, flags, 0, dirfd, 0, addr, int(mode), uint32(uflags|unix.O_LARGEFILE))
}

func (ur *ioUring) close(ctx context.Context, flags uint8, fd int) error {
	return ur.submit3(ContextWithUserdata(ctx, ioUringFd(fd)), IORING_OP_CLOSE, flags, 0, fd, 0, 0)
}

func (ur *ioUring) statx(ctx context.Context, flags uint8, dirfd int, path string, uflags int, mask int, stat *unix.Statx_t) error {
	ptr, err := unix.BytePtrFromString(path)
	if err != nil {
		return err
	}
	addr := uint64(uintptr(unsafe.Pointer(ptr)))
	off := uint64(uintptr(unsafe.Pointer(stat)))

	return ur.submit6(ContextWithUserdata(ctx, ioUringFd(dirfd)), IORING_OP_STATX, flags, 0, dirfd, off, addr, mask, uint32(uflags))
}

func (ur *ioUring) read(ctx context.Context, flags uint8, fd int, p []byte) error {
	if p == nil || len(p) < 1 {
		return ErrInvalidParam
	}

	opcode := IORING_OP_READ
	addr := uint64(uintptr(unsafe.Pointer(unsafe.SliceData(p))))

	return ur.submit3(ContextWithUserdata(ctx, ioUringFd(fd)), opcode, flags, 0, fd, addr, len(p))
}

func (ur *ioUring) readWithBufferSelect(ctx context.Context, flags uint8, fd int, n int, group uint16) error {
	opcode := IORING_OP_READ
	flags |= IOSQE_BUFFER_SELECT

	return ur.submit9(ContextWithUserdata(ctx, ioUringFd(fd)), opcode, flags, 0, fd, 0, 0, n, 0, group, 0, 0)
}

func (ur *ioUring) write(ctx context.Context, flags uint8, fd int, p []byte, n int) error {
	if p == nil || len(p) < 1 {
		return ErrInvalidParam
	}

	opcode := IORING_OP_WRITE
	addr := uint64(uintptr(unsafe.Pointer(unsafe.SliceData(p))))

	return ur.submit3(ContextWithUserdata(ctx, ioUringFd(fd)), opcode, flags, 0, fd, addr, n)
}

func (ur *ioUring) fadvise(ctx context.Context, flags uint8, fd int, offset int64, n int, advice int) error {
	return ur.submit6(ContextWithUserdata(ctx, ioUringFd(fd)), IORING_OP_FADVISE, flags, 0, fd, uint64(offset), 0, n, uint32(advice))
}

func (ur *ioUring) madvise(ctx context.Context, flags uint8, b []byte, advice int) error {
	addr := uint64(uintptr(unsafe.Pointer(unsafe.SliceData(b))))
	return ur.submit6(ctx, IORING_OP_MADVISE, flags, 0, 0, 0, addr, len(b), uint32(advice))
}

const (
	IORING_RECVSEND_POLL_FIRST = 1 << iota
	IORING_RECV_MULTISHOT
	IORING_RECVSEND_FIXED_BUF
	IORING_SEND_ZC_REPORT_USAGE
	IORING_RECVSEND_BUNDLE
)

func (ur *ioUring) send(ctx context.Context, flags uint8, fd int, p []byte) error {
	if p == nil || len(p) < 1 {
		return ErrInvalidParam
	}
	opcode := IORING_OP_SEND
	addr := uint64(uintptr(unsafe.Pointer(unsafe.SliceData(p))))

	return ur.submit3(ContextWithUserdata(ctx, ioUringFd(fd)), opcode, flags, 0, fd, addr, len(p))
}

func (ur *ioUring) receive(ctx context.Context, flags uint8, ioprio uint16, fd int, p []byte) error {
	if p == nil || len(p) < 1 {
		return ErrInvalidParam
	}
	opcode := IORING_OP_RECV
	addr := uint64(uintptr(unsafe.Pointer(unsafe.SliceData(p))))

	return ur.submit6(ContextWithUserdata(ctx, ioUringFd(fd)), opcode, flags, ioprio, fd, 0, addr, len(p), unix.MSG_WAITALL)
}

func (ur *ioUring) receiveWithBufferSelect(ctx context.Context, flags uint8, ioprio uint16, fd int, n int, group uint16) error {
	opcode := IORING_OP_RECV
	flags |= IOSQE_BUFFER_SELECT
	ctx = ContextWithUserdata(ctx, bufferGroupIndex(group))

	return ur.submit9(ContextWithUserdata(ctx, ioUringFd(fd)), opcode, flags, ioprio, fd, 0, 0, n, 0, group, 0, 0)
}

func (ur *ioUring) openAt2(ctx context.Context, flags uint8, dirfd int, pathname string, how *unix.OpenHow) error {
	ptr, err := unix.BytePtrFromString(pathname)
	if err != nil {
		return err
	}
	addr := uint64(uintptr(unsafe.Pointer(ptr)))
	off := uint64(uintptr(unsafe.Pointer(how)))

	return ur.submit6(ContextWithUserdata(ctx, ioUringFd(dirfd)), IORING_OP_OPENAT2, flags, 0, dirfd, off, addr, unix.SizeofOpenHow, 0)
}

func (ur *ioUring) epollCtl(ctx context.Context, flags uint8, epfd int, op int, fd int, events uint32) error {
	e := unix.EpollEvent{Events: events, Fd: int32(fd)}
	opcode := IORING_OP_EPOLL_CTL
	addr := uint64(uintptr(unsafe.Pointer(&e)))

	return ur.submit6(ContextWithUserdata(ctx, ioUringFd(epfd)), opcode, flags, 0, epfd, uint64(fd), addr, op, 0)
}

func (ur *ioUring) epollAdd(ctx context.Context, epfd int, fd int, events uint32) error {
	return ur.epollCtl(ContextWithUserdata(ctx, ioUringFd(epfd)), 0, epfd, unix.EPOLL_CTL_ADD, fd, events)
}

func (ur *ioUring) epollMod(ctx context.Context, epfd int, fd int, events uint32) error {
	return ur.epollCtl(ContextWithUserdata(ctx, ioUringFd(epfd)), 0, epfd, unix.EPOLL_CTL_MOD, fd, events)
}

func (ur *ioUring) epollDel(ctx context.Context, epfd int, fd int) error {
	return ur.epollCtl(ContextWithUserdata(ctx, ioUringFd(epfd)), 0, epfd, unix.EPOLL_CTL_DEL, fd, 0)
}

func (ur *ioUring) splice(ctx context.Context, flags uint8, rfd int, rOff *int64, wfd int, wOff *int64, n int, uflags int) error {
	rOffVal, wOffVal := uint64(math.MaxUint64), uint64(math.MaxUint64)
	if rOff != nil {
		rOffVal = uint64(*rOff)
	}
	if wOff != nil {
		wOffVal = uint64(*wOff)
	}
	return ur.submit9(ContextWithUserdata(ctx, ioUringFd(wfd)), IORING_OP_SPLICE, flags, 0, wfd, wOffVal, rOffVal, n, uint32(uflags), 0, 0, rfd)
}

func (ur *ioUring) provideBuffers(ctx context.Context, flags uint8, num int, starting int, addr unsafe.Pointer, size int, group uint16) error {
	return ur.submit9(ctx, IORING_OP_PROVIDE_BUFFERS, flags, 0, num, uint64(starting), uint64(uintptr(addr)), size, 0, group, 0, 0)
}

func (ur *ioUring) removeBuffers(ctx context.Context, flags uint8, num int, group uint16) error {
	return ur.submit9(ctx, IORING_OP_REMOVE_BUFFERS, flags, 0, num, 0, 0, 0, 0, group, 0, 0)
}

func (ur *ioUring) tee(ctx context.Context, flags uint8, rfd int, wfd int, n int, uflags int) error {
	return ur.submit9(ContextWithUserdata(ctx, ioUringFd(wfd)), IORING_OP_TEE, flags, 0, wfd, 0, 0, n, uint32(uflags), 0, 0, rfd)
}

func (ur *ioUring) shutdown(ctx context.Context, flags uint8, fd int, how int) error {
	return ur.submit3(ContextWithUserdata(ctx, ioUringFd(fd)), IORING_OP_SHUTDOWN, flags, 0, fd, 0, how)
}

func (ur *ioUring) renameAt(ctx context.Context, flags uint8, oldPath, newPath string, uflags int) error {
	oldPtr, err := unix.BytePtrFromString(oldPath)
	if err != nil {
		return err
	}
	oldAddr := uint64(uintptr(unsafe.Pointer(oldPtr)))
	newPtr, err := unix.BytePtrFromString(newPath)
	if err != nil {
		return err
	}
	newAddr := uint64(uintptr(unsafe.Pointer(newPtr)))

	return ur.submit6(ctx, IORING_OP_RENAMEAT, flags, 0, unix.AT_FDCWD, newAddr, oldAddr, unix.AT_FDCWD, uint32(uflags))
}

func (ur *ioUring) unlinkAt(ctx context.Context, flags uint8, dirfd int, pathname string, uflags int) error {
	ptr, err := unix.BytePtrFromString(pathname)
	if err != nil {
		return err
	}
	addr := uint64(uintptr(unsafe.Pointer(ptr)))

	return ur.submit6(ContextWithUserdata(ctx, ioUringFd(dirfd)), IORING_OP_UNLINKAT, flags, 0, dirfd, 0, addr, 0, uint32(uflags))
}

func (ur *ioUring) mkdirAt(ctx context.Context, flags uint8, dirfd int, pathname string, uflags int, mode uint32) error {
	ptr, err := unix.BytePtrFromString(pathname)
	if err != nil {
		return err
	}
	addr := uint64(uintptr(unsafe.Pointer(ptr)))

	return ur.submit6(ContextWithUserdata(ctx, ioUringFd(dirfd)), IORING_OP_MKDIRAT, flags, 0, dirfd, 0, addr, int(mode), uint32(uflags))
}

func (ur *ioUring) symlinkAt(ctx context.Context, flags uint8, oldPath string, newDirfd int, newPath string) error {
	ptr, err := unix.BytePtrFromString(oldPath)
	if err != nil {
		return err
	}
	oldAddr := uint64(uintptr(unsafe.Pointer(ptr)))

	ptr, err = unix.BytePtrFromString(newPath)
	if err != nil {
		return err
	}
	newAddr := uint64(uintptr(unsafe.Pointer(ptr)))

	return ur.submit6(ContextWithUserdata(ctx, ioUringFd(newDirfd)), IORING_OP_SYMLINKAT, flags, 0, newDirfd, newAddr, oldAddr, 0, 0)
}

func (ur *ioUring) linkAt(ctx context.Context, flags uint8, oldDirfd int, oldPath string, newDirfd int, newPath string, uflags int) error {
	ptr, err := unix.BytePtrFromString(oldPath)
	if err != nil {
		return err
	}
	oldAddr := uint64(uintptr(unsafe.Pointer(ptr)))

	ptr, err = unix.BytePtrFromString(newPath)
	if err != nil {
		return err
	}
	newAddr := uint64(uintptr(unsafe.Pointer(ptr)))

	return ur.submit6(ContextWithUserdata(ctx, ioUringFd(newDirfd)), IORING_OP_LINKAT, flags, 0, oldDirfd, newAddr, oldAddr, newDirfd, uint32(uflags))
}

func (ur *ioUring) msgRing(ctx context.Context, flags uint8, targetFd int, userData int64, res int32) error {
	return ur.submit6(ContextWithUserdata(ctx, ioUringFd(targetFd)), IORING_OP_MSG_RING, flags, 0, targetFd, uint64(userData), 0, int(res), 0)
}

func (ur *ioUring) socket(ctx context.Context, flags uint8, domain, typ, proto int, fileIndex uint32) error {
	return ur.submit9(ctx, IORING_OP_SOCKET, flags, 0, domain, uint64(typ), 0, proto, 0, 0, 0, int(fileIndex))
}

func (ur *ioUring) sendZeroCopy(ctx context.Context, flags uint8, fd int, p []byte, msgFlags uint32) error {
	addr := uint64(uintptr(unsafe.Pointer(unsafe.SliceData(p))))
	return ur.submit6(ContextWithUserdata(ctx, ioUringFd(fd)), IORING_OP_SEND_ZC, flags, 0, fd, 0, addr, len(p), msgFlags)
}

func (ur *ioUring) sendToZeroCopy(ctx context.Context, flags uint8, target Sockaddr, p []byte, msgFlags uint32) error {
	addr2, addrLen, err := sockaddr(target)
	if err != nil {
		return err
	}
	addr := uint64(uintptr(unsafe.Pointer(unsafe.SliceData(p))))
	return ur.submit9(ctx, IORING_OP_SEND_ZC, flags, 0, 0, uint64(uintptr(addr2)), addr, len(p), msgFlags, 0, 0, addrLen<<16)
}
