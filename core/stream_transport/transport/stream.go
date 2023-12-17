package transport

import (
	"context"
	"github.com/orbit-w/golib/bases/packet"
	"github.com/orbit-w/orbit-net/core/stream_transport/transport_err"
	"io"
	"sync/atomic"
)

/*
   @Author: orbit-w
   @File: stream
   @2023 11月 周日 16:09
*/

type Stream struct {
	id     int64
	state  StreamState
	ts     *TcpServer
	tc     *TcpClient
	rb     *ReceiveBuf
	recvCh <-chan StreamMsg
	ctx    context.Context
	cancel context.CancelFunc
}

type StreamMsg struct {
	buf packet.IPacket
	err error
}

func NewStream(_id int64, initialSize int, f context.Context, _ts *TcpServer, _tc *TcpClient) *Stream {
	rb := NewReceiveBuf(initialSize)
	ctx, cancel := context.WithCancel(f)
	return &Stream{
		id:     _id,
		ts:     _ts,
		tc:     _tc,
		rb:     rb,
		recvCh: rb.get(),
		ctx:    ctx,
		cancel: cancel,
	}
}

func (s *Stream) Id() int64 {
	return s.id
}

func (s *Stream) write(msg StreamMsg) {
	_ = s.rb.put(msg)
}

func (s *Stream) Read() (packet.IPacket, error) {
	select {
	case msg, ok := <-s.recvCh:
		if !ok {
			return nil, transport_err.ErrCancel
		}
		if msg.err != nil {
			return msg.buf, msg.err
		}
		s.rb.load()
		return msg.buf, nil
	case <-s.ctx.Done(): //TODO:
		return nil, transport_err.ErrCancel
	}
}

func (s *Stream) GetErr() error {
	return s.rb.err
}

func (s *Stream) OnClose() {
	_ = s.rb.put(StreamMsg{
		err: transport_err.ErrCancel,
	})
}

func (s *Stream) OnElegantlyClose() {
	_ = s.rb.put(StreamMsg{
		err: io.EOF,
	})
}

func (s *Stream) Close() {
	s.rb.OnClose()
}

func (s *Stream) GetCtx() context.Context {
	return s.ctx
}

func (s *Stream) swapState(st StreamState) StreamState {
	return StreamState(atomic.SwapUint32((*uint32)(&s.state), uint32(st)))
}

func (s *Stream) compareAndSwapState(old, new StreamState) bool {
	return atomic.CompareAndSwapUint32((*uint32)(&s.state), uint32(old), uint32(new))
}

func (s *Stream) getState() StreamState {
	return StreamState(atomic.LoadUint32((*uint32)(&s.state)))
}
