package configs

import (
	"fmt"
	"net/url"
)

type Lognitor struct {
	host     *url.URL
	grpcHost *url.URL
	token    string
	grpc     bool
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
		host:     u,
		token:    token,
		grpcHost: grpcUrl,
		grpc:     false,
	}, nil
}

// EnableGrpc enables grpc for the lognitor
// Attention if you enable grpc, need recreate the lognitor writer
func (l *Lognitor) EnableGrpc() {
	l.grpc = true
}

// Host returns the host of the lognitor
func (l *Lognitor) Host() string {
	return l.host.String()
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
