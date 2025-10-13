// ©Hayabusa Cloud Co., Ltd. 2025. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

//go:build amd64 || amd64p32

package sox

//go:noescape
func mFence()

var memoryBarrier = mFence
