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
	"testing"
	"time"
)

func Test_Transport(t *testing.T) {
	host := "127.0.0.1:6800"
	listener, err := net.Listen("tcp", host)
	assert.NoError(t, err)
	log.Println("start serve...")
	server := new(stream_transport.Server)
	server.Serve(listener, func(stream stream_transport.IStreamServer) error {
		for {
			in, err := stream.Recv()
			if err != nil {
				if transport_err.IsClosedConnError(err) {
					break
				}
				log.Println("conn read stream failed: ", err.Error())
				break
			}
			log.Println("receive message from client: ", in.Data()[0])
			if err = stream.Send(in); err != nil {
				log.Println("server response failed: ", err.Error())
			}
		}
		return nil
	})

	conn := stream_transport.DialWithOps(host, "node_0", stream_transport.DialOption{
		CurrentNodeId: "node_1",
	})
	defer func() {
		_ = conn.Close()
	}()

	ctx := context.Background()
	stream, err := conn.NewStream(ctx)
	assert.NoError(t, err)

	go func() {
		for {
			in, err := stream.Recv()
			if err != nil {
				if errors.Is(err, transport_err.ErrCancel) || errors.Is(err, io.EOF) {
					log.Println("Recv failed: ", err.Error())
				} else {
					log.Println("Recv failed: ", err.Error())
				}
				break
			}
			log.Println("recv response: ", in.Data()[0])
		}
	}()

	w := packet.Writer()
	w.Write([]byte{1})
	_ = stream.Send(w)

	time.Sleep(time.Second * 10)
	_ = stream.CloseSend()
	time.Sleep(time.Second * 5)
}
