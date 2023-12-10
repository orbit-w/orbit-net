package transport

import (
	"context"
	"fmt"
	"github.com/orbit-w/golib/bases/packet"
	"github.com/orbit-w/orbit-net/core/stream_transport/metadata"
	"io"
	"net"
	"time"
)

/*
   @Author: orbit-w
   @File: tcp_server
   @2023 11月 周日 21:03
*/

type TcpServer struct {
	authed        bool
	streamRBSize  int
	conn          net.Conn
	framer        *Framer
	codec         *TcpCodec
	sw            *SenderWrapper
	buf           *ControlBuffer
	activeStreams *Streams
	ctx           context.Context
	cancel        context.CancelFunc

	streamHandler func(transport IServerTransport, stream *Stream)
}

func NewTcpServer(ctx context.Context, _conn net.Conn, ops *ConnOption) *TcpServer {
	if ctx == nil {
		ctx = context.Background()
	}
	cCtx, cancel := context.WithCancel(ctx)

	ts := &TcpServer{
		streamRBSize:  ops.StreamRecvBufSize,
		conn:          _conn,
		codec:         NewTcpCodec(ops.MaxIncomingPacket, false),
		streamHandler: ops.StreamHandler,
		activeStreams: NewStreams(),
		framer:        new(Framer),
		ctx:           cCtx,
		cancel:        cancel,
	}

	sw := NewSender(ts.SendData)
	ts.sw = sw
	ts.buf = NewControlBuffer(ops.MaxIncomingPacket, ts.sw)

	go ts.HandleLoop()
	return ts
}

func (ts *TcpServer) Write(stream *Stream, data packet.IPacket) (err error) {
	fp := ts.framer.Encode(&Frame{
		Type:     FrameReplyRaw,
		StreamId: stream.Id(),
		Data:     data,
	})
	err = ts.buf.Set(fp)
	fp.Return()
	return
}

// SendData 隐事调用 body.Return
// 消息编码协议: size<int32> | gzipped<bool> | body<bytes>
func (ts *TcpServer) SendData(body packet.IPacket) error {
	pack, err := ts.codec.EncodeBody(body)
	if err != nil {
		return err
	}
	defer pack.Return()
	if err = ts.conn.SetWriteDeadline(time.Now().Add(WriteTimeout)); err != nil {
		return err
	}
	_, err = ts.conn.Write(pack.Data())
	return err
}

func (ts *TcpServer) CloseStream(streamId int64) {
	_, ok := ts.activeStreams.GetAndDel(streamId)
	if ok {
		data := ts.framer.Encode(&Frame{
			StreamId: streamId,
			Type:     FrameCleanStream,
		})
		_ = ts.buf.Set(data)
		data.Return()
	}
}

func (ts *TcpServer) Close(reason string) error {
	return ts.conn.Close()
}

func (ts *TcpServer) HandleLoop() {
	header := headPool.Get().(*Buffer)
	buffer := bodyPool.Get().(*Buffer)
	defer func() {
		headPool.Put(header)
		bodyPool.Put(buffer)
	}()

	var (
		err  error
		data packet.IPacket
	)

	defer func() {
		ts.activeStreams.Close(func(stream *Stream) {
			stream.OnClose()
		})
		ts.buf.OnClose()
		if ts.conn != nil {
			_ = ts.conn.Close()
		}
		if err == io.EOF || IsClosedConnError(err) {
			//连接正常断开
		} else {
			fmt.Println(fmt.Errorf("tcp_conn disconnected: %s", err.Error()))
		}
	}()

	for {
		select {
		case <-ts.ctx.Done():
			return
		default:
			data, err = ts.codec.BlockDecode(ts.conn, header.Bytes, buffer.Bytes)
			if err != nil {
				return
			}
			if err = ts.OnData(data); err != nil {
				//TODO: 错误处理？
				return
			}
		}
	}
}

func (ts *TcpServer) OnData(data packet.IPacket) error {
	defer data.Return()
	for len(data.Remain()) > 0 {
		if bytes, err := data.ReadBytes32(); err != nil {
			reader := packet.Reader(bytes)
			frame, err := ts.framer.Decode(reader)
			if err != nil {
				return err
			}
			ts.HandleData(&frame)
		}
	}
	return nil
}

func (ts *TcpServer) HandleData(in *Frame) {
	switch in.Type {
	case FrameRaw:
		ts.handleRawFrame(in)
	case FrameStreamHeartbeat:
		ack := ts.framer.Encode(&Frame{
			Type: FrameStreamHeartbeat,
		})
		_ = ts.buf.Set(ack)
		ack.Return()
	case FrameStartStream:
		ts.handleStartFrame(in)
	case FrameCleanStream:
		ts.handleCleanFrame(in.StreamId)
	}
}

func (ts *TcpServer) handleRawFrame(in *Frame) {
	streamId := in.StreamId
	if in.End {
		stream, ok := ts.activeStreams.GetAndDel(streamId)
		if ok {
			stream.OnClose()
		}

		ack := ts.framer.Encode(&Frame{
			Type:     FrameCliHalfClosedAck,
			StreamId: streamId,
		})
		defer ack.Return()
		_ = ts.buf.Set(ack)
		return
	}

	msg := StreamMsg{}
	data := in.Data
	if data != nil && data.Len() > 0 {
		stream, ok := ts.activeStreams.Get(streamId)
		msg.buf = data
		if ok {
			stream.write(msg)
		}
	}
}

func (ts *TcpServer) handleStartFrame(in *Frame) {
	if ts.activeStreams.Exist(in.StreamId) {
		return
	}

	md := metadata.MD{}
	if err := metadata.Unmarshal(in.Data, &md); err != nil {
		//TODO: 敏感信息解析失败后处理？
		fmt.Println("metadata unmarshal failed: ", err.Error())
	}

	in.Data.Return()
	in.Data = nil

	ctx := metadata.NewMetaContext(ts.ctx, md)
	stream := NewStream(in.StreamId, ts.streamRBSize, ctx, ts, nil)
	ts.activeStreams.Reg(in.StreamId, stream)

	go func() {
		ts.streamHandler(ts, stream)
	}()
}

func (ts *TcpServer) handleCleanFrame(streamId int64) {
	stream, ok := ts.activeStreams.GetAndDel(streamId)
	if ok {
		stream.OnClose()
	}
}
