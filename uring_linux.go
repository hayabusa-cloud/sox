// ©Hayabusa Cloud Co., Ltd. 2022. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

//go:build linux

package sox

import (
	"context"
	"sync/atomic"
	"time"
	"unsafe"

	"golang.org/x/sys/unix"
)

const (
	IORING_SETUP_IOPOLL             = 1 << 0
	IORING_SETUP_SQPOLL             = 1 << 1
	IORING_SETUP_SQ_AFF             = 1 << 2
	IORING_SETUP_CQSIZE             = 1 << 3
	IORING_SETUP_CLAMP              = 1 << 4
	IORING_SETUP_ATTACH_WQ          = 1 << 5
	IORING_SETUP_R_DISABLED         = 1 << 6
	IORING_SETUP_SUBMIT_ALL         = 1 << 7
	IORING_SETUP_COOP_TASKRUN       = 1 << 8
	IORING_SETUP_TASKRUN_FLAG       = 1 << 9
	IORING_SETUP_SQE128             = 1 << 10
	IORING_SETUP_CQE32              = 1 << 11
	IORING_SETUP_SINGLE_ISSUER      = 1 << 12
	IORING_SETUP_DEFER_TASKRUN      = 1 << 13
	IORING_SETUP_NO_MMAP            = 1 << 14
	IORING_SETUP_REGISTERED_FD_ONLY = 1 << 15
	IORING_SETUP_NO_SQARRAY         = 1 << 16
	IORING_SETUP_HYBRID_IOPOLL      = 1 << 17
)

const (
	IORING_FEAT_SINGLE_MMAP = 1 << iota
	IORING_FEAT_NODROP
	IORING_FEAT_SUBMIT_STABLE
	IORING_FEAT_RW_CUR_POS
	IORING_FEAT_CUR_PERSONALITY
	IORING_FEAT_FAST_POLL
	IORING_FEAT_POLL_32BITS
	IORING_FEAT_SQPOLL_NONFIXED
	IORING_FEAT_EXT_ARG
	IORING_FEAT_NATIVE_WORKERS
	IORING_FEAT_RSRC_TAGS
	IORING_FEAT_CQE_SKIP
	IORING_FEAT_LINKED_FILE
	IORING_FEAT_REG_REG_RING
)

const (
	IORING_ENTER_GETEVENTS       = 1 << 0
	IORING_ENTER_SQ_WAKEUP       = 1 << 1
	IORING_ENTER_SQ_WAIT         = 1 << 2
	IORING_ENTER_EXT_ARG         = 1 << 3
	IORING_ENTER_REGISTERED_RING = 1 << 4
)

const (
	IORING_OFF_SQ_RING    int64 = 0
	IORING_OFF_CQ_RING    int64 = 0x8000000
	IORING_OFF_SQES       int64 = 0x10000000
	IORING_OFF_PBUF_RING        = 0x80000000
	IORING_OFF_PBUF_SHIFT       = 16
	IORING_OFF_MMAP_MASK        = 0xf8000000
)

const (
	IORING_SQ_NEED_WAKEUP = 1 << iota
	IORING_SQ_CQ_OVERFLOW
	IORING_SQ_TASKRUN
)

const (
	IOSQE_FIXED_FILE = 1 << iota
	IOSQE_IO_DRAIN
	IOSQE_IO_LINK
	IOSQE_IO_HARDLINK
	IOSQE_ASYNC
	IOSQE_BUFFER_SELECT
	IOSQE_CQE_SKIP_SUCCESS
)

const (
	IORING_POLL_ADD_MULTI = 1 << iota
	IORING_POLL_UPDATE_EVENTS
	IORING_POLL_UPDATE_USER_DATA
	IORING_POLL_ADD_LEVEL
)

const (
	IORING_ASYNC_CANCEL_ALL = 1 << iota
	IORING_ASYNC_CANCEL_FD
	IORING_ASYNC_CANCEL_ANY
	IORING_ASYNC_CANCEL_FD_FIXED
	IORING_ASYNC_CANCEL_USERDATA
	IORING_ASYNC_CANCEL_OP
)

const (
	IORING_CQE_F_BUFFER = 1 << iota
	IORING_CQE_F_MORE
	IORING_CQE_F_SOCK_NONEMPTY
	IORING_CQE_F_NOTIF
)

const (
	IORING_CQE_BUFFER_SHIFT = 16
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
	IORING_REGISTER_PBUF_STATUS
	IORING_REGISTER_NAPI
	IORING_UNREGISTER_NAPI
)

const (
	IO_URING_OP_SUPPORTED = 1 << 0
)

