package log

import (
	"fmt"
	"golang.org/x/net/context"
	"os"
	"path"
	"strings"
	"sync"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"gopkg.in/natefinch/lumberjack.v2"
)

type Logger struct {
	logger *zap.Logger
}

var Log *Logger

var lso sync.Once

func GetLogger(ctx context.Context) *Logger {
	lso.Do(func() {
		Log = newLog(ctx)
	})
	return Log
}

func newLog(ctx context.Context) *Logger {
	logdir := ctx.Value("logdir")
	logpath := path.Join(logdir.(string), "")

	appname := ctx.Value("appname")

	level := ctx.Value("loglevel")
	if level == nil {
		level = "info"
	}

	level = strings.ToLower(level.(string))
	loglevel := zapcore.InfoLevel
	switch level {
	case "debug":
		loglevel = zapcore.DebugLevel
	case "warn":
		loglevel = zapcore.WarnLevel
	case "error":
		loglevel = zapcore.ErrorLevel
	default:
		loglevel = zapcore.InfoLevel
	}

	hook := lumberjack.Logger{
		Filename:   logpath, // 日志文件路径
		MaxSize:    32,      // 每个日志文件保存的最大尺寸 单位：M
		MaxBackups: 3,       // 日志文件最多保存多少个备份
		MaxAge:     7,       // 文件最多保存多少天
		Compress:   true,    // 是否压缩
	}

	encoderConfig := zapcore.EncoderConfig{
		TimeKey:        "time",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "linenum",
		MessageKey:     "msg",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.LowercaseLevelEncoder,  // 小写编码器
		EncodeTime:     zapcore.ISO8601TimeEncoder,     // ISO8601 UTC 时间格式
		EncodeDuration: zapcore.SecondsDurationEncoder, //
		EncodeCaller:   zapcore.FullCallerEncoder,      // 全路径编码器
		EncodeName:     zapcore.FullNameEncoder,
	}

	// 设置日志级别
	atomicLevel := zap.NewAtomicLevel()
	atomicLevel.SetLevel(loglevel)

	core := zapcore.NewCore(
		zapcore.NewJSONEncoder(encoderConfig),                                           // 编码器配置
		zapcore.NewMultiWriteSyncer(zapcore.AddSync(os.Stdout), zapcore.AddSync(&hook)), // 打印到控制台和文件
		atomicLevel, // 日志级别
	)

	// 开启开发模式，堆栈跟踪
	caller := zap.AddCaller()
	// 开启文件及行号
	development := zap.Development()
	// 设置初始化字段
	filed := zap.Fields(zap.String("serviceName", appname.(string)))
	// 构造日志
	return &Logger{
		zap.New(core, caller, development, filed),
	}
}

func (log Logger) Debug(format string, a ...interface{}) {
	msg := ""

	if a == nil {
		msg = fmt.Sprintf(format)
	} else {
		msg = fmt.Sprintf(format, a)
	}

	log.logger.Debug(msg)
}

func (log Logger) Info(format string, a ...interface{}) {
	msg := ""

	if a == nil {
		msg = fmt.Sprintf(format)
	} else {
		msg = fmt.Sprintf(format, a)
	}

	log.logger.Info(msg)
}

func (log Logger) Warn(format string, a ...interface{}) {
	msg := ""

	if a == nil {
		msg = fmt.Sprintf(format)
	} else {
		msg = fmt.Sprintf(format, a)
	}

	log.logger.Warn(msg)
}

func (log Logger) Error(err error, format string, a ...interface{}) {
	msg := ""

	if a == nil {
		msg = fmt.Sprintf(format)
	} else {
		msg = fmt.Sprintf(format, a)
	}

	log.logger.Error(msg, zap.Error(err))
}

func (log Logger) Close() {
	log.logger.Sync()
}
