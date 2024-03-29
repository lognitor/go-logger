package logger

import "time"

type Log struct {
	Time    time.Time     `json:"time"`
	Level   string        `json:"level"`
	Prefix  string        `json:"prefix"`
	Message any           `json:"message"`
	Agent   string        `json:"agent"`
	Trace   []Frame       `json:"trace"`
	Source  FrameWithCode `json:"source"`
}

type Frame struct {
	Path string `json:"path"`
	Line int    `json:"line"`
	Func string `json:"func"`
}

type FrameWithCode struct {
	Frame
	Code []string `json:"code"`
}
