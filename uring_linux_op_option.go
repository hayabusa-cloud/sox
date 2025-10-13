// ©Hayabusa Cloud Co., Ltd. 2025. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

//go:build linux

package sox

import (
	"os"
	"time"
)

type UringOpOption struct {
	Flags           uint8
	Callback        uint8
	Ioprio          uint16
	FileIndex       uint32
	Backlog         int
	ReadBufferSize  int
	WriteBufferSize int
	Duration        time.Duration
	FileMode        os.FileMode
	Advice          int
	Offset          int64
	N               *int
	Count           int
}

const (
	uringOpFlagsNone = iota
)
const (
	uringOpCallbackNone = iota
	uringOpCallbackBind
	uringOpCallbackListen
)
const (
	uringOpIoprioNone = iota
)

var defaultUringOpOption = UringOpOption{
	Flags:           uringOpFlagsNone,
	Callback:        uringOpCallbackNone,
	Ioprio:          uringOpIoprioNone,
	Backlog:         defaultBacklog,
	ReadBufferSize:  bufferSizeDefault,
	WriteBufferSize: bufferSizeDefault,
	Duration:        jiffy,
	FileMode:        0644,
	N:               nil,
	Count:           1,
}

type UringOpOptionFunc func(opt *UringOpOption)

func (uo *UringOpOption) Apply(opts ...UringOpOptionFunc) {
	for _, f := range opts {
		f(uo)
	}
}
