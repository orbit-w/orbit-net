package transport

import (
	"github.com/orbit-w/golib/bases/packet"
	"github.com/orbit-w/orbit-net/core/unbounded"
)

/*
   @Author: orbit-w
   @File: sender
   @2023 11月 周日 19:52
*/

type SenderWrapper struct {
	sender  func(body packet.IPacket) error
	channel IUnboundedChan[sendParams]
}

type sendParams struct {
	buf packet.IPacket
}

func NewSender(sender func(body packet.IPacket) error) *SenderWrapper {
	ins := &SenderWrapper{
		sender:  sender,
		channel: unbounded.New[sendParams](64),
	}

	go func() {
		defer func() {
			if x := recover(); x != nil {

			}
		}()

		ins.channel.Receive(func(msg sendParams) bool {
			_ = ins.sender(msg.buf)
			return false
		})
	}()

	return ins
}

func (ins *SenderWrapper) Send(data packet.IPacket) error {
	return ins.channel.Send(sendParams{data})
}

func (ins *SenderWrapper) OnClose() {
	ins.channel.Close()
}
