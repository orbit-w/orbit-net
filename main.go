package orbit_net

import (
	"context"
	"errors"
	"github.com/orbit-w/orbit-net/core/stream_transport"
	"github.com/orbit-w/orbit-net/core/stream_transport/transport_err"
	"io"
	"log"
	"net"
)

func StreamTransportClient() {
	conn := stream_transport.DialWithOps("remote_addr", "", stream_transport.DialOption{
		CurrentNodeId: "node_1",
	})
	defer func() {
		_ = conn.Close()
	}()

	ctx := context.Background()
	stream, err := conn.NewStream(ctx)
	if err != nil {
		panic(err.Error())
	}

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

	_ = stream.Send(nil)
	_ = stream.CloseSend()
}

func StreamTransportServer(host string) {
	listener, err := net.Listen("tcp", host)
	if err != nil {
		panic(err.Error())
	}

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
			in.Return()
		}
		return nil
	})
}
