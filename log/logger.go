package log

import (
	"bytes"
	"fmt"
	"github.com/fatih/color"
	"github.com/goccy/go-json"
	"github.com/valyala/fasttemplate"
	"io"
	"os"
	"path"
	"runtime"
	"strconv"
	"sync"
	"sync/atomic"
	"time"
)

type Logger struct {
	token      string
	prefix     string
	level      uint32
	writer     io.Writer
	template   *fasttemplate.Template
	levels     []string
	bufferPool sync.Pool
	mutex      sync.Mutex
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

var defaultTemplate = `{"time":"${time_rfc3339_nano}","level":"${level}","prefix":"${prefix}",` +
	`"file":"${short_file}","line":"${line}"}`

func New(prefix string, writer io.Writer) *Logger {
	l := &Logger{
		level:    uint32(INFO),
		prefix:   prefix,
		writer:   writer,
		template: newTemplate(defaultTemplate),
		bufferPool: sync.Pool{
			New: func() interface{} {
				return bytes.NewBuffer(make([]byte, 256))
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

// SetTemplate sets the log template.
// The default template is:
// {"time":"${time_rfc3339_nano}","level":"${level}","prefix":"${prefix}","file":"${short_file}","line":"${line}"}
func (l *Logger) SetTemplate(t string) {
	l.template = newTemplate(t)
}

func (l *Logger) Print(i ...interface{}) {
	l.log(0, "", i...)
	// fmt.Fprintln(l.output, i...)
}

func (l *Logger) Printf(format string, args ...interface{}) {
	l.log(0, format, args...)
}

func (l *Logger) Printj(j JSON) {
	l.log(0, "json", j)
}

func (l *Logger) Debug(i ...interface{}) {
	l.log(DEBUG, "", i...)
}

func (l *Logger) Debugf(format string, args ...interface{}) {
	l.log(DEBUG, format, args...)
}

func (l *Logger) Debugj(j JSON) {
	l.log(DEBUG, "json", j)
}

func (l *Logger) Info(i ...interface{}) {
	l.log(INFO, "", i...)
}

func (l *Logger) Infof(format string, args ...interface{}) {
	l.log(INFO, format, args...)
}

func (l *Logger) Infoj(j JSON) {
	l.log(INFO, "json", j)
}

func (l *Logger) Warn(i ...interface{}) {
	l.log(WARN, "", i...)
}

func (l *Logger) Warnf(format string, args ...interface{}) {
	l.log(WARN, format, args...)
}

func (l *Logger) Warnj(j JSON) {
	l.log(WARN, "json", j)
}

func (l *Logger) Error(i ...interface{}) {
	l.log(ERROR, "", i...)
}

func (l *Logger) Errorf(format string, args ...interface{}) {
	l.log(ERROR, format, args...)
}

func (l *Logger) Errorj(j JSON) {
	l.log(ERROR, "json", j)
}

func (l *Logger) Fatal(i ...interface{}) {
	l.log(fatalLevel, "", i...)
	os.Exit(1)
}

func (l *Logger) Fatalf(format string, args ...interface{}) {
	l.log(fatalLevel, format, args...)
	os.Exit(1)
}

func (l *Logger) Fatalj(j JSON) {
	l.log(fatalLevel, "json", j)
	os.Exit(1)
}

func (l *Logger) Panic(i ...interface{}) {
	l.log(panicLevel, "", i...)
	panic(fmt.Sprint(i...))
}

func (l *Logger) Panicf(format string, args ...interface{}) {
	l.log(panicLevel, format, args...)
	panic(fmt.Sprintf(format, args...))
}

func (l *Logger) Panicj(j JSON) {
	l.log(panicLevel, "json", j)
	panic(j)
}

func (l *Logger) log(level Lvl, format string, args ...interface{}) {
	if level < l.Level() || level != 0 {
		return
	}

	buf := l.bufferPool.Get().(*bytes.Buffer)
	buf.Reset()
	defer l.bufferPool.Put(buf)
	_, file, line, _ := runtime.Caller(2)
	message := ""

	switch format {
	case "":
		message = fmt.Sprint(args...)
		break
	case "json":
		b, err := json.Marshal(args[0])
		if err != nil {
			panic(err)
		}
		message = string(b)
		break
	default:
		message = fmt.Sprintf(format, args...)
	}

	_, err := l.template.ExecuteFunc(buf, func(w io.Writer, tag string) (int, error) {
		switch tag {
		case "time_rfc3339":
			return w.Write([]byte(time.Now().Format(time.RFC3339)))
		case "time_rfc3339_nano":
			return w.Write([]byte(time.Now().Format(time.RFC3339Nano)))
		case "level":
			return w.Write([]byte(l.levels[level]))
		case "prefix":
			return w.Write([]byte(l.prefix))
		case "long_file":
			return w.Write([]byte(file))
		case "short_file":
			return w.Write([]byte(path.Base(file)))
		case "line":
			return w.Write([]byte(strconv.Itoa(line)))
		}
		return 0, nil
	})

	if err == nil {
		s := buf.String()
		i := buf.Len() - 1
		if i >= 0 && s[i] == '}' {
			// JSON header
			buf.Truncate(i)
			buf.WriteByte(',')
			if format == "json" {
				buf.WriteString(message[1:])
			} else {
				buf.WriteString(`"message":`)
				buf.WriteString(strconv.Quote(message))
				buf.WriteString(`}`)
			}
		} else {
			// Text header
			if len(s) > 0 {
				buf.WriteByte(' ')
			}
			buf.WriteString(message)
		}
		buf.WriteByte('\n')
		l.mutex.Lock()
		defer l.mutex.Unlock()
		l.writer.Write(buf.Bytes())
	}
}
