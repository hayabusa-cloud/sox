// ©Hayabusa Cloud Co., Ltd. 2025. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

//go:build linux

package sox

import (
	"context"
	"encoding/binary"
	"time"
	"unsafe"

	"golang.org/x/sys/cpu"
	"golang.org/x/sys/unix"
)

const (
	_ = 1 << (iota + 7)
	_
	UringEntriesPico
	UringEntriesNano
	UringEntriesMicro
	UringEntriesSmall
	UringEntriesMedium
	UringEntriesLarge
	UringEntriesHuge
)

type UringOptions struct {
	Entries                 int
	LockedBufferMem         int
	ReadBufferSize          int
	ReadBufferNum           int
	ReadBufferGidOffset     int
	WriteBufferSize         int
	WriteBufferNum          int
	MultiSizeBuffer         int
	MultiIssuers            bool
	NotifySucceed           bool
	IndirectSubmissionQueue bool
}

func NewUring(options ...func(options *UringOptions)) (*Uring, error) {
	opt := defaultUringOptions
	for _, option := range options {
		option(&opt)
	}
	setupOpts := []func(params *ioUringParams){ioUringDisabledOptions}
	if opt.MultiIssuers {
		setupOpts = append(setupOpts, func(params *ioUringParams) {
			params.flags |= IORING_SETUP_COOP_TASKRUN
		})
	} else {
		setupOpts = append(setupOpts, func(params *ioUringParams) {
			params.flags |= IORING_SETUP_SINGLE_ISSUER
			params.flags |= IORING_SETUP_DEFER_TASKRUN
		})
	}
	if !opt.IndirectSubmissionQueue {
		setupOpts = append(setupOpts, ioUringNoSQArrayOptions)
	}
	r, err := newIoUring(opt.Entries, setupOpts...)
	if err != nil {
		return nil, err
	}
	rFlags, wFlags := uint8(0), uint8(0)
	if r.feature(IORING_FEAT_CQE_SKIP) && !opt.NotifySucceed {
		wFlags |= IOSQE_CQE_SKIP_SUCCESS
	}
	ret := Uring{
		ioUring:          r,
		UringOptions:     &opt,
		bufferRings:      newUringBufferRings(),
		readLikeOpFlags:  rFlags,
		writeLikeOpFlags: wFlags,
	}

	if opt.MultiSizeBuffer > 0 {
		ret.bufferGroups = newUringBufferGroups(opt.MultiSizeBuffer)
		ret.bufferGroups.setGIDOffset(opt.ReadBufferGidOffset)
	} else {
		ret.buffers = newUringProvideBuffers(opt.ReadBufferSize, opt.ReadBufferNum)
		ret.buffers.setGIDOffset(opt.ReadBufferGidOffset)
	}
	feat := UringFeatures{
		SQEntries:         int(r.params.sqEntries),
		CQEntries:         int(r.params.cqEntries),
		UserdataByteOrder: binary.LittleEndian,
	}
	if cpu.IsBigEndian {
		feat.UserdataByteOrder = binary.BigEndian
	}
	ret.Features = &feat

	return &ret, nil
}

type UringFeatures struct {
	SQEntries         int
	CQEntries         int
	UserdataByteOrder binary.ByteOrder
}

type Uring struct {
	*ioUring
	*UringOptions
	Features *UringFeatures

	buffers      *uringProvideBuffers
	bufferGroups *uringProvideBufferGroups
	bufferRings  *uringBufferRings

	buffersPool *RegisterBufferPool

	readLikeOpFlags  uint8
	writeLikeOpFlags uint8
}

func (ur *Uring) Start(ctx context.Context) error {
	// register probes
	probe := ioUringProbe{}
	err := ur.registerProbe(&probe)
	if err != nil {
		return err
	}
	// assume all ops up to linux 6.12 are available
	// check the availability of ops released after linux 6.12
	for _, op := range ur.ops {
		switch op.op {
		}
	}

	// register buffers
	if ur.LockedBufferMem > (1 << 16) {
		maximizeMemoryLock()
	}
	regBufNum := min(registerBufferNum, ur.LockedBufferMem/registerBufferSize)
	ur.buffersPool = NewRegisterBufferPool(regBufNum)
	ur.buffersPool.Fill(func() RegisterBuffer { return RegisterBuffer{} })
	regBufAddr := unsafe.Pointer(unsafe.SliceData(ur.buffersPool.items))
	err = ur.registerBuffers(regBufAddr, regBufNum, registerBufferSize)
	if err != nil {
		return err
	}

	// provide buffers
	if ur.buffers != nil {
		err = ur.bufferRings.registerBuffers(ur.ioUring, ur.buffers)
		if err != nil {
			return err
		}
	} else if ur.bufferGroups != nil {
		err = ur.bufferRings.registerGroups(ur.ioUring, ur.bufferGroups)
		if err != nil {
			return err
		}
	}
	ur.bufferRings.advance(ur.ioUring)

	// enable ring
	err = ur.enable()
	if err != nil {
		return err
	}

	return nil
}

