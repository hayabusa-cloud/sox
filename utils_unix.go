// ©Hayabusa Cloud Co., Ltd. 2022. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

//go:build unix

package sox

import (
	"golang.org/x/sys/unix"
	"unsafe"
)

type ioVec = unix.Iovec

func errFromUnixErrno(err error) error {
	if err == nil {
		return nil
	}
	errno, ok := err.(unix.Errno)
	if !ok {
		return err
	}
	switch errno {
	case unix.EINTR:
		return ErrInterruptedSyscall
	case unix.EAGAIN:
		return ErrTemporarilyUnavailable
	case unix.EINPROGRESS:
		return ErrInProgress
	case unix.EFAULT:
		return ErrFaultParams
	case unix.EINVAL:
		return ErrInvalidParam
	case unix.EMFILE:
		return ErrProcessFileLimit
	case unix.ENFILE:
		return ErrSystemFileLimit
	case unix.ENODEV:
		return ErrNoDevice
	case unix.ENOMEM:
		return ErrNoAvailableMemory
	case unix.EPERM:
		return ErrNoPermission
	default:
		return errno
	}
}

func ioVecFromBytesSlice(iov [][]byte) (addr uintptr, n int) {
	vec := make([]unix.Iovec, len(iov))
	for i := range len(iov) {
		vec[i] = unix.Iovec{Base: unsafe.SliceData(iov[i]), Len: uint64(len(iov[i]))}
	}
	addr, n = uintptr(unsafe.Pointer(unsafe.SliceData(vec))), len(vec)

	return
}

func ioVecAddrLen(vec []ioVec) (addr uintptr, n int) {
	addr, n = uintptr(unsafe.Pointer(unsafe.SliceData(vec))), len(vec)
	return
}

func ioVecFromPicoBuffers(buffers []PicoBuffer) []ioVec {
	vec := make([]ioVec, len(buffers))
	for i := range len(buffers) {
		vec[i] = ioVec{Base: (*byte)(unsafe.Pointer(&buffers[i])), Len: BufferSizePico}
	}

	return vec
}

func ioVecFromNanoBuffers(buffers []NanoBuffer) []ioVec {
	vec := make([]ioVec, len(buffers))
	for i := range len(buffers) {
		vec[i] = ioVec{Base: (*byte)(unsafe.Pointer(&buffers[i])), Len: BufferSizeNano}
	}

	return vec
}

func ioVecFromMicroBuffers(buffers []MicroBuffer) []ioVec {
	vec := make([]ioVec, len(buffers))
	for i := range buffers {
		vec[i] = ioVec{Base: (*byte)(unsafe.Pointer(&buffers[i])), Len: BufferSizeMicro}
	}

	return vec
}

func ioVecFromSmallBuffers(buffers []SmallBuffer) []ioVec {
	vec := make([]ioVec, len(buffers))
	for i := range buffers {
		vec[i] = ioVec{Base: (*byte)(unsafe.Pointer(&buffers[i])), Len: BufferSizeSmall}
	}

	return vec
}

func ioVecFromMediumBuffers(buffers []MediumBuffer) []ioVec {
	vec := make([]ioVec, len(buffers))
	for i := range buffers {
		vec[i] = ioVec{Base: (*byte)(unsafe.Pointer(&buffers[i])), Len: BufferSizeMedium}
	}

	return vec
}

func ioVecFromLargeBuffers(buffers []LargeBuffer) []ioVec {
	vec := make([]ioVec, len(buffers))
	for i := range buffers {
		vec[i] = ioVec{Base: (*byte)(unsafe.Pointer(&buffers[i])), Len: BufferSizeLarge}
	}

	return vec
}

func ioVecFromHugeBuffers(buffers []HugeBuffer) []ioVec {
	vec := make([]ioVec, len(buffers))
	for i := range buffers {
		vec[i] = ioVec{Base: (*byte)(unsafe.Pointer(&buffers[i])), Len: BufferSizeHuge}
	}

	return vec
}

func ioVecFromGiantBuffers(buffers []GiantBuffer) []ioVec {
	vec := make([]ioVec, len(buffers))
	for i := range buffers {
		vec[i] = ioVec{Base: (*byte)(unsafe.Pointer(&buffers[i])), Len: BufferSizeGiant}
	}

	return vec
}
