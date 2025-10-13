// ©Hayabusa Cloud Co., Ltd. 2024. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

//go:build unix

package sox

import "golang.org/x/sys/unix"

func maximizeMemoryLock() {
	rlim := unix.Rlimit{}
	err := unix.Getrlimit(unix.RLIMIT_MEMLOCK, &rlim)
	if err != nil {
		panic(err)
	}
	if rlim.Cur >= rlim.Max {
		return
	}
	rlim.Cur = rlim.Max
	err = unix.Setrlimit(unix.RLIMIT_MEMLOCK, &rlim)
	if err != nil {
		panic(err)
	}
}
