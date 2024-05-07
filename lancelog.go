package lancelog

import (
	"bytes"
	"fmt"
	"github.com/sirupsen/logrus"
	"io"
	"path"
	"runtime"
	"sort"
	"strings"
	"time"
)

var myLog *logrus.Logger

// 只导出以下常用方法
var (
	Trace func(args ...interface{})
	Debug func(args ...interface{})
	Info  func(args ...interface{})
	Warn  func(args ...interface{})
	Error func(args ...interface{})
	Fatal func(args ...interface{})
	Panic func(args ...interface{})
)

func init() {

	myLog = logrus.New()

	myLog.SetFormatter(&NestFormatter{
		TrimMessages:    true,                              //去掉消息中的空格
		HideKeys:        true,                              //隐藏key
		TimestampFormat: "2006-01-02 15:04:05",             //时间格式
		FieldsOrder:     []string{"component", "category"}, //字段排序
		CallerFirst:     true,                              //先打印调用者信息
		NoFieldsColors:  true,                              //不给自己的字段加颜色
		CustomCallerFormatter: func(f *runtime.Frame) string { //自定义文件函数等信息

			s := strings.Split(f.Function, ".")
			funcName := s[len(s)-1]
			return fmt.Sprintf(" [%s:%d][%s()]", path.Base(f.File), f.Line, funcName)
		},
	})
	myLog.SetReportCaller(true)

	Trace = myLog.Trace
	Debug = myLog.Debug
	Info = myLog.Info
	Warn = myLog.Warn
	Error = myLog.Error
	Fatal = myLog.Fatal
	Panic = myLog.Panic
	Info("=== lance log init success ===")
}

// 额外的配置需求
type Logger = logrus.Logger
type Formatter = logrus.Formatter
type Level = logrus.Level
type Hook = logrus.Hook
type Entry = logrus.Entry
type Fields = logrus.Fields
type LogFunction = logrus.LogFunction

const (
	PanicLevel = logrus.PanicLevel
	FatalLevel = logrus.FatalLevel
	ErrorLevel = logrus.ErrorLevel
	WarnLevel  = logrus.WarnLevel
	InfoLevel  = logrus.InfoLevel
	DebugLevel = logrus.DebugLevel
	TraceLevel = logrus.TraceLevel
)

func SetOutput(out io.Writer) {
	myLog.SetOutput(out)
}

func SetFormatter(formatter Formatter) {
	myLog.SetFormatter(formatter)
}

func SetReportCaller(include bool) {
	myLog.SetReportCaller(include)
}

func SetLevel(level Level) {
	myLog.SetLevel(level)
}

func GetLevel() Level {
	return myLog.GetLevel()
}

func IsLevelEnabled(level Level) bool {
	return myLog.IsLevelEnabled(level)
}

func AddHook(hook Hook) {
	myLog.AddHook(hook)
}

func WithFields(fields Fields) *Entry {
	return myLog.WithFields(fields)
}

// NestFormatter 自定义日志格式，为了精简直接定义  from "github.com/antonfisher/nested-logrus-formatter"
type NestFormatter struct {
	// FieldsOrder - default: fields sorted alphabetically
	FieldsOrder []string

	// TimestampFormat - default: time.StampMilli = "Jan _2 15:04:05.000"
	TimestampFormat string

	// HideKeys - show [fieldValue] instead of [fieldKey:fieldValue]
	HideKeys bool

	// NoColors - disable colors
	NoColors bool

	// NoFieldsColors - apply colors only to the level, default is level + fields
	NoFieldsColors bool

	// NoFieldsSpace - no space between fields
	NoFieldsSpace bool

	// ShowFullLevel - show a full level [WARNING] instead of [WARN]
	ShowFullLevel bool

	// NoUppercaseLevel - no upper case for level value
	NoUppercaseLevel bool

	// TrimMessages - trim whitespaces on messages
	TrimMessages bool

	// CallerFirst - print caller info first
	CallerFirst bool

	// CustomCallerFormatter - set custom formatter for caller info
	CustomCallerFormatter func(*runtime.Frame) string
}

