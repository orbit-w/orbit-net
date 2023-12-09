package stream_transport

/*
   @Author: orbit-w
   @File: base
   @2023 11月 周五 16:42
*/

const (
	CC_Running = iota
	CC_Done
)

const (
	DefaultMaxStreamPacket      = 1048576
	DefaultStreamRecvBufferSize = 512
	RpcMaxIncomingPacket        = 1048576
	MaxIncomingPacket           = 1<<18 - 1
)
