package stream_transport

import (
	"context"
	"github.com/orbit-w/orbit-net/core/stream_transport/transport"
	"sync/atomic"
)

/*
   @Author: orbit-w
   @File: stream_conn
   @2023 11月 周五 16:43
*/

type ClientConn struct {
	state int32
	ct    transport.IClientTransport
}

func DialWithOps(op *transport.DialOptions) *ClientConn {
	conn := new(ClientConn)
	conn.ct = transport.NewTcpClient(op)
	return conn
}

func (cc *ClientConn) NewStream(ctx context.Context) (*StreamClient, error) {
	stream, err := cc.ct.NewStream(ctx, DefaultStreamRecvBufferSize)
	if err != nil {
		return nil, err
	}
	return &StreamClient{
		Id:     stream.Id(),
		conn:   cc,
		stream: stream,
		ct:     cc.ct,
	}, nil
}

func (cc *ClientConn) Close() error {
	atomic.StoreInt32(&cc.state, CC_Done)
	return cc.ct.Close("")
}
