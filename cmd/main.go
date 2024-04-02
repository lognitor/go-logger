package main

import (
	"context"
	"github.com/lognitor/go-logger/configs"
	"github.com/lognitor/go-logger/logger"
	"github.com/lognitor/go-logger/writers"
	"log"
	"time"
)

func main() {
	cfg, err := configs.NewLognitor("localhost:50051", "localhost:8080", "sometoken")
	if err != nil {
		log.Fatalf("failed to create lognitor config: %v", err)
	}

	writer, err := writers.NewLognitorWriter(context.Background(), cfg)
	if err != nil {
		log.Fatalf("failed to create lognitor writer: %v", err)
	}

	l := logger.New(writer, "test")
	test(l)
}

func test(l *logger.Logger) {
	for i := 0; i < 10; i++ {
		l.Infof("hello there %d", i)
	}

	time.Sleep(time.Second * 10)
}
