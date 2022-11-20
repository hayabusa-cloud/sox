package sox

import "context"

type Ctx struct {
	context.Context
	fd int
}