// Format an mlog entry
func (f *NestFormatter) Format(entry *logrus.Entry) ([]byte, error) {
	levelColor := getColorByLevel(entry.Level)

	timestampFormat := f.TimestampFormat
	if timestampFormat == "" {
		timestampFormat = time.StampMilli
	}

	// output buffer
	b := &bytes.Buffer{}

	// write time
	b.WriteString(entry.Time.Format(timestampFormat))

	// write level
	var level string
	if f.NoUppercaseLevel {
		level = entry.Level.String()
	} else {
		level = strings.ToUpper(entry.Level.String())
	}

	if f.CallerFirst {
		f.writeCaller(b, entry)
	}

	if !f.NoColors {
		fmt.Fprintf(b, "\x1b[%dm", levelColor)
	}

	b.WriteString(" [")
	if f.ShowFullLevel {
		b.WriteString(level)
	} else {
		b.WriteString(level[:4])
	}
	b.WriteString("]")

	if !f.NoFieldsSpace {
		b.WriteString(" ")
	}

	if !f.NoColors && f.NoFieldsColors {
		b.WriteString("\x1b[0m")
	}

	// write fields
	if f.FieldsOrder == nil {
		f.writeFields(b, entry)
	} else {
		f.writeOrderedFields(b, entry)
	}

	if f.NoFieldsSpace {
		b.WriteString(" ")
	}

	if !f.NoColors && !f.NoFieldsColors {
		b.WriteString("\x1b[0m")
	}

	// write message
	if f.TrimMessages {
		b.WriteString(strings.TrimSpace(entry.Message))
	} else {
		b.WriteString(entry.Message)
	}

	if !f.CallerFirst {
		f.writeCaller(b, entry)
	}

	b.WriteByte('\n')

	return b.Bytes(), nil
}

func (f *NestFormatter) writeCaller(b *bytes.Buffer, entry *logrus.Entry) {
	if entry.HasCaller() {
		if f.CustomCallerFormatter != nil {
			fmt.Fprintf(b, f.CustomCallerFormatter(entry.Caller))
		} else {
			fmt.Fprintf(
				b,
				" (%s:%d %s)",
				entry.Caller.File,
				entry.Caller.Line,
				entry.Caller.Function,
			)
		}
	}
}

func (f *NestFormatter) writeFields(b *bytes.Buffer, entry *logrus.Entry) {
	if len(entry.Data) != 0 {
		fields := make([]string, 0, len(entry.Data))
		for field := range entry.Data {
			fields = append(fields, field)
		}

		sort.Strings(fields)

		for _, field := range fields {
			f.writeField(b, entry, field)
		}
	}
}

func (f *NestFormatter) writeOrderedFields(b *bytes.Buffer, entry *logrus.Entry) {
	length := len(entry.Data)
	foundFieldsMap := map[string]bool{}
	for _, field := range f.FieldsOrder {
		if _, ok := entry.Data[field]; ok {
			foundFieldsMap[field] = true
			length--
			f.writeField(b, entry, field)
		}
	}

	if length > 0 {
		notFoundFields := make([]string, 0, length)
		for field := range entry.Data {
			if foundFieldsMap[field] == false {
				notFoundFields = append(notFoundFields, field)
			}
		}

		sort.Strings(notFoundFields)

		for _, field := range notFoundFields {
			f.writeField(b, entry, field)
		}
	}
}

func (f *NestFormatter) writeField(b *bytes.Buffer, entry *logrus.Entry, field string) {
	if f.HideKeys {
		fmt.Fprintf(b, "[%v]", entry.Data[field])
	} else {
		fmt.Fprintf(b, "[%s:%v]", field, entry.Data[field])
	}

	if !f.NoFieldsSpace {
		b.WriteString(" ")
	}
}

const (
	colorRed    = 31
	colorYellow = 33
	colorBlue   = 36
	colorGray   = 37
)

func getColorByLevel(level Level) int {
	switch level {
	case DebugLevel, TraceLevel:
		return colorGray
	case WarnLevel:
		return colorYellow
	case ErrorLevel, FatalLevel, PanicLevel:
		return colorRed
	default:
		return colorBlue
	}
}