type UringEventData struct {
	Operation uint8
	OpFlags   uint8
	Buffer    bufferGroupIndex // buf_index or buf_group
	FD        ioUringFd
	UserData  ioUringUserdata
}

func (ur *Uring) Wait(events []*UringEvent) (n int, err error) {
	err = ur.ioUring.enter()
	if err != nil {
		return 0, err
	}
	for n < len(events) {
		cqe, err := ur.ioUring.wait()
		if err != nil && err != ErrTemporarilyUnavailable {
			return n, err
		} else if err == ErrTemporarilyUnavailable {
			return n, ErrTemporarilyUnavailable
		}
		events[n] = (*UringEvent)(unsafe.Pointer(cqe))
		n++
	}

	return n, nil
}

func (ur *Uring) Socket(ctx context.Context, network, address string, options ...UringOpOptionFunc) error {
	flags, callback, i := ur.socketOptions(options)
	switch network {
	case "ip", "ip4", "ip6":
		a, err := ResolveIPAddr(network, address)
		if err != nil {
			return err
		}
		sa, err := sockaddrData(ipAddrToSockaddr(network, a))
		if err != nil {
			return err
		}
		ctx = ContextWithUserdata(ctx, ur.socketOpUserdata(callback, networkIPFamily(network, a.IP), sa))
		return ur.socket(ctx, flags, int(ipFamily(a.IP)), unix.SOCK_RAW|unix.SOCK_CLOEXEC, unix.IPPROTO_RAW, i)
	case "tcp", "tcp4", "tcp6":
		a, err := ResolveTCPAddr(network, address)
		if err != nil {
			return err
		}
		sa, err := sockaddrData(tcpAddrToSockaddr(network, a))
		if err != nil {
			return err
		}
		ctx = ContextWithUserdata(ctx, ur.socketOpUserdata(callback, networkIPFamily(network, a.IP), sa))
		return ur.socket(ctx, flags, int(ipFamily(a.IP)), unix.SOCK_STREAM|unix.SOCK_CLOEXEC, unix.IPPROTO_TCP, i)
	case "udp", "udp4", "udp6":
		a, err := ResolveUDPAddr(network, address)
		if err != nil {
			return err
		}
		sa, err := sockaddrData(udpAddrToSockaddr(network, a))
		if err != nil {
			return err
		}
		ctx = ContextWithUserdata(ctx, ur.socketOpUserdata(callback, networkIPFamily(network, a.IP), sa))
		return ur.socket(ctx, flags, int(ipFamily(a.IP)), unix.SOCK_DGRAM|unix.SOCK_CLOEXEC, unix.IPPROTO_UDP, i)
	case "sctp", "sctp4", "sctp6":
		a, err := ResolveSCTPAddr(network, address)
		if err != nil {
			return err
		}
		sa, err := sockaddrData(sctpAddrToSockaddr(network, a))
		if err != nil {
			return err
		}
		ctx = ContextWithUserdata(ctx, ur.socketOpUserdata(callback, networkIPFamily(network, a.IP), sa))
		return ur.socket(ctx, flags, int(ipFamily(a.IP)), unix.SOCK_SEQPACKET|unix.SOCK_CLOEXEC, unix.IPPROTO_SCTP, i)
	case "unix", "unixgram", "unixpacket":
		a, err := ResolveUnixAddr(network, address)
		if err != nil {
			return err
		}
		sa, err := unixSockaddrData(unixAddrToSockaddr(a))
		if err != nil {
			return err
		}
		ctx = ContextWithUserdata(ctx, ur.socketOpUserdata(callback, NetworkUnix, sa))
		if network == "unixgram" {
			return ur.socket(ctx, flags, unix.AF_UNIX, unix.SOCK_DGRAM|unix.SOCK_CLOEXEC, 0, i)
		} else if network == "unixpacket" {
			return ur.socket(ctx, flags, unix.AF_UNIX, unix.SOCK_SEQPACKET|unix.SOCK_CLOEXEC, 0, i)
		}
		return ur.socket(ctx, flags, unix.AF_UNIX, unix.SOCK_STREAM|unix.SOCK_CLOEXEC, 0, i)
	}
	return UnknownNetworkError(network)
}

