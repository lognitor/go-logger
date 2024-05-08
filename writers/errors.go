package writers

import "errors"

var (
	ErrWriterIsClosed      = errors.New("lognitor writer is closed")
	ErrRetryWorkerIsClosed = errors.New("retry for lognitor writer is closed")
)
