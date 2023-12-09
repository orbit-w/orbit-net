package stream_transport

import (
	"context"
	"github.com/orbit-w/golib/bases/packet"
	"github.com/orbit-w/orbit-net/core/stream_transport/transport"
)

/*
   @Author: orbit-w
   @File: stream_server
   @2023 11月 周五 16:53
*/

type ServerStream struct {
	id     int64 //stream Id
	stream *transport.Stream
	st     transport.IServerTransport
}

func NewServerStream(s *transport.Stream, _st transport.IServerTransport) *ServerStream {
	return &ServerStream{
		id:     s.Id(),
		stream: s,
		st:     _st,
	}
}

func (ss *ServerStream) Send(data packet.IPacket) error {
	return ss.st.Write(ss.stream, data)
}

func (ss *ServerStream) Recv() (packet.IPacket, error) {
	return ss.stream.Read()
}

func (ss *ServerStream) Id() int64 {
	return ss.id
}

func (ss *ServerStream) Close(reason string) error {
	return ss.st.Close(reason)
}

func (ss *ServerStream) Context() context.Context {
	return ss.stream.GetCtx()
}
