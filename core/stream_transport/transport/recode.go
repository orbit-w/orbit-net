package transport

import (
	"errors"
	"fmt"
)

/*
   @Author: orbit-w
   @File: recode
   @2023 11月 周日 14:19
*/

var (
	ErrCancel         = errors.New("error code: context canceled")
	ErrStreamShutdown = errors.New("error code: stream shutdown")
)

func ExceedMaxIncomingPacket(size uint32) error {
	return errors.New(fmt.Sprintf("exceed max incoming packet size: %d", size))
}

func ReadBodyFailed(err error) error {
	return errors.New(fmt.Sprintf("read body failed: %s", err.Error()))
}

func ReceiveBufPutErr(err error) error {
	return errors.New(fmt.Sprintf("receiveBuf put failed: %s", err.Error()))
}

func NewStreamBufSetErr(err error) error {
	return errors.New(fmt.Sprintf("NewStream set failed: %s", err.Error()))
}
