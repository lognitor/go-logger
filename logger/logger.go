package logger

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"runtime"
	"sync"
	"sync/atomic"
	"time"
)

const version = "go-sdk/v0.0.1"

// A Logger represents an active logging object that generates json lines of
// output to an [io.Writer]. Each logging operation makes a single call to
// the Writer's Write method. A Logger can be used simultaneously from
// multiple goroutines; it guarantees to serialize access to the Writer.
// TODO: Support not json output
type Logger struct {
	prefix     string
	level      uint32
	writer     io.WriteCloser
	levels     []string
	bufferPool sync.Pool
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

// New create new logger instance
func New(writer io.WriteCloser, prefix string) *Logger {
	l := &Logger{
		level:  uint32(INFO),
		prefix: prefix,
		writer: writer,
		bufferPool: sync.Pool{
			New: func() interface{} {
				return bytes.NewBuffer(make([]byte, 0, 2048))
			},
		},
	}
	l.initLevels()
	return l
}

func (l *Logger) Close() error {
	return l.writer.Close()
}

func (l *Logger) initLevels() {
	//blue := color.New(color.FgBlue).SprintFunc()
	//green := color.New(color.FgGreen).SprintFunc()
	//yellow := color.New(color.FgYellow).SprintFunc()
	//red := color.New(color.FgRed).SprintFunc()
	//
	//redBg := color.New(color.BgRed).SprintFunc()
	//yellowBg := color.New(color.BgYellow).SprintFunc()

	l.levels = []string{
		"-",
		"DEBUG",
		"INFO",
		"WARN",
		"ERROR",
		"",
		"PANIC",
		"FATAL",
	}
}

// Level returns the current logger level.
func (l *Logger) Level() Lvl {
	return Lvl(atomic.LoadUint32(&l.level))
}

// SetLevel sets the logger level.
func (l *Logger) SetLevel(level Lvl) {
	atomic.StoreUint32(&l.level, uint32(level))
}

// Prefix returns the current logger prefix.
func (l *Logger) Prefix() string {
	return l.prefix
}

// SetPrefix sets the logger prefix.
func (l *Logger) SetPrefix(p string) {
	l.prefix = p
}

// Writer returns the current writer.
func (l *Logger) Writer() io.Writer {
	return l.writer
}

// SetOutput sets the output destination for the logger.
func (l *Logger) SetOutput(w io.WriteCloser) {
	l.writer = w
	//if w, ok := w.(*os.File); !ok || !isatty.IsTerminal(w.Fd()) {
	//	l.DisableColor()
	//}
}

// Print calls l.Output to print to the logger.
func (l *Logger) Print(i ...any) {
	l.log(0, i...)
}

// Printf calls l.Output to print to the logger with a format.
func (l *Logger) Printf(format string, args ...any) {
	l.logf(0, format, args...)
}

// Debug calls l.Output to print to the logger.
func (l *Logger) Debug(i ...any) {
	l.log(DEBUG, i...)
}

// Debugf calls l.Output to print to the logger with a format.
func (l *Logger) Debugf(format string, args ...any) {
	l.logf(DEBUG, format, args...)
}

// Info calls l.Output to print to the logger info level.
func (l *Logger) Info(i ...any) {
	l.log(INFO, i...)
}

// Infof calls l.Output to print to the logger info level with a format.
func (l *Logger) Infof(format string, args ...any) {
	l.logf(INFO, format, args...)
}

// Warn calls l.Output to print to the logger warn level.
func (l *Logger) Warn(i ...any) {
	l.log(WARN, i...)
}

// Warnf calls l.Output to print to the logger warn level with a format.
func (l *Logger) Warnf(format string, args ...any) {
	l.logf(WARN, format, args...)
}

// Error calls l.Output to print to the logger error level.
func (l *Logger) Error(i ...any) {
	l.log(ERROR, i...)
}

// Errorf calls l.Output to print to the logger error level with a format.
func (l *Logger) Errorf(format string, args ...any) {
	l.logf(ERROR, format, args...)
}

// Fatal calls l.Output to print to the logger fatal level.
func (l *Logger) Fatal(i ...any) {
	l.log(fatalLevel, i...)
	os.Exit(1)
}

// Fatalf calls l.Output to print to the logger fatal level with a format.
func (l *Logger) Fatalf(format string, args ...any) {
	l.logf(fatalLevel, format, args...)
	os.Exit(1)
}

// Panic calls l.Output to print to the logger panic level and then panic.
func (l *Logger) Panic(i ...any) {
	l.log(panicLevel, i...)
	panic(fmt.Sprint(i...))
}

// Panicf calls l.Output to print to the logger panic level with a format and then panic.
func (l *Logger) Panicf(format string, args ...any) {
	l.logf(panicLevel, format, args...)
	panic(fmt.Sprintf(format, args...))
}

func (l *Logger) log(level Lvl, d ...any) {
	var (
		msg []byte
		err error
	)

	if len(d) == 1 {
		for _, v := range d {
			if str, ok := v.(string); ok {
				msg = []byte(str)
				break
			}

			if msg, err = json.Marshal(v); err != nil {
				return
			}
		}
	}

	if len(d) > 1 {
		msg, err = json.Marshal(d)
		if err != nil {
			return
		}
	}

	l.logf(level, string(msg))
}

func (l *Logger) logf(level Lvl, format string, args ...any) {
	if level < l.Level() {
		return
	}
	message := ""

	buf := l.bufferPool.Get().(*bytes.Buffer)
	defer func() {
		buf.Reset()
		l.bufferPool.Put(buf)
	}()

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

	log := Log{
		Time:    time.Now(),
		Level:   l.levels[level],
		Prefix:  l.prefix,
		Message: message,
		Agent:   version,
		Trace:   t,
		Source: FrameWithCode{
			Frame: Frame{
				Path: file,
				Line: line,
				Func: funcName,
			},
			Code: source,
		},
	}

	if err := json.NewEncoder(buf).Encode(log); err != nil {
		return
	}

	buf.WriteByte('\n')

	data := make([]byte, buf.Len())
	copy(data, buf.Bytes())

	if _, err := l.writer.Write(data); err != nil {
		return
	}
}
