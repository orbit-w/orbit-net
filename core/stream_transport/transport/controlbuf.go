package transport

import (
	"github.com/orbit-w/golib/bases/misc/number_utils"
	"github.com/orbit-w/golib/bases/packet"
	"github.com/orbit-w/orbit-net/core/stream_transport/transport_err"
	"log"
	"sync"
)

/*
   @Author: orbit-w
   @File: control_buf
   @2023 11月 周日 17:21
*/

type ControlBuffer struct {
	consumerWaiting bool
	state           int8
	max             uint32
	length          int
	buffer          packet.IPacket
	mu              sync.Mutex
	sw              *SenderWrapper

	ch    chan struct{}
	close chan struct{}
}

func NewControlBuffer(max uint32, _sw *SenderWrapper) *ControlBuffer {
	ins := &ControlBuffer{
		state:           TypeWorking,
		consumerWaiting: false,
		max:             max,
		buffer:          packet.New(),
		mu:              sync.Mutex{},
		ch:              make(chan struct{}, 1),
		close:           make(chan struct{}, 1),
		sw:              _sw,
	}
	go ins.flush()
	ins.Kick()
	return ins
}

func BuildControlBuffer(buf *ControlBuffer, max uint32) {
	buf.max = max
	buf.ch = make(chan struct{}, 1)
	buf.buffer = packet.NewWithInitialSize(2048)
	buf.mu = sync.Mutex{}
}

func (ins *ControlBuffer) Run(_sw *SenderWrapper) {
	ins.mu.Lock()
	if ins.state == TypeStopped {
		ins.mu.Unlock()
		return
	}
	ins.sw = _sw
	ins.close = make(chan struct{}, 1)
	ins.state = TypeWorking
	ins.mu.Unlock()

	go ins.flush()
	ins.Kick()
}

func (ins *ControlBuffer) Kick() {
	var kick bool
	if ins.state == TypeWorking && ins.consumerWaiting {
		kick = true
		ins.consumerWaiting = false
	}
	if kick {
		select {
		case ins.ch <- struct{}{}:
		default:

		}
	}
}

func (ins *ControlBuffer) Set(buf packet.IPacket) error {
	ins.mu.Lock()
	if ins.state == TypeStopped {
		ins.mu.Unlock()
		return transport_err.ErrRpcDisconnected
	}
	var kick bool
	ins.length++
	d := buf.Data()
	ins.buffer.WriteBytes32(d)
	if ins.consumerWaiting {
		kick = true
		ins.consumerWaiting = false
	}
	ins.mu.Unlock()
	if kick {
		select {
		case ins.ch <- struct{}{}:
		default:
		}
	}
	return nil
}

// OnClose TODO: Is it reasonable to reject stream input immediately?
func (ins *ControlBuffer) OnClose() {
	ins.mu.Lock()
	defer ins.mu.Unlock()
	ins.state = TypeStopped
	close(ins.close)
}

func (ins *ControlBuffer) flush() {
	defer func() {
		if x := recover(); x != nil {
		}
		ins.safeReturn()
	}()

FLUSH:
	ins.mu.Lock()
	for ins.state == TypeWorking && !ins.isEmpty() {
		w := packet.Writer()
		size := number_utils.Min[int](BatchLimit, ins.length)
		for i := 0; i < size; i++ {
			length, _ := ins.buffer.NextBytesSize32()
			if uint32(w.Len())+uint32(length)+4 > ins.max {
				break
			}
			ins.length--
			data, _ := ins.buffer.ReadBytes32()
			w.WriteBytes32(data)
		}

		if ins.sw.Send(w) {
			log.Println("sw send failed")
		} else {
			log.Println("sw send ...")
		}
	}

	ins.consumerWaiting = true
	ins.buffer.Reset()
	ins.mu.Unlock()
	select {
	case <-ins.ch:
		goto FLUSH
	case <-ins.close:
		return
	}
}

func (ins *ControlBuffer) safeReturn() {
	ins.mu.Lock()
	defer ins.mu.Unlock()
	ins.state = TypeStopped
	ins.sw.OnClose()
	ins.buffer.Reset()
	close(ins.ch)
	ins.buffer.Return()
	ins.buffer = nil
}

func (ins *ControlBuffer) isEmpty() bool {
	return ins.length == 0
}
