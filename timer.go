// ©Hayabusa Cloud Co., Ltd. 2022. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package sox

import (
	"io"
	"time"
)

const jiffies = time.Millisecond

type Timer interface {
	Now() time.Time
	io.ReadCloser
}
