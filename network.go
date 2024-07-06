// ©Hayabusa Cloud Co., Ltd. 2022. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package sox

import (
	"encoding/binary"
	"errors"
	"io"
	"net"
)

var (

	// ErrInterruptedSyscall represents an error indicating an interrupted system call.
	ErrInterruptedSyscall = errors.New("interrupted system call")

	// ErrTemporarilyUnavailable represents an error indicating that a resource is temporarily unavailable.
	// In many cases, when ErrTemporarilyUnavailable is returned, the caller would block.
	ErrTemporarilyUnavailable = errors.New("resource temporarily unavailable")

	// ErrInProgress represents an error indicating that an operation is in progress.
	ErrInProgress = errors.New("in progress")

	// ErrFaultParams represents an error indicating fault parameters.
	ErrFaultParams = errors.New("fault parameters")

	// ErrInvalidParam is an error indicating that a parameter passed to a function is invalid.
	ErrInvalidParam = errors.New("invalid param")

	// ErrProcessFileLimit represents an error indicating that the process has reached its open file descriptor limit.
	ErrProcessFileLimit = errors.New("process open fd limit")

	// ErrSystemFileLimit represents an error indicating that the system has reached its open file descriptor limit.
	ErrSystemFileLimit = errors.New("system open fd limit")

	// ErrNoDevice represents an error indicating that there is no device available.
	ErrNoDevice = errors.New("no device")

	// ErrNoAvailableMemory represents an error indicating that there is no available kernel memory.
	ErrNoAvailableMemory = errors.New("no available kernel memory")

	// ErrNoPermission represents an error indicating that the operation is not permitted.
	ErrNoPermission = errors.New("operation not permitted")

	// ErrCanceled represents an error indicating that an operation was canceled.
	ErrCanceled = errors.New("operation canceled")
)

// NetworkType is a custom type used to represent network types in the code.
type NetworkType int

const (
	// NetworkUnix represents a constant value used to indicate the Unix network type.
	NetworkUnix NetworkType = 1
	// NetworkIPv4 represents a constant value used to indicate the IPv4 network type.
	NetworkIPv4 NetworkType = 2
	// NetworkIPv6 represents a constant value used to indicate the IPv6 network type.
	NetworkIPv6 NetworkType = 10
)

// Socket is a generic network socket
type Socket interface {
	// Fd returns the file descriptor associated with the socket.
	Fd() int

	// NetworkType returns the network type associated with the socket.
	NetworkType() NetworkType

	// Protocol returns the underlying protocol of the Socket.
	// It indicates the transmission protocol features.
	Protocol() UnderlyingProtocol

	// Reader is an interface that represents an object that can read data.
	io.Reader

	// Writer is an interface that represents a data writer.
	io.Writer

	// Closer is an interface implemented by objects that can be closed.
	io.Closer
}

// ListenerSocket is an interface that represents a network socket capable of accepting incoming connections.
// It extends the Socket interface and adds the Accept method which returns a new socket for the accepted connection.
type ListenerSocket interface {
	Socket
	// Accept accepts incoming connections and returns a new socket for the accepted connection.
	Accept() (Socket, error)
}

// Listener represents a generic network listener.
// It is an alias for the net.Listener type.
type Listener = net.Listener

// Conn is an alias for net.Conn type. It represents a generic network connection.
// It has methods for accessing the local and remote addresses, setting deadlines,
// and shutting down the connection.
//
// Conn is used in various examples and functions in the codebase.
// Some of the usage examples include:
// - Accept: accepts a new connection from the listener
// - NewTCPConn: creates a new TCP connection
// - NewUDPConn: creates a new UDP connection
// - NewSCTPConn: creates a new SCTP connection
// - NewUnixConn: creates a new Unix connection
type Conn = net.Conn

// Addr is an alias for the net.Addr type. It represents a network address.
type Addr = net.Addr

// OpError is an error that represents an operation error along with
// the operation name, network protocol, source address, target address,
// and the underlying error. It is an alias for net.OpError.
type OpError = net.OpError

// AddrError is an error type that represents an address-related error.
type AddrError = net.AddrError

// InvalidAddrError is an error type that represents an invalid network address.
type InvalidAddrError = net.InvalidAddrError

// UnknownNetworkError represents an error when an unknown network is encountered.
// The error is typically returned when unexpected network types are encountered in network-related operations.
type UnknownNetworkError = net.UnknownNetworkError

var (
	// DefaultResolver is a variable that holds the default resolver implementation
	// for performing network resolution in the Go standard library.
	DefaultResolver = net.DefaultResolver
)

var (
	// NetworkByteOrder represents the network byte order, which is big-endian.
	NetworkByteOrder = binary.BigEndian
)

// GetFd gets the file descriptor of a given object. It checks if the object implements the PollFd interface,
// and if so, returns the result of calling the Fd method on it. If the object implements the file interface,
// it retrieves the file using the File method and returns its file descriptor. If none of these conditions are met,
// it returns -1.
func GetFd(x any) int {
	if x, ok := x.(PollFd); ok {
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
