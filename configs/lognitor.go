package configs

import (
	"fmt"
	"net/url"
	"time"
)

type Lognitor struct {
	httpHost    *url.URL
	httpTimeout time.Duration
	grpcHost    *url.URL
	grpcTimeout time.Duration
	token       string
	grpc        bool
}

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
		httpHost:    u,
		httpTimeout: time.Second * 3,
		token:       token,
		grpcHost:    grpcUrl,
		grpcTimeout: time.Second * 3,
		grpc:        false,
	}, nil
}

// EnableGrpc enables grpc for the lognitor
// Attention if you enable grpc, need recreate the lognitor writer
func (l *Lognitor) EnableGrpc() {
	l.grpc = true
}

// SetHttpTimeout sets timeout for the HTTP requests
func (l *Lognitor) SetHttpTimeout(timeout time.Duration) {
	l.httpTimeout = timeout
}

// SetGrpcTimeout sets timeout for the gRPC requests
func (l *Lognitor) SetGrpcTimeout(timeout time.Duration) {
	l.grpcTimeout = timeout
}

// HttpHost returns the host of the lognitor
func (l *Lognitor) HttpHost() string {
	return l.httpHost.String()
}

// HttpTimeout returns the http requests timeout
func (l *Lognitor) HttpTimeout() time.Duration {
	return l.httpTimeout
}

// Token returns the token of the lognitor
func (l *Lognitor) Token() string {
	return l.token
}

// IsGrpc returns if the lognitor is using grpc
func (l *Lognitor) IsGrpc() bool {
	return l.grpc
}

// GrpcHost returns the grpc host of the lognitor
func (l *Lognitor) GrpcHost() string {
	return l.grpcHost.String()
}

// GrpcTimeout returns the grpc requests timeout
func (l *Lognitor) GrpcTimeout() time.Duration {
	return l.grpcTimeout
}
