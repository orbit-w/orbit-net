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

type IClientConn interface {
	// NewStream creates a new Stream for the client side.
	// To ensure resources are not leaked due to the stream returned, one of the following
	// actions must be performed:
	//  1. Call Close on the ClientConn.
	//  2. Call RecvMsg until a non-nil error is returned.
	//  3. Receive a non-nil, non-io.EOF error from Header or SendMsg.
	NewStream(ctx context.Context) (*StreamClient, error)

	// Close tears down the ClientConn and underlying the connection.
	Close() error
}

type ClientConn struct {
	state int32
	ct    transport.IClientTransport
}

type DialOption struct {
	IsBlock           bool
	IsGzip            bool
	MaxIncomingPacket uint32 //default: 1<<18 - 1
	CurrentNodeId     string
	DisconnectHandler func(nodeId string)
}

func DialWithOps(remoteAddr, remoteId string, op ...DialOption) IClientConn {
	conn := new(ClientConn)
	bop := parseOp(remoteAddr, remoteId, op...)
	conn.ct = transport.NewTcpClient(bop)
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

func parseOp(remoteAddr, remoteId string, ops ...DialOption) transport.DialOption {
	var (
		maxIncoming uint32 = MaxIncomingPacket
		isBlock     bool   //block 模式目前不支持
		isGzip      bool   //gzip 压缩目前不支持
		dh          func(nodeId string)
	)

	if len(ops) > 0 {
		op := ops[0]
		if op.MaxIncomingPacket > 0 {
			maxIncoming = op.MaxIncomingPacket
		}
		if op.IsGzip {
			isGzip = op.IsGzip
		}
		dh = op.DisconnectHandler
	}

	return transport.DialOption{
		RemoteAddr:        remoteAddr,
		RemoteNodeId:      remoteId,
		MaxIncomingPacket: maxIncoming,
		IsBlock:           isBlock,
		IsGzip:            isGzip,
		DisconnectHandler: dh,
	}
}
