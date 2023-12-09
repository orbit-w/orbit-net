package stream_transport

import "sync/atomic"

/*
   @Author: orbit-w
   @File: control
   @2023 11月 周五 15:44
*/

type WriteQuota struct {
	quota int32
	ch    chan struct{}
	done  <-chan struct{}
}

func (wq *WriteQuota) Get(sz int32) error {
	for {
		if atomic.LoadInt32(&wq.quota) > 0 {
			atomic.AddInt32(&wq.quota, -sz)
		}
		select {
		case <-wq.ch:
			continue
		case <-wq.done:
			return ErrStreamQuotaEmpty
		}
	}
}

func (wq *WriteQuota) Replenish(n int) {
	sz := int32(n)
	a := atomic.AddInt32(&wq.quota, sz)
	b := a - sz
	if b <= 0 && a > 0 {
		select {
		case wq.ch <- struct{}{}:
		default:

		}
	}
}
