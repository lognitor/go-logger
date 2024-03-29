package writers

type ConfigLognitorInterface interface {
	Host() string
	GrpcHost() string
	Token() string
	IsGrpc() bool
}
