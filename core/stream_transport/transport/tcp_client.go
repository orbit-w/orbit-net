package transport

import (
	"context"
	"fmt"
	"github.com/orbit-w/golib/bases/misc/number_utils"
	"github.com/orbit-w/golib/bases/packet"
	"github.com/orbit-w/orbit-net/core/stream_transport/metadata"
	"github.com/orbit-w/orbit-net/core/stream_transport/transport_err"
	"go.uber.org/atomic"
	"io"
	"log"
	"net"
	"sync"
	"time"
)

/*
   @Author: orbit-w
   @File: tcp_client
   @2023 11月 周日 16:32
*/

type TcpClient struct {
	mu            sync.Mutex
	state         atomic.Uint32
	lastAck       atomic.Int64
	remoteAddr    string
	remoteNodeId  string
	currentNodeId string
	ctx           context.Context
	cancel        context.CancelFunc
	framer        *Framer
	codec         TcpCodec
	conn          net.Conn
	buf           *ControlBuffer
	sw            *SenderWrapper
	streams       *Streams
	dHandle       func(remoteNodeId string)
}

func NewTcpClient(_ops DialOption) IClientTransport {
	ctx, cancel := context.WithCancel(context.Background())
	buf := new(ControlBuffer)
	BuildControlBuffer(buf, _ops.MaxIncomingPacket)
	tc := &TcpClient{
		mu:            sync.Mutex{},
		remoteAddr:    _ops.RemoteAddr,
		remoteNodeId:  _ops.RemoteNodeId,
		currentNodeId: _ops.CurrentNodeId,
		dHandle:       _ops.DisconnectHandler,
		buf:           buf,
		ctx:           ctx,
		cancel:        cancel,
		framer:        new(Framer),
		streams:       NewStreams(),
		codec:         NewTcpCodec(_ops.MaxIncomingPacket, _ops.IsGzip),
	}

	if _ops.IsBlock {
		//TODO: 阻塞模式？
	}
	tc.DialWithOps(_ops)
	return tc
}

// Write TcpClient obj does not implicitly call IPacket.Return to return the
// packet to the pool, and the user needs to explicitly call it.
func (tc *TcpClient) Write(s *Stream, pack packet.IPacket, isLast bool) error {
	switch {
	case isLast:
		if !s.compareAndSwapState(StreamActive, StreamWriteDone) {
			return transport_err.ErrStreamDone
		}
	case s.getState() != StreamActive:
		return transport_err.ErrStreamDone
	}
	var reader packet.IPacket
	if pack != nil && len(pack.Remain()) > 0 {
		reader = packet.Reader(pack.Data())
	}
	frame := Frame{
		Type:     FrameRaw,
		StreamId: s.Id(),
		Data:     reader,
		End:      isLast,
	}
	fp := tc.framer.Encode(&frame)
	err := tc.buf.Set(fp)
	fp.Return()
	return err
}

func (tc *TcpClient) WriteData(s *Stream, data []byte, isLast bool) error {
	switch {
	case isLast:
		if !s.compareAndSwapState(StreamActive, StreamWriteDone) {
			return transport_err.ErrStreamDone
		}
	case s.getState() != StreamActive:
		return transport_err.ErrStreamDone
	}
	reader := packet.Reader(data)
	frame := Frame{
		Type:     FrameRaw,
		StreamId: s.Id(),
		Data:     reader,
		End:      isLast,
	}
	fp := tc.framer.Encode(&frame)
	err := tc.buf.Set(fp)
	fp.Return()
	return err
}

func (tc *TcpClient) Close(reason string) error {
	if tc.conn != nil {
		_ = tc.conn.Close()
	}
	return nil
}

func (tc *TcpClient) NewStream(ctx context.Context, initialSize int) (*Stream, error) {
	streamId := tc.streams.StreamId()
	stream := NewStream(streamId, initialSize, tc.ctx, nil, tc)
	md, _ := metadata.FromMetaContext(ctx)
	data, err := metadata.Marshal(md)
	if err != nil {
		return nil, err
	}
	fp := tc.framer.Encode(&Frame{
		Type:     FrameStartStream,
		StreamId: streamId,
		Data:     packet.Reader(data),
	})
	defer fp.Return()

	tc.mu.Lock()
	defer tc.mu.Unlock()
	if tc.state.Load() == StatusDisconnected {
		return nil, transport_err.ErrCancel
	}

	if err = tc.buf.Set(fp); err != nil {
		return nil, transport_err.NewStreamBufSetErr(err)
	}
	tc.streams.Reg(streamId, stream)
	return stream, nil
}

func (tc *TcpClient) CloseStream(streamId int64) {
	tc.mu.Lock()
	tc.streams.Del(streamId)
	tc.mu.Unlock()
}

func (tc *TcpClient) DialWithOps(ops DialOption) {
	go tc.handleDial(ops)
}

func (tc *TcpClient) handleDial(_ DialOption) {
	defer func() {
		if tc.dHandle != nil {
			tc.dHandle(tc.remoteNodeId)
		}
		tc.buf.OnClose()
	}()

	task := func() error {
		return tc.dial()
	}

	//When the number of failed connection attempts reaches the upper limit,
	//the conn state will be set to the 'disconnected' state,
	//and all virtual streams will be closed.
	if err := withRetry(task); err != nil {
		tc.mu.Lock()
		defer tc.mu.Unlock()
		tc.state.Store(StatusDisconnected)
		tc.streams.Close(func(stream *Stream) {
			stream.OnClose()
		})
		return
	}

	defer tc.handleDisconnected()

	tc.state.Store(StatusConnected)
	tc.lastAck.Store(0)
	tc.sw = NewSender(tc.SendData)
	tc.buf.Run(tc.sw)
	go tc.keepalive()
	<-tc.ctx.Done()
}

