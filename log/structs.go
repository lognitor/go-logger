package log

type Frame struct {
	Path string `json:"path"`
	Line int    `json:"line"`
	Func string `json:"func"`
}

type FrameWithCode struct {
	Frame
	Code []string `json:"code"`
}
