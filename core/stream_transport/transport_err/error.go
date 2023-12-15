package transport_err

import (
	"errors"
	"fmt"
	"strings"
)

var (
	ErrRpcDisconnected  = errors.New("error_rpc_disconnected")
	ErrRpcDisconnectedP = "error_rpc_disconnected"
	ErrStreamDone       = errors.New("error_the_stream_is_done")
	ErrMaxOfRetry       = errors.New(`error_max_of_retry`)
	ErrCancel           = errors.New("transport_err code: context canceled")
	ErrStreamShutdown   = errors.New("transport_err code: stream shutdown")
	ErrStreamQuotaEmpty = errors.New("err_stream_quota_empty")
)

func IsErrRpcDisconnected(err error) bool {
	return err != nil && strings.Contains(err.Error(), ErrRpcDisconnectedP)
}

func IsClosedConnError(err error) bool {
	/*
		`use of closed file or network connection` (Go ver > 1.8, internal/pool.ErrClosing)
		`mux: listener closed` (cmux.ErrListenerClosed)
	*/
	return err != nil && strings.Contains(err.Error(), "closed")
}

func IsCancelError(err error) bool {
	return err != nil && strings.Contains(err.Error(), "context canceled")
}

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
