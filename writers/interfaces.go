package writers

import "time"

type (
	ConfigLognitorInterface interface {
		ConfigHttp
		ConfigGrpc
		ConfigRetry
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

	ConfigRetry interface {
		NoCount() bool
		RetryCount() int
		RetryDelay() time.Duration
	}
)