const (
	ioUringDefaultEntries = 0x2000

	ioUringDefaultSqThreadCPU  = 1
	ioUringDefaultSqThreadIdle = 5 * time.Second
)

type ioUring struct {
	_ noCopy

	sl     Spinlock
	params *ioUringParams

	sq     ioUringSq
	cq     ioUringCq
	ringFd int
	ops    []ioUringProbeOp
	ctxs   []ioUringCtx
	bufs   [][]byte
}

type (
	ioUringFd       int32
	ioUringUserdata []byte
)

type ioUringCtx struct {
	op       uint8
	flags    uint8
	buf      bufferGroupIndex // buf_index or buf_group
	fd       ioUringFd
	userdata ioUringUserdata
}

func (ctx *ioUringCtx) getOp() (uint8, uint8) {
	return ctx.op, ctx.flags
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
		sl:     Spinlock{},
		params: params,

		sq: ioUringSq{
			ringSz: params.sqOff.array + uint32(unsafe.Sizeof(uint32(0)))*params.sqEntries,
		},
		cq: ioUringCq{
			ringSz: params.cqOff.cqes + uint32(unsafe.Sizeof(uint32(0)))*params.cqEntries,
		},
		ringFd: fd,
		bufs:   [][]byte{},
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
	uring.sq.array = unsafe.Slice((*uint32)(unsafe.Pointer(ptr+uintptr(params.sqOff.array))), int(params.sqEntries))
	b, err = unix.Mmap(uring.ringFd, IORING_OFF_SQES, int(params.sqEntries)*int(unsafe.Sizeof(ioUringSqe{})), unix.PROT_READ|unix.PROT_WRITE|unix.PROT_EXEC, unix.MAP_SHARED|unix.MAP_POPULATE)
	if err != nil {
		return uring, errFromUnixErrno(err)
	}
	uring.sq.sqes = unsafe.Slice((*ioUringSqe)(unsafe.Pointer(&b[0])), int(params.sqEntries))

	uring.cq.kHead = (*uint32)(unsafe.Pointer(ptr + uintptr(params.cqOff.head)))
	uring.cq.kTail = (*uint32)(unsafe.Pointer(ptr + uintptr(params.cqOff.tail)))
	uring.cq.kRingMask = (*uint32)(unsafe.Pointer(ptr + uintptr(params.cqOff.ringMask)))
	uring.cq.kRingEntries = (*uint32)(unsafe.Pointer(ptr + uintptr(params.cqOff.ringEntries)))
	uring.cq.kOverflow = (*uint32)(unsafe.Pointer(ptr + uintptr(params.cqOff.overflow)))
	uring.cq.cqes = unsafe.Slice((*ioUringCqe)(unsafe.Pointer(ptr+uintptr(params.cqOff.cqes))), int(params.cqEntries))

	uring.ctxs = make([]ioUringCtx, *uring.cq.kRingEntries)

	return uring, nil
}

type ioUringProbe struct {
	lastOp uint8
	opsLen uint8
	resv   uint16
	resv2  [3]uint32
	ops    [256]ioUringProbeOp
}

type ioUringProbeOp struct {
	op    uint8
	resv  uint8
	flags uint16
	resv2 uint32
}

func (ur *ioUring) registerProbe(probe *ioUringProbe) error {
	addr, n := uintptr(unsafe.Pointer(probe)), uintptr(len(probe.ops))
	_, _, errno := unix.Syscall6(unix.SYS_IO_URING_REGISTER, uintptr(ur.ringFd), IORING_REGISTER_PROBE, addr, n, 0, 0)
	if errno != 0 {
		return errFromUnixErrno(errno)
	}
	ur.ops = make([]ioUringProbeOp, 0, probe.opsLen)
	for i := range probe.opsLen {
		if probe.ops[i].flags&IO_URING_OP_SUPPORTED < IO_URING_OP_SUPPORTED {
			continue
		}
		ur.ops = append(ur.ops, probe.ops[i])
	}

	return nil
}

func (ur *ioUring) registerBuffers(addr unsafe.Pointer, n, size int) error {
	if ur.bufs != nil && len(ur.bufs) > 0 {
		panic("io-uring buffers already registered")
	}
	if n < 1 || size < 1 || 0 != n&(n-1) || 0 != size&(size-1) {
		return ErrInvalidParam
	}
	vectors := make([]ioVec, 0, n)
	ur.bufs = make([][]byte, 0, n)
	for i := range n {
		base := unsafe.Add(addr, i*size)
		vectors = append(vectors, ioVec{Base: (*byte)(base), Len: uint64(size)})
		ur.bufs = append(ur.bufs, unsafe.Slice((*byte)(base), size))
	}
	data := uintptr(unsafe.Pointer(unsafe.SliceData(vectors)))
	reg := ioUringRSrcRegister{nr: uint32(n), data: uint64(data)}
	regPtr, regSize := uintptr(unsafe.Pointer(&reg)), unsafe.Sizeof(reg)
	_, _, errno := unix.Syscall6(unix.SYS_IO_URING_REGISTER, uintptr(ur.ringFd), IORING_REGISTER_BUFFERS2, regPtr, regSize, 0, 0)
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
	ur.bufs = [][]byte{}

	return nil
}

