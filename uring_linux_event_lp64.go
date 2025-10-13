// ©Hayabusa Cloud Co., Ltd. 2025. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

//go:build amd64 || arm64 || loong64 || mips64

package sox

type UringEvent struct {
	*UringEventData
	Result int32
	Flags  uint32
}
