// ©Hayabusa Cloud Co., Ltd. 2022. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package sox

import (
	"errors"
	"io"
	"net"
)

var (
	ErrInterruptedSyscall     = errors.New("interrupted system call")
	ErrTemporarilyUnavailable = errors.New("resource temporarily unavailable")
	ErrInProgress             = errors.New("in progress")
	ErrFaultParams            = errors.New("fault parameters")
	ErrInvalidParam           = errors.New("invalid param")
	ErrProcessFileLimit       = errors.New("process open fd limit")
	ErrSystemFileLimit        = errors.New("system open fd limit")
	ErrNoDevice               = errors.New("no device")
	ErrNoAvailableMemory      = errors.New("no available kernel memory")
	ErrNoPermission           = errors.New("operation not permitted")
)

type NetworkType int

const (
	NetworkUnix = 1
	NetworkIPv4 = 2
	NetworkIPv6 = 10
)

type Socket interface {
	Fd() int
	Protocol() UnderlyingProtocol
	io.Reader
	io.Writer
	io.Closer
}

type Listener = net.Listener
type Conn = net.Conn
type Addr = net.Addr
type OpError = net.OpError
type AddrError = net.AddrError
type InvalidAddrError = net.InvalidAddrError
type UnknownNetworkError = net.UnknownNetworkError

func GetFd(x any) int {
	if x, ok := x.(pollFd); ok {
		return x.Fd()
	}
	if x, ok := x.(file); ok {
		f, err := x.File()
		if err != nil {
			return -1
		}
		return int(f.Fd())
	}

	return -1
}
