package transport

import (
	"github.com/alphadose/zenq/v2"
	"github.com/orbit-w/golib/bases/packet"
	"log"
	"runtime/debug"
)

/*
   @Author: orbit-w
   @File: sender
   @2023 11月 周日 19:52
*/

type SenderWrapper struct {
	sender func(body packet.IPacket) error
	zq     *zenq.ZenQ[sendParams]
}

type sendParams struct {
	buf packet.IPacket
}

func NewSender(code int8, sender func(body packet.IPacket) error) *SenderWrapper {
	ins := &SenderWrapper{
		sender: sender,
		zq:     zenq.New[sendParams](2048),
	}

	go func() {
		defer func() {
			if x := recover(); x != nil {
				debug.PrintStack()
			}
		}()

		for {
			msg, open := ins.zq.Read()
			if !open {
				log.Println("sender break")
				break
			}
			_ = ins.sender(msg.buf)
			log.Printf("%v recv message", code)
		}
	}()

	return ins
}

func (ins *SenderWrapper) Send(data packet.IPacket) (writeClose bool) {
	writeClose = ins.zq.Write(sendParams{data})
	return
}

func (ins *SenderWrapper) OnClose() {
	ins.zq.Close()
}
