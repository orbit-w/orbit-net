package stream_transport

import (
	"context"
	"github.com/orbit-w/golib/bases/packet"
)

/*
   @Author: orbit-w
   @File: stream
   @2023 11月 周五 17:06
*/

type IStreamClient interface {
	Recv() (packet.IPacket, error)
	// Send IStreamClient does not implicitly call IPacket.Return to return the
	// packet to the pool, and the user needs to explicitly call it.
	Send(pack packet.IPacket) error
	SendData(data []byte) error
	// CloseSend closes the send direction of the stream. It closes the stream
	// when non-nil transport_err is met. It is also not safe to call CloseSend
	// concurrently with SendMsg.
	CloseSend() error
}

type IStreamServer interface {
	// Send IStreamServer does not implicitly call IPacket.Return to return the
	// packet to the pool, and the user needs to explicitly call it.
	Send(pack packet.IPacket) error
	SendData(data []byte) error
	Recv() (packet.IPacket, error)
	Id() int64
	Close(reason string) error
	Context() context.Context
}
