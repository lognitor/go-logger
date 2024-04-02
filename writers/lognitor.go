package writers

import (
	"bytes"
	"context"
	"fmt"
	"github.com/lognitor/entrypoint/pkg/transport/grpc/entrypoint"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"net/http"
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
		in    chan []byte
		host  string
		token string
		http  *http.Client
		grpc  struct {
			conn    *grpc.ClientConn
			client  EntrypointClient
			timeout time.Duration
		}
	}
)

// NewLognitorWriter creates a new io.Writer
// Attention! Creating this io.Writer launches the Worker listening channel.
// Avoid unnecessary creation operations!
func NewLognitorWriter(ctx context.Context, config ConfigLognitorInterface) (*LognitorWriter, error) {
	w := &LognitorWriter{
		in:    make(chan []byte, 1000),
		host:  config.Host(),
		token: config.Token(),
		http: &http.Client{
			Timeout: time.Second * 3,
		},
	}

	if config.IsGrpc() {
		if err := w.initGRPC(ctx, config.GrpcHost()); err != nil {
			return nil, err
		}
	}

	go w.worker()

	return w, nil
}

func (w *LognitorWriter) initGRPC(ctx context.Context, host string) error {
	conn, err := grpc.DialContext(ctx, host, grpc.WithTransportCredentials(insecure.NewCredentials()), grpc.WithBlock())
	if err != nil {
		return fmt.Errorf("failed to connect GRPC: %s", err)
	}

	client := entrypoint.NewEntrypointClient(conn)
	w.grpc.conn = conn
	w.grpc.client = client

	return nil
}

// Write writes log to the channel
func (w *LognitorWriter) Write(p []byte) (n int, err error) {
	go w.writeToChannel(p)

	return len(p), nil
}

// Close closes the channel and connection for destroying the worker
func (w *LognitorWriter) Close() {
	close(w.in)
	if w.grpc.conn != nil {
		w.grpc.conn.Close()
	}
}

func (w *LognitorWriter) writeToChannel(b []byte) {
	w.in <- b
}

func (w *LognitorWriter) worker() {
	for r := range w.in {
		if err := w.sendRequest(r); err != nil {
			continue
		}
	}
}

func (w *LognitorWriter) sendRequest(b []byte) error {
	if w.grpc.client != nil {
		return w.sendGRPC(b)
	}

	return w.sendHTTP(b)
}

func (w *LognitorWriter) sendHTTP(b []byte) error {
	req, err := http.NewRequest(http.MethodPost, w.host, bytes.NewBuffer(b))
	if err != nil {
		return fmt.Errorf("failed to create request: %s", err)
	}

	req.Header.Set("Authorization", w.token)

	resp, err := w.http.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %s", err)
	}
	defer resp.Body.Close()

	return nil
}

func (w *LognitorWriter) sendGRPC(b []byte) error {
	ctx, cancel := context.WithTimeout(context.Background(), w.grpc.timeout)
	defer cancel()

	_, err := w.grpc.client.WriteLogSync(ctx, &entrypoint.PayloadRequest{
		Message: b,
	})

	return err
}
