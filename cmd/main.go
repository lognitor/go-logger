package main

import (
	"github.com/lognitor/go-logger/logger"
	"os"
)

func main() {
	logger := logger.New(os.Stdout, "test")

	logger.Info("hi there")
	logger.Info("hi there 2")
}

func test() {

}
