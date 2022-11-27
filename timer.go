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
