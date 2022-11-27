package sox

import "context"

func ContextWithUserdata(parent context.Context, userdata any) context.Context {
	return context.WithValue(parent, "__SOX_USERDATA__", userdata)
}

func ContextUserdata[T any](parent context.Context) (ret T) {
	val := parent.Value("__SOX_USERDATA__")
	if val == nil {
		return
	}
	ok := false
	if ret, ok = val.(T); ok {
		return ret
	}
	return
}

type fdGetter interface {
	getFD() (fd int)
}

type fdSetter interface {
	setFD(fd int)
}

type fdCtx struct {
	context.Context
	fd int
}

func (ctx *fdCtx) setFD(fd int) {
	ctx.fd = fd
}

func (ctx *fdCtx) getFD() (fd int) {
	return ctx.fd
}

func contextWithFD(parent context.Context, fd int) context.Context {
	if fdc, ok := parent.(fdSetter); ok {
		fdc.setFD(fd)
		return parent
	}

	return &fdCtx{Context: parent, fd: fd}
}

func contextFD(ctx context.Context) int {
	if fdc, ok := ctx.(fdGetter); ok {
		return fdc.getFD()
	}

	return -1
}
