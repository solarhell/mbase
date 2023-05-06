package mycontext

import (
	"errors"
	"os"
	"path/filepath"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	_L = zap.NewNop()
	_S = _L.Sugar()
)

func init() {
	var _, err = UseDefaultLogger()
	if err != nil {
		panic(err)
	}
}

// L 返回context内部的默认全局zap.Logger
func L() *zap.Logger {
	return _L
}

// S 返回context内部的默认全局zap.SugaredLogger
func S() *zap.SugaredLogger {
	return _S
}

// ReplaceLogger 用给定的zap.Logger替换context内部的默认全局zap.Logger和zap.SugaredLogger
func ReplaceLogger(logger *zap.Logger) func() {
	var prev = _L
	_L = logger
	_S = logger.Sugar()
	return func() { ReplaceLogger(prev) }
}

// UseDefaultLogger 使用预定义的简单zap.Logger替换全局默认zap.Logger。
func UseDefaultLogger() (func(), error) {
	var logger, err = NewSimpleLogger("info", "stderr", "console", false)
	if err != nil {
		return nil, err
	}
	return ReplaceLogger(logger), nil
}

// UseDevelopLogger 使用预定义的简单zap.Logger替换全局默认zap.Logger，适合于开发、测试、写简单的工具时用。
func UseDevelopLogger() (func(), error) {
	var logger, err = NewSimpleLogger("debug", "log", "console", false)
	if err != nil {
		return nil, err
	}
	return ReplaceLogger(logger), nil
}

// UseSimpleLogger 使用简单的默认风格logger替换掉全局的zap.Logger
func UseSimpleLogger(level, outpath, encoding string, showCaller bool) (func(), error) {
	var logger, err = NewSimpleLogger(level, outpath, encoding, showCaller)
	if err != nil {
		return nil, err
	}
	return ReplaceLogger(logger), nil
}

// NewSimpleLogger 生成并返回一个简单的默认风格的zap.Logger
func NewSimpleLogger(level, outpath, encoding string, showCaller bool) (*zap.Logger, error) {
	var zlevel zapcore.Level
	switch level {
	case "debug":
		zlevel = zap.DebugLevel
	case "info":
		zlevel = zap.InfoLevel
	case "warn":
		zlevel = zap.WarnLevel
	case "error":
		zlevel = zap.ErrorLevel
	case "panic":
		zlevel = zap.PanicLevel
	case "fatal":
		zlevel = zap.FatalLevel
	default:
		return nil, errors.New("Unexpected log level " + level)
	}

	if dir := filepath.Dir(outpath); dir != "." && dir != ".." && dir != "/" {
		if _, e := os.Stat(dir); errors.Is(e, os.ErrNotExist) {
			if e := os.MkdirAll(dir, 0755); e != nil {
				return nil, e
			}
		}
	}

	var zcfg = zap.Config{
		Level:            zap.NewAtomicLevelAt(zlevel),
		OutputPaths:      []string{outpath},
		ErrorOutputPaths: []string{outpath},
		Encoding:         encoding,
		DisableCaller:    !showCaller,

		EncoderConfig: zapcore.EncoderConfig{
			MessageKey:     "M",
			LevelKey:       "L",
			TimeKey:        "T",
			NameKey:        "N",
			CallerKey:      "Caller",
			FunctionKey:    zapcore.OmitKey,
			StacktraceKey:  "Stack",
			SkipLineEnding: false,
			LineEnding:     zapcore.DefaultLineEnding,
			EncodeLevel:    zapcore.LowercaseLevelEncoder,
			EncodeTime: func(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
				enc.AppendString(t.Format("2006-01-02 15:04:05.000"))
			},
			EncodeDuration:      zapcore.NanosDurationEncoder,
			EncodeCaller:        zapcore.ShortCallerEncoder,
			EncodeName:          zapcore.FullNameEncoder,
			NewReflectedEncoder: nil,
			ConsoleSeparator:    "",
		},
	}

	return zcfg.Build()
}
