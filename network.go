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
	ErrNoPermission           = errors.New("operation not permiited")
)

type NetworkType int

const (
	NetworkUnix = 1
	NetworkIPv4 = 2
	NetworkIPv6 = 10
)

type Socket interface {
	FD() int
	Protocol() UnderlyingProtocol
	io.Reader
	io.Writer
	io.Closer
}

type Conn = net.Conn
type Addr = net.Addr
type OpError = net.OpError
type AddrError = net.AddrError
type InvalidAddrError = net.InvalidAddrError
type UnknownNetworkError = net.UnknownNetworkError
