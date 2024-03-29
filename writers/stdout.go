package writers

import (
	"github.com/mattn/go-colorable"
	"io"
	"os"
)

func GetColorStdout() io.Writer {
	return colorable.NewColorableStdout()
}

func GetDefaultStdout() io.Writer {
	return os.Stdout
}