func (tc *TcpClient) SendData(data packet.IPacket) error {
	err := tc.sendData(data)
	if err != nil {
		log.Println("[TcpClient] [func: SendData] exec failed: ", err.Error())
		if tc.conn != nil {
			_ = tc.conn.Close()
		}
	}
	return err
}

func (tc *TcpClient) sendData(data packet.IPacket) error {
	body := tc.codec.EncodeBody(data)
	if err := tc.conn.SetWriteDeadline(time.Now().Add(WriteTimeout)); err != nil {
		body.Return()
		return err
	}
	_, err := tc.conn.Write(body.Data())
	body.Return()
	return err
}

func (tc *TcpClient) handleDisconnected() {
	if tc.state.CompareAndSwap(StatusConnected, StatusDisconnected) {
		tc.streams.Close(func(stream *Stream) {
			stream.OnClose()
		})
	}
}

func (tc *TcpClient) dial() error {
	conn, err := net.Dial("tcp", tc.remoteAddr)
	if err != nil {
		log.Println("[TcpClient] dial failed: ", err.Error())
		return err
	}

	tc.conn = conn
	go tc.reader()
	return nil
}

func (tc *TcpClient) reader() {
	header := make([]byte, HeadLen)
	body := make([]byte, tc.codec.maxIncomingSize)

	var (
		err   error
		in    packet.IPacket
		bytes []byte
	)

	defer func() {
		if tc.conn != nil {
			_ = tc.conn.Close()
		}
		if tc.cancel != nil {
			tc.cancel()
		}

		if err != nil {
			if !(err == io.EOF || transport_err.IsClosedConnError(err)) {
				log.Println(fmt.Errorf("tcp %s disconnected: %s", tc.remoteAddr, err.Error()))
			}
		}
	}()

	tc.ack()

	for {
		in, err = tc.recv(header, body)
		if err != nil {
			return
		}

		tc.ack()
		for len(in.Remain()) > 0 {
			bytes, err = in.ReadBytes32()
			if err != nil {
				break
			}
			reader := packet.Reader(bytes)
			_ = tc.decodeRspAndDispatch(reader)
		}
	}
}

func (tc *TcpClient) recv(header []byte, body []byte) (packet.IPacket, error) {
	in, err := tc.codec.BlockDecode(tc.conn, header, body)
	if err != nil {
		return nil, err
	}
	return in, err
}

func (tc *TcpClient) decodeRspAndDispatch(data packet.IPacket) error {
	frame, err := tc.framer.Decode(data)
	if err != nil {
		log.Println("[TcpClient] [func: decodeRspAndDispatch] failed: ", err.Error())
		return err
	}

	if frame.Type == FrameStreamHeartbeat {
		return nil
	}
	tc.HandleData(&frame)
	return nil
}

func (tc *TcpClient) HandleData(in *Frame) {
	switch in.Type {
	case FrameReplyRaw:
		tc.handleReplyRaw(in)
	case FrameCleanStream:
		tc.handleCleanStream(in.StreamId)
	case FrameCliHalfClosedAck:
		tc.handleElegantlyClosedStream(in.StreamId)
	}
}

func (tc *TcpClient) handleReplyRaw(in *Frame) {
	data := in.Data
	if len(data.Remain()) > 0 {
		msg := StreamMsg{}
		msg.buf = packet.Reader(data.Remain())
		if stream, ok := tc.streams.Get(in.StreamId); ok {
			stream.write(msg)
		}
	}
}

func (tc *TcpClient) handleCleanStream(streamId int64) {
	stream, ok := tc.streams.GetAndDel(streamId)
	if ok {
		stream.OnElegantlyClose()
	}
}

// 优雅的关闭stream
func (tc *TcpClient) handleElegantlyClosedStream(streamId int64) {
	stream, ok := tc.streams.GetAndDel(streamId)
	if ok {
		stream.OnElegantlyClose()
	}
}

func (tc *TcpClient) keepalive() {
	ticker := time.NewTicker(time.Second)
	ping := tc.framer.Encode(&Frame{
		Type:     FrameStreamHeartbeat,
		StreamId: 0,
	})

	defer ping.Return()

	prev := time.Now().Unix()
	timeout := time.Duration(0)
	outstandingPing := false

	for {
		select {
		case <-ticker.C:
			la := tc.lastAck.Load()
			if la > prev {
				prev = la
				ticker.Reset(time.Duration(la-time.Now().Unix()) + AckInterval)
				outstandingPing = false
				continue
			}

			if outstandingPing && timeout <= 0 {
				log.Println("[TcpClient] no heartbeat: ", tc.remoteAddr)
				_ = tc.conn.Close()
				return
			}

			if !outstandingPing {
				_ = tc.buf.Set(ping)
				outstandingPing = true
				timeout = PingTimeOut
			}
			sd := number_utils.Min[time.Duration](AckInterval, timeout)
			timeout -= sd
			ticker.Reset(sd)
		case <-tc.ctx.Done():
			return
		}
	}
}

func (tc *TcpClient) ack() {
	tc.lastAck.Store(time.Now().Unix())
}

func (tc *TcpClient) StateCompareAndSwap(old, new uint32) bool {
	return tc.state.CompareAndSwap(old, new)
}

func withRetry(handle func() error) error {
	retried := int32(0)
	for {
		err := handle()
		if err == nil {
			return nil
		}
		time.Sleep(time.Millisecond * time.Duration(100<<retried))
		if retried >= MaxRetried {
			return transport_err.ErrMaxOfRetry
		}
		retried++
	}
}
