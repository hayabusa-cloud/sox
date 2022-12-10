//go:build linux

package sox

import (
	"context"
	"golang.org/x/sys/unix"
	"reflect"
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
	IORING_REGISTER_BUFFERS uintptr = iota
	IORING_UNREGISTER_BUFFERS
	IORING_REGISTER_FILES
	IORING_UNREGISTER_FILES
	IORING_REGISTER_EVENTFD
	IORING_UNREGISTER_EVENTFD
	IORING_REGISTER_FILES_UPDATE
	IORING_REGISTER_EVENTFD_ASYNC
	IORING_REGISTER_PROBE
	IORING_REGISTER_PERSONALITY
	IORING_UNREGISTER_PERSONALITY
	IORING_REGISTER_RESTRICTIONS
	IORING_REGISTER_ENABLE_RINGS
	IORING_REGISTER_FILES2
	IORING_REGISTER_FILES_UPDATE2
	IORING_REGISTER_BUFFERS2
	IORING_REGISTER_BUFFERS_UPDATE
	IORING_REGISTER_IOWQ_AFF
	IORING_UNREGISTER_IOWQ_AFF
	IORING_REGISTER_IOWQ_MAX_WORKERS
	IORING_REGISTER_RING_FDS
	IORING_UNREGISTER_RING_FDS
	IORING_REGISTER_PBUF_RING
	IORING_UNREGISTER_PBUF_RING
	IORING_REGISTER_SYNC_CANCEL
	IORING_REGISTER_FILE_ALLOC_RANGE
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
	bufs   Buffers
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
		bufs:   Buffers{},
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

func (ur *ioUring) registerBuffers(n, size int) error {
	if ur.bufs != nil && len(ur.bufs) > 0 {
		panic("io-uring buffers already registered")
	}
	if n < 1 || size < 1 || n != n&(n-1) || size != size&(size-1) {
		return ErrInvalidParam
	}
	ur.bufs = NewBuffers(n, size)
	addr, n := ioVecFromBytesSlice(ur.bufs)
	_, _, errno := unix.Syscall6(unix.SYS_IO_URING_REGISTER, uintptr(ur.ringFd), IORING_REGISTER_BUFFERS, addr, uintptr(n), 0, 0)
	if errno != 0 {
		return errFromUnixErrno(errno)
	}

	return nil
}

func (ur *ioUring) unregisterBuffers() error {
	if ur.bufs == nil || len(ur.bufs) < 1 {
		panic("no io-uring buffers registered")
	}
	_, _, errno := unix.Syscall6(unix.SYS_IO_URING_REGISTER, uintptr(ur.ringFd), IORING_UNREGISTER_BUFFERS, 0, 0, 0, 0)
	if errno != 0 {
		return errFromUnixErrno(errno)
	}
	ur.bufs = Buffers{}

	return nil
}

func (ur *ioUring) registerPoller(p *epoll) (int, error) {
	efd, err := unix.Eventfd(0, unix.EFD_NONBLOCK|unix.EFD_CLOEXEC)
	if err != nil {
		return 0, errFromUnixErrno(err)
	}

	_, _, errno := unix.Syscall6(unix.SYS_IO_URING_REGISTER, uintptr(ur.ringFd), IORING_REGISTER_EVENTFD_ASYNC, uintptr(unsafe.Pointer(&efd)), 1, 0, 0)
	if errno != 0 {
		return 0, errFromUnixErrno(errno)
	}

	err = p.add(efd, unix.EPOLLIN|unix.EPOLLET)
	if err != nil {
		return 0, err
	}

	return efd, nil
}

func (ur *ioUring) submit(ctx context.Context, op uint8, fd int, off uint64, addr uint64, n int, uflags uint32) error {
	for sw := NewSpinWait().SetLevel(spinWaitLevelAtomic); !sw.Closed(); sw.Once() {
		if !ur.sqLock.CompareAndSwap(false, true) {
			continue
		}
		break
	}
	defer ur.sqLock.Store(false)

	h, t := *ur.sq.kHead, *ur.sq.kTail
	if (t+1)&*ur.sq.kRingMask == h {
		return ErrTemporarilyUnavailable
	}

	e := &ur.sq.sqes[t]
	e.opcode = op
	e.flags = IOSQE_ASYNC
	e.fd = int32(fd)
	e.off = off
	e.addr = addr
	e.len = uint32(n)
	e.uflags = uflags
	e.userData = uint64(uintptr(unsafe.Pointer(&ctx)))

	*ur.sq.kTail = (t + 1) & (*ur.sq.kRingMask)

	return nil
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

func (ur *ioUring) poll(n int) error {
	if ur.params.flags&IORING_SETUP_IOPOLL == 0 {
		return nil
	}
	_, err := ioUringEnter(ur.ringFd, 0, uintptr(n), IORING_ENTER_GETEVENTS)

	return err
}

func (ur *ioUring) wait() (*ioUringCqe, error) {
	for sw := NewSpinWait().SetLevel(spinWaitLevelAtomic); !sw.Closed(); sw.Once() {
		h, t := atomic.LoadUint32(ur.cq.kHead), atomic.LoadUint32(ur.cq.kTail)
		if h == t {
			break
		}

		e := &ur.cq.cqes[h]
		ok := atomic.CompareAndSwapUint32(ur.cq.kHead, h, (h+1)&(*ur.cq.kRingMask))
		if !ok {
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

func (cqe *ioUringCqe) Context() context.Context {
	if cqe.userData == 0 {
		return context.Background()
	}
	return *(*context.Context)(unsafe.Pointer(uintptr(cqe.userData)))
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
