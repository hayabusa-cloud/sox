// ©Hayabusa Cloud Co., Ltd. 2025. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

//go:build mips64p32

package sox

type UringEvent struct {
	_ uint32
	*UringEventData
	Result int32
	Flags  uint32
}
