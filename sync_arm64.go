// ©Hayabusa Cloud Co., Ltd. 2025. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

//go:build arm || arm64 || armbe || arm64be

package sox

//go:noescape
func dmbSy()

var memoryBarrier = dmbSy
