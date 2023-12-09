package transport

import (
	"github.com/orbit-w/golib/bases/container/ring_buffer"
	"sync"
)

/*
   @Author: orbit-w
   @File: control
   @2023 11月 周日 16:11
*/

type ReceiveBuf struct {
	c   chan StreamMsg
	mu  sync.Mutex
	buf *ring_buffer.RingBuffer[StreamMsg]
	err error
}

func NewReceiveBuf(size int) *ReceiveBuf {
	return &ReceiveBuf{
		mu:  sync.Mutex{},
		c:   make(chan StreamMsg, 1),
		buf: ring_buffer.New[StreamMsg](size),
	}
}

func (rb *ReceiveBuf) OnClose() {
	rb.mu.Lock()
	defer rb.mu.Unlock()
	rb.err = ErrStreamShutdown
	close(rb.c)
	rb.buf.Contract()
}

func (rb *ReceiveBuf) put(r StreamMsg) error {
	rb.mu.Lock()
	defer rb.mu.Unlock()
	if rb.err != nil {
		return ReceiveBufPutErr(rb.err)
	}

	if r.err != nil {
		rb.err = r.err
	}
	if rb.buf.IsEmpty() {
		select {
		case rb.c <- r:
			return nil
		default:
		}
	}
	rb.buf.Push(r)
	return nil
}

func (rb *ReceiveBuf) load() {
	rb.mu.Lock()
	if !rb.buf.IsEmpty() {
		r, _ := rb.buf.Peek()
		select {
		case rb.c <- r:
			_, _ = rb.buf.Pop()
		default:
		}
	}
	rb.buf.Contract()
	rb.mu.Unlock()
}

func (rb *ReceiveBuf) get() <-chan StreamMsg {
	return rb.c
}
