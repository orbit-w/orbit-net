package transport

import (
	err "github.com/orbit-w/orbit-net/core/stream_transport/transport_err"
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
	buf []StreamMsg
	err error
}

func NewReceiveBuf(size int) *ReceiveBuf {
	return &ReceiveBuf{
		mu:  sync.Mutex{},
		c:   make(chan StreamMsg, 1),
		buf: make([]StreamMsg, 0),
	}
}

func (rb *ReceiveBuf) OnClose() {
	rb.mu.Lock()
	defer rb.mu.Unlock()
	rb.err = err.ErrStreamShutdown
	close(rb.c)
}

func (rb *ReceiveBuf) put(r StreamMsg) error {
	rb.mu.Lock()
	if rb.err != nil {
		rb.mu.Unlock()
		return err.ReceiveBufPutErr(rb.err)
	}

	if r.err != nil {
		rb.err = r.err
	}
	if len(rb.buf) == 0 {
		select {
		case rb.c <- r:
			rb.mu.Unlock()
			return nil
		default:
		}
	}
	rb.buf = append(rb.buf, r)
	rb.mu.Unlock()
	return nil
}

func (rb *ReceiveBuf) load() {
	rb.mu.Lock()
	if len(rb.buf) > 0 {
		select {
		case rb.c <- rb.buf[0]:
			rb.buf[0] = StreamMsg{}
			rb.buf = rb.buf[1:]
		default:
		}
	}
	rb.mu.Unlock()
}

func (rb *ReceiveBuf) get() <-chan StreamMsg {
	return rb.c
}