func (ur *ioUring) registerBufRing(entries int, groupID uint16) (*ioUringBufRing, error) {
	if entries < 1 || entries > (1<<15) {
		panic("entries must be between 1 and 32768")
	}
	entries--
	entries |= entries >> 1
	entries |= entries >> 2
	entries |= entries >> 4
	entries |= entries >> 8
	entries++
	s := AlignedMem(entries * int(unsafe.Sizeof(ioUringBuf{})))
	ringAddr := uintptr(unsafe.Pointer(unsafe.SliceData(s)))
	r := (*ioUringBufRing)(unsafe.Pointer(ringAddr))
	reg := ioUringBufReg{
		ringAddr:    uint64(ringAddr),
		ringEntries: uint32(entries),
		bgid:        groupID,
	}
	addr := uintptr(unsafe.Pointer(&reg))
	_, _, errno := unix.Syscall6(unix.SYS_IO_URING_REGISTER, uintptr(ur.ringFd), IORING_REGISTER_PBUF_RING, addr, 1, 0, 0)
	if errno != 0 {
		return r, errFromUnixErrno(errno)
	}
	return r, nil
}

func (ur *ioUring) unregisterBufRing(groupID uint16) error {
	reg := ioUringBufReg{bgid: groupID}
	addr := uintptr(unsafe.Pointer(&reg))
	_, _, errno := unix.Syscall6(unix.SYS_IO_URING_REGISTER, uintptr(ur.ringFd), IORING_UNREGISTER_PBUF_RING, addr, 1, 0, 0)
	if errno != 0 {
		return errFromUnixErrno(errno)
	}
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

func (ur *ioUring) feature(feat uint32) bool {
	return feat == ur.params.features&feat
}

func (ur *ioUring) enable() error {
	_, _, errno := unix.Syscall6(unix.SYS_IO_URING_REGISTER, uintptr(ur.ringFd), IORING_REGISTER_ENABLE_RINGS, 0, 0, 0, 0)
	if errno != 0 {
		return errFromUnixErrno(errno)
	}
	return nil
}

func (ur *ioUring) directSubmission(ctx context.Context, fn func(e *ioUringSqe)) error {
	ur.sl.Lock()
	defer ur.sl.Unlock()

	h, t := *ur.sq.kHead, *ur.sq.kTail
	if (t+1)&*ur.sq.kRingMask == h {
		return ErrTemporarilyUnavailable
	}

	e := &ur.sq.sqes[t&*ur.sq.kRingMask]
	fn(e)
	c := &ur.ctxs[t&*ur.cq.kRingMask]
	c.op = e.opcode
	c.flags = e.flags
	c.buf = ContextUserdata[bufferGroupIndex](ctx)
	c.fd = ContextUserdata[ioUringFd](ctx)
	c.userdata = ContextUserdata[ioUringUserdata](ctx)
	e.userData = uint64(uintptr(unsafe.Pointer(c)))
	*ur.sq.kTail++

	return nil
}

func (ur *ioUring) indirectSubmission(ctx context.Context, fn func(e *ioUringSqe)) error {
	ur.sl.Lock()
	defer ur.sl.Unlock()

	h, t := *ur.sq.kHead, *ur.sq.kTail
	if (t+1)&*ur.sq.kRingMask == h {
		return ErrTemporarilyUnavailable
	}

	e := &ur.sq.sqes[t&*ur.sq.kRingMask]
	fn(e)
	c := &ur.ctxs[t&*ur.cq.kRingMask]
	c.op = e.opcode
	c.flags = e.flags
	c.buf = ContextUserdata[bufferGroupIndex](ctx)
	c.fd = ContextUserdata[ioUringFd](ctx)
	c.userdata = ContextUserdata[ioUringUserdata](ctx)
	e.userData = uint64(uintptr(unsafe.Pointer(c)))

	ur.sq.array[t&*ur.sq.kRingMask] = t & *ur.sq.kRingMask
	*ur.sq.kTail++

	return nil
}

func (ur *ioUring) submit3(ctx context.Context, op uint8, flags uint8, ioprio uint16, fd int, addr uint64, n int) error {
	setSqe := func(e *ioUringSqe) {
		e.opcode = op
		e.flags = flags
		e.ioprio = ioprio
		e.fd = int32(fd)
		e.addr = addr
		e.len = uint32(n)
	}
	if ur.params.flags&IORING_SETUP_NO_SQARRAY == IORING_SETUP_NO_SQARRAY {
		return ur.directSubmission(ctx, setSqe)
	}
	return ur.indirectSubmission(ctx, setSqe)
}

func (ur *ioUring) submit6(ctx context.Context, op uint8, flags uint8, ioprio uint16, fd int, off uint64, addr uint64, n int, uflags uint32) error {
	setSqe := func(e *ioUringSqe) {
		e.opcode = op
		e.flags = flags
		e.ioprio = ioprio
		e.fd = int32(fd)
		e.off = off
		e.addr = addr
		e.len = uint32(n)
		e.uflags = uflags
	}
	if ur.params.flags&IORING_SETUP_NO_SQARRAY == IORING_SETUP_NO_SQARRAY {
		return ur.directSubmission(ctx, setSqe)
	}
	return ur.indirectSubmission(ctx, setSqe)
}

func (ur *ioUring) submit9(ctx context.Context, op uint8, flags uint8, ioprio uint16, fd int, off uint64, spliceOffIn uint64, n int, uflags uint32, bufGroup uint16, personality uint16, spliceFdIn int) error {
	setSqe := func(e *ioUringSqe) {
		e.opcode = op
		e.flags = flags
		e.ioprio = ioprio
		e.fd = int32(fd)
		e.off = off
		e.addr = spliceOffIn
		e.len = uint32(n)
		e.uflags = uflags
		e.userData = uint64(uintptr(unsafe.Pointer(&ctx)))
		e.bufIndex = bufGroup
		e.personality = personality
		e.spliceFdIn = int32(spliceFdIn)
	}
	if ur.params.flags&IORING_SETUP_NO_SQARRAY == IORING_SETUP_NO_SQARRAY {
		return ur.directSubmission(ctx, setSqe)
	}
	return ur.indirectSubmission(ctx, setSqe)
}

func (ur *ioUring) sqCount() int {
	return int((*ur.sq.kTail - *ur.sq.kHead) & *ur.sq.kRingMask)
}

func (ur *ioUring) enter() error {
	if atomic.LoadUint32(ur.sq.kFlags)&IORING_SQ_NEED_WAKEUP == IORING_SQ_NEED_WAKEUP {
		_, err := ioUringEnter(ur.ringFd, uintptr(ur.params.sqEntries), 0, IORING_ENTER_SQ_WAKEUP)
		if err != nil {
			return err
		}
	}
	ur.sl.Lock()
	defer ur.sl.Unlock()
	if (ur.params.flags&IORING_SETUP_SQPOLL == 0) && *ur.sq.kHead != *ur.sq.kTail {
		n := (*ur.sq.kTail - *ur.sq.kHead) & *ur.sq.kRingMask
		_, err := ioUringEnter(ur.ringFd, uintptr(n), 0, IORING_ENTER_GETEVENTS)
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
	submit := (*ur.sq.kTail - *ur.sq.kHead) & *ur.sq.kRingMask
	_, err := ioUringEnter(ur.ringFd, uintptr(submit), uintptr(n), IORING_ENTER_GETEVENTS)

	return err
}

func (ur *ioUring) wait() (*ioUringCqe, error) {
	sw := SpinWait{}
	for {
		h, t := atomic.LoadUint32(ur.cq.kHead), atomic.LoadUint32(ur.cq.kTail)
		if h == t {
			break
		}

		e := &ur.cq.cqes[h&*ur.cq.kRingMask]
		ok := atomic.CompareAndSwapUint32(ur.cq.kHead, h, h+1)
		if ok {
			return e, nil
		}
		sw.Once()
	}

	return nil, ErrTemporarilyUnavailable
}

func (ur *ioUring) cqAdvance(nr uint32) {
	if nr == 0 {
		return
	}
	ur.cq.advance(nr)
}

func (ur *ioUring) bufRingInit(br *ioUringBufRing) {
	br.tail = 0
}

func (ur *ioUring) bufRingAdd(br *ioUringBufRing, addr uintptr, n int, bid uint16, mask, offset uintptr) {
	add := ioUringBufSize * ((uintptr(br.tail) + offset) & mask)
	buf := (*ioUringBuf)(unsafe.Add(unsafe.Pointer(br), add))
	buf.addr = uint64(addr)
	buf.len = uint32(n)
	buf.bid = bid
}

func (ur *ioUring) bufRingAdvance(br *ioUringBufRing, count int) {
	br.tail += uint16(count)
}

func (ur *ioUring) bufRingAvailable(br *ioUringBufRing, bgid uint16) int {
	head, ret := uint16(0), 0
	ret = ur.bufRingHead(bgid, &head)
	if ret > 0 {
		return ret
	}
	return int(br.tail - head)
}

func (ur *ioUring) bufRingHead(groupID uint16, head *uint16) int {
	status := ioUringBufStatus{bufGroup: uint32(groupID)}

	ret, _, errno := unix.Syscall6(unix.SYS_IO_URING_REGISTER, uintptr(ur.ringFd), IORING_REGISTER_PBUF_STATUS, uintptr(unsafe.Pointer(&status)), 1, 0, 0)
	if ret != 0 {
		return int(errno)
	}
	*head = uint16(status.head)
	return 0
}

func (ur *ioUring) bufRingCQAdvance(br *ioUringBufRing, count int) {
	ur.bufRingAdvance(br, count)
	ur.cqAdvance(uint32(count))
}

type ioUringRSrcRegister struct {
	nr    uint32
	resv  uint32
	resv2 uint64
	data  uint64
	tags  uint64
}

type ioUringBufReg struct {
	ringAddr    uint64
	ringEntries uint32
	bgid        uint16
	pad         uint16
	_           [3]uint64
}

type ioUringBufStatus struct {
	bufGroup uint32
	head     uint32
	_        [8]uint32
}

type ioUringBuf struct {
	addr uint64
	len  uint32
	bid  uint16
	tail uint16
}

var ioUringBufSize = unsafe.Sizeof(ioUringBuf{})

type ioUringBufRing ioUringBuf

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

func (cq *ioUringCq) advance(nr uint32) {
	atomic.AddUint32(cq.kTail, nr)
}

type ioUringCqe struct {
	userData uint64
	res      int32
	flags    uint32
}

func (cqe *ioUringCqe) context() *ioUringCtx {
	if cqe.userData == 0 {
		return nil
	}
	return (*ioUringCtx)(unsafe.Pointer(uintptr(cqe.userData)))
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
	ioUringDefaultParams   = &ioUringParams{}
	ioUringDisabledOptions = func(params *ioUringParams) {
		params.flags |= IORING_SETUP_R_DISABLED
	}
	ioUringNoSQArrayOptions = func(params *ioUringParams) {
		params.flags |= IORING_SETUP_NO_SQARRAY
	}
	ioUringIoPollOptions = func(params *ioUringParams) {
		params.flags |= IORING_SETUP_IOPOLL
		params.flags &= ^uint32(IORING_SETUP_COOP_TASKRUN)
		params.flags &= ^uint32(IORING_SETUP_TASKRUN_FLAG)
		params.flags &= ^uint32(IORING_SETUP_DEFER_TASKRUN)
	}
	ioUringHybridIoPollOptions = func(params *ioUringParams) {
		params.flags |= IORING_SETUP_IOPOLL | IORING_SETUP_HYBRID_IOPOLL
		params.flags &= ^uint32(IORING_SETUP_COOP_TASKRUN)
		params.flags &= ^uint32(IORING_SETUP_TASKRUN_FLAG)
		params.flags &= ^uint32(IORING_SETUP_DEFER_TASKRUN)
	}
	// sq poll mode is not recommended
	ioUringSqPollOptions = func(params *ioUringParams) {
		params.flags |= IORING_SETUP_SQPOLL | IORING_SETUP_SQ_AFF
		params.sqThreadCPU = ioUringDefaultSqThreadCPU
		params.sqThreadIdle = uint32(ioUringDefaultSqThreadIdle.Milliseconds())
	}
)

func ioUringSetup(entries uint32, params *ioUringParams) (fd int, err error) {
	r1, _, errno := unix.Syscall(
		unix.SYS_IO_URING_SETUP,
		uintptr(entries),
		uintptr(unsafe.Pointer(params)),
		0,
	)
	if errno != 0 {
		err = errFromUnixErrno(errno)
		return
	}
	unix.CloseOnExec(int(r1))
	return int(r1), nil
}

func ioUringEnter(fd int, toSubmit uintptr, minComplete uintptr, flags uintptr) (n int, err error) {
	result, _, errno := unix.Syscall6(unix.SYS_IO_URING_ENTER, uintptr(fd), toSubmit, minComplete, flags, 0, 0)
	if errno != 0 {
		return int(result), errFromUnixErrno(errno)
	}
	return int(result), nil
}