func (ur *Uring) SocketRaw(ctx context.Context, domain, typ, proto int, options ...UringOpOptionFunc) error {
	flags, callback, i := ur.socketOptions(options)
	ctx = ContextWithUserdata(ctx, ur.socketOpUserdata(callback, NetworkType(domain), nil))
	return ur.socket(ctx, flags, domain, typ, proto, i)
}

func (ur *Uring) TCP4Socket(ctx context.Context, options ...UringOpOptionFunc) error {
	return ur.SocketRaw(ctx, unix.AF_INET, unix.SOCK_STREAM|unix.SOCK_CLOEXEC, unix.IPPROTO_TCP, options...)
}

func (ur *Uring) TCP6Socket(ctx context.Context, options ...UringOpOptionFunc) error {
	return ur.SocketRaw(ctx, unix.AF_INET6, unix.SOCK_STREAM|unix.SOCK_CLOEXEC, unix.IPPROTO_TCP, options...)
}

func (ur *Uring) UDP4Socket(ctx context.Context, options ...UringOpOptionFunc) error {
	return ur.SocketRaw(ctx, unix.AF_INET, unix.SOCK_DGRAM|unix.SOCK_CLOEXEC, unix.IPPROTO_UDP, options...)
}

func (ur *Uring) UDP6Socket(ctx context.Context, options ...UringOpOptionFunc) error {
	return ur.SocketRaw(ctx, unix.AF_INET6, unix.SOCK_DGRAM|unix.SOCK_CLOEXEC, unix.IPPROTO_UDP, options...)
}

func (ur *Uring) UDPLITE4Socket(ctx context.Context, options ...UringOpOptionFunc) error {
	return ur.SocketRaw(ctx, unix.AF_INET, unix.SOCK_DGRAM|unix.SOCK_CLOEXEC, unix.IPPROTO_UDPLITE, options...)
}

func (ur *Uring) UDPLITE6Socket(ctx context.Context, options ...UringOpOptionFunc) error {
	return ur.SocketRaw(ctx, unix.AF_INET6, unix.SOCK_DGRAM|unix.SOCK_CLOEXEC, unix.IPPROTO_UDPLITE, options...)
}

func (ur *Uring) SCTP4Socket(ctx context.Context, options ...UringOpOptionFunc) error {
	return ur.SocketRaw(ctx, unix.AF_INET, unix.SOCK_SEQPACKET|unix.SOCK_CLOEXEC, unix.IPPROTO_SCTP, options...)
}

func (ur *Uring) SCTP6Socket(ctx context.Context, options ...UringOpOptionFunc) error {
	return ur.SocketRaw(ctx, unix.AF_INET6, unix.SOCK_SEQPACKET|unix.SOCK_CLOEXEC, unix.IPPROTO_SCTP, options...)
}

func (ur *Uring) UnixSocket(ctx context.Context, options ...UringOpOptionFunc) error {
	return ur.SocketRaw(ctx, unix.AF_LOCAL, unix.SOCK_SEQPACKET|unix.SOCK_CLOEXEC, 0, options...)
}

func (ur *Uring) Bind(ctx context.Context, so Socket, addr Addr, options ...UringOpOptionFunc) error {
	flags, callback := ur.bindOptions(options)
	sa := AddrToSockaddr(addr)
	sab, err := sockaddrData(sa)
	if err != nil {
		return err
	}
	ctx = ContextWithUserdata(ctx, ur.bindOpUserdata(callback, so.NetworkType(), sab))
	return ur.bind(ctx, flags, so.Fd(), AddrToSockaddr(addr))
}

// Listen starts listening on a given socket with the specified options and manages I/O operations using io_uring.
// Listen starts socket listening with given options
func (ur *Uring) Listen(ctx context.Context, so Socket, options ...UringOpOptionFunc) error {
	flags, backlog := ur.listenOptions(options)
	return ur.listen(ctx, flags, so.Fd(), backlog)
}

// Accept accepts a new connection from listener
func (ur *Uring) Accept(ctx context.Context, listener Socket, options ...UringOpOptionFunc) error {
	flags, ioprio := ur.acceptOptions(options)
	ctx = ContextWithUserdata(ctx, ur.acceptOpUserData(listener.NetworkType()))
	return ur.accept(ctx, flags, ioprio, listener.Fd())
}

