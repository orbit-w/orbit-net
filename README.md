# orbit-net

## stream_transport:
Streaming mode is a high-level abstraction of the concept of virtual links. The binding relationship between virtual links and physical links is n:1,

The original intention of the design is to reduce signal lock collisions in high concurrency scenarios, and also to reasonably reduce the number of connections and release memory pressure.

## Client
```go

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


```