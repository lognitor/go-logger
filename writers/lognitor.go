package writers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/lognitor/entrypoint/pkg/transport/grpc/entrypoint"
	"github.com/lognitor/go-logger/logger"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"net/http"
	"sync"
	"time"
)

// LognitorWriter is a writer that sends logs to lognitor
type (
	// EntrypointClient interface of Lognitor gRPC api client
	EntrypointClient interface {
		WriteLogSync(ctx context.Context, in *entrypoint.PayloadRequest, opts ...grpc.CallOption) (*entrypoint.PayloadReply, error)
		WriteLogAsync(ctx context.Context, in *entrypoint.PayloadRequest, opts ...grpc.CallOption) (*entrypoint.PayloadReply, error)
	}

	// LognitorWriter writer with meta information
	LognitorWriter struct {
		// ctx instance of context for writer
		ctx context.Context
		// cancel is the cancel function for ctx
		cancel context.CancelCauseFunc
		// wait group for added logs
		wg sync.WaitGroup
		// in is input from the logger
		in chan []byte
		// token is the user token
		token string
		// retry contains chan for receive errors
		retry retry
		// http represents meta information for http requests
		http httpInfo
		// grpc represents meta information for grpc requests
		grpc grpcInfo
	}

	retry struct {
		ctx     context.Context
		cancel  context.CancelCauseFunc
		wg      sync.WaitGroup
		in      chan []byte
		noCount bool
		count   int
		delay   time.Duration
	}

	httpInfo struct {
		host   string
		client *http.Client
	}

	grpcInfo struct {
		conn    *grpc.ClientConn
		client  EntrypointClient
		timeout time.Duration
	}
)

// NewLognitorWriter creates a new io.Writer
// Attention! Creating this io.Writer launches the Worker listening channel.
// Avoid unnecessary creation operations!
func NewLognitorWriter(ctx context.Context, config ConfigLognitorInterface) (*LognitorWriter, error) {
	w := &LognitorWriter{
		in:    make(chan []byte, 1000),
		token: config.Token(),
		retry: retry{
			in:      make(chan []byte, 100),
			noCount: config.NoCount(),
			count:   config.RetryCount(),
			delay:   config.RetryDelay(),
		},
	}

	w.ctx, w.cancel = context.WithCancelCause(context.Background())
	w.retry.ctx, w.retry.cancel = context.WithCancelCause(context.Background())

	if config.IsGrpc() {
		if err := w.initGRPC(ctx, config.GrpcHost(), config.GrpcTimeout()); err != nil {
			return nil, err
		}
	}

	go w.worker()
	go w.retryWorker()

	return w, nil
}

func (w *LognitorWriter) initGRPC(ctx context.Context, host string, timeout time.Duration) error {
	conn, err := grpc.DialContext(
		ctx,
		host,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock(),
	)
	if err != nil {
		return fmt.Errorf("failed to connect GRPC: %s", err)
	}

	client := entrypoint.NewEntrypointClient(conn)
	w.grpc.conn = conn
	w.grpc.client = client
	w.grpc.timeout = timeout

	return nil
}

// Write writes log to the channel
func (w *LognitorWriter) Write(p []byte) (n int, err error) {
	// checking on the canceled context or any other errors
	if err = context.Cause(w.ctx); err != nil {
		return
	}

	w.wg.Add(1)
	go w.writeToChannel(p)

	return len(p), nil
}

// Close closes the channel and connection for destroying the worker
// also wait until current logs in channel will be done
func (w *LognitorWriter) Close() error {
	// checking on the canceled context or any other errors
	if err := context.Cause(w.ctx); err != nil {
		return err
	}

	w.wg.Wait()
	w.retry.wg.Wait()
	//close channel for incoming logs
	close(w.in)
	//wait until remaining logs will be done
	<-w.ctx.Done()
	//close retry channel
	close(w.retry.in)
	//wait until retries done
	<-w.retry.ctx.Done()

	if w.grpc.conn != nil {
		if err := w.grpc.conn.Close(); err != nil {
			return fmt.Errorf("failed to close gRPC connection: %w", err)
		}
	}

	return nil
}

func (w *LognitorWriter) writeToChannel(b []byte) {
	defer w.wg.Done()

	w.in <- b
}

func (w *LognitorWriter) writeToRetry(b []byte) {
	defer w.wg.Done()

	w.retry.in <- b
}

func (w *LognitorWriter) worker() {
	for r := range w.in {
		if err := w.sendRequest(r); err != nil {
			w.wg.Add(1)
			go w.writeToRetry(r)
		}
	}

	w.cancel(ErrWriterIsClosed)
}

func (w *LognitorWriter) retryWorker() {
	for r := range w.retry.in {
		w.retry.wg.Add(1)
		go w.sendWithRetry(r)
	}

	w.retry.cancel(ErrRetryWorkerIsClosed)
}

func (w *LognitorWriter) sendRequest(b []byte) error {
	if w.grpc.client != nil {
		return w.sendGRPC(b)
	}

	return w.sendHTTP(b)
}

func (w *LognitorWriter) sendWithRetry(b []byte) {
	defer w.retry.wg.Done()

	if w.retry.noCount {
		err := w.sendRequest(b)

		for err != nil {
			err = w.sendRequest(b)
			<-time.After(w.retry.delay)
		}
	} else {
		for try := 0; try < w.retry.count; try++ {
			if err := w.sendRequest(b); err != nil {
				<-time.After(w.retry.delay)
				continue
			}

			return
		}
	}
}

func (w *LognitorWriter) sendHTTP(b []byte) error {
	req, err := http.NewRequest(http.MethodPost, w.http.host, bytes.NewBuffer(b))
	if err != nil {
		return fmt.Errorf("failed to create request: %s", err)
	}

	req.Header.Set("X-TOKEN", w.token)

	resp, err := w.http.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %s", err)
	}
	defer resp.Body.Close()

	return nil
}

func (w *LognitorWriter) sendGRPC(b []byte) error {
	log := new(logger.Log)

	if err := json.Unmarshal(b, log); err != nil {
		return fmt.Errorf("failed to unmarshal log: %s", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), w.grpc.timeout)
	defer cancel()

	req := &entrypoint.PayloadRequest{
		Level:   log.Level,
		Prefix:  log.Prefix,
		Message: log.Message,
		Token:   w.token,
		Trace:   make([]*entrypoint.Frame, 0, len(log.Trace)),
		Source: &entrypoint.FrameWithCode{
			Path: log.Source.Path,
			Line: uint32(log.Source.Line),
			Func: log.Source.Func,
			Code: log.Source.Code,
		},
	}

	for _, trace := range log.Trace {
		req.Trace = append(req.Trace, &entrypoint.Frame{
			Path: trace.Path,
			Line: uint32(trace.Line),
			Func: trace.Func,
		})
	}

	if _, err := w.grpc.client.WriteLogSync(ctx, req); err != nil {
		return err
	}

	return nil
}