// Connect initiates socket connection to remote address
func (ur *Uring) Connect(ctx context.Context, so Socket, remote Addr, options ...UringOpOptionFunc) error {
	flags := ur.operationOptions(options)
	return ur.connect(ctx, flags, so.Fd(), AddrToSockaddr(remote))
}

// Receive performs socket receive operation
func (ur *Uring) Receive(ctx context.Context, so Socket, b []byte, options ...UringOpOptionFunc) error {
	if b == nil {
		flags, ioprio, bufSize, bufGroup := ur.receiveWithBufferSelectOptions(so, options)
		return ur.receiveWithBufferSelect(ctx, flags|ur.readLikeOpFlags, ioprio, so.Fd(), bufSize, bufGroup)
	}
	flags, ioprio, offset, n := ur.receiveOptions(b, options)
	return ur.receive(ctx, flags|ur.readLikeOpFlags, ioprio, so.Fd(), b, uint64(offset), n)
}

// Send writes data to socket
func (ur *Uring) Send(ctx context.Context, so Socket, p []byte, options ...UringOpOptionFunc) error {
	flags, ioprio, offset, n := ur.sendOptions(p, options)
	return ur.send(ctx, flags|ur.writeLikeOpFlags, ioprio, so.Fd(), p, uint64(offset), n)
}

// Timeout submits timeout request with specified duration
func (ur *Uring) Timeout(ctx context.Context, d time.Duration, options ...UringOpOptionFunc) error {
	flags, cnt := ur.timeoutOptions(options)
	nano := d.Nanoseconds()
	return ur.timeout(ctx, flags, cnt, &unix.Timespec{Sec: nano / int64(time.Second), Nsec: nano % int64(time.Second)}, 0)
}

// Shutdown gracefully closes socket
func (ur *Uring) Shutdown(ctx context.Context, so Socket, how int, options ...UringOpOptionFunc) error {
	flags := ur.operationOptions(options)
	return ur.shutdown(ctx, flags, so.Fd(), how)
}

// Nop submits no-op request
func (ur *Uring) Nop(ctx context.Context, options ...UringOpOptionFunc) error {
	flags := ur.operationOptions(options)
	return ur.nop(ctx, flags)
}

// Close closes file descriptor
func (ur *Uring) Close(ctx context.Context, file PollCloser, options ...UringOpOptionFunc) error {
	flags := ur.operationOptions(options)
	return ur.close(ctx, flags, file.Fd())
}

// Read performs read operation
func (ur *Uring) Read(ctx context.Context, fd int, b []byte, options ...UringOpOptionFunc) error {
	flags, ioprio, offset, n := ur.readOptions(b, options)
	return ur.read(ctx, flags|ur.readLikeOpFlags, ioprio, fd, b, uint64(offset), n)
}

// Write performs write operation
func (ur *Uring) Write(ctx context.Context, fd int, b []byte, n int, options ...UringOpOptionFunc) error {
	flags, ioprio, offset, n := ur.writeOptions(b, options)
	return ur.write(ctx, flags|ur.writeLikeOpFlags, ioprio, fd, b, uint64(offset), n)
}

// Splice transfers data between file descriptors
func (ur *Uring) Splice(ctx context.Context, fdIn, fdOut int, len int, options ...UringOpOptionFunc) error {
	flags := ur.operationOptions(options)
	return ur.splice(ctx, flags, fdIn, nil, fdOut, nil, len, 0)
}

// Tee duplicates data between pipes
func (ur *Uring) Tee(ctx context.Context, fdIn, fdOut int, length int, options ...UringOpOptionFunc) error {
	flags := ur.operationOptions(options)
	return ur.tee(ctx, flags, fdIn, fdOut, length, 0)
}

// Sync performs file sync operation
func (ur *Uring) Sync(ctx context.Context, fd int, options ...UringOpOptionFunc) error {
	flags := ur.operationOptions(options)
	return ur.fsync(ctx, flags, fd)
}

// ReadV performs vectored read operation
func (ur *Uring) ReadV(ctx context.Context, fd int, iovs [][]byte, options ...UringOpOptionFunc) error {
	flags := ur.operationOptions(options)
	return ur.readv(ctx, flags|ur.readLikeOpFlags, fd, iovs)
}

// WriteV performs vectored write operation
func (ur *Uring) WriteV(ctx context.Context, fd int, iovs [][]byte, options ...UringOpOptionFunc) error {
	flags := ur.operationOptions(options)
	return ur.writev(ctx, flags|ur.writeLikeOpFlags, fd, iovs)
}

