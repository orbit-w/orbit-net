package transport

import "sync"

/*
   @Author: orbit-w
   @File: streams
   @2023 11月 周日 16:08
*/

type Streams struct {
	rw      sync.RWMutex
	streams map[int64]*Stream
}

func NewStreams() *Streams {
	return &Streams{
		rw:      sync.RWMutex{},
		streams: make(map[int64]*Stream, 1<<3),
	}
}

func (ins *Streams) Get(id int64) (*Stream, bool) {
	ins.rw.RLock()
	s, ok := ins.streams[id]
	ins.rw.RUnlock()
	return s, ok
}

func (ins *Streams) Exist(id int64) (exist bool) {
	ins.rw.RLock()
	_, exist = ins.streams[id]
	ins.rw.RUnlock()
	return
}

func (ins *Streams) Len() int {
	return len(ins.streams)
}

func (ins *Streams) Reg(id int64, s *Stream) {
	ins.rw.Lock()
	ins.streams[id] = s
	ins.rw.Unlock()
}

func (ins *Streams) Del(id int64) {
	ins.rw.Lock()
	delete(ins.streams, id)
	ins.rw.Unlock()
}

func (ins *Streams) GetAndDel(id int64) (*Stream, bool) {
	ins.rw.Lock()
	s, exist := ins.streams[id]
	if exist {
		delete(ins.streams, id)
	}
	ins.rw.Unlock()
	return s, exist
}

func (ins *Streams) Range(iter func(stream *Stream)) {
	ins.rw.RLock()
	for k := range ins.streams {
		stream := ins.streams[k]
		iter(stream)
	}
	ins.rw.RUnlock()
}

func (ins *Streams) Close(onClose func(stream *Stream)) {
	ins.rw.Lock()
	defer ins.rw.Unlock()
	for k := range ins.streams {
		stream := ins.streams[k]
		onClose(stream)
	}
	ins.streams = make(map[int64]*Stream)
}
