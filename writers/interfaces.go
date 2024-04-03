package writers

import "time"

type (
	ConfigLognitorInterface interface {
		ConfigHttp
		ConfigGrpc
		Token() string
	}

	ConfigHttp interface {
		HttpHost() string
		HttpTimeout() time.Duration
	}

	ConfigGrpc interface {
		GrpcHost() string
		GrpcTimeout() time.Duration
		IsGrpc() bool
	}
)
