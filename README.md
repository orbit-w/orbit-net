# orbit-net

## StreamTransport
StreamTransport is a multiplexing transport library for Golang
StreamTransport implements multiple virtual network connections to reuse one physical network link. 

The StreamTransport transport layer relies on TCP underlying connections to provide reliability and ordering 
and to provide stream-oriented multiplexing.

## Advantage
    * High concurrency: Reducing lock collisions and blocking situations
    * Asynchronous: Send messages asynchronously without blocking the caller
    * Batch: Transfer messages in batches
    * Efficient and safe

## Benchmark

```
[11:33:59] [master ✖] ❱❱❱ go test -v -run=^$ -bench .
goos: darwin
goarch: arm64
pkg: github.com/orbit-w/orbit-net/test
Benchmark_StreamSend_Test
Benchmark_StreamSend_Test/BenchmarkStreamSend
Benchmark_StreamSend_Test/BenchmarkStreamSend-8         	15436710	        77.77 ns/op
Benchmark_ConcurrencyStreamSend_Test
Benchmark_ConcurrencyStreamSend_Test-8                  	16750544	        62.02 ns/op	     147 B/op	       1 allocs/op
PASS
ok  	github.com/orbit-w/orbit-net/test	33.839s
```

## Client
```go
package main

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

## Server
```go
package main

import (
    "context"
    "errors"
    "github.com/orbit-w/orbit-net/core/stream_transport"
    "github.com/orbit-w/orbit-net/core/stream_transport/transport_err"
    "io"
    "log"
    "net"
)

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
```