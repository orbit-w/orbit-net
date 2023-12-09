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
	Send(data packet.IPacket) error
	CloseSend() error
}

type IStreamServer interface {
	Send(data packet.IPacket) error
	Recv() (packet.IPacket, error)
	Id() int64
	Close(reason string) error
	Context() context.Context
}
