package log

import (
	"bufio"
	"os"
	"runtime"
)

func Trace(skip int) []Frame {
	frames := make([]Frame, 0, 10)
	for {
		pc, path, line, ok := runtime.Caller(skip)
		if !ok {
			break
		}
		fn := runtime.FuncForPC(pc)
		frame := Frame{
			Func: fn.Name(),
			Line: line,
			Path: path,
		}
		frames = append(frames, frame)
		skip++
	}

	return frames
}

func getSource(f string, line int) []string {
	file, err := os.Open(f)
	if err != nil {
		return nil
	}
	defer file.Close()

	start := line - 4
	if start < 0 {
		start = 0
	}
	end := line + 2
	scanner := bufio.NewScanner(file)
	lines := make([]string, 0, 7)

	for i := 0; scanner.Scan(); i++ {
		if i >= start && i <= end {
			lines = append(lines, scanner.Text())
		}
	}

	return lines
}
