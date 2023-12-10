package orbit_net

import (
	"context"
	"errors"
	"github.com/orbit-w/orbit-net/core/stream_transport"
	"github.com/orbit-w/orbit-net/core/stream_transport/transport"
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
				if errors.Is(err, transport.ErrCancel) {
					break
				}
			}
		}
	}()

	_ = stream.Send(nil)
	_ = stream.CloseSend()
}
