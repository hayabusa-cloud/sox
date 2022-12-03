package sox

import "context"

type userdataGetter[T any] interface {
	getUserdata() T
}

type userdataSetter[T any] interface {
	setUserdata(data T)
}

type innerCtxGetter interface {
	innerCtx() context.Context
}

type userdataCtx[T any] struct {
	context.Context
	userdata T
}

func (ctx *userdataCtx[T]) getUserdata() T {
	return ctx.userdata
}

func (ctx *userdataCtx[T]) setUserdata(data T) {
	ctx.userdata = data
}

func (ctx *userdataCtx[T]) innerCtx() context.Context {
	return ctx.Context
}

func ContextWithUserdata[T any](parent context.Context, userdata T) context.Context {
	if uc, ok := parent.(userdataSetter[T]); ok {
		uc.setUserdata(userdata)
		return parent
	}

	return &userdataCtx[T]{Context: parent, userdata: userdata}
}

func ContextUserdata[T any](ctx context.Context) (ret T) {
	for ctx != nil {
		if uc, ok := ctx.(userdataGetter[T]); ok {
			return uc.getUserdata()
		}
		if uc, ok := ctx.(innerCtxGetter); ok {
			ctx = uc.innerCtx()
		} else {
			break
		}
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

func (ctx *fdCtx) getFD() (fd int) {
	return ctx.fd
}

func (ctx *fdCtx) setFD(fd int) {
	ctx.fd = fd
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
