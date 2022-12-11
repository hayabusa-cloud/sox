// ©Hayabusa Cloud Co., Ltd. 2022. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package sox

import (
	"context"
	"time"
)

// Interface is the interface of sox events notification library
type Interface interface {
	// AddListen adds listen event on the given listener with the given event handler
	AddListen(listener Listener, handler AcceptedHandler)
	// AddIO adds io event with the given event handlers
	AddIO(dispatch DispatchHandler, message MessageHandler, written WrittenHandler, closed ClosedHandler)
	// AddTimer adds timer event with the given event handler
	AddTimer(ticked TickedHandler)
	// Serve starts serving
	Serve() error
	// Poll waits for events. The d parameter specifies the duration that Poll will block
	// d == 0 means Poll method will return immediately even if there is no events came
	// d < 0 means Poll method will block forever or until there are any events
	Poll(d time.Duration) error
}

// Options represents
type Options struct {
	// UserPoll sets if it's need to call Interface.Poll to wait events manually
	// set UserPoll to true might improve performance of some IO operations
	UserPoll bool
	// Parallel sets the number of goroutines that do handle the events. Default value is 0
	// Parallel <= 0 means the events will be handled at the same goroutine as the polling goroutine
	// Parallel == 1 means the events will be handled in series at one another worker goroutine
	// Parallel >= 2 means the events will be handled in parallel at Parallel worker goroutines
	// It is possible to specify which worker will be used to handle the event
	// by implement your customized DispatchHandler
	Parallel int
}

var defaultOptions = Options{}

// New creates and returns a new event loop with given options
func New(options ...func(option *Options)) (evLoop Interface, err error) {
	panic("todo")
}

// ListenAndServe listens on the given network and address, and then calls
// Serve method with the given handler to handle incoming connect requests
func ListenAndServe(network string, address string, handler AcceptedHandler) error {
	panic("todo")
}

// Serve accepts and handles incoming connections on the listener l with the given handler
// Serve method works like this:
//
//	evLoop, _ = New()
//	evLoop.AddAccepted(listener, handler)
//	return evLoop.Serve()
func Serve(listener Listener, handler AcceptedHandler) error {
	panic("todo")
}

// AcceptedHandler handles the listener accepted new connections event
type AcceptedHandler interface {
	ServeAccepted(conn Conn, listener Listener)
}

// ConnectedHandler handles the client connected to remote server event
type ConnectedHandler interface {
	ServeConnected(conn Conn)
}

// ClosedHandler handles connection closed event
type ClosedHandler interface {
	ServeDisconnected(lfd int, rfd int)
}

// DispatchHandler chooses the MessageHandler which will be used to handle the incoming message
type DispatchHandler interface {
	ServeDispatch(ctx context.Context, reader PollReader) MessageHandler
}

// MessageHandler is the most often used handle. It handles the incoming request messages
type MessageHandler interface {
	ServeMessage(ctx context.Context, reply PollWriter, request PollReader)
}

// WrittenHandler handles send message completed events
type WrittenHandler interface {
	ServeWritten(ctx context.Context, writer PollWriter)
}

// TickedHandler handles timer events
type TickedHandler interface {
	ServeMessage(at time.Time)
}
