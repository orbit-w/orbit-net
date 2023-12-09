package transport

import (
	"context"
	"github.com/orbit-w/golib/bases/packet"
)

/*
   @Author: orbit-w
   @File: transport
   @2023 11月 周日 17:01
*/

type IServerTransport interface {
	Write(stream *Stream, data packet.IPacket) error
	Close(reason string) error
	CloseStream(streamId int64)
}

type IClientTransport interface {
	Write(stream *Stream, data packet.IPacket, isLast bool) error
	Close(reason string) error
	NewStream(ctx context.Context, initialSize int) (*Stream, error)
	CloseStream(streamId int64)
}

type DialOptions struct {
	RemoteNodeId      string
	CurrentNodeId     string
	RemoteAddr        string
	MaxIncomingPacket uint32
	IsBlock           bool
	IsGzip            bool
	DisconnectHandler func(nodeId string)
}

type ConnOptions struct {
	MaxIncomingPacket uint32
	StreamRecvBufSize int
	StreamHandler     func(st IServerTransport, s *Stream)
}
