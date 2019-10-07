package xload

import (
	"context"
	"errors"
	"time"
)

// Picker picks one element from the collection and returns it.
// Pick should return nil when the Picker is unable to pick an element from the collection.
// See buffer.Do for usage.
type Picker interface {
	Pick(collection interface{}) interface{}
}

// Operation is the function that buffers requests, based on a collection of inputs.
// transforming single inserts into a batchinsert and returning the array of
// inserted items is an example of Operation.
// The operation is responsible for asserting the type of the inputs, aggregating
// them to run a single request and return the response.
type Operation = func(context.Context, ...interface{}) (interface{}, error)

// Buffer is a type capable of buffering operations.
type Buffer struct {
	op      Operation
	pending chan *request
	ctx     context.Context
}

// run is the function called on each cycle.
func (b *Buffer) run(ctx context.Context, reqs []*request) {

	fragments := make([]interface{}, len(reqs))
	var ignore int
	for i, r := range reqs {
		if r == nil {
			ignore++
			continue
		}
		fragments[i] = r.in
	}

	out, err := b.op(ctx, fragments[:len(fragments)-ignore]...)
	for _, r := range reqs {
		if r == nil {
			continue
		}
		r.out, r.err = out, err
		close(r.done)
	}
}

// NewBuffer prepares a buffer. The context is passed to operations for processing.
// The buffer stops accepting new requests when the context is done.
// The buffer runs a batch of request at the given frequency, or when the pending request
// number match the size, whichever comes first.
func NewBuffer(ctx context.Context, op Operation, size int, freq time.Duration) *Buffer {

	if size <= 0 {
		panic(errors.New("non-positive size for NewBuffer"))
	}

	b := &Buffer{
		op:      op,
		pending: make(chan *request, size),
		ctx:     ctx,
	}

	timer := time.NewTimer(freq)

	cycle := func(buf []*request) int {
		if buf[0] != nil {
			b.run(b.ctx, buf)
		}
		for k := range buf {
			buf[k] = nil
		}

		timer.Stop()
		select {
		case <-timer.C:
		default:

		}
		timer.Reset(freq)

		return 0
	}

	go func() {

		buf := make([]*request, size)
		var i int

		var closed bool
		closeOnce := func() {
			if !closed {
				closed = true
				close(b.pending)
			}
		}

		for {
			select {

			case <-ctx.Done():
				closeOnce()

			case <-timer.C:
				i = cycle(buf)

			case r, ok := <-b.pending:
				if !ok {
					cycle(buf)
					timer.Stop()
					return
				}

				buf[i] = r
				i++
				if i >= size {
					i = cycle(buf)
				}
			}
		}
	}()

	return b
}

// Do returns the response associated to the input after calling Pick(),
// or the whole buffered response if Picker is not implemented.
// It uses no context, as the inputs are buffered to generate a single request,
// and no context should be prefered.
// Do returns an error when the buffer's context is done to prevent dead requests.
func (b *Buffer) Do(v interface{}) (interface{}, error) {
	r := newRequest(v)
	select {
	case <-b.ctx.Done():
		return nil, b.ctx.Err()
	default:
		select {
		case b.pending <- r:
			return r.res()
		}
	}
}

// request holds the reference of the input, and provides a res() method that returns the response.
type request struct {
	in   interface{}
	out  interface{}
	done chan struct{}
	err  error
}

func newRequest(v interface{}) *request {
	return &request{
		in:   v,
		done: make(chan struct{}),
	}
}

// res blocks until the response is available, then returns it.
func (r *request) res() (interface{}, error) {
	<-r.done
	if r.err != nil {
		return nil, r.err
	}

	if p, ok := r.in.(Picker); ok {
		return p.Pick(r.out), nil
	}

	return r.out, r.err
}
