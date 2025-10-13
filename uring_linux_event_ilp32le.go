// ©Hayabusa Cloud Co., Ltd. 2025. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

//go:build amd64p32 || mips64p32le

package sox

type UringEvent struct {
	*UringEventData
	_      uint32
	Result int32
	Flags  uint32
}
