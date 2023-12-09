package stream_transport

import (
	"github.com/orbit-w/golib/bases/packet"
	"github.com/orbit-w/orbit-net/core/stream_transport/transport"
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

func (sc *StreamClient) Send(data packet.IPacket) error {
	if data == nil {
		return nil
	}
	return sc.ct.Write(sc.stream, data, false)
}

func (sc *StreamClient) CloseSend() error {
	if sc.sentLast {
		return nil
	}
	sc.sentLast = true
	if err := sc.ct.Write(sc.stream, nil, true); err != nil {
		if !transport.IsErrRpcDisconnected(err) {
			return err
		}
		return nil
	}
	return nil
}

// Close application layer active control call Close()
func (sc *StreamClient) Close() error {
	sc.conn.ct.CloseStream(sc.Id)
	sc.stream.Close()
	sc.stream = nil
	sc.conn = nil
	return nil
}
