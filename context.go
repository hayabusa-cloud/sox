// ©Hayabusa Cloud Co., Ltd. 2022. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

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
