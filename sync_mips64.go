// ©Hayabusa Cloud Co., Ltd. 2025. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

//go:build mips || mipsle || mips64 || mips64p32 || mips64le || mips64p32le

package sox

//go:noescape
func sync()

var memoryBarrier = sync
