package configs

import (
	"fmt"
	"net/url"
	"time"
)

type (
	Lognitor struct {
		Http  http
		Grpc  grpc
		Retry retry
		token string
	}

	grpc struct {
		url     *url.URL
		timeout time.Duration
		enabled bool
	}

	http struct {
		url     *url.URL
		timeout time.Duration
	}

	retry struct {
		delay time.Duration
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
		Http: http{
			url:     u,
			timeout: time.Second * 3,
		},
		Grpc: grpc{
			url:     grpcUrl,
			timeout: time.Second * 3,
		},
		Retry: retry{
			delay: time.Second * 2,
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

// RetryDelay returns delay between retries
func (l *Lognitor) RetryDelay() time.Duration {
	return l.Retry.delay
}
