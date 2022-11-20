//go:build linux

package sox

import (
	"golang.org/x/sys/unix"
	"reflect"
	"runtime"
	"sync/atomic"
	"syscall"
	"time"
	"unsafe"
)

const (
	IORING_SETUP_IOPOLL = 1 << 0
	IORING_SETUP_SQPOLL = 1 << 1
	IORING_SETUP_SQ_AFF = 1 << 2
)

const (
	IORING_ENTER_GETEVENTS       = 1 << 0
	IORING_ENTER_SQ_WAKEUP       = 1 << 1
	IORING_ENTER_SQ_WAIT         = 1 << 2
	IORING_ENTER_EXT_ARG         = 1 << 3
	IORING_ENTER_REGISTERED_RING = 1 << 4
)

const (
	IORING_OFF_SQ_RING int64 = 0
	IORING_OFF_CQ_RING int64 = 0x8000000
	IORING_OFF_SQES    int64 = 0x10000000
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

const (
	IORING_SQ_NEED_WAKEUP = 1 << iota
	IORING_SQ_CQ_OVERFLOW
)

const (
	IOSQE_FIXED_FILE = 1 << iota
	IOSQE_IO_DRAIN
	IOSQE_IO_LINK
	IOSQE_IO_HARDLINK
	IOSQE_ASYNC
)

const (
	REGISTER_EVENTFD_ASYNC uintptr = 7
)

const (
	ioUringDefaultEntries = 0x2000

	ioUringDefaultSqThreadCPU  = 1
	ioUringDefaultSqThreadIdle = 5 * time.Second
)

type ioUring struct {
	params *ioUringParams

	sq     ioUringSq
	sqLock atomic.Bool
	cq     ioUringCq
	ringFd int
}

func newIoUring(entries int, opts ...func(params *ioUringParams)) (*ioUring, error) {
	if entries < 1 {
		return nil, ErrInvalidParam
	}

	params := new(ioUringParams)
	*params = *ioUringDefaultParams
	for _, opt := range opts {
		opt(params)
	}

	fd, err := ioUringSetup(uint32(entries), params)
	if err != nil {
		return nil, err
	}

	uring := &ioUring{
		params: params,

		sq: ioUringSq{
			ringSz: params.sqOff.array + uint32(unsafe.Sizeof(uint32(0)))*params.sqEntries,
		},
		sqLock: atomic.Bool{},
		cq: ioUringCq{
			ringSz: params.cqOff.cqes + uint32(unsafe.Sizeof(uint32(0)))*params.cqEntries,
		},
		ringFd: fd,
	}

	b, err := unix.Mmap(uring.ringFd, IORING_OFF_SQ_RING, int(uring.sq.ringSz), unix.PROT_READ|unix.PROT_WRITE|unix.PROT_EXEC, unix.MAP_SHARED|unix.MAP_POPULATE)
	if err != nil {
		return uring, errFromUnixErrno(err)
	}
	ptr := uintptr(unsafe.Pointer(&b[0]))
	uring.sq.kHead = (*uint32)(unsafe.Pointer(ptr + uintptr(params.sqOff.head)))
	uring.sq.kTail = (*uint32)(unsafe.Pointer(ptr + uintptr(params.sqOff.tail)))
	uring.sq.kRingMask = (*uint32)(unsafe.Pointer(ptr + uintptr(params.sqOff.ringMask)))
	uring.sq.kRingEntries = (*uint32)(unsafe.Pointer(ptr + uintptr(params.sqOff.ringEntries)))
	uring.sq.kFlags = (*uint32)(unsafe.Pointer(ptr + uintptr(params.sqOff.flags)))
	uring.sq.kDropped = (*uint32)(unsafe.Pointer(ptr + uintptr(params.sqOff.dropped)))
	uring.sq.array = *(*[]uint32)(
		unsafe.Pointer(&reflect.SliceHeader{
			Data: ptr + uintptr(params.sqOff.array),
			Len:  int(params.sqEntries),
			Cap:  int(params.sqEntries),
		}))
	b, err = unix.Mmap(uring.ringFd, IORING_OFF_SQES, int(params.sqEntries)*int(unsafe.Sizeof(ioUringSqe{})), unix.PROT_READ|unix.PROT_WRITE|unix.PROT_EXEC, unix.MAP_SHARED|unix.MAP_POPULATE)
	if err != nil {
		return uring, errFromUnixErrno(err)
	}
	uring.sq.sqes = *(*[]ioUringSqe)(
		unsafe.Pointer(&reflect.SliceHeader{
			Data: uintptr(unsafe.Pointer(&b[0])),
			Len:  int(params.sqEntries),
			Cap:  int(params.sqEntries),
		}))

	b, err = unix.Mmap(uring.ringFd, IORING_OFF_CQ_RING, int(uring.cq.ringSz), unix.PROT_READ|unix.PROT_WRITE|unix.PROT_EXEC, unix.MAP_SHARED|unix.MAP_POPULATE)
	if err != nil {
		return uring, errFromUnixErrno(err)
	}
	ptr = uintptr(unsafe.Pointer(&b[0]))
	uring.cq.kHead = (*uint32)(unsafe.Pointer(ptr + uintptr(params.cqOff.head)))
	uring.cq.kTail = (*uint32)(unsafe.Pointer(ptr + uintptr(params.cqOff.tail)))
	uring.cq.kRingMask = (*uint32)(unsafe.Pointer(ptr + uintptr(params.cqOff.ringMask)))
	uring.cq.kRingEntries = (*uint32)(unsafe.Pointer(ptr + uintptr(params.cqOff.ringEntries)))
	uring.cq.kOverflow = (*uint32)(unsafe.Pointer(ptr + uintptr(params.cqOff.overflow)))

	uring.cq.cqes = *(*[]ioUringCqe)(
		unsafe.Pointer(&reflect.SliceHeader{
			Data: ptr + uintptr(params.cqOff.cqes),
			Len:  int(params.cqEntries),
			Cap:  int(params.cqEntries),
		}))

	return uring, nil
}

func (ur *ioUring) nop(fd int, ctx *Ctx) error {
	return ur.submit(IORING_OP_NOP, fd, 0, 0, 0, 0, ctx)
}

func (ur *ioUring) readv(fd int, iov [][]byte, ctx *Ctx) error {
	if iov == nil || len(iov) < 1 {
		return ErrInvalidParam
	}
	opcode := IORING_OP_READV
	addr, n := ioVecFromSliceOfBytes(iov)

	return ur.submit(opcode, fd, 0, addr, n, unix.MSG_WAITALL, ctx)
}

func (ur *ioUring) writev(fd int, iov [][]byte, ctx *Ctx) error {
	if iov == nil || len(iov) < 1 {
		return ErrInvalidParam
	}

	opcode := IORING_OP_WRITEV
	addr, n := ioVecFromSliceOfBytes(iov)

	return ur.submit(opcode, fd, 0, addr, n, 0, ctx)
}

func (ur *ioUring) fsync(fd int, ctx *Ctx) error {
	return ur.submit(IORING_OP_FSYNC, fd, 0, 0, 0, 0, ctx)
}

func (ur *ioUring) accept(fd int, ctx *Ctx) error {
	return ur.submit(IORING_OP_ACCEPT, fd, 0, 0, 0, unix.SOCK_NONBLOCK|unix.SOCK_CLOEXEC, ctx)
}

func (ur *ioUring) close(fd int, ctx *Ctx) error {
	return ur.submit(IORING_OP_CLOSE, fd, 0, 0, 0, 0, ctx)
}

func (ur *ioUring) read(fd int, p []byte, ctx *Ctx) error {
	if p == nil || len(p) < 1 {
		return ErrInvalidParam
	}

	opcode := IORING_OP_READ
	addr := uint64(uintptr(unsafe.Pointer(&p[0])))

	return ur.submit(opcode, fd, 0, addr, len(p), 0, ctx)
}

func (ur *ioUring) write(fd int, p []byte, n int, ctx *Ctx) error {
	if p == nil || len(p) < 1 {
		return ErrInvalidParam
	}

	opcode := IORING_OP_WRITE
	addr := uint64(uintptr(unsafe.Pointer(&p[0])))

	return ur.submit(opcode, fd, 0, addr, n, 0, ctx)
}

func (ur *ioUring) send(fd int, p []byte, ctx *Ctx) error {
	if p == nil || len(p) < 1 {
		return ErrInvalidParam
	}
	opcode := IORING_OP_SEND
	addr := uint64(uintptr(unsafe.Pointer(&p[0])))

	return ur.submit(opcode, fd, 0, addr, len(p), 0, ctx)
}

func (ur *ioUring) receive(fd int, p []byte, ctx *Ctx) error {
	if p == nil || len(p) < 1 {
		return ErrInvalidParam
	}
	opcode := IORING_OP_RECV
	addr := uint64(uintptr(unsafe.Pointer(&p[0])))

	return ur.submit(opcode, fd, 0, addr, len(p), unix.MSG_WAITALL, ctx)
}

func (ur *ioUring) epollCtl(epfd int, op int, fd int, events uint32, ctx *Ctx) error {
	e := unix.EpollEvent{Events: events, Fd: int32(fd)}
	opcode := IORING_OP_EPOLL_CTL
	addr := uint64(uintptr(unsafe.Pointer(&e)))

	return ur.submit(opcode, epfd, uint64(fd), addr, op, 0, ctx)
}

func (ur *ioUring) submit(op uint8, fd int, off uint64, addr uint64, n int, uflags uint32, ctx *Ctx) error {
	for !ur.sqLock.CompareAndSwap(false, true) {
		runtime.Gosched()
		continue
	}
	defer ur.sqLock.Store(false)

	for {
		h, t := *ur.sq.kHead, *ur.sq.kTail
		if (t+1)&*ur.sq.kRingMask == h {
			break
		}

		e := &ur.sq.sqes[t]
		e.opcode = op
		e.flags = IOSQE_ASYNC
		e.fd = int32(fd)
		e.off = off
		e.addr = addr
		e.len = uint32(n)
		e.uflags = uflags
		e.userData = uint64(uintptr(unsafe.Pointer(ctx)))

		*ur.sq.kTail = (t + 1) & (*ur.sq.kRingMask)

		return nil
	}

	return ErrTemporarilyUnavailable
}

func (ur *ioUring) enter() error {
	if atomic.LoadUint32(ur.sq.kFlags)&IORING_SQ_NEED_WAKEUP != 0 {
		_, err := ioUringEnter(ur.ringFd, uintptr(ur.params.sqEntries), 0, IORING_ENTER_SQ_WAKEUP)
		if err != nil {
			return err
		}
	}
	if (ur.params.flags&IORING_SETUP_SQPOLL == 0) && *ur.sq.kHead != *ur.sq.kTail {
		_, err := ioUringEnter(ur.ringFd, uintptr(ur.params.sqEntries), 0, 0)
		if err != nil {
			return err
		}
	}

	return nil
}

func (ur *ioUring) wait() (*ioUringCqe, error) {
	for {
		h, t := atomic.LoadUint32(ur.cq.kHead), atomic.LoadUint32(ur.cq.kTail)
		if h == t {
			break
		}

		e := &ur.cq.cqes[h]
		ok := atomic.CompareAndSwapUint32(ur.cq.kHead, h, (h+1)&(*ur.cq.kRingMask))
		if !ok {
			runtime.Gosched()
			continue
		}

		return e, nil
	}

	return nil, ErrTemporarilyUnavailable
}

type ioUringSq struct {
	kHead        *uint32
	kTail        *uint32
	kRingMask    *uint32
	kRingEntries *uint32
	kDropped     *uint32
	kFlags       *uint32
	array        []uint32
	sqes         []ioUringSqe

	ringSz uint32
}
type ioUringSqe struct {
	opcode   uint8
	flags    uint8
	ioprio   uint16
	fd       int32
	off      uint64
	addr     uint64
	len      uint32
	uflags   uint32
	userData uint64

	bufIndex    uint16
	personality uint16
	spliceFdIn  int32
	pad         [2]uint64
}

type ioUringCq struct {
	kHead        *uint32
	kTail        *uint32
	kRingMask    *uint32
	kRingEntries *uint32
	kOverflow    *uint32
	cqes         []ioUringCqe

	ringSz uint32
}

type ioUringCqe struct {
	userData uint64
	res      int32
	flags    uint32
}

type ioSqRingOffsets struct {
	head        uint32
	tail        uint32
	ringMask    uint32
	ringEntries uint32
	flags       uint32
	dropped     uint32
	array       uint32
	resv        [3]uint32
}

type ioCqRingOffsets struct {
	head        uint32
	tail        uint32
	ringMask    uint32
	ringEntries uint32
	overflow    uint32
	cqes        uint32
	flags       uint32
	resv        [3]uint32
}

type ioUringParams struct {
	sqEntries    uint32
	cqEntries    uint32
	flags        uint32
	sqThreadCPU  uint32
	sqThreadIdle uint32
	features     uint32
	wqFd         uint32
	resv         [3]uint32
	sqOff        ioSqRingOffsets
	cqOff        ioCqRingOffsets
}

var (
	ioUringDefaultParams = &ioUringParams{}
	ioUringIoPollOptions = func(params *ioUringParams) {
		params.flags |= IORING_SETUP_IOPOLL
	}
	ioUringSqPollOptions = func(params *ioUringParams) {
		params.flags |= IORING_SETUP_SQPOLL | IORING_SETUP_SQ_AFF
		params.sqThreadCPU = ioUringDefaultSqThreadCPU
		params.sqThreadIdle = uint32(ioUringDefaultSqThreadIdle.Milliseconds())
	}
)

func ioUringSetup(entries uint32, params *ioUringParams) (fd int, err error) {
	r1, _, errno := syscall.Syscall(
		unix.SYS_IO_URING_SETUP,
		uintptr(entries),
		uintptr(unsafe.Pointer(params)),
		0,
	)
	if errno != 0 {
		err = errFromUnixErrno(errno)
		return
	}
	fd, err = int(r1), nil
	return
}

func ioUringEnter(fd int, toSubmit uintptr, minComplete uintptr, flags uintptr) (n int, err error) {
	result, _, errno := unix.Syscall6(unix.SYS_IO_URING_ENTER, uintptr(fd), toSubmit, minComplete, flags, 0, 0)
	if errno != 0 {
		return int(result), errFromUnixErrno(errno)
	}
	return int(result), nil
}
