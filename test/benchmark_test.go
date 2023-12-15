package test

import (
	"context"
	"errors"
	"github.com/orbit-w/golib/bases/packet"
	"github.com/orbit-w/orbit-net/core/stream_transport"
	"github.com/orbit-w/orbit-net/core/stream_transport/transport_err"
	"github.com/stretchr/testify/assert"
	"io"
	"log"
	"net"
	"sync"
	"testing"
	"time"
)

var (
	once = new(sync.Once)
)

func Benchmark_StreamSend_Test(b *testing.B) {
	host := "127.0.0.1:6800"
	Serve(b, host)
	conn := stream_transport.DialWithOps(host, "node_0", stream_transport.DialOption{
		CurrentNodeId: "node_1",
	})
	stream := NewStream(b, conn)
	w := packet.Writer()
	w.Write([]byte{1})
	b.Run("BenchmarkStreamSend", func(b *testing.B) {
		b.ResetTimer()
		b.StartTimer()
		defer b.StopTimer()
		for i := 0; i < b.N; i++ {
			_ = stream.Send(w)
		}
	})

	b.StopTimer()
	time.Sleep(time.Second * 5)
	//_ = conn.Close()
}

func Benchmark_ConcurrencyStreamSend_Test(b *testing.B) {
	host := "127.0.0.1:6800"
	Serve(b, host)

	w := packet.Writer()
	w.Write([]byte{1})
	b.RunParallel(func(pb *testing.PB) {
		conn := stream_transport.DialWithOps(host, "node_0", stream_transport.DialOption{
			CurrentNodeId: "node_1",
		})
		stream := NewStream(b, conn)
		b.ResetTimer()
		b.StartTimer()
		b.ReportAllocs()

		for pb.Next() {
			_ = stream.Send(w)
		}
	})

	b.StopTimer()
	time.Sleep(time.Second * 5)
	//_ = conn.Close()
}

func NewStream(b *testing.B, conn stream_transport.IClientConn) stream_transport.IStreamClient {
	ctx := context.Background()
	stream, err := conn.NewStream(ctx)
	assert.NoError(b, err)
	go func() {
		for {
			_, err := stream.Recv()
			if err != nil {
				if errors.Is(err, transport_err.ErrCancel) || errors.Is(err, io.EOF) {
				} else {
					log.Println("Recv failed: ", err.Error())
				}
				break
			}
		}
	}()
	return stream
}

func Serve(b *testing.B, host string) {
	once.Do(func() {
		listener, err := net.Listen("tcp", host)
		assert.NoError(b, err)
		server := new(stream_transport.Server)
		server.Serve(listener, func(stream stream_transport.IStreamServer) error {
			for {
				_, err := stream.Recv()
				if err != nil {
					if transport_err.IsClosedConnError(err) || transport_err.IsCancelError(err) {
						break
					}
					log.Println("conn read stream failed: ", err.Error())
					break
				}
			}
			return nil
		})
	})
}
