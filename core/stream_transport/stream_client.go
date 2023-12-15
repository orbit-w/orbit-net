package stream_transport

import (
	"github.com/orbit-w/golib/bases/packet"
	"github.com/orbit-w/orbit-net/core/stream_transport/transport"
	"github.com/orbit-w/orbit-net/core/stream_transport/transport_err"
)

/*
   @Author: orbit-w
   @File: stream_client
   @2023 11月 周五 16:45
*/

type StreamClient struct {
	sentLast bool
	Id       int64
	conn     *ClientConn
	stream   *transport.Stream
	ct       transport.IClientTransport
}

func (sc *StreamClient) Recv() (packet.IPacket, error) {
	return sc.stream.Read()
}

func (sc *StreamClient) Send(pack packet.IPacket) error {
	if pack == nil {
		return nil
	}
	return sc.ct.Write(sc.stream, pack, false)
}

func (sc *StreamClient) SendData(data []byte) error {
	if data == nil {
		return nil
	}
	return sc.ct.WriteData(sc.stream, data, false)
}

func (sc *StreamClient) CloseSend() error {
	if sc.sentLast {
		return nil
	}
	sc.sentLast = true
	if err := sc.ct.Write(sc.stream, nil, true); err != nil {
		if !transport_err.IsErrRpcDisconnected(err) {
			return err
		}
		return nil
	}
	return nil
}
