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
	Write(stream *Stream, pack packet.IPacket) error
	WriteData(stream *Stream, data []byte) (err error)
	Close(reason string) error
	CloseStream(streamId int64)
}

type IClientTransport interface {
	Write(stream *Stream, pack packet.IPacket, isLast bool) error
	WriteData(s *Stream, data []byte, isLast bool) error
	Close(reason string) error
	NewStream(ctx context.Context, initialSize int) (*Stream, error)
	CloseStream(streamId int64)
}

type DialOption struct {
	RemoteNodeId      string
	CurrentNodeId     string
	RemoteAddr        string
	MaxIncomingPacket uint32
	IsBlock           bool
	IsGzip            bool
	DisconnectHandler func(nodeId string)
}

type ConnOption struct {
	MaxIncomingPacket uint32
	StreamRecvBufSize int
	StreamHandler     func(st IServerTransport, s *Stream)
}