// ReadFixed performs read with fixed buffer
func (ur *Uring) ReadFixed(ctx context.Context, fd int, bufIndex int, options ...UringOpOptionFunc) ([]byte, error) {
	flags, offset, n := ur.readFixedOptions(bufIndex, options)
	return ur.readFixed(ctx, flags|ur.readLikeOpFlags, fd, bufIndex, uint64(offset), n)
}

// WriteFixed performs write with fixed buffer
func (ur *Uring) WriteFixed(ctx context.Context, fd int, bufIndex int, n int, options ...UringOpOptionFunc) error {
	flags, offset := ur.writeFixedOptions(options)
	return ur.writeFixed(ctx, flags|ur.writeLikeOpFlags, fd, bufIndex, uint64(offset), n)
}

// OpenAt opens file at given path relative to directory fd
func (ur *Uring) OpenAt(ctx context.Context, dirfd int, pathname string, flags int, mode uint32, options ...UringOpOptionFunc) error {
	opFlags := ur.operationOptions(options)
	return ur.openAt(ctx, opFlags, dirfd, pathname, flags, mode)
}

// FileAdvise provides advice about file access patterns
func (ur *Uring) FileAdvise(ctx context.Context, fd int, offset int64, length int, advice int, options ...UringOpOptionFunc) error {
	flags := ur.operationOptions(options)
	return ur.fadvise(ctx, flags, fd, uint64(offset), length, advice)
}

// SendMsg sends message with control data
func (ur *Uring) SendMsg(ctx context.Context, so PollFd, buffers [][]byte, oob []byte, to Addr, options ...UringOpOptionFunc) error {
	flags, ioprio := ur.sendmsgOptions(buffers, options)
	var sa unix.Sockaddr
	if to != nil {
		sa = AddrToSockaddr(to)
	}
	return ur.sendmsg(ctx, flags|ur.writeLikeOpFlags, ioprio, so.Fd(), buffers, oob, sa)
}

// RecvMsg receives message with control data
func (ur *Uring) RecvMsg(ctx context.Context, so PollFd, buffers [][]byte, oob []byte, options ...UringOpOptionFunc) error {
	flags, ioprio := ur.recvmsgOptions(buffers, options)
	return ur.recvmsg(ctx, flags|ur.readLikeOpFlags, ioprio, so.Fd(), buffers, oob)
}

// PollAdd adds file descriptor to poll set
func (ur *Uring) PollAdd(ctx context.Context, fd int, events int, options ...UringOpOptionFunc) error {
	flags := ur.operationOptions(options)
	return ur.pollAdd(ctx, flags, fd, 0, events)
}

// PollRemove removes file descriptor from poll set
func (ur *Uring) PollRemove(ctx context.Context, options ...UringOpOptionFunc) error {
	flags := ur.operationOptions(options)
	return ur.pollRemove(ctx, flags)
}

// TimeoutRemove removes timeout request
func (ur *Uring) TimeoutRemove(ctx context.Context, userData unsafe.Pointer, options ...UringOpOptionFunc) error {
	flags := ur.operationOptions(options)
	return ur.timeoutRemove(ctx, flags, userData, 0)
}

// AsyncCancel cancels pending async operation
func (ur *Uring) AsyncCancel(ctx context.Context, userdata *ioUringCtx, options ...UringOpOptionFunc) error {
	flags := ur.operationOptions(options)
	return ur.asyncCancel(ctx, flags, userdata)
}

// Fallocate allocates space for file
func (ur *Uring) Fallocate(ctx context.Context, fd int, mode uint32, offset int64, length int64, options ...UringOpOptionFunc) error {
	flags := ur.operationOptions(options)
	return ur.fAllocate(ctx, flags, fd, mode, uint64(offset), length)
}

// SyncFileRange syncs file range to storage
func (ur *Uring) SyncFileRange(ctx context.Context, fd int, offset int64, length int, flags int, options ...UringOpOptionFunc) error {
	opFlags := ur.operationOptions(options)
	return ur.syncFileRange(ctx, opFlags, fd, uint64(offset), length, flags)
}

// LinkTimeout creates linked timeout operation
func (ur *Uring) LinkTimeout(ctx context.Context, d time.Duration, options ...UringOpOptionFunc) error {
	flags := ur.operationOptions(options)
	nano := d.Nanoseconds()
	return ur.linkTimeout(ctx, flags, &unix.Timespec{Sec: nano / int64(time.Second), Nsec: nano % int64(time.Second)}, 0)
}
