package log

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/fatih/color"
	"io"
	"os"
	"runtime"
	"sync"
	"sync/atomic"
	"time"
)

const version = "go-sdk/v0.0.1"

type Logger struct {
	token      string
	prefix     string
	level      uint32
	writer     io.Writer
	levels     []string
	bufferPool sync.Pool
	mutex      sync.Mutex
	source     bool
}

type (
	Lvl  uint8
	JSON map[string]any
)

const (
	DEBUG Lvl = iota + 1
	INFO
	WARN
	ERROR
	OFF
	panicLevel
	fatalLevel
)

func New(writer io.Writer, prefix string) *Logger {
	l := &Logger{
		level:  uint32(INFO),
		prefix: prefix,
		writer: writer,
		bufferPool: sync.Pool{
			New: func() interface{} {
				return bytes.NewBuffer(make([]byte, 2048))
			},
		},
	}
	l.initLevels()
	return l
}

func (l *Logger) initLevels() {
	blue := color.New(color.FgBlue).SprintFunc()
	green := color.New(color.FgGreen).SprintFunc()
	yellow := color.New(color.FgYellow).SprintFunc()
	red := color.New(color.FgRed).SprintFunc()

	redBg := color.New(color.BgRed).SprintFunc()
	yellowBg := color.New(color.BgYellow).SprintFunc()

	l.levels = []string{
		"-",
		blue("DEBUG"),
		green("INFO"),
		yellow("WARN"),
		red("ERROR"),
		"",
		yellowBg("PANIC"),
		redBg("FATAL"),
	}
}

// Level returns the current log level.
func (l *Logger) Level() Lvl {
	return Lvl(atomic.LoadUint32(&l.level))
}

// SetLevel sets the log level.
func (l *Logger) SetLevel(level Lvl) {
	atomic.StoreUint32(&l.level, uint32(level))
}

// Prefix returns the current log prefix.
func (l *Logger) Prefix() string {
	return l.prefix
}

// SetPrefix sets the log prefix.
func (l *Logger) SetPrefix(p string) {
	l.prefix = p
}

// Writer returns the current writer.
func (l *Logger) Writer() io.Writer {
	return l.writer
}

// SetOutput sets the output destination for the logger.
// TODO: Add support for disable colorable.Writer
func (l *Logger) SetOutput(w io.Writer) {
	l.writer = w
	//if w, ok := w.(*os.File); !ok || !isatty.IsTerminal(w.Fd()) {
	//	l.DisableColor()
	//}
}

// Print calls l.Output to print to the logger.
func (l *Logger) Print(i ...interface{}) {
	l.log(0, "", i...)
}

// Printf calls l.Output to print to the logger with a format.
func (l *Logger) Printf(format string, args ...interface{}) {
	l.log(0, format, args...)
}

// Debug calls l.Output to print to the logger.
func (l *Logger) Debug(i ...interface{}) {
	l.log(DEBUG, "", i...)
}

// Debugf calls l.Output to print to the logger with a format.
func (l *Logger) Debugf(format string, args ...interface{}) {
	l.log(DEBUG, format, args...)
}

// Info calls l.Output to print to the logger info level.
func (l *Logger) Info(i ...interface{}) {
	l.log(INFO, "", i...)
}

// Infof calls l.Output to print to the logger info level with a format.
func (l *Logger) Infof(format string, args ...interface{}) {
	l.log(INFO, format, args...)
}

// Warn calls l.Output to print to the logger warn level.
func (l *Logger) Warn(i ...interface{}) {
	l.log(WARN, "", i...)
}

// Warnf calls l.Output to print to the logger warn level with a format.
func (l *Logger) Warnf(format string, args ...interface{}) {
	l.log(WARN, format, args...)
}

// Error calls l.Output to print to the logger error level.
func (l *Logger) Error(i ...interface{}) {
	l.log(ERROR, "", i...)
}

// Errorf calls l.Output to print to the logger error level with a format.
func (l *Logger) Errorf(format string, args ...interface{}) {
	l.log(ERROR, format, args...)
}

// Fatal calls l.Output to print to the logger fatal level.
func (l *Logger) Fatal(i ...interface{}) {
	l.log(fatalLevel, "", i...)
	os.Exit(1)
}

// Fatalf calls l.Output to print to the logger fatal level with a format.
func (l *Logger) Fatalf(format string, args ...interface{}) {
	l.log(fatalLevel, format, args...)
	os.Exit(1)
}

// Panic calls l.Output to print to the logger panic level and then panic.
func (l *Logger) Panic(i ...interface{}) {
	l.log(panicLevel, "", i...)
	panic(fmt.Sprint(i...))
}

// Panicf calls l.Output to print to the logger panic level with a format and then panic.
func (l *Logger) Panicf(format string, args ...interface{}) {
	l.log(panicLevel, format, args...)
	panic(fmt.Sprintf(format, args...))
}

func (l *Logger) log(level Lvl, format string, args ...interface{}) {
	if level < l.Level() {
		return
	}
	message := ""

	out := struct {
		Time    time.Time     `json:"time"`
		Level   string        `json:"level"`
		Prefix  string        `json:"prefix"`
		Message any           `json:"message"`
		Agent   string        `json:"agent"`
		Trace   []Frame       `json:"trace"`
		Source  FrameWithCode `json:"source"`
	}{}

	buf := l.bufferPool.Get().(*bytes.Buffer)
	buf.Reset()
	defer l.bufferPool.Put(buf)

	pc, file, line, _ := runtime.Caller(2)
	funcName := runtime.FuncForPC(pc).Name()
	source := getSource(file, line)
	t := Trace(2)

	switch format {
	case "":
		message = fmt.Sprint(args...)
		break
	default:
		message = fmt.Sprintf(format, args...)
	}

	out.Time = time.Now()
	out.Level = l.levels[level]
	out.Prefix = l.prefix
	out.Message = message
	out.Agent = version
	out.Trace = t
	out.Source.Path = file
	out.Source.Line = line
	out.Source.Func = funcName
	out.Source.Code = source

	b, _ := json.Marshal(out)

	buf.WriteString(fmt.Sprintf("[%s] ", l.prefix))
	buf.WriteString(l.levels[level])
	buf.WriteString("\t")
	buf.WriteString(string(b))

	buf.WriteByte('\n')
	l.mutex.Lock()
	defer l.mutex.Unlock()
	l.writer.Write(buf.Bytes())
}
