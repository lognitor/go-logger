package main

import (
	"context"
	"fmt"
	"github.com/lognitor/go-logger/configs"
	"github.com/lognitor/go-logger/logger"
	"github.com/lognitor/go-logger/writers"
	"io"
	"log"
	"os"
	"time"
)

type user struct {
	Id   string `json:"id"`
	Body string `json:"body"`
	File []byte `json:"file"`
}

func main() {
	cfg, err := configs.NewLognitor("local.entrypoint.lognitor.io:4443", "https://local.entrypoint.lognitor.io", "sometoken")
	if err != nil {
		log.Fatalf("failed to create lognitor config: %v", err)
	}
	cfg.EnableGrpc()
	cfg.SetGrpcTimeout(time.Second * 10)
	cfg.SetRetryCount(2)
	cfg.SetRetryDelay(time.Millisecond * 300)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	writer, err := writers.NewLognitorWriter(ctx, cfg)
	if err != nil {
		log.Fatalf("failed to create lognitor writer: %v", err)
	}

	l := logger.New(writer, "test")
	defer func() {
		if err = l.Close(); err != nil {
			log.Fatalf("failed to close logger: %s", err)
		}
	}()

	for i := 0; i <= 5; i++ {
		go test(l)
	}

	<-ctx.Done()
}

func test(l *logger.Logger) {
	for i := 0; i < 3; i++ {
		f, err := os.Open("cmd/examples/base/1.txt")
		if err != nil {
			log.Fatalf("failed to open file: %s", err)
		}

		data, err := io.ReadAll(f)
		if err != nil {
			log.Fatalf("failed to read file")
		}

		u := user{
			Id:   fmt.Sprintf("%d", i),
			Body: fmt.Sprintf("body %d", i),
			File: data,
		}

		l.Info(u)
	}
}
