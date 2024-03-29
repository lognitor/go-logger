package writers

import (
	"fmt"
	"github.com/lognitor/entrypoint/pkg/transport/grpc/entrypoint"
	"google.golang.org/grpc"
	"net/http"
	"time"
)

// LognitorWriter is a writer that sends logs to lognitor
type LognitorWriter struct {
	in    chan []byte
	host  string
	token string
	http  *http.Client
	grpc  struct {
		conn   *grpc.ClientConn
		client *entrypoint.EntrypointClient
	}
}

// NewLognitorWriter creates a new io.Writer
// Attention! Creating this io.Writer launches the Worker listening channel.
// Avoid unnecessary creation operations!
func NewLognitorWriter(config ConfigLognitorInterface) (*LognitorWriter, error) {
	w := &LognitorWriter{
		in:    make(chan []byte, 1000),
		host:  config.Host(),
		token: config.Token(),
		http: &http.Client{
			Timeout: time.Second * 3,
		},
	}

	if config.IsGrpc() {
		w.initGRPC(config.GrpcHost())
	}

	go w.worker()

	return w, nil
}

func (w *LognitorWriter) initGRPC(host string) error {
	conn, err := grpc.Dial(host)
	if err != nil {
		return fmt.Errorf("failed to connect GRPC: %s", err)
	}

	client := entrypoint.NewEntrypointClient(conn)
	w.grpc.conn = conn
	w.grpc.client = &client

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
	//http.Post(w.host)

	return nil
}
