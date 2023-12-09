package stream_transport

import (
	"context"
	"github.com/orbit-w/orbit-net/core/stream_transport/transport"
	"net"
	"sync"
	"time"
)

/*
   @Author: orbit-w
   @File: server
   @2023 11月 周五 17:04
*/

type Server struct {
	isGzip   bool
	ccu      int32
	host     string
	listener net.Listener
	rw       sync.RWMutex
	ctx      context.Context
	cancel   context.CancelFunc

	handle func(stream IStreamServer) error
}

type AcceptorOptions struct {
	MaxIncomingPacket uint32
	IsGzip            bool
}

func (ins *Server) Serve(listener net.Listener, _handle func(stream IStreamServer) error, ops ...AcceptorOptions) {
	op := parseAndWrapOP(ops...)
	transport.NewBodyPool(op.MaxIncomingPacket)
	ctx, cancel := context.WithCancel(context.Background())
	ins.host = ""
	ins.isGzip = op.IsGzip
	ins.ctx = ctx
	ins.cancel = cancel
	ins.handle = _handle
	ins.listener = listener
	go ins.acceptLoop()
}

func (ins *Server) acceptLoop() {
	for {
		conn, err := ins.listener.Accept()
		if err != nil {
			select {
			case <-ins.ctx.Done():
				return
			default:
				time.Sleep(100 * time.Millisecond)
				continue
			}
		}

		transport.NewTcpServer(ins.ctx, conn, &transport.ConnOptions{
			StreamHandler:     ins.handleStream,
			MaxIncomingPacket: MaxIncomingPacket,
			StreamRecvBufSize: DefaultStreamRecvBufferSize,
		})
	}
}

func (ins *Server) handleStream(ts transport.IServerTransport, stream *transport.Stream) {
	serverStream := NewServerStream(stream, ts)
	defer func() {
		err := stream.GetErr()
		if err == nil {
			ts.CloseStream(stream.Id())
		}
		stream.Close()
	}()

	if appErr := ins.handle(serverStream); appErr != nil {
		//TODO:
	}
}

func parseAndWrapOP(ops ...AcceptorOptions) AcceptorOptions {
	var op AcceptorOptions
	if len(ops) > 0 {
		op = ops[0]
	}
	if op.MaxIncomingPacket == 0 {
		op.MaxIncomingPacket = RpcMaxIncomingPacket
	}
	return op
}
