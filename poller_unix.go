// ©Hayabusa Cloud Co., Ltd. 2024. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

//go:build unix

package sox

import (
	"golang.org/x/sys/unix"
)

// PollSignalfd is the interface that groups Fd and ReadSiginfo method
type PollSignalfd interface {
	pollFd
	// ReadSiginfo reads and returns the came signal info
	ReadSiginfo() (sig unix.Signal, code int, err error)
}
