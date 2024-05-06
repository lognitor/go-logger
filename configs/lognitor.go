package configs

import (
	"fmt"
	"net/url"
	"time"
)

type (
	Lognitor struct {
		Http  Http
		Grpc  Grpc
		Retry Retry
		token string
	}

	Grpc struct {
		url     *url.URL
		timeout time.Duration
		enabled bool
	}

	Http struct {
		url     *url.URL
		timeout time.Duration
	}

	Retry struct {
		noCount bool
		count   int
		delay   time.Duration
	}
)

// NewLognitor creates a new Lognitor config
func NewLognitor(grpcHost, httpHost, token string) (*Lognitor, error) {
	u, err := url.Parse(httpHost)
	if err != nil {
		return nil, fmt.Errorf("invalid http host: %s", httpHost)
	}

	grpcUrl, err := url.Parse(grpcHost)
	if err != nil {
		return nil, fmt.Errorf("invalid grpc host: %s", grpcHost)
	}

	return &Lognitor{
		Http: Http{
			url:     u,
			timeout: time.Second * 3,
		},
		Grpc: Grpc{
			url:     grpcUrl,
			timeout: time.Second * 3,
		},
		Retry: Retry{
			count: 2,
			delay: time.Second * 5,
		},
		token: token,
	}, nil
}

// EnableGrpc enables grpc for the lognitor
// Attention if you enable grpc, need recreate the lognitor writer
func (l *Lognitor) EnableGrpc() {
	l.Grpc.enabled = true
}

// SetHttpTimeout sets timeout for the HTTP requests
func (l *Lognitor) SetHttpTimeout(timeout time.Duration) {
	l.Http.timeout = timeout
}

// SetGrpcTimeout sets timeout for the gRPC requests
func (l *Lognitor) SetGrpcTimeout(timeout time.Duration) {
	l.Grpc.timeout = timeout
}

// EnableNoCount set flag for retry while error will be nil
func (l *Lognitor) EnableNoCount() {
	l.Retry.noCount = true
}

// SetRetryCount sets count of retries for each failed log
func (l *Lognitor) SetRetryCount(count int) {
	l.Retry.count = count
}

// SetRetryDelay sets delay between retries
func (l *Lognitor) SetRetryDelay(delay time.Duration) {
	l.Retry.delay = delay
}

// HttpHost returns the host of the lognitor
func (l *Lognitor) HttpHost() string {
	return l.Http.url.String()
}

// HttpTimeout returns the http requests timeout
func (l *Lognitor) HttpTimeout() time.Duration {
	return l.Http.timeout
}

// Token returns the token of the lognitor
func (l *Lognitor) Token() string {
	return l.token
}

// IsGrpc returns if the lognitor is using grpc
func (l *Lognitor) IsGrpc() bool {
	return l.Grpc.enabled
}

// GrpcHost returns the grpc host of the lognitor
func (l *Lognitor) GrpcHost() string {
	return l.Grpc.url.String()
}

// GrpcTimeout returns the grpc requests timeout
func (l *Lognitor) GrpcTimeout() time.Duration {
	return l.Grpc.timeout
}

// NoCount returns value of flag that mean retry while error will be nil
func (l *Lognitor) NoCount() bool {
	return l.Retry.noCount
}

// RetryCount returns retry count for each failed log
func (l *Lognitor) RetryCount() int {
	return l.Retry.count
}

// RetryDelay returns delay between retries
func (l *Lognitor) RetryDelay() time.Duration {
	return l.Retry.delay
}
